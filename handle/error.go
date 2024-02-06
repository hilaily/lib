package handle

import (
	"errors"
	"fmt"
)

var (
	_ IErr = &Err{}
)

type IErr interface {
	ErrCode() int // add E prefix, because Code, Msg may be used for structure field.
	ErrMsg() string
	ExtraInfo() string
	Error() string
}

// New create a new error
func New(code int, msg string, extra ...string) *Err {
	return NewErr(code, msg, extra...)
}

func NewErr(code int, msg string, extra ...string) *Err {
	e := &Err{
		Code: code,
		Msg:  msg,
	}
	if len(extra) > 0 {
		e.Extra = extra[0]
	}
	return e
}

type Err struct {
	Code  int
	Msg   string
	Extra string
}

func (e *Err) ErrCode() int {
	return e.Code
}

func (e *Err) ErrMsg() string {
	return e.Msg
}

func (e *Err) ExtraInfo() string {
	return e.Extra
}

// Clone a new err with new msg, code is the same
func (e *Err) Clone(format string, a ...any) *Err {
	return &Err{
		Code:  e.Code,
		Msg:   fmt.Sprintf(format, a...),
		Extra: e.Extra,
	}
}

func (e *Err) Format(a ...any) *Err {
	return e.Clone(e.Msg, a...)
}

func (e *Err) SetExtra(format string, a ...any) *Err {
	return &Err{
		Code:  e.Code,
		Msg:   e.Msg,
		Extra: fmt.Sprintf(format, a...),
	}
}

func (e *Err) Error() string {
	return fmt.Sprintf("code: %d, msg: %s, extra: %s", e.Code, e.Msg, e.Extra)
}

// Unwrap unwrap error to get custom err object
//
//nolint:errorlint
func Unwrap[T any](err error) (T, bool) {
	var current = err
	var last = err
	for {
		current = errors.Unwrap(current)
		if current == nil {
			err = last
			break
		}
		last = current
	}
	v, ok := err.(T)
	return v, ok
}
