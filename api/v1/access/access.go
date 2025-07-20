package access

import (
	"github.com/vovka1200/tpss-go-back/api"
	"github.com/vovka1200/tpss-go-back/api/v1/access/groups"
	"github.com/vovka1200/tpss-go-back/api/v1/access/users"
)

type Access struct {
	Users  users.Users   `json:"users"`
	Groups groups.Groups `json:"groups"`
}

func (a *Access) Register(methods api.Methods) {
	a.Users.Register(methods)
	a.Groups.Register(methods)
}
