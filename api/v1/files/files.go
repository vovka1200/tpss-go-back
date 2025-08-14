package files

import (
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"github.com/vovka1200/pgme"
	"github.com/vovka1200/tpss-go-back/api/v1/access/users/user"
	"io"
	"strings"
)

type Files struct {
}

type saveHandler func(*pgme.Database, string, string, []byte) (string, error)

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
			return err
		}
	} else {
		log.Error(err)
		return err
	}
	return nil
}

func (f *Files) PostHandler(ctx *fasthttp.RequestCtx, db *pgme.Database) error {
	space := ctx.UserValue("space").(string)
	if form, err := ctx.Request.MultipartForm(); err == nil {
		var ids []string
		for _, filesArray := range form.File {
			for _, f := range filesArray {
				log.WithFields(log.Fields{
					"space": space,
					"name":  f.Filename,
					"size":  f.Size,
					"mime":  f.Header.Get("Content-Type"),
					"ip":    ctx.Conn().RemoteAddr(),
				}).Debug("Файл")
				if content, err := f.Open(); err == nil {
					if buf, err := io.ReadAll(content); err == nil {
						if writer := serve(space); writer != nil {
							if id, err := writer(db, f.Filename, f.Header.Get("Content-Type"), buf); err == nil {
								ids = append(ids, id)
							}
						}
					}
				}
			}
		}
		ctx.Response.SetBody([]byte(strings.Join(ids, "\n")))
		return nil
	} else {
		log.Error(err)
		return err
	}
}

func serve(space string) saveHandler {
	if space == "avatar" {
		return user.UploadAvatar
	}
	return nil
}
