package handle

import (
	"fmt"
	"math"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type IBizErr interface {
	ErrCode() int // add Err prefix, because Code, Msg may be used for other thing.
	ErrMsg() string
	ExtraInfo() string
}

type GinWriteErr = func(c *gin.Context, httpStatus, code int, msg ...string)

type Gin struct {
	defaultHandler
}

func (g *Gin) Read(c *gin.Context, args interface{}) {
	g.defaultHandler.Read(c.Request, args)
}

func (g *Gin) Return(c *gin.Context, data interface{}, err error, msg string) {
	if err != nil {
		g.WriteError(c, err)
		return
	}
	g.WriteOK(c, data)
}

// WriteOKCode write ok with code
func (g *Gin) WriteOK(c *gin.Context, data interface{}, codes ...int) {
	code := 0
	if len(codes) > 0 {
		code = codes[0]
	}
	c.JSON(200, map[string]interface{}{
		"code": code,
		"data": data,
	})
}

func (g *Gin) WriteError(c *gin.Context, err error) {
	v, ok := Unwrap[IErr](err)
	if ok {
		// it is a biz error
		g.WriteBizErr(c, v)
		return
	}
	g.WriteSysErr(c, err.Error())
}

func (g *Gin) WriteBizErr(c *gin.Context, b IBizErr) {
	c.JSON(200, map[string]interface{}{
		"code":  b.ErrCode(),
		"msg":   b.ErrMsg(),
		"extra": b.ExtraInfo(),
	})
}

func (g *Gin) WriteSysErr(c *gin.Context, extraInfo string) {
	c.JSON(500, map[string]interface{}{
		"code":  500,
		"msg":   "系统错误，请联系管理员",
		"extra": extraInfo,
	})
}

func (g *Gin) WriteCliErr(c *gin.Context, extraInfo string) {
	c.JSON(400, map[string]interface{}{
		"code":  400,
		"msg":   "无操作权限",
		"extra": extraInfo,
	})
}

func (g *Gin) Write(c *gin.Context, httpCode int, data any) {
	c.JSON(httpCode, data)

}

func (g *Gin) WriteErr(c *gin.Context, httpStatus, code int, msg ...string) {
	l := len(msg)
	m := map[string]interface{}{
		"code": code,
		"msg":  "",
	}
	if l >= 2 {
		m["msg"] = msg[0]
		m["cause"] = msg[1]
	} else if l == 1 {
		m["msg"] = msg[0]
	}

	c.JSON(httpStatus, m)
}

// CatchMiddleware a middleware that catches panic status and writes response.
func (g *Gin) CatchMiddleware() gin.HandlerFunc {
	return g.Catch(g.WriteErr)
}

func (g *Gin) Catch(writeErr GinWriteErr) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			r := recover()
			ex, err := catch(r)
			if err != nil {
				trace := make([]byte, 1<<16)
				n := runtime.Stack(trace, false)
				s := fmt.Sprintf("panic: '%v', \nMethod: %s, TCEURL Path: %s, \nStack Trace:\n %s", r, c.Request.Method, c.Request.URL.Path,
					string(trace[:int(math.Min(float64(n), float64(7000)))]))
				// 输出详细的桟信息
				logrus.Errorln(s)
				writeErr(c, 500, 500, "内部系统错误，请联系管理员", "")
			}
			if ex != nil {
				httpCode := ex.code
				// these http code are not defined
				// https://github.com/golang/go/blob/78e99761fc4bf1f5370f912b8a4594789c2f09f8/src/net/http/server.go#L1098 (function checkWriteHeaderCode)
				if httpCode < 200 || httpCode >= 600 {
					httpCode = 500
				}
				if ex.originError != nil {
					writeErr(c, httpCode, ex.code, ex.msg, ex.originError.Error())
				} else {
					writeErr(c, httpCode, ex.code, ex.msg, "")
				}
			}
		}()
		c.Next()
	}
}
