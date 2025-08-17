package user

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"github.com/vovka1200/pgme"
	"io"
	"strings"
)

func UploadAvatar(req *fasthttp.RequestCtx, db *pgme.Database) (string, error) {
	ctx := context.Background()
	if conn, err := db.NewConnection(ctx); err == nil {
		defer db.Disconnect(conn)
		id := req.UserValue("id").(string)
		var ids []string
		if form, err := req.Request.MultipartForm(); err == nil {
			for _, filesArray := range form.File {
				for _, f := range filesArray {
					log.WithFields(log.Fields{
						"name": f.Filename,
						"size": f.Size,
						"mime": f.Header.Get("Content-Type"),
						"ip":   req.Conn().RemoteAddr(),
					}).Debug("Файл")
					if content, err := f.Open(); err == nil {
						if buf, err := io.ReadAll(content); err == nil {
							row := conn.QueryRow(ctx, `
								WITH avatar AS (
									INSERT INTO files.avatars (name,mime,body) 
									VALUES ($2,$3,$4) 
									RETURNING id
								), users AS (
									UPDATE access.users
									SET avatar_id=(SELECT id FROM avatar)
									WHERE id=$1
								)
								SELECT id 
								FROM avatar
								`,
								id,
								f.Filename,
								f.Header.Get("Content-Type"),
								buf,
							)
							var id string
							if err := row.Scan(&id); err == nil {
								ids = append(ids, id)
							} else {
								log.Error(err)
								return "", err
							}
						}
					}
				}
			}
		}
		return strings.Join(ids, ","), nil
	} else {
		log.Error(err)
		return "", err
	}
}
