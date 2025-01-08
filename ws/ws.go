package ws

import (
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

var (
	ErrUnknownMessageType   = fmt.Errorf("unknown message type")
	ErrInvalidMessageFormat = fmt.Errorf("invalid message format")
)

// WSHandler 是一个通用的 WebSocket 处理器
type WSHandler struct {
	upgrader websocket.Upgrader
	// 存储所有活跃的连接
	connections sync.Map
	// 消息处理器映射
	handler      func(conn *websocket.Conn, payload []byte) error
	errorHandler func(conn *websocket.Conn, err error)
}

// NewWSHandler 创建一个新的 WebSocket 处理器
func NewWSHandler() *WSHandler {
	return &WSHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源，生产环境中应该更严格
			},
		},
	}
}

// RegisterHandler 注册消息处理器
func (h *WSHandler) RegisterHandler(handler func(conn *websocket.Conn, payload []byte) error) {
	h.handler = handler
}

func (h *WSHandler) RegisterErrorHandler(handler func(conn *websocket.Conn, err error)) {
	h.errorHandler = handler
}

// HandleConnection 处理 WebSocket 连接
func (h *WSHandler) HandleConnection(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.WithError(err).Error("Failed to upgrade connection")
		return
	}

	// 生成连接 ID 并存储连接
	connID := h.generateConnID()
	h.connections.Store(connID, conn)

	defer func() {
		conn.Close()
		h.connections.Delete(connID)
	}()

	for {
		messageType, message, err := conn.ReadMessage()
		// logrus.WithField("msg", string(message)).Debug("Received message")
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.WithError(err).Error("WebSocket read error")
			}
			if h.errorHandler != nil {
				h.errorHandler(conn, err)
			}
			return
		}

		// 只处理文本消息
		if messageType != websocket.TextMessage {
			logrus.WithField("type", messageType).Warn("Received non-text message")
			continue
		}

		if h.handler != nil {
			if err := h.handler(conn, message); err != nil {
				logrus.WithError(err).Error("Handler error")
				if h.errorHandler != nil {
					h.errorHandler(conn, err)
				}
			}
		}
	}
}

// SendError 发送错误响应
func (h *WSHandler) SendError(conn *websocket.Conn, err error) {
	if h.errorHandler != nil {
		h.errorHandler(conn, err)
		return
	}
	if err := conn.WriteJSON(map[string]string{"error": err.Error()}); err != nil {
		logrus.WithError(err).Error("Failed to send error response")
	}
}

// Broadcast 向所有连接广播消息
func (h *WSHandler) Broadcast(message interface{}) {
	h.connections.Range(func(key, value interface{}) bool {
		conn := value.(*websocket.Conn)
		if err := conn.WriteJSON(message); err != nil {
			h.SendError(conn, err)
		}
		return true
	})
}

func (h *WSHandler) generateConnID() string {
	// generateConnID 生成唯一的连接 ID
	return "conn_" + time.Now().Format("20060102150405.000") + strconv.Itoa(rand.Intn(100000000))
}
