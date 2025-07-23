package group

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
	"github.com/vovka1200/pgme"
	"github.com/vovka1200/tpss-go-back/api"
	"github.com/vovka1200/tpss-go-back/api/v1/access/matrix"
	"github.com/vovka1200/tpss-go-back/jsonrpc2"
	"github.com/vovka1200/tpss-go-back/websocket"
)

type Rules struct {
}

type Params struct {
	GroupId string `json:"group_id"`
}

type Response struct {
	Rules matrix.Rules `json:"rules"`
}

func (r *Rules) Register(methods api.Methods) {
	methods["access.groups.group.rules.list"] = r.HandleList
}

func (r *Rules) HandleList(db *pgme.Database, state *websocket.State, data json.RawMessage) (any, jsonrpc2.Error) {
	params := Params{}
	var err error
	var conn *pgxpool.Conn
	if err = jsonrpc2.UnmarshalParams[Params](data, &params); err == nil {
		log.WithFields(log.Fields{
			"group_id": params.GroupId,
		}).Info("Поиск")
		ctx := context.Background()
		if conn, err = db.NewConnection(ctx); err == nil {
			defer db.Disconnect(conn)
			rows, _ := conn.Query(ctx, `
				SELECT 
				    o.name AS object, 
				    r.access
				FROM access.rules r
				JOIN access.objects o ON o.id=object_id
				WHERE group_id=$1
				`,
				params.GroupId,
			)
			response := Response{}
			if response.Rules, err = pgx.CollectRows[matrix.Rule](rows, pgx.RowToStructByNameLax[matrix.Rule]); err == nil {
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
