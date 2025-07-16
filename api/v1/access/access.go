package access

import (
	"github.com/vovka1200/tpss-go-back/api/v1/access/login"
	"github.com/vovka1200/tpss-go-back/api/v1/access/users"
	"github.com/vovka1200/tpss-go-back/jsonrpc2"
)

type Access struct {
	Login login.Login `json:"login"`
	Users users.Users `json:"users"`
}

func (a *Access) Register(methods jsonrpc2.Methods) {
	a.Login.Register(methods)
	a.Users.Register(methods)
}
