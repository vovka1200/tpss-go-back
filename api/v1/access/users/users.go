package users

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
	"github.com/vovka1200/pgme"
	"github.com/vovka1200/tpss-go-back/api"
	"github.com/vovka1200/tpss-go-back/api/v1/access/users/user"
	"github.com/vovka1200/tpss-go-back/jsonrpc2"
	"github.com/vovka1200/tpss-go-back/websocket"
)

type Users struct {
	User user.User
}

type Params struct {
	Id     string `json:"id"`
	Filter string `json:"filter"`
}

type Response struct {
	Users []user.User `json:"users"`
}

func (u *Users) Register(methods api.Methods) {
	methods["access.users.list"] = u.HandleList
	u.User.Register(methods)
}

func (u *Users) HandleList(db *pgme.Database, state *websocket.State, data json.RawMessage) (any, jsonrpc2.Error) {
	return jsonrpc2.QueryWithParams[Params](db, data, func(ctx context.Context, conn *pgxpool.Conn, params Params) (any, jsonrpc2.Error) {
		log.WithFields(log.Fields{
			"id":     params.Id,
			"filter": params.Filter,
		}).Info("Поиск")
		rows, _ := conn.Query(ctx, `
				SELECT 
				    u.id, 
				    u.name, 
				    u.username,
				    u.created,
				    array_agg(g.name) AS groups
				FROM access.users u
				JOIN access.members m ON m.user_id=u.id
				JOIN access.groups g ON g.id=m.group_id
				WHERE (u.name ~* $2::text OR g.name ~* $2::text)
				  AND ($1='' OR u.id=$1::uuid) 
				GROUP BY 1,2,3,4`,
			params.Id,
			params.Filter,
		)
		var err error
		response := Response{}
		if response.Users, err = pgx.CollectRows[user.User](rows, pgx.RowToStructByNameLax[user.User]); err == nil {
			log.WithFields(log.Fields{
				"filter": params.Filter,
				"count":  len(response.Users),
			}).Info("Результат")
			return response, nil
		} else {
			return nil, &jsonrpc2.RPCError{
				Code:    jsonrpc2.InternalError,
				Message: err.Error(),
			}
		}
	})
}
