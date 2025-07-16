package server

import (
	"encoding/json"
	"github.com/fasthttp/router"
	"github.com/fasthttp/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"github.com/vovka1200/pgme"
	v1 "github.com/vovka1200/tpss-go-back/api/v1"
	"github.com/vovka1200/tpss-go-back/jsonrpc2"
	"time"
)

type Server struct {
	Listen   string `json:"listen"`
	upgrader *websocket.FastHTTPUpgrader
	API      struct {
		V1 v1.API `json:"v1"`
	} `json:"API"`
	Database pgme.Database `json:"database"`
}

func (s *Server) Run() {
	log.Info("Запуск сервера")
	if err := s.Database.InitPool(); err != nil {
		log.Fatal(err)
	}
	log.Info("База данных подключена")
	s.API.V1.Register()
	log.Info("API v1 инициализирован")
	s.upgrader = &websocket.FastHTTPUpgrader{
		CheckOrigin: checkOrigin,
	}
	r := router.New()
	r.GET("/api/v1", s.webSocketHandlerV1)
	log.Info("Роутер инициализирован")
	if err := fasthttp.ListenAndServe(s.Listen, r.Handler); err != nil {
		log.Fatal(err)
	}
}

func (s *Server) webSocketHandlerV1(ctx *fasthttp.RequestCtx) {
	log.WithFields(log.Fields{
		"addr": ctx.RemoteAddr(),
	}).Info("Соединение")
	if err := s.upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		conn.SetCloseHandler(closeHandler)
		userId := ""
		unAuthorized := true
		responseQuitChan, responseChan := responseLoop(conn)
		defer close(responseQuitChan)
		keepAliveChan := keepAliveLoop(conn)
		defer close(keepAliveChan)
		unAuthorizedChan := unAuthorizedLoop(responseChan)
		defer close(unAuthorizedChan)
		for {
			if msgType, msg, err := conn.ReadMessage(); err == nil {
				if msgType == websocket.TextMessage {
					response := jsonrpc2.Response{}
					userId, response = s.API.V1.Handler(conn, &s.Database, userId, msg)
					if response.Error == nil && userId != "" && unAuthorized {
						unAuthorized = false
						unAuthorizedChan <- true
						log.WithFields(log.Fields{
							"user": userId,
							"addr": conn.RemoteAddr(),
						}).Info("Соединение авторизовано")
					}
					if buffer, err := json.Marshal(response); err == nil {
						time.Sleep(1 * time.Second)
						responseChan <- buffer
					} else {
						log.WithFields(log.Fields{
							"user": userId,
							"addr": conn.RemoteAddr(),
						}).Error(err)
					}
				} else if msgType == websocket.PingMessage {
					if err := conn.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(3*time.Second)); err != nil {
						log.WithFields(log.Fields{
							"type": msgType,
							"addr": conn.RemoteAddr(),
						}).Error(err)
						break
					}
				} else {
					log.WithFields(log.Fields{
						"type": msgType,
						"addr": conn.RemoteAddr(),
					}).Error("Ошибка типа сообщения")
					break
				}
			} else {
				log.WithFields(log.Fields{
					"type": msgType,
					"addr": conn.RemoteAddr(),
				}).Error(err)
				break
			}
		}
		keepAliveChan <- true
		if userId != "" {
			unAuthorizedChan <- true
		}
		responseQuitChan <- true
		log.WithFields(log.Fields{
			"user": userId,
			"addr": conn.RemoteAddr(),
		}).Debug("Соединение завершено")
		_ = conn.Close()
	}); err != nil {
		log.Error(err)
	}
}

func closeHandler(code int, text string) error {
	log.Debugf("Соединение закрыто: code=%v msg=%v", code, text)
	return nil
}

func keepAliveLoop(conn *websocket.Conn) chan bool {
	quitChan := make(chan bool)
	go func() {
		for {
			select {
			case <-quitChan:
				return
			case <-time.After(10 * time.Second):
				if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(3*time.Second)); err != nil {
					log.Error(err)
					return
				}
			}
		}
	}()
	return quitChan
}

func checkOrigin(ctx *fasthttp.RequestCtx) bool {
	return true
}

func unAuthorizedLoop(responseLoop chan []byte) chan bool {
	quitChan := make(chan bool)
	buffer, _ := jsonrpc2.Marshal(
		jsonrpc2.NewError(401, "UnAuthorized"),
	)
	go func() {
		for {
			select {
			case <-quitChan:
				return
			case <-time.After(3 * time.Second):
				responseLoop <- buffer
			}
		}
	}()
	return quitChan
}

func responseLoop(conn *websocket.Conn) (chan bool, chan []byte) {
	bufferChan := make(chan []byte)
	quitChan := make(chan bool)
	go func() {
		for {
			select {
			case <-quitChan:
				return
			case buffer := <-bufferChan:
				if err := conn.WriteMessage(websocket.TextMessage, buffer); err != nil {
					log.WithFields(log.Fields{
						"addr": conn.RemoteAddr(),
					}).Error(err)
				}
			}
		}
	}()
	return quitChan, bufferChan
}
