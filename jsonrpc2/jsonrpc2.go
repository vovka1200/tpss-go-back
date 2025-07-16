package jsonrpc2

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

const MethodNotFound int = -32601
const InvalidParams int = -32602
const ParseError int = -32700
const InternalError int = -32603
const Unauthorized int = 401

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
	Error   Error           `json:"error,omitempty"`
}

type RPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type Error *RPCError

func UnmarshalParams[T any](data []byte, v *T) error {
	if data == nil {
		return nil
	}
	return json.Unmarshal(data, v)
}

func Marshal(v any) (json.RawMessage, Error) {
	if b, err := json.Marshal(v); err == nil {
		return b, nil
	} else {
		log.Error(err)
		return nil, &RPCError{
			Code:    InternalError,
			Message: err.Error(),
		}
	}
}

func NewResponse(id any) Response {
	return Response{
		JSONRPC: "2.0",
		ID:      id,
	}
}

func NewErrorResponse(id any, code int, message string) Response {
	return Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
	}
}

func Handler(msg []byte, handler func(request Request) (any, Error)) ([]byte, error) {
	var request Request
	var response Response
	var err error
	var buffer []byte
	var result any
	if err = json.Unmarshal(msg, &request); err == nil {
		response = NewResponse(request.ID)
		if result, response.Error = handler(request); err == nil {
			if response.Result, err = json.Marshal(result); err == nil {
				if buffer, err = json.Marshal(response); err == nil {
					return buffer, nil
				}
			}
		}
	} else {
		response = NewErrorResponse(request.ID, ParseError, err.Error())
	}
	log.Error(err)
	if buffer, err := json.Marshal(response); err == nil {
		return buffer, nil
	} else {
		log.Error(err)
		return nil, err
	}
}
