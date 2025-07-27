package settings

import (
	"github.com/vovka1200/tpss-go-back/api"
	"github.com/vovka1200/tpss-go-back/api/v1/settings/params"
)

type Settings struct {
	Params params.CustomParams `json:"params"`
}

func (s *Settings) Register(methods api.Methods) {
	s.Params.Register(methods)
}
