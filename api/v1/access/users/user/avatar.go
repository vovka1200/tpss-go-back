package user

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/vovka1200/pgme"
)

func UploadAvatar(db *pgme.Database, name string, mime string, body []byte) (string, error) {
	ctx := context.Background()
	if conn, err := db.NewConnection(ctx); err == nil {
		defer db.Disconnect(conn)
		row := conn.QueryRow(ctx, `
			INSERT INTO files.avatars (name,mime,body) 
			VALUES ($1,$2,$3) 
			RETURNING id`,
			name,
			mime,
			body,
		)
		var id string
		if err := row.Scan(&id); err == nil {
			return id, nil
		} else {
			log.Error(err)
			return "", err
		}
	} else {
		log.Error(err)
		return "", err
	}
}
