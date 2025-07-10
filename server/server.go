package server

import (
	"github.com/fasthttp/router"
	"github.com/fasthttp/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	v1 "github.com/vovka1200/tpss-go-back/api/v1"
)

type Server struct {
	Listen   string `json:"listen"`
	upgrader *websocket.FastHTTPUpgrader
	API      struct {
		V1 v1.API `json:"v1"`
	} `json:"API"`
}

func (s *Server) Run() {
	log.Info("Запуск сервера")
	s.upgrader = &websocket.FastHTTPUpgrader{
		CheckOrigin: s.checkOrigin,
	}
	r := router.New()
	r.GET("/api/v1", s.webSocketHandlerV1)
	log.Info("Роутер инициализирован")
	if err := fasthttp.ListenAndServe(s.Listen, r.Handler); err != nil {
		log.Fatal(err)
	}
}

func (s *Server) webSocketHandlerV1(ctx *fasthttp.RequestCtx) {
	if err := s.upgrader.Upgrade(ctx, s.API.V1.Handler); err != nil {
		log.Error(err)
	}
}

func (s *Server) checkOrigin(ctx *fasthttp.RequestCtx) bool {
	return true
}
