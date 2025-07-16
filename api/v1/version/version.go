package version

import (
	"encoding/json"
	"github.com/vovka1200/pgme"
	"github.com/vovka1200/tpss-go-back/api"
	"github.com/vovka1200/tpss-go-back/jsonrpc2"
	"github.com/vovka1200/tpss-go-back/websocket"
)

const Number string = "1.0.0"

type Version struct {
}

type Params struct {
	Integers []int `json:"integers,omitempty"`
}

type Answer struct {
	Version string `json:"version"`
}

func (v *Version) Register(methods api.Methods) {
	methods["version"] = v.Handler
}

func (v *Version) Handler(db *pgme.Database, state *websocket.State, data json.RawMessage) (any, jsonrpc2.Error) {
	return Answer{
		Version: Number,
	}, nil
}
