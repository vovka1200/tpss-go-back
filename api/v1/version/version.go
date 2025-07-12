package version

import (
	"encoding/json"
	"github.com/vovka1200/tpss-go-back/jsonrpc2"
)

const Number string = "1.0"

type Version struct {
}

type Params struct {
	Integers []int `json:"integers,omitempty"`
}

type Answer struct {
	Version string `json:"version"`
}

func (v *Version) Register(methods jsonrpc2.Methods) {
	methods["version"] = v.Handler
}

func (v *Version) Handler(data json.RawMessage) (json.RawMessage, *jsonrpc2.Error) {
	return jsonrpc2.Marshal(Answer{
		Version: Number,
	})
}
