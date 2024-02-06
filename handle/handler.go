package handle

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bytedance/go-tagexpr/v2/binding"
)

const (
	applicationJSON = "application/json"
)

type defaultHandler struct{}

// Read binds the request parameters and validates them,
// throws error status to response if there is error.
func (d *defaultHandler) Read(req *http.Request, args interface{}) {
	err := binding.BindAndValidate(args, req, nil)
	if err != nil {
		d.Throw(400, "fail to bind or validate request parameter", err)
	}
}

func (d *defaultHandler) Return(r http.ResponseWriter, data interface{}, err error, msg string) {
	d.Check(err, msg)
	d.WriteOK(r, data)
}

func (d *defaultHandler) WriteOK(r http.ResponseWriter, data interface{}) {
	d.writeJSON(r, 200, data)
}

// WriteErr writes a response with default struct.
func (d *defaultHandler) WriteErr(c http.ResponseWriter, httpStatus int, code int, msg string) {
	d.defaultWriteErrWithCause(c, httpStatus, code, msg, "")
}

// Throw 抛出异常
func (d *defaultHandler) Throw(code int, msg string, err ...error) {
	panic(newException(code, msg, err...))
}

// Check 检测错误
func (d *defaultHandler) Check(err error, msg ...string) {
	if err != nil {
		var m string
		if len(msg) > 0 {
			m = msg[0]
		} else {
			m = err.Error()
		}
		panic(newException(500, m, err))
	}
}

func (d *defaultHandler) defaultWriteErrWithCause(c http.ResponseWriter, httpStatus int, code int, msg, cause string) {
	d.writeJSON(c, httpStatus, map[string]interface{}{
		"code":  code,
		"msg":   msg,
		"cause": cause,
	})
}

func (d *defaultHandler) writeJSON(c http.ResponseWriter, httpStatus int, data interface{}) {
	c.WriteHeader(httpStatus)
	c.Header().Set("Content-Type", applicationJSON)
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		panic("marshal response: " + err.Error())
	}
	_, _ = c.Write(jsonBytes)
}

func catch(r interface{}) (*Exception, error) {
	switch v := r.(type) {
	case nil:
		return nil, nil
	case *Exception:
		if v == nil {
			v = new(Exception)
		}
		return v, nil
	case error:
		return nil, v
	default:
		return nil, fmt.Errorf("%v", r)
	}
}
