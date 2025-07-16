package login

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	log "github.com/sirupsen/logrus"
	"github.com/vovka1200/pgme"
	"github.com/vovka1200/tpss-go-back/api/v1/access/users"
	"github.com/vovka1200/tpss-go-back/jsonrpc2"
)

const Method = "access.login"

type Login struct {
}

type Params struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Response struct {
	Account users.User `json:"account"`
}

func (l *Login) Register(methods jsonrpc2.Methods) {
	methods[Method] = l.Handler
}

func (l *Login) Handler(db *pgme.Database, userId string, data json.RawMessage) (any, *jsonrpc2.Error) {
	params := Params{}
	if err := jsonrpc2.UnmarshalParams[Params](data, &params); err == nil {
		log.WithFields(log.Fields{
			"username": params.Username,
		}).Info("Login")
		ctx := context.Background()
		if conn, err := db.NewConnection(ctx); err == nil {
			defer db.Disconnect(conn)
			rows, _ := conn.Query(ctx, `
				SELECT 
				    u.id, 
				    u.name, 
				    u.username,
				    array_agg(g.name) AS groups
				FROM access.users u
				JOIN access.members m ON m.user_id=u.id
				JOIN access.groups g ON g.id=m.group_id
				WHERE username=$1
				  AND password=crypt($2,password)
				GROUP BY 1,2,3`,
				params.Username,
				params.Password,
			)
			response := Response{}
			if response.Account, err = pgx.CollectOneRow[users.User](rows, pgx.RowToStructByNameLax[users.User]); err == nil {
				log.WithFields(log.Fields{
					"username": params.Username,
				}).Info("Авторизован")
				return response, nil
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
