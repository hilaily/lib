package handle

// New creates a handling status with code, msg and cause.
// NOTE:
//  code=0 means no error
func newException(code int, msg string, errs ...error) *Exception {
	s := &Exception{
		code: code,
		msg:  msg,
	}
	if len(errs) > 0 {
		s.originError = errs[0]
	}
	return s
}

type Exception struct {
	code        int
	msg         string
	originError error
	*stack
}
