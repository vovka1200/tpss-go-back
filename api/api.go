package api

import (
	"encoding/json"
	"github.com/vovka1200/pgme"
	"github.com/vovka1200/tpss-go-back/jsonrpc2"
	"github.com/vovka1200/tpss-go-back/websocket"
)

type Handler func(*pgme.Database, *websocket.State, json.RawMessage) (any, jsonrpc2.Error)
type Methods map[string]Handler
