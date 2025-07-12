package server

import (
	"encoding/json"
	"github.com/fasthttp/router"
	"github.com/fasthttp/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	v1 "github.com/vovka1200/tpss-go-back/api/v1"
	"github.com/vovka1200/tpss-go-back/jsonrpc2"
	"strings"
	"time"
)

type Server struct {
	Listen   string `json:"listen"`
	upgrader *websocket.FastHTTPUpgrader
	API      struct {
		V1 v1.API `json:"v1"`
	} `json:"API"`
}

func (s *Server) Run() {
	log.Info("Запуск сервера")
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
		keepAliveQuitChan := sendingKeepAlive(conn)
		sendUnAuthorizedChan := sendingUnAuthorized(conn)
		authorized := false
		for {
			if msgType, msg, err := conn.ReadMessage(); err == nil {
				if msgType == websocket.TextMessage {
					response := s.API.V1.Handler(conn, authorized, msg)
					if response.Error == nil && !authorized {
						authorized = true
						sendUnAuthorizedChan <- true
						log.WithFields(log.Fields{
							"addr": conn.RemoteAddr(),
						}).Info("Соединение авторизовано")
					}
					if buffer, err := json.Marshal(response); err == nil {
						if err := conn.WriteMessage(websocket.TextMessage, buffer); err != nil {
							log.WithFields(log.Fields{
								"addr": conn.RemoteAddr(),
							}).Error(err)
						}
					} else {
						log.WithFields(log.Fields{
							"addr": conn.RemoteAddr(),
						}).Error(err)
					}
				} else if msgType == websocket.PingMessage {
					conn.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(3*time.Second))
				} else {
					log.WithFields(log.Fields{
						"type": msgType,
						"addr": conn.RemoteAddr(),
					}).Error("Ошибка типа сообщения")
					break
				}
			} else {
				if !strings.Contains(err.Error(), "1005") {
					log.WithFields(log.Fields{
						"addr": conn.RemoteAddr(),
					}).Error(err)
				}
				break
			}
		}
		keepAliveQuitChan <- true
		close(keepAliveQuitChan)
		if !authorized {
			sendUnAuthorizedChan <- true
		}
		close(sendUnAuthorizedChan)
		conn.Close()
	}); err != nil {
		log.Error(err)
	}
}

func closeHandler(code int, text string) error {
	log.Debugf("Соединение закрыто: code=%v msg=%v", code, text)
	return nil
}

func sendingKeepAlive(conn *websocket.Conn) chan bool {
	quitChan := make(chan bool)
	go func() {
		for {
			select {
			case <-quitChan:
				return
			case <-time.After(10 * time.Second):
				conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(3*time.Second))
			}
		}
	}()
	return quitChan
}

func checkOrigin(ctx *fasthttp.RequestCtx) bool {
	return true
}

func sendingUnAuthorized(conn *websocket.Conn) chan bool {
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
				if err := conn.WriteMessage(websocket.TextMessage, buffer); err != nil {
					log.WithFields(log.Fields{
						"addr": conn.RemoteAddr(),
					}).Error(err)
				}
			}
		}
	}()
	return quitChan
}
