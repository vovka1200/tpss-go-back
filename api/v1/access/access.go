package access

import (
	"github.com/vovka1200/tpss-go-back/jsonrpc2"
)

type Access struct {
	Login Login
}

func (a *Access) Register(methods jsonrpc2.Methods) {
	methods["login"] = a.Login.Handler
}
