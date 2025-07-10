package jsonrpc2

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

const MethodNotFound int = -32601
const InvalidParams int = -32602
const ParseError int = -32700
const InternalError int = -32603

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      any             `json:"id"`
}

type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

type Error struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type Handler func(json.RawMessage) (json.RawMessage, *Error)
type Methods map[string]Handler

func UnmarshalParams(data []byte, v any) *Error {
	if data == nil {
		return nil
	}
	if err := json.Unmarshal(data, v); err != nil {
		log.Error(err)
		return &Error{
			Code:    InvalidParams,
			Message: err.Error(),
		}
	}
	return nil
}

func Marshal(v any) (json.RawMessage, *Error) {
	if b, err := json.Marshal(v); err == nil {
		return b, nil
	} else {
		log.Error(err)
		return nil, &Error{
			Code:    InternalError,
			Message: err.Error(),
		}
	}
}
