package group

import (
	"github.com/vovka1200/tpss-go-back/api"
	"time"
)

type Group struct {
	Id      string    `json:"id"`
	Name    string    `json:"name"`
	Created time.Time `json:"created"`
	Members []string  `json:"members"`
	Rules   Rules     `json:"rules"`
}

func (g *Group) Register(methods api.Methods) {
	g.Rules.Register(methods)
}
