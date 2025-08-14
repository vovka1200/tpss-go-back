package matrix

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

type Rules []Rule

type Rule struct {
	Object string   `json:"object"`
	Access []string `json:"access"`
}

type Params struct {
}

type Response struct {
	Rules Rules `json:"matrix"`
}

type Matrix struct {
}

func (m *Matrix) Register(methods api.Methods) {
	methods["access.matrix"] = m.HandleList
}

func (m *Matrix) HandleList(db *pgme.Database, state *websocket.State, data json.RawMessage) (any, jsonrpc2.Error) {
	return jsonrpc2.QueryWithParams[Params](db, data, func(ctx context.Context, conn *pgxpool.Conn, params Params) (any, jsonrpc2.Error) {
		log.WithFields(log.Fields{
			"user": state.UserId,
			"ip":   state.Conn.RemoteAddr(),
		}).Info("Параметры")
		rows, _ := conn.Query(ctx, `
			SELECT   
				o.name AS object,
				r.access 
			FROM access.rules r
			JOIN access.objects o ON o.id = r.object_id
			JOIN access.members m ON m.group_id=r.group_id
			WHERE m.user_id=$1`,
			state.UserId,
		)
		var err error
		response := Response{}
		if response.Rules, err = pgx.CollectRows[Rule](rows, pgx.RowToStructByNameLax[Rule]); err == nil {
			return response, nil
		} else {
			return nil, &jsonrpc2.RPCError{
				Code:    jsonrpc2.InternalError,
				Message: err.Error(),
			}
		}
	})
}
