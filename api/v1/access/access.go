package access

import (
	"github.com/vovka1200/tpss-go-back/api"
	"github.com/vovka1200/tpss-go-back/api/v1/access/login"
	"github.com/vovka1200/tpss-go-back/api/v1/access/users"
)

type Access struct {
	Login login.Login `json:"login"`
	Users users.Users `json:"users"`
}

func (a *Access) Register(methods api.Methods) {
	a.Login.Register(methods)
	a.Users.Register(methods)
}
