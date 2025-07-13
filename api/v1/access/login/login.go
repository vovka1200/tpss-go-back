package login

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	log "github.com/sirupsen/logrus"
	"github.com/vovka1200/pgme"
	"github.com/vovka1200/tpss-go-back/jsonrpc2"
)

type Login struct {
}

type Params struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Answer struct {
	Authorized bool `json:"authorized",query:"authorized"`
}

func (l *Login) Handler(db *pgme.Database, data json.RawMessage) (any, *jsonrpc2.Error) {
	params := Params{}
	if err := jsonrpc2.UnmarshalParams[Params](data, &params); err == nil {
		log.WithFields(log.Fields{
			"username": params.Username,
		}).Info("Login")
		ctx := context.Background()
		if conn, err := db.NewConnection(ctx); err == nil {
			defer db.Disconnect(conn)
			rows, _ := conn.Query(ctx, `
				SELECT true AS authorized
				FROM access.users
				WHERE username=$1
				  AND password=crypt($2,password)
				LIMIT 1;`,
				params.Username,
				params.Password,
			)
			if answer, err := pgx.CollectOneRow[Answer](rows, pgx.RowToStructByNameLax[Answer]); err == nil {
				log.WithFields(log.Fields{
					"username": params.Username,
				}).Info("Авторизован")
				return answer, nil
			} else {
				log.Error(err)
				return nil, &jsonrpc2.Error{
					Code:    401,
					Message: "Access denied",
				}
			}
		} else {
			log.Error(err)
			return nil, &jsonrpc2.Error{
				Code:    500,
				Message: err.Error(),
			}
		}
	} else {
		return nil, err
	}
}
