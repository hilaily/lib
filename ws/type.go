package ws

import (
	"encoding/json"
	"fmt"
)

// WebSocket消息类型
const (
	MessageTypePing = "ping"
	MessageTypePong = "pong"
)

var (
	ErrUnknownMessageType   = fmt.Errorf("unknown message type")
	ErrInvalidMessageFormat = fmt.Errorf("invalid message format")
)

type ClientOptions = func(*Client) error
type ServerOptions = func(*Server) error

// RequestMessage 命令消息
type RequestMessage struct {
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data,omitempty"`
	Timestamp int64           `json:"timestamp,omitempty"`
	Code      int             `json:"code,omitempty"`
	Message   string          `json:"message,omitempty"`
}

// ResponseMessage 响应消息
type ResponseMessage struct {
	Type      string `json:"type"`
	Data      any    `json:"data,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
	Code      int    `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
}

type IWriter interface {
	WriteJSON(message any) error
	Broadcast(message any)
}

type IClientWriter interface {
	WriteJSON(message any) error
}

type HandlerFunc func(writer IWriter, message json.RawMessage, connID string) error
type ClientHandlerFunc func(writer IClientWriter, message json.RawMessage) error
