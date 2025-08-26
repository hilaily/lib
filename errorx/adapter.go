package errorx

import (
	"context"
	"errors"
	"maps"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/structpb"
)

type ConnectRPCAdapter struct{}

func (a *ConnectRPCAdapter) ToConnectRpcError(e *Err) *connect.Error {
	// 使用正确的 gRPC 状态码
	connectCode := a.ToConnectCode(e)
	connectErr := connect.NewError(connectCode, errors.New(e.message))

	// 如果有底层错误，添加到详细信息中
	details := map[string]any{}
	if e.err != nil {
		details["error"] = e.err.Error()
	}
	if len(e.extra) > 0 {
		maps.Copy(details, e.extra)

		// 如果有堆栈信息，也添加进去（仅在开发环境）
		if stackTrace := e.StackTrace(); stackTrace != "" {
			details["stack_trace"] = stackTrace
		}
	}

	if len(details) > 0 {
		if v, err := structpb.NewValue(details); err == nil {
			if detail, err := connect.NewErrorDetail(v); err == nil {
				connectErr.AddDetail(detail)
			}
		}
		// wrap 一下是因为 connectErr.Error() 不会返回 detail 里的信息，导致这个方法返回的 e 执行 e.Error() 里面只有 code 和 message 内容。
		// e.err = fmt.Errorf("%s, %w", e.Error(), connectErr)
	}
	return connectErr
}

func (a *ConnectRPCAdapter) ConnectRpcAuthInterceptor() connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			res, err := next(ctx, req)
			if err != nil {
				e, ok := err.(*Err)
				if ok {
					return res, a.ToConnectRpcError(e)
				}
				return res, err
			}
			return res, nil
		})
	}
	return connect.UnaryInterceptorFunc(interceptor)
}

func (a *ConnectRPCAdapter) ToConnectCode(e *Err) connect.Code {
	switch e.code {
	case CodeBadRequest, CodeInvalidArgument:
		return connect.CodeInvalidArgument // INVALID_ARGUMENT
	case CodeAuth:
		return connect.CodeUnauthenticated // UNAUTHENTICATED
	case CodePermissionDenied:
		return connect.CodePermissionDenied // PERMISSION_DENIED
	case CodeNotFound:
		return connect.CodeNotFound // NOT_FOUND
	case CodeAlreadyExists:
		return connect.CodeAlreadyExists // ALREADY_EXISTS
	case CodeCanceled:
		return connect.CodeCanceled // CANCELLED
	case CodeTimeout:
		return connect.CodeDeadlineExceeded // DEADLINE_EXCEEDED
	case CodeInternal, CodeUnknown:
		return connect.CodeInternal // INTERNAL
	default:
		return connect.Code(e.code)
	}
}

// HTTPAdapter 用于适配传统的 HTTP 错误处理
// type HTTPAdapter struct{}

// func (a *HTTPAdapter) Adapt(e *Err) error {
// 	// 可以返回包含 HTTP 状态码的自定义错误类型
// 	return &HTTPError{
// 		StatusCode: e.GetHTTPStatusCode(),
// 		Code:       e.Code,
// 		Message:    e.Message,
// 		Err:        e.err,
// 	}
// }

// // HTTPError HTTP 错误类型
// type HTTPError struct {
// 	StatusCode int
// 	Code       uint32
// 	Message    string
// 	Err        error
// }

// func (h *HTTPError) Error() string {
// 	if h.Err != nil {
// 		return fmt.Sprintf("HTTP %d (%d): %s: %s", h.StatusCode, h.Code, h.Message, h.Err.Error())
// 	}
// 	return fmt.Sprintf("HTTP %d (%d): %s", h.StatusCode, h.Code, h.Message)
// }
