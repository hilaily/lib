package logrustool

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	// RequestIDHeaderKey 是 HTTP 请求头中存储请求 ID 的键
	RequestIDHeaderKey = "X-Request-ID"
	// RequestIDKey 是 Gin 上下文中存储请求 ID 的键
	RequestIDKey = "request_id"
)

// RequestIDMiddleware 是一个 Gin 中间件，用于生成和处理请求 ID
func RequestIDMiddleware(options ...func(*RequestIDMiddlewareOptions)) gin.HandlerFunc {
	opt := &RequestIDMiddlewareOptions{
		Generator: GenerateRequestID,
	}
	for _, option := range options {
		option(opt)
	}
	return func(c *gin.Context) {
		// 从请求头中获取请求 ID，如果不存在则生成新的
		requestID := c.GetHeader(RequestIDHeaderKey)
		if requestID == "" {
			requestID = opt.Generator()
			c.Header(RequestIDHeaderKey, requestID)
		}

		// 将请求 ID 添加到 Gin 上下文
		c.Set(RequestIDKey, requestID)

		c.Next()
	}
}

// RequestIDMiddlewareOptions 是 RequestIDMiddleware 的选项
type RequestIDMiddlewareOptions struct {
	// Generator 是生成请求 ID 的函数
	Generator func() string
}

// WithGenerator 设置生成请求 ID 的函数
func WithGenerator(generator func() string) func(*RequestIDMiddlewareOptions) {
	return func(o *RequestIDMiddlewareOptions) {
		o.Generator = generator
	}
}

// GetRequestID 从 Gin 上下文中获取请求 ID
func GetRequestID(c *gin.Context) string {
	if id, exists := c.Get(RequestIDKey); exists {
		if requestID, ok := id.(string); ok {
			return requestID
		}
	}
	return ""
}

// RequestLogger 返回一个带有请求 ID 的 logrus.Entry
func RequestLogger(c *gin.Context) *logrus.Entry {
	return logrus.WithField(RequestIDKey, GetRequestID(c))
}

type RequestIDHook struct{}

func (h *RequestIDHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *RequestIDHook) Fire(entry *logrus.Entry) error {
	// 从 entry 上下文中获取请求 ID
	if entry.Context != nil {
		if requestID, ok := entry.Context.Value(RequestIDKey).(string); ok {
			entry.Data[RequestIDKey] = requestID
		}
	}
	return nil
}

// GenerateRequestID 生成一个唯一的请求 ID
func GenerateRequestID() string {
	return uuid.New().String()
}
