package v1

import (
	"encoding/json"
	"github.com/fasthttp/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/vovka1200/tpss-go-back/api/v1/version"
	"github.com/vovka1200/tpss-go-back/jsonrpc2"
	"reflect"
)

type APIHandler interface {
	Handler(data json.RawMessage) (json.RawMessage, *jsonrpc2.Error)
}

type API struct {
	methods jsonrpc2.Methods
	Version version.Version `api:"version"`
}

func (api *API) Handler(conn *websocket.Conn) {
	for {
		if msgType, msg, err := conn.ReadMessage(); err == nil {
			if msgType == websocket.TextMessage {

				var req jsonrpc2.Request
				result := jsonrpc2.Response{
					JSONRPC: "2.0",
				}

				if err := json.Unmarshal(msg, &req); err == nil {
					log.WithFields(log.Fields{
						"id":     req.ID,
						"method": req.Method,
						"addr":   conn.RemoteAddr(),
					}).Debug("Запрос")
					result.ID = req.ID

					result.Result, result.Error = api.call(req.Method, req.Params)

				} else {
					log.Error(err)
					result.ID = nil
					result.Error = &jsonrpc2.Error{
						Code:    jsonrpc2.ParseError,
						Message: err.Error(),
					}
				}

				if buffer, err := json.Marshal(result); err == nil {
					if err := conn.WriteMessage(websocket.TextMessage, buffer); err != nil {
						log.Error(err)
						break
					}
				} else {
					log.Error(err)
					break
				}
			} else {
				log.WithFields(log.Fields{
					"type": msgType,
				}).Error("Ошибка типа сообщения")
				break
			}
		} else {
			log.Error(err)
			break
		}
	}
}

func (api *API) call(method string, params json.RawMessage) (json.RawMessage, *jsonrpc2.Error) {
	data := reflect.ValueOf(api).Elem()
	for i := 0; i < data.NumField(); i++ {
		field := data.Field(i)
		if field.Kind() == reflect.Struct {
			if tag, ok := data.Type().Field(i).Tag.Lookup("api"); ok && tag == method {
				if register := field.Addr().MethodByName("Handler"); register.IsValid() {
					values := register.Call([]reflect.Value{reflect.ValueOf(params)})
					return values[0].Interface().(json.RawMessage), values[1].Interface().(*jsonrpc2.Error)
				}
			}
		}
	}
	return nil, &jsonrpc2.Error{
		Code:    jsonrpc2.MethodNotFound,
		Message: "Method not found",
	}
}
