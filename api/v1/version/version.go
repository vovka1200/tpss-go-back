package version

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/vovka1200/tpss-go-back/jsonrpc2"
)

const Number string = "1.0"

type Version struct {
	Version string `json:"version"`
}

type Params struct {
	Integers []int `json:"integers,omitempty"`
}

func (v *Version) Register(methods jsonrpc2.Methods) {
	methods["version"] = v.Handler
}

func (v *Version) Handler(data json.RawMessage) (json.RawMessage, *jsonrpc2.Error) {
	if _, err := v.params(data); err == nil {
		v.Version = Number
		return jsonrpc2.Marshal(v)
	} else {
		return nil, err
	}
}

func (v *Version) params(data json.RawMessage) (*Params, *jsonrpc2.Error) {
	p := &Params{}
	if err := jsonrpc2.UnmarshalParams(data, p); err == nil {
		log.Debugf("%+v", p)
		return p, nil
	} else {
		return nil, err
	}
}
