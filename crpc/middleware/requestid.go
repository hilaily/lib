package middleware

import (
	"context"

	"connectrpc.com/connect"
	"github.com/hilaily/lib/logrustool"
)

func RequestIDMiddleware() connect.UnaryInterceptorFunc {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			requestID := req.Header().Get(logrustool.RequestIDKey)
			if requestID == "" {
				requestID = logrustool.GenerateRequestID()
				req.Header().Set(logrustool.RequestIDKey, requestID)
			}

			// 将请求 ID 添加到上下文
			ctx = context.WithValue(ctx, logrustool.RequestIDKey, requestID)
			res, err := next(ctx, req)
			if res != nil {
				res.Header().Set(logrustool.RequestIDHeaderKey, requestID)
			}
			return res, err
		}
	})
}
