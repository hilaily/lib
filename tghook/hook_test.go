package tghook

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	token  = "929779696:AAHnxgRxEgLHuemzLd78xBioNN5taZk3zck"
	chatID = 463930964
)

func TestHook(t *testing.T) {
	namespace := "test project"
	logrus.AddHook(NewHook(token, chatID, namespace, nil, nil))
	logrus.Error("test error")

	logrus.AddHook(NewHook(token, chatID, namespace, nil, map[string]interface{}{
		"env": "dev",
	}))
	logrus.Error("test error")
	time.Sleep(time.Second)
}
