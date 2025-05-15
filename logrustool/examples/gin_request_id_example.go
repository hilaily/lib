package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hilaily/lib/logrustool"
	"github.com/sirupsen/logrus"
)

func main() {
	// 配置 logrus
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	// 创建 Gin 引擎
	r := gin.New()

	// 使用自定义中间件
	r.Use(logrustool.RequestIDMiddleware())
	r.Use(gin.Recovery())

	// 定义路由
	r.GET("/hello", func(c *gin.Context) {
		// 使用带有请求 ID 的 logger
		logger := logrustool.RequestLogger(c)

		// 所有日志都会自动包含请求 ID
		logger.Info("处理 /hello 请求")
		logger.WithField("custom_field", "custom_value").Info("带有自定义字段的日志")

		// 也可以直接使用 logrus，请求 ID 也会被添加
		logrus.Info("使用全局 logrus")

		c.JSON(http.StatusOK, gin.H{
			"message":    "Hello, world!",
			"request_id": logrustool.GetRequestID(c),
		})
	})

	// 启动服务器
	logrus.Info("服务启动在 :8080")
	r.Run(":8080")
}
