package server

import (
	"github.com/fasthttp/router"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"github.com/vovka1200/pgme"
	v1 "github.com/vovka1200/tpss-go-back/api/v1"
	"github.com/vovka1200/tpss-go-back/api/v1/access/users/user"
	"github.com/vovka1200/tpss-go-back/api/v1/files"
	"github.com/vovka1200/tpss-go-back/jsonrpc2"
	"github.com/vovka1200/tpss-go-back/websocket"
)

type Server struct {
	Listen string `json:"listen"`
	API    struct {
		V1 v1.API `json:"v1"`
	} `json:"API"`
	Database pgme.Database `json:"database"`
	Files    files.Files
}

func (s *Server) Run() {
	log.Info("Запуск сервера")
	if err := s.Database.InitPool(); err != nil {
		log.Fatal(err)
	}
	log.Info("База данных подключена")
	s.API.V1.Register()
	log.Info("API v1 инициализирован")
	r := router.New()
	r.GET("/api/v1", s.webSocketHandlerV1)
	r.GET("/api/v1/file/{id}", func(ctx *fasthttp.RequestCtx) {
		s.Files.GetHandler(ctx, &s.Database)
	})
	r.POST("/api/v1/users/{id}/avatar", func(ctx *fasthttp.RequestCtx) {
		if ids, err := user.UploadAvatar(ctx, &s.Database); err == nil {
			ctx.SetBody([]byte(ids))
		} else {
			ctx.SetStatusCode(500)
			ctx.SetBody([]byte(err.Error()))
		}
	})
	log.Info("Роутер инициализирован")
	if err := fasthttp.ListenAndServe(s.Listen, r.Handler); err != nil {
		log.Fatal(err)
	}
}

func (s *Server) webSocketHandlerV1(ctx *fasthttp.RequestCtx) {
	log.WithFields(log.Fields{
		"addr": ctx.RemoteAddr(),
	}).Info("Соединение")
	// websockets layer
	if err := websocket.Handler(ctx, func(state *websocket.State, msg []byte) ([]byte, error) {
		// jsonrpc layer
		return jsonrpc2.Handler(msg, func(request jsonrpc2.Request) (any, jsonrpc2.Error) {
			log.WithFields(log.Fields{
				"id":     request.ID,
				"method": request.Method,
				"ip":     ctx.Conn().RemoteAddr(),
			}).Debug("Запрос")
			// API layer
			return s.API.V1.Handler(state, &s.Database, request.Method, request.Params)
		})
	}); err != nil {
		log.Error(err)
	}
}
