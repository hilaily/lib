# handle

提供方便的解析 HTTP 请求，返回 HTTP 响应的方法。

#Gin 使用

## 注册中间件

```go
g := &handler.Gin{}
r.Group("/",g.CatchMwGin)
```

## 使用

```go
type debugHandler struct{
  handler.Gin // 内嵌 gin 类型的 handler 结构体
}

func (d *debugHandler) GetUser(c *gin.Context) {
	args := &types.SearchOpt{}
	d.Read(c, args) // 读取请求
	res, err := d.store.GetUsers(args)
	d.Check(err, "") // 判断 error
  r, err := d.TestUser()
  if err != nil{
    d.WriteErr(c, 503, 1001, "验证用户失败")
  }
	d.WriteOK(c, res) // 返回响应
}
```

## Read

```
Read(c *gin.Context, args interface{})
```

## Return

```
Return(c *gin.Context, data interface{}, err error, msg string)
```

## WriteOK

```
WriteOK(c *gin.Context, data interface{})
```

## WriteErr

```
WriteErr(c *gin.Context, httpStatus, code int, msg ...string)
```