package errorx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
)

var (
	// biz error - 使用常量定义错误码
	CodeBadRequest       uint32 = 400
	CodeInvalidArgument  uint32 = 400
	CodeAuth             uint32 = 401
	CodePermissionDenied uint32 = 403
	CodeNotFound         uint32 = 404
	CodeAlreadyExists    uint32 = 409
	CodeCanceled         uint32 = 499

	// system error
	CodeInternal uint32 = 500
	CodeUnknown  uint32 = 500
	CodeTimeout  uint32 = 504

	// 预定义错误类型 - 这些是模板，不应该直接修改
	BadRequest       = &Err{code: CodeBadRequest, message: "bad_request"}
	InvalidArgument  = &Err{code: CodeInvalidArgument, message: "invalid_argument"}
	Auth             = &Err{code: CodeAuth, message: "auth_failed"}
	PermissionDenied = &Err{code: CodePermissionDenied, message: "permission_denied"}
	NotFound         = &Err{code: CodeNotFound, message: "not_found"}
	AlreadyExists    = &Err{code: CodeAlreadyExists, message: "already_exists"}
	Canceled         = &Err{code: CodeCanceled, message: "operation_canceled"}

	// system error
	Internal = &Err{code: CodeInternal, message: "internal_server_error"}
	Unknown  = &Err{code: CodeUnknown, message: "unknown_error"}
	Timeout  = &Err{code: CodeTimeout, message: "operation_timeout"}
)

// New 创建一个新的错误
func New(code uint32, message string) *Err {
	return &Err{code: code, message: message}
}

// Errorf 创建一个内部服务器错误
func Errorf(format string, msg ...any) *Err {
	return Internal.SetErr(fmt.Errorf(format, msg...))
}

type Err struct {
	code    uint32
	message string
	err     error
	params  map[string]any
	extra   map[string]any
}

func (e *Err) MarshalJSON() ([]byte, error) {
	detail := map[string]any{}
	if e.err != nil {
		detail["error"] = e.err.Error()
	}
	if len(e.extra) > 0 {
		detail["extra"] = e.extra
	}
	if len(e.params) > 0 {
		detail["params"] = e.params
	}

	return json.Marshal(map[string]any{
		"code":    e.code,
		"message": e.message,
		"detail":  detail,
	})
}

func (e *Err) Error() string {
	buf := bytes.NewBufferString(fmt.Sprintf("code: %d, message: %s", e.code, e.message))
	if e.err != nil {
		buf.WriteString(fmt.Sprintf(", err: %s", e.err.Error()))
	}
	if len(e.extra) > 0 {
		buf.WriteString(fmt.Sprintf(", extra: %v", e.extra))
	}
	return buf.String()
}

func (e *Err) GetUnderlyingError() error {
	return e.err
}

// Clone 创建错误的副本，解决并发安全问题
func (e *Err) Clone() *Err {
	return &Err{
		code:    e.code,
		message: e.message,
		err:     e.err,
		extra:   e.extra,
	}
}

// SetMsg 设置错误消息（返回新的副本）
func (e *Err) SetMsg(format string, msg ...any) *Err {
	clone := e.Clone()
	clone.message = fmt.Sprintf(format, msg...)
	return clone
}

// SetErr 设置底层错误（返回新的副本）
func (e *Err) SetErr(err error) *Err {
	clone := e.Clone()
	clone.err = err
	return clone
}

// Unwrap 返回底层错误，支持 errors.Is 和 errors.As
func (e *Err) Unwrap() error {
	return e.err
}

func (e *Err) GetCode() uint32 {
	return e.code
}

func (e *Err) GetMsg() string {
	return e.message
}

func (e *Err) GetParams() map[string]any {
	return e.params
}

// WithStack 添加堆栈信息
func (e *Err) WithStack() *Err {
	clone := e.Clone()
	if clone.extra == nil {
		clone.extra = make(map[string]any)
	}
	clone.extra["stack"] = e.stackTrace()
	return clone
}

func (e *Err) WithExtra(key string, value any) *Err {
	clone := e.Clone()
	if clone.extra == nil {
		clone.extra = make(map[string]any)
	}
	clone.extra[key] = value
	return clone
}

func (e *Err) WithParams(params map[string]any) *Err {
	clone := e.Clone()
	clone.params = params
	return clone
}

// StackTrace 获取格式化的堆栈跟踪
func (e *Err) stackTrace() string {
	stack := make([]uintptr, 32)
	n := runtime.Callers(3, stack)
	frames := runtime.CallersFrames(stack[:n])
	var trace string
	for {
		frame, more := frames.Next()
		trace += fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
		if !more {
			break
		}
	}
	return trace
}

// Is 检查错误类型，支持 errors.Is
func (e *Err) Is(target error) bool {
	if t, ok := target.(*Err); ok {
		return e.code == t.code
	}
	return false
}
