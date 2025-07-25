package user

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
	"github.com/vovka1200/pgme"
	"github.com/vovka1200/tpss-go-back/api"
	"github.com/vovka1200/tpss-go-back/api/v1/access/matrix"
	"github.com/vovka1200/tpss-go-back/jsonrpc2"
	"github.com/vovka1200/tpss-go-back/websocket"
	"time"
)

const AuthenticationMethod = "access.users.user.login"

type User struct {
	Id       string    `json:"id"`
	Username string    `json:"username"`
	Name     string    `json:"name"`
	Created  time.Time `json:"created"`
	Groups   []string  `json:"groups"`
}

func (u *User) Register(methods api.Methods) {
	methods[AuthenticationMethod] = u.HandleAuthentication
}

type AuthorizeParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

type AuthorizeResponse struct {
	Account User         `json:"account"`
	Matrix  matrix.Rules `json:"matrix"`
	Token   string       `json:"token"`
}

func (u *User) HandleAuthentication(db *pgme.Database, state *websocket.State, data json.RawMessage) (any, jsonrpc2.Error) {
	params := AuthorizeParams{}
	var err error
	var conn *pgxpool.Conn
	if err = jsonrpc2.UnmarshalParams[AuthorizeParams](data, &params); err == nil {
		log.WithFields(log.Fields{
			"username": params.Username,
			"token":    params.Token,
			"ip":       state.Conn.RemoteAddr(),
		}).Info("Параметры")
		ctx := context.Background()
		if conn, err = db.NewConnection(ctx); err == nil {
			defer db.Disconnect(conn)
			rows, _ := conn.Query(ctx, `
				SELECT 
				    jsonb_build_object(
				    	'id',u.id, 
				    	'name',u.name, 
				    	'username',u.username,
				    	'groups',jsonb_agg(g.name),
				    	'created', u.created
				    ) AS account,
				    (SELECT jsonb_agg(
							jsonb_build_object(
								'object', o.name,
								'access', r.access
							)
						)
				     FROM access.rules r
				     JOIN access.objects o ON o.id = r.object_id
					 JOIN access.members m ON m.group_id=r.group_id
				     WHERE m.user_id=u.id
				    ) AS matrix,
				    (SELECT token FROM access.add_session(u.id)) as token
				FROM access.users u
				JOIN access.members m ON m.user_id=u.id
				JOIN access.groups g ON g.id=m.group_id
				LEFT JOIN access.sessions s ON s.user_id=u.id
				WHERE username=$1 AND password=crypt($2,password) 
				   OR 
				      s.token=$3 AND NOT s.archived
				GROUP BY u.id`,
				params.Username,
				params.Password,
				params.Token,
			)
			response := AuthorizeResponse{}
			if response, err = pgx.CollectOneRow[AuthorizeResponse](rows, pgx.RowToStructByNameLax[AuthorizeResponse]); err == nil {
				state.UserId = response.Account.Id
				log.WithFields(log.Fields{
					"username": params.Username,
					"ip":       state.Conn.RemoteAddr(),
				}).Info("Авторизован")
				return response, nil
			} else {
				log.Error(err)
				if errors.Is(err, pgx.ErrNoRows) {
					return nil, &jsonrpc2.RPCError{
						Code:    jsonrpc2.Unauthorized,
						Message: "authorization failed",
					}
				}
			}
		}
	} else {
		log.Error(err)
		return nil, &jsonrpc2.RPCError{
			Code:    jsonrpc2.InvalidParams,
			Message: err.Error(),
		}
	}
	log.Error(err)
	return nil, &jsonrpc2.RPCError{
		Code:    500,
		Message: err.Error(),
	}
}
