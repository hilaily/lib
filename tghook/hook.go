package tghook

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	tg "gopkg.in/telegram-bot-api.v4"
)

// Hook structure
type Hook struct {
	Client    *tg.BotAPI
	ClientID  int64
	namespace string
	levels    []logrus.Level
	ext       map[string]interface{}
	extString string
}

// NewHook init
func NewHook(apiKey string, clientID int64, namespace string, levels []logrus.Level, ext map[string]interface{}) *Hook {
	client, _ := tg.NewBotAPI(
		apiKey,
	)
	if len(levels) == 0 {
		levels = []logrus.Level{
			logrus.ErrorLevel,
			logrus.FatalLevel,
			logrus.PanicLevel,
		}
	}

	hook := &Hook{
		client,
		clientID,
		namespace,
		levels,
		ext,
		"",
	}

	if len(ext) > 0 {
		en, _ := json.MarshalIndent(ext, "", "\t")
		hook.extString = string(en)
	}
	return hook
}

// Fire routine
func (hook *Hook) Fire(logEntry *logrus.Entry) error {
	var notifyErr string

	if err, ok := logEntry.Data["error"].(error); ok {
		notifyErr = err.Error()
	} else {
		notifyErr = logEntry.Message
	}

	msg := tg.MessageConfig{}
	msg.ChatID = hook.ClientID

	if hook.extString == "" {
		msg.Text = fmt.Sprintf("namespace: %s\n\nlevel: %s\nmsg: %s",
			hook.namespace,
			strings.ToUpper(logEntry.Level.String()),
			notifyErr,
		)
	} else {
		msg.Text = fmt.Sprintf("namespace: %s\n\nlevel: %s\nmsg: %s\n\next\n%s",
			hook.namespace,
			strings.ToUpper(logEntry.Level.String()),
			notifyErr,
			hook.extString,
		)
	}
	logEntry.Logger.Debug(msg)
	result, err := hook.Client.Send(msg)
	logEntry.Logger.Debug(result)

	if err != nil {
		logEntry.Logger.WithFields(logrus.Fields{
			"source": "telegram",
			"error":  err,
		}).Warn("Failed to send error to Telegram")
	}

	return nil
}

// Levels setting
func (hook *Hook) Levels() []logrus.Level {
	return hook.levels
}
