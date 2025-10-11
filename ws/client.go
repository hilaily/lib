package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hilaily/kit/pool"
	"github.com/sirupsen/logrus"
)

var (
	defaultConnectConfig = ConnectConfig{
		// 默认配置
		reconnectEnabled:      true,
		reconnectInitialDelay: 500 * time.Millisecond,
		reconnectMaxDelay:     10 * time.Second,
		maxReconnectRetries:   -1,
		readTimeout:           60 * time.Second,
		heartbeatInterval:     30 * time.Second,
	}
)

type Client struct {
	conn        *websocket.Conn
	url         string
	mu          sync.RWMutex
	isConnected bool

	// 内部状态
	reconnecting bool
	pool         pool.IPool

	connectConfig *ConnectConfig

	// 事件处理器
	// onOperationUpdate func(data OperationData)
	onError      func(err error)
	onConnect    func()
	onDisconnect func()

	handler map[string]ClientHandlerFunc

	// 控制通道
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}

	log *logrus.Entry
}

// 创建新的WebSocket客户端
func NewClient(_url string, ops ...ClientOptions) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())

	w := &Client{
		url:           _url,
		connectConfig: &defaultConnectConfig,
		ctx:           ctx,
		cancel:        cancel,
		done:          make(chan struct{}),
		pool:          pool.NewPool(10),
		log:           logrus.WithField("module", "ws"),
	}
	for _, opt := range ops {
		if err := opt(w); err != nil {
			return nil, err
		}
	}

	w.startHeartbeat()
	err := w.connect()
	if err != nil {
		return nil, err
	}
	return w, nil
}

// 断开连接
func (c *Client) Disconnect() error {
	c.cancel()

	c.mu.Lock()
	defer c.mu.Unlock()

	// 关闭后不再自动重连
	c.connectConfig.reconnectEnabled = false

	if c.conn != nil {
		c.isConnected = false
		err := c.conn.Close()
		c.conn = nil
		if err != nil {
			return err
		}
	}

	// 安全关闭 done（避免重复关闭 panic）
	select {
	case <-c.done:
		// 已关闭
	default:
		close(c.done)
	}

	return nil
}

// 检查是否已连接
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isConnected
}

// 发送消息
func (c *Client) SendMessage(message any) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isConnected || c.conn == nil {
		return fmt.Errorf("websocket not connected, isConnected: %v, conn is empty: %v", c.isConnected, c.conn == nil)
	}

	err := c.conn.WriteJSON(message)
	if err != nil {
		return fmt.Errorf("send message failed %w", err)
	}
	return nil
}

func (c *Client) RegisterHandler(msgType string, handler ClientHandlerFunc) {
	if _, exists := c.handler[msgType]; exists {
		c.log.WithField("msgType", msgType).Panic("Handler already registered")
		return
	}
	c.handler[msgType] = handler
}

// 等待连接关闭
func (c *Client) Wait() {
	<-c.done
}

// 连接到WebSocket服务器
func (c *Client) connect() error {
	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("parse URL: %w", err)
	}

	c.log.Infof("连接到WebSocket服务器: %s", c.url)

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("dial websocket: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.isConnected = true
	c.mu.Unlock()

	// 启动消息处理协程
	go c.readMessages()

	// 触发连接事件
	if c.onConnect != nil {
		c.onConnect()
	}

	return nil
}

// 在 readMessages 中添加心跳处理
func (c *Client) startHeartbeat() {
	interval := c.connectConfig.heartbeatInterval
	if interval <= 0 {
		interval = 30 * time.Second
	}
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := c.SendMessage(&ToMessage{Type: MessageTypePing}); err != nil {
					c.log.Errorf("发送心跳失败: %v", err)
				}
			case <-c.ctx.Done():
				return
			}
		}
	}()
}

// 读取消息的协程
func (c *Client) readMessages() {
	var err error
	defer func() {
		if err != nil {
			c.log.Errorf("readMessages defer, err: %v", err)
		}
		c.mu.Lock()
		c.isConnected = false
		c.mu.Unlock()

		if c.onDisconnect != nil {
			c.onDisconnect()
		}

		// 根据上下文与配置决定是否自动重连
		select {
		case <-c.ctx.Done():
			// 终止：关闭 done
			select {
			case <-c.done:
			default:
				close(c.done)
			}
		default:
			if c.connectConfig.reconnectEnabled {
				go c.reconnectLoop()
			} else {
				// 不重连：关闭 done
				select {
				case <-c.done:
				default:
					close(c.done)
				}
			}
		}
	}()

	for {
		select {
		case <-c.ctx.Done():
			err = fmt.Errorf("readMessages context done")
			return
		default:
			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()

			if conn == nil {
				err = fmt.Errorf("readMessages conn is nil")
				return
			}

			// 设置读取超时
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))

			_, message, e := conn.ReadMessage()
			if e != nil {
				if websocket.IsUnexpectedCloseError(e, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.log.Errorf("WebSocket读取错误: %v", e)
					if c.onError != nil {
						c.onError(e)
					}
				}
				err = fmt.Errorf("readMessages failed, %v", e)
				return
			}

			c.handleMessage(message)
		}
	}
}

// 重连循环：指数回退直到成功或达到最大重试次数
func (c *Client) reconnectLoop() {
	c.mu.Lock()
	if c.reconnecting || c.ctx.Err() != nil {
		c.mu.Unlock()
		return
	}
	c.reconnecting = true
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.reconnecting = false
		c.mu.Unlock()
	}()

	attempt := 0
	delay := c.connectConfig.reconnectInitialDelay
	if delay <= 0 {
		delay = 500 * time.Millisecond
	}
	maxDelay := c.connectConfig.reconnectMaxDelay
	if maxDelay <= 0 {
		maxDelay = 10 * time.Second
	}

	for {
		if c.ctx.Err() != nil {
			return
		}

		// 尝试连接
		c.log.Infof("WebSocket重连，第%d次", attempt+1)
		if err := c.tryReconnect(); err == nil {
			return
		} else {
			c.log.WithError(err).Warnf("WebSocket重连失败，第%d次，%s后重试", attempt+1, delay)
		}

		attempt++
		if c.connectConfig.maxReconnectRetries > 0 && attempt >= c.connectConfig.maxReconnectRetries {
			c.log.Warn("达到最大重连次数，停止重连")
			// 不再重连：关闭 done
			select {
			case <-c.done:
			default:
				close(c.done)
			}
			return
		}

		timer := time.NewTimer(delay)
		select {
		case <-timer.C:
		case <-c.ctx.Done():
			timer.Stop()
			return
		}

		// 指数回退
		delay *= 2
		if delay > maxDelay {
			delay = maxDelay
		}
	}
}

func (c *Client) tryReconnect() error {
	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("parse URL: %w", err)
	}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("dial websocket: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.isConnected = true
	c.mu.Unlock()

	c.log.Info("WebSocket重连成功")

	// 触发连接事件
	if c.onConnect != nil {
		c.onConnect()
	}

	go c.readMessages()
	return nil
}

// 处理接收到的消息
func (c *Client) handleMessage(message []byte) {
	var baseMsg FromMessage
	if err := json.Unmarshal(message, &baseMsg); err != nil {
		c.log.WithError(err).Errorf("解析消息失败: %v", err)
		if c.onError != nil {
			c.onError(err)
		}
		return
	}

	// c.log.Debugf("收到消息: %v", baseMsg)
	msgType := baseMsg.Type
	switch msgType {
	case MessageTypePong:
		return
	default:
		handler, exists := c.handler[msgType]
		if exists {
			c.pool.Go(
				func() {
					err := handler(c.getWriter(), baseMsg.Data)
					if err != nil {
						c.log.WithError(err).Errorf("处理消息失败: %v", err)
						c.log.Debugf("处理消息失败: %v, msgType: %s, data: %s", err, msgType, string(baseMsg.Data))
					}
				},
			)
		} else {
			c.log.WithField("msgType", msgType).Warn("未找到处理程序")
		}
	}
}

func (c *Client) getWriter() IClientWriter {
	return &clientWriter{c: c}
}

type clientWriter struct {
	c *Client
}

func (w *clientWriter) WriteJSON(message any) error {
	return w.c.conn.WriteJSON(message)
}

// 重连配置
type ConnectConfig struct {
	reconnectEnabled      bool
	reconnectInitialDelay time.Duration
	reconnectMaxDelay     time.Duration
	maxReconnectRetries   int // -1 表示无限重试
	readTimeout           time.Duration
	heartbeatInterval     time.Duration
}
