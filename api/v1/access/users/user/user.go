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
	Account        User         `json:"account"`
	Matrix         matrix.Rules `json:"matrix"`
	Token          string       `json:"token"`
	TokenLiveUntil time.Time    `json:"token_live_until" db:"token_live_until"`
}

func (u *User) HandleAuthentication(db *pgme.Database, state *websocket.State, data json.RawMessage) (any, jsonrpc2.Error) {
	return jsonrpc2.QueryWithParams[AuthorizeParams](db, data, func(ctx context.Context, conn *pgxpool.Conn, params AuthorizeParams) (any, jsonrpc2.Error) {
		log.WithFields(log.Fields{
			"username": params.Username,
			"token":    params.Token,
			"ip":       state.Conn.RemoteAddr(),
		}).Info("Параметры")
		rows, _ := conn.Query(ctx, `
				SELECT 
				    jsonb_build_object(
				    	'id',u.id, 
				    	'name',u.name, 
				    	'username',u.username,
				    	'groups',jsonb_agg(g.name),
				    	'created', u.created
				    ) AS account,
				    (SELECT token FROM access.add_session(u.id)) as token,
				    now()+'1d'::interval as token_live_until
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
		if response, err := pgx.CollectOneRow[AuthorizeResponse](rows, pgx.RowToStructByNameLax[AuthorizeResponse]); err == nil {
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
					Message: "Требуется аутентификация",
				}
			} else {
				return nil, &jsonrpc2.RPCError{
					Code:    jsonrpc2.InternalError,
					Message: err.Error(),
				}
			}
		}
	})
}
