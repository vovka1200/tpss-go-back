package clients

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
	"github.com/vovka1200/pgme"
	"github.com/vovka1200/tpss-go-back/api"
	"github.com/vovka1200/tpss-go-back/api/v1/crm/clients/client"
	"github.com/vovka1200/tpss-go-back/jsonrpc2"
	"github.com/vovka1200/tpss-go-back/websocket"
)

type Clients struct {
}

type Params struct {
	Filter string `json:"filter"`
}

type Response struct {
	Clients []client.Client `json:"clients"`
}

func (c *Clients) Register(methods api.Methods) {
	methods["crm.clients.list"] = c.HandleList
}

func (c *Clients) HandleList(db *pgme.Database, state *websocket.State, data json.RawMessage) (any, jsonrpc2.Error) {
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
				    c.id, 
				    c.name, 
				    c.created
				FROM crm.clients c
				WHERE (c.name ~* $1::text OR c.name ~* $1::text)
				LIMIT 100
				`,
				params.Filter,
			)
			response := Response{}
			if response.Clients, err = pgx.CollectRows[client.Client](rows, pgx.RowToStructByNameLax[client.Client]); err == nil {
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
