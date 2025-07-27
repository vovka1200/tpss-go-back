package websocket

import (
	ws "github.com/fasthttp/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"github.com/vovka1200/tpss-go-back/jsonrpc2"
	"time"
)

type ConnectionHandler func(*State)
type MessageHandler func(*State, []byte) ([]byte, error)

type State struct {
	Conn   *ws.Conn
	UserId string
}

func Handler(ctx *fasthttp.RequestCtx, handler MessageHandler) error {
	return upgrade(ctx, func(state *State) {
		sendWorkerQuit, sendChan := sendWorker(state.Conn)
		defer close(sendWorkerQuit)
		keepAliveWorkerQuit := keepAliveWorker(state.Conn)
		defer close(keepAliveWorkerQuit)
		sendChan <- unauthorized()
		for {
			if msgType, msg, err := state.Conn.ReadMessage(); err == nil {
				if response, err := handler(state, msg); err == nil {
					sendChan <- response
				} else {
					log.WithFields(log.Fields{
						"user": state.UserId,
						"addr": state.Conn.RemoteAddr(),
					}).Error(err)
				}
			} else if msgType == ws.PingMessage {
				if err := state.Conn.WriteControl(ws.PongMessage, []byte{}, time.Now().Add(3*time.Second)); err != nil {
					log.WithFields(log.Fields{
						"type": msgType,
						"user": state.UserId,
						"addr": state.Conn.RemoteAddr(),
					}).Error(err)
					break
				}
			} else {
				log.WithFields(log.Fields{
					"type": msgType,
					"user": state.UserId,
					"addr": state.Conn.RemoteAddr(),
				}).Error("Ошибка типа сообщения")
				break
			}
		}
		keepAliveWorkerQuit <- true
		sendWorkerQuit <- true
		log.WithFields(log.Fields{
			"user": state.UserId,
			"addr": state.Conn.RemoteAddr(),
		}).Debug("Соединение завершено")
		_ = state.Conn.Close()
	})
}

func upgrade(ctx *fasthttp.RequestCtx, handler ConnectionHandler) error {
	u := &ws.FastHTTPUpgrader{
		CheckOrigin: checkOrigin,
	}
	err := u.Upgrade(ctx, func(conn *ws.Conn) {
		conn.SetCloseHandler(closeHandler)
		state := State{
			Conn: conn,
		}
		handler(&state)
	})
	return err
}

func checkOrigin(ctx *fasthttp.RequestCtx) bool {
	return true
}

func closeHandler(code int, text string) error {
	log.Debugf("Соединение закрыто: code=%v msg=%v", code, text)
	return nil
}

func keepAliveWorker(conn *ws.Conn) chan bool {
	quitChan := make(chan bool)
	go func() {
		for {
			select {
			case <-quitChan:
				return
			case <-time.After(10 * time.Second):
				if err := conn.WriteControl(ws.PingMessage, []byte{}, time.Now().Add(3*time.Second)); err != nil {
					log.Error(err)
					return
				}
			}
		}
	}()
	return quitChan
}

func sendWorker(conn *ws.Conn) (chan bool, chan []byte) {
	bufferChan := make(chan []byte)
	quitChan := make(chan bool)
	go func() {
		for {
			select {
			case <-quitChan:
				return
			case buffer := <-bufferChan:
				if err := conn.WriteMessage(ws.TextMessage, buffer); err != nil {
					log.WithFields(log.Fields{
						"addr": conn.RemoteAddr(),
					}).Error(err)
				}
			}
		}
	}()
	return quitChan, bufferChan
}

func unauthorized() []byte {
	buffer, _ := jsonrpc2.Marshal(
		jsonrpc2.NewErrorResponse(nil, jsonrpc2.Unauthorized, "Требуется аутентификация"),
	)
	return buffer
}
