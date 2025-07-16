package users

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
	"github.com/vovka1200/pgme"
	"github.com/vovka1200/tpss-go-back/api"
	"github.com/vovka1200/tpss-go-back/jsonrpc2"
	"github.com/vovka1200/tpss-go-back/websocket"
)

type Users struct {
}

type Params struct {
	Filter string `json:"filter"`
}

type Response struct {
	Users []User `json:"users"`
}

func (u *Users) Register(methods api.Methods) {
	methods["access.users"] = u.Handler
}

func (u *Users) Handler(db *pgme.Database, state *websocket.State, data json.RawMessage) (any, jsonrpc2.Error) {
	params := Params{}
	var err error
	var conn *pgxpool.Conn
	if err = jsonrpc2.UnmarshalParams[Params](data, &params); err == nil {
		log.WithFields(log.Fields{
			"filter": params.Filter,
		}).Info("Поиск")
		ctx := context.Background()
		if conn, err = db.NewConnection(ctx); err == nil {
			defer db.Disconnect(conn)
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
				WHERE (u.name ~* $1::text OR g.name ~* $1::text)
				GROUP BY 1,2,3,4`,
				params.Filter,
			)
			response := Response{}
			if response.Users, err = pgx.CollectRows[User](rows, pgx.RowToStructByNameLax[User]); err == nil {
				return response, nil
			}
		}
	}
	log.Error(err)
	return nil, &jsonrpc2.RPCError{
		Code:    500,
		Message: err.Error(),
	}
}
