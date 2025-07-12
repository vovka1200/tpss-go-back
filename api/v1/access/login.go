package access

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/vovka1200/tpss-go-back/jsonrpc2"
)

type Login struct {
}

type Params struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Answer struct {
	Authorized bool `json:"authorized"`
}

func (l *Login) Handler(data json.RawMessage) (json.RawMessage, *jsonrpc2.Error) {
	params := Params{}
	if err := jsonrpc2.UnmarshalParams[Params](data, &params); err == nil {
		log.WithFields(log.Fields{
			"username": params.Username,
		}).Info("Login")
		if params.Username == "test" {
			return jsonrpc2.Marshal(Answer{
				Authorized: true,
			})
		} else {
			return nil, &jsonrpc2.Error{
				Code:    jsonrpc2.Unauthorized,
				Message: "UnAuthorized",
			}
		}
	} else {
		return nil, err
	}
}
