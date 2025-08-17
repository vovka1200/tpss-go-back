package group

import (
	"github.com/vovka1200/tpss-go-back/api"
	"github.com/vovka1200/tpss-go-back/api/v1/entities"
)

type Group struct {
	entities.Entity
	Members []string `json:"members"`
	Rules   Rules    `json:"rules"`
}

func (g *Group) Register(methods api.Methods) {
	g.Rules.Register(methods)
}
