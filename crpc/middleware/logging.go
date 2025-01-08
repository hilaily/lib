package middleware

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/sirupsen/logrus"
)

// NewLoggingInterceptor 创建一个日志拦截器
func LoggingMiddleware() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			// 记录开始时间
			start := time.Now()

			// 获取请求信息
			procedure := req.Spec().Procedure

			// 打印请求信息
			logrus.WithFields(logrus.Fields{
				"procedure": procedure,
				"headers":   req.Header(),
				"data":      req.Any(),
			}).Info("收到请求")

			// 调用下一个处理器
			res, err := next(ctx, req)

			// 计算处理时间
			duration := time.Since(start)

			// 打印响应信息
			fields := logrus.Fields{
				"procedure": procedure,
				"duration":  duration,
			}

			if err != nil {
				fields["error"] = err.Error()
				logrus.WithFields(fields).Error("请求失败")
			} else {
				fields["status"] = "success"
				logrus.WithFields(fields).Info("请求完成")
			}

			return res, err
		})
	}
}
