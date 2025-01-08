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
// receiveData is the data type of the message received, it should not be a pointer
type WSHandler[recvData any] struct {
	upgrader websocket.Upgrader
	// 存储所有活跃的连接
	connections sync.Map
	// 消息处理器映射
	handler      func(conn *websocket.Conn, data recvData) error
	errorHandler func(conn *websocket.Conn, err error)
}

// NewWSHandler 创建一个新的 WebSocket 处理器
// errorHandler is to process error, it can be nil, there is a default error handler
func NewHandler[recvData any](handler func(conn *websocket.Conn, data recvData) error, errorHandler func(conn *websocket.Conn, err error)) *WSHandler[recvData] {
	return &WSHandler[recvData]{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源，生产环境中应该更严格
			},
		},
		handler:      handler,
		errorHandler: errorHandler,
	}
}

// HandleConnection 处理 WebSocket 连接
func (h *WSHandler[recvData]) HandleConnection(c *gin.Context) {
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
		var message recvData
		err := conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.WithError(err).Error("WebSocket read error")
			}
			if h.errorHandler != nil {
				h.errorHandler(conn, err)
			}
			return
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

// Broadcast 向所有连接广播消息
func (h *WSHandler[recvData]) Broadcast(message recvData) {
	h.connections.Range(func(key, value interface{}) bool {
		conn := value.(*websocket.Conn)
		if err := conn.WriteJSON(message); err != nil {
			h.sendError(conn, err)
		}
		return true
	})
}

// SendError 发送错误响应
func (h *WSHandler[recvData]) sendError(conn *websocket.Conn, err error) {
	if h.errorHandler != nil {
		h.errorHandler(conn, err)
		return
	}
	if err := conn.WriteJSON(map[string]string{"error": err.Error()}); err != nil {
		logrus.WithError(err).Error("[ws] Failed to send error response")
	}
}

func (h *WSHandler[recvData]) generateConnID() string {
	// generateConnID 生成唯一的连接 ID
	return "conn_" + time.Now().Format("20060102150405.000") + strconv.Itoa(rand.Intn(100000000))
}
