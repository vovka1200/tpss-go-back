package v1

import (
	"encoding/json"
	"github.com/fasthttp/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/vovka1200/tpss-go-back/api/v1/version"
	"github.com/vovka1200/tpss-go-back/jsonrpc2"
)

type APIHandler interface {
	Handler(data json.RawMessage) (json.RawMessage, *jsonrpc2.Error)
}

type API struct {
	methods jsonrpc2.Methods
	Version version.Version `api:"version"`
}

func (api *API) Handler(conn *websocket.Conn, msg []byte) []byte {

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

		if method, ok := api.methods[req.Method]; ok {
			result.Result, result.Error = method(req.Params)
		}

	} else {
		log.Error(err)
		result.ID = nil
		result.Error = &jsonrpc2.Error{
			Code:    jsonrpc2.ParseError,
			Message: err.Error(),
		}
	}

	if buffer, err := json.Marshal(result); err == nil {
		return buffer
	} else {
		log.Error(err)
		return nil
	}

}

func (api *API) Register() {
	api.methods = make(jsonrpc2.Methods)
	api.Version.Register(api.methods)
}
