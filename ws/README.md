# WS

## example

```go
ws.GET("/ws/test", HandleWs)

func HandleWs(c *gin.Context) {
	wsHandler := NewWSHandler()
	wsHandler.RegisterHandler(func(conn *websocket.Conn, payload []byte) error {
		return nil
	})
	wsHandler.RegisterErrorHandler(func(conn *websocket.Conn, err error) {
		logrus.WithError(err).Error("WebSocket error")
	})
	wsHandler.HandleConnection(c)
}

```
