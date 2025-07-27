package params

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

type CustomParam struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Default string `json:"default"`
}

type CustomParams struct {
}

type Params struct {
	Filter string `json:"filter"`
}

type Response struct {
	Params []CustomParam `json:"params"`
}

func (p *CustomParams) Register(methods api.Methods) {
	methods["settings.params.list"] = p.HandleList
}

func (p *CustomParams) HandleList(db *pgme.Database, state *websocket.State, data json.RawMessage) (any, jsonrpc2.Error) {
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
				    p.name,
					p.type,
					COALESCE(p.default,'') AS default
				FROM customs.params p
				WHERE p.name ~* $1::text
				   OR p.type::text ~* $1::text
				   OR p.default ~* $1::text
				LIMIT 100
				`,
				params.Filter,
			)
			response := Response{}
			if response.Params, err = pgx.CollectRows[CustomParam](rows, pgx.RowToStructByNameLax[CustomParam]); err == nil {
				log.WithFields(log.Fields{
					"filter": params.Filter,
					"count":  len(response.Params),
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
