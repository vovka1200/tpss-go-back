package v1

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	log "github.com/sirupsen/logrus"
	"github.com/vovka1200/pgme"
	common "github.com/vovka1200/tpss-go-back/api"
	"github.com/vovka1200/tpss-go-back/api/v1/access"
	"github.com/vovka1200/tpss-go-back/api/v1/access/matrix"
	"github.com/vovka1200/tpss-go-back/api/v1/access/users/user"
	"github.com/vovka1200/tpss-go-back/api/v1/crm"
	"github.com/vovka1200/tpss-go-back/api/v1/settings"
	"github.com/vovka1200/tpss-go-back/api/v1/version"
	"github.com/vovka1200/tpss-go-back/jsonrpc2"
	"github.com/vovka1200/tpss-go-back/websocket"
)

type API struct {
	methods  common.Methods
	Version  version.Version
	Access   access.Access
	CRM      crm.CRM
	Settings settings.Settings
}

func (api *API) Handler(state *websocket.State, db *pgme.Database, method string, params json.RawMessage) (any, jsonrpc2.Error) {
	if state.UserId != "" {
		// Если соединение авторизовано
		if handler, ok := api.methods[method]; ok {
			if method == matrix.Method || api.allowed(state, db, method) {
				return handler(db, state, params)
			} else {
				return nil, &jsonrpc2.RPCError{
					Code:    jsonrpc2.AccessDenied,
					Message: "Method access denied",
				}
			}
		} else {
			log.WithFields(log.Fields{
				"method": method,
				"addr":   state.Conn.RemoteAddr(),
			}).Error("Метод не найден")
			return nil, &jsonrpc2.RPCError{
				Code:    jsonrpc2.MethodNotFound,
				Message: "Method not found",
			}
		}
	} else {
		// Иначе если запрос аутентификации
		if method == user.AuthenticationMethod {
			var result any
			var err jsonrpc2.Error
			if result, err = api.methods[method](db, state, params); result != nil {
				if result.(user.AuthorizeResponse).Account.Id != "" {
					state.UserId = result.(user.AuthorizeResponse).Account.Id
				}
			}
			return result, err
		}
		if method == version.Method {
			return api.methods[method](db, state, params)
		}
		return nil, &jsonrpc2.RPCError{
			Code:    jsonrpc2.Unauthorized,
			Message: "Требуется аутентификация",
		}
	}
}

func (api *API) Register() {
	api.methods = make(common.Methods)
	api.Version.Register(api.methods)
	api.Access.Register(api.methods)
	api.CRM.Register(api.methods)
	api.Settings.Register(api.methods)
}

func (api *API) allowed(state *websocket.State, db *pgme.Database, method string) bool {
	ctx := context.Background()
	if conn, err := db.NewConnection(ctx); err == nil {
		defer db.Disconnect(conn)
		if rows, _ := conn.Query(ctx, `
			SELECT m.object,
			       m.access
			FROM access.matrix m
			WHERE m.user_id=$1 
			  AND m.object=$2
		`,
			state.UserId,
			method,
		); err == nil {
			if rule, err := pgx.CollectOneRow[matrix.Rule](rows, pgx.RowToStructByName[matrix.Rule]); err == nil {
				return rule.Object == method && len(rule.Access) > 0
			} else {
				log.Error(err)
			}
		}
	}
	return false
}
