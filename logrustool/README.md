# LogrusTool

这个包提供了一些 Logrus 的扩展工具，包括 RequestID 钩子、日志级别设置等功能。

## RequestID 功能

`RequestID` 功能可以让你在 Gin Web 服务中为每个请求自动生成一个唯一的请求 ID，并在所有日志中包含这个 ID。

### 安装

```bash
go get github.com/hilaily/lib/logrustool
```

### 在 Gin 中使用 RequestID

```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/hilaily/lib/logrustool"
	"github.com/sirupsen/logrus"
)

func main() {
	// 配置 logrus
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// 创建 Gin 引擎
	r := gin.New()

	// 使用 RequestID 中间件
	r.Use(logrustool.RequestIDMiddleware())

	// 定义路由
	r.GET("/example", func(c *gin.Context) {
		// 方法 1: 获取带有请求 ID 的 logger
		logger := logrustool.RequestLogger(c)
		logger.Info("这条日志会包含请求 ID")

		// 方法 2: 直接使用 logrus，请求 ID 也会被自动添加
		logrus.Info("这条日志也会包含请求 ID")

		// 获取当前请求的 ID
		requestID := logrustool.GetRequestID(c)

		c.JSON(200, gin.H{
			"request_id": requestID,
		})
	})

	r.Run(":8080")
}
```

### 特性

1. 自动为每个请求生成唯一的请求 ID (UUID)
2. 支持从请求头 `X-Request-ID` 获取已有的请求 ID
3. 将请求 ID 添加到响应头
4. 每条日志都会自动包含请求 ID
5. 提供简便的方法获取当前请求的 ID

### 输出示例

请求日志 (JSON 格式):

```json
{
  "level": "info",
  "msg": "处理 /hello 请求",
  "request_id": "a7e0f9b3-9c0d-4b5e-8f6a-1c2d3e4f5a6b",
  "time": "2023-05-10T15:04:05Z"
}
```

## 其他功能

...（其他功能说明）
