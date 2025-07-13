package v1

import (
	"encoding/json"
	"github.com/fasthttp/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/vovka1200/tpss-go-back/api/v1/access"
	"github.com/vovka1200/tpss-go-back/api/v1/version"
	"github.com/vovka1200/tpss-go-back/jsonrpc2"
)

type API struct {
	methods jsonrpc2.Methods
	Version version.Version
	Access  access.Access
}

func (api *API) Handler(conn *websocket.Conn, authorized bool, msg []byte) jsonrpc2.Response {

	var req jsonrpc2.Request
	response := jsonrpc2.Response{
		JSONRPC: "2.0",
	}

	if err := json.Unmarshal(msg, &req); err == nil {
		log.WithFields(log.Fields{
			"id":     req.ID,
			"method": req.Method,
			"addr":   conn.RemoteAddr(),
		}).Debug("Запрос")
		response.ID = req.ID
		var result any

		if authorized {
			if method, ok := api.methods[req.Method]; ok {
				result, response.Error = method(req.Params)
			} else {
				response.Error = &jsonrpc2.Error{
					Code:    jsonrpc2.MethodNotFound,
					Message: "Method not found",
				}
			}
		} else {
			if req.Method == "login" {
				result, response.Error = api.methods["login"](req.Params)
			} else {
				response.Error = &jsonrpc2.Error{
					Code:    401,
					Message: "Unauthorized",
				}
			}
		}
		if response.Result, err = json.Marshal(result); err != nil {
			log.Error(err)
			response.Error = &jsonrpc2.Error{
				Code:    500,
				Message: "Internal error",
			}
		}
	} else {
		log.Error(err)
		response.ID = nil
		response.Error = &jsonrpc2.Error{
			Code:    jsonrpc2.ParseError,
			Message: err.Error(),
		}
	}

	return response
}

func (api *API) Register() {
	api.methods = make(jsonrpc2.Methods)
	api.Version.Register(api.methods)
	api.Access.Register(api.methods)
}
