package files

import (
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"github.com/vovka1200/pgme"
	"strings"
)

type Files struct {
}

func (f *Files) GetHandler(ctx *fasthttp.RequestCtx, db *pgme.Database) error {
	id := ctx.UserValue("id").(string)
	log.WithFields(log.Fields{
		"id": id,
		"ip": ctx.Conn().RemoteAddr(),
	}).Debug("Файл")
	if conn, err := db.NewConnection(ctx); err == nil {
		defer db.Disconnect(conn)
		row := conn.QueryRow(ctx, "SELECT mime,body FROM files.files WHERE id=$1", id)
		var body []byte
		var mime string
		if err := row.Scan(&mime, &body); err == nil {
			log.WithFields(log.Fields{
				"id":   id,
				"size": len(body),
				"mime": mime,
				"ip":   ctx.Conn().RemoteAddr(),
			}).Debug("Файл")
			ctx.Response.Header.SetContentType(mime)
			ctx.Response.SetBody(body)
			return nil
		} else {
			log.Error(err)
			if strings.Contains(err.Error(), "no rows") {
				ctx.SetStatusCode(404)
			}
			if strings.Contains(err.Error(), "syntax") {
				ctx.SetStatusCode(400)
			}
			return err
		}
	} else {
		log.Error(err)
		return err
	}
}
