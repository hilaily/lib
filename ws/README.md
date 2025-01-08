# WS

## example

```go
ws.GET("/ws/test", HandleWs)

func HandleWs(c *gin.Context) {
	wsHandler := NewHandler(
		func(conn *websocket.Conn, payload []byte) error {
			logrus.WithField("msg", string(payload)).Info("Received message")
			return nil
		},
		func(conn *websocket.Conn, err error) {
			logrus.WithError(err).Error("WebSocket error")
		},
	)
	wsHandler.HandleConnection(c)
}

```
