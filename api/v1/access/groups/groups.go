package groups

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
	"github.com/vovka1200/pgme"
	"github.com/vovka1200/tpss-go-back/api"
	"github.com/vovka1200/tpss-go-back/api/v1/access/groups/group"
	"github.com/vovka1200/tpss-go-back/jsonrpc2"
	"github.com/vovka1200/tpss-go-back/websocket"
)

type Groups struct {
	Group group.Group `json:"group"`
}

type Params struct {
	Filter string `json:"filter"`
}

type Response struct {
	Groups []group.Group `json:"groups"`
}

func (g *Groups) Register(methods api.Methods) {
	methods["access.groups.list"] = g.HandleList
	g.Group.Register(methods)
}

func (g *Groups) HandleList(db *pgme.Database, state *websocket.State, data json.RawMessage) (any, jsonrpc2.Error) {
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
				    g.id, 
				    g.name, 
				    g.created,
				    array_agg(u.name) AS members
				FROM access.groups g
				LEFT JOIN access.members m ON m.group_id=g.id
				LEFT JOIN access.users u ON u.id=m.user_id
				WHERE (g.name ~* $1::text OR g.name ~* $1::text)
				GROUP BY 1,2,3`,
				params.Filter,
			)
			response := Response{}
			if response.Groups, err = pgx.CollectRows[group.Group](rows, pgx.RowToStructByNameLax[group.Group]); err == nil {
				log.WithFields(log.Fields{
					"filter": params.Filter,
					"count":  len(response.Groups),
				}).Info("Результат")
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
