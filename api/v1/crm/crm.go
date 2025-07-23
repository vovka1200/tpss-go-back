package crm

import (
	"github.com/vovka1200/tpss-go-back/api"
	"github.com/vovka1200/tpss-go-back/api/v1/crm/clients"
)

type CRM struct {
	Clients clients.Clients `json:"clients"`
}

func (c *CRM) Register(methods api.Methods) {
	c.Clients.Register(methods)
}
