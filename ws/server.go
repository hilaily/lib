package ws

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// WSHandler 是一个通用的 WebSocket 处理器
type Server struct {
	upgrader websocket.Upgrader
	// 存储所有活跃的连接
	connections sync.Map
	// 消息处理器映射
	handler map[string]HandlerFunc
	// errorHandler func(conn *websocket.Conn, code common.AIAPPErrCode, err error, conversationID string)
	// 连接状态管理
	connState sync.Map
	log       *logrus.Entry
}

// NewServer 创建一个新的 WebSocket 处理器
func NewServer(ops ...ServerOptions) (*Server, error) {
	s := &Server{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源，生产环境中应该更严格
			},
		},
		handler: make(map[string]HandlerFunc),
	}
	for _, opt := range ops {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	if s.log == nil {
		s.log = logrus.WithField("module", "ws")
	}
	if _, exists := s.handler[MessageTypePing]; !exists {
		s.RegisterHandler(MessageTypePing, s.handlPingMessage)
	}
	return s, nil
}

// RegisterHandler 注册消息处理器
func (h *Server) RegisterHandler(msgType string, handler HandlerFunc) {
	if _, exists := h.handler[msgType]; exists {
		h.log.WithField("msgType", msgType).Panic("Handler already registered")
		return
	}
	h.handler[msgType] = handler
}

// Broadcast 向所有连接广播消息
func (h *Server) Broadcast(message any) {
	h.connections.Range(func(key, value any) bool {
		conn := value.(*websocket.Conn)
		if err := conn.WriteJSON(message); err != nil {
			h.log.WithError(err).Error("Broadcast failed")
		}
		return true
	})
}

func (h *Server) GinHandler() gin.HandlerFunc {
	return gin.WrapF(h.HandleConnection)
}

// HandleConnection 处理 WebSocket 连接
func (h *Server) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.WithError(err).Error("Failed to upgrade connection")
		return
	}

	connID := h.generateConnID()
	h.connections.Store(connID, conn)
	h.connState.Store(connID, true) // Mark connection as active

	// Create a done channel for cleanup coordination
	done := make(chan struct{})
	defer func() {
		close(done)
		conn.Close()
		h.connections.Delete(connID)
		h.connState.Delete(connID)
	}()

	for {
		// Check if connection is still active
		if active, ok := h.connState.Load(connID); !ok || !active.(bool) {
			return
		}

		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.log.WithError(err).Error("WebSocket read error")
			}
			// Mark connection as inactive before handling error
			h.connState.Store(connID, false)
			// h.errorHandler(conn, common.AIAPPErrCode_AGENT_WS_READ_FAILED, err, connID)
			// If it's a close error, return immediately
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived, websocket.CloseNormalClosure) {
				h.log.WithError(err).Info("WebSocket connection closed")
				return
			}
			continue
		}

		// 只处理文本消息
		if messageType != websocket.TextMessage {
			h.log.WithField("type", messageType).Warn("Received non-text message")
			// h.errorHandler(conn, common.AIAPPErrCode_AGENT_WS_TEXT_ONLY, err, "-1")
			continue
		}

		// 处理消息
		if err := h.handleMessage(conn, message, connID); err != nil {
			h.log.WithError(err).Error("Message handling failed")
			continue
		}
	}
}

// handleMessage 处理接收到的消息
func (h *Server) handleMessage(conn *websocket.Conn, message []byte, connID string) error {
	// 尝试解析为命令消息
	var typeMsg FromMessage
	err := json.Unmarshal(message, &typeMsg)
	if err != nil {
		return fmt.Errorf("failed to parse message: %w, data: %s", err, string(message))
	}

	if handler, exists := h.handler[typeMsg.Type]; exists {
		return handler(h.getWriter(conn), typeMsg.Data, connID)
	}

	return fmt.Errorf("unknown message type: %s, %w", typeMsg.Type, ErrUnknownMessageType)
}

func (h *Server) handlPingMessage(writer IWriter, message json.RawMessage, connID string) error {
	return writer.WriteJSON(ToMessage{
		Type: MessageTypePong,
	})
}

func (h *Server) getWriter(conn *websocket.Conn) IWriter {
	return &writer{
		c:    h,
		conn: conn,
	}
}

func (h *Server) generateConnID() string {
	// generateConnID 生成唯一的连接 ID
	return "conn_" + time.Now().Format("20060102150405.000") + strconv.Itoa(rand.Intn(100000000))
}

type writer struct {
	conn *websocket.Conn
	c    *Server
}

func (w *writer) WriteJSON(message any) error {
	return w.conn.WriteJSON(message)
}

func (w *writer) Broadcast(message any) {
	w.c.Broadcast(message)
}
