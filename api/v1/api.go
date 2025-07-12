package v1

import (
	"encoding/json"
	"github.com/fasthttp/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/vovka1200/tpss-go-back/api/v1/access"
	"github.com/vovka1200/tpss-go-back/api/v1/version"
	"github.com/vovka1200/tpss-go-back/jsonrpc2"
)

type APIHandler interface {
	Handler(data json.RawMessage) (json.RawMessage, *jsonrpc2.Error)
}

type API struct {
	methods jsonrpc2.Methods
	Version version.Version
	Access  access.Access
}

func (api *API) Handler(conn *websocket.Conn, authorized bool, msg []byte) jsonrpc2.Response {

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

		if authorized {
			if method, ok := api.methods[req.Method]; ok {
				result.Result, result.Error = method(req.Params)
			} else {
				result.Error = &jsonrpc2.Error{
					Code:    jsonrpc2.MethodNotFound,
					Message: "Method not found",
				}
			}
		} else {
			if req.Method == "login" {
				result.Result, result.Error = api.methods[req.Method](req.Params)
			} else {
				result.Error = &jsonrpc2.Error{
					Code:    401,
					Message: "Unauthorized",
				}
			}
		}

	} else {
		log.Error(err)
		result.ID = nil
		result.Error = &jsonrpc2.Error{
			Code:    jsonrpc2.ParseError,
			Message: err.Error(),
		}
	}

	return result
}

func (api *API) Register() {
	api.methods = make(jsonrpc2.Methods)
	api.Version.Register(api.methods)
	api.Access.Register(api.methods)
}
