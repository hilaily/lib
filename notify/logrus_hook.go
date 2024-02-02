package notify

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

func NewHookByENV() (*ErrorHook, error) {
	projectName := os.Getenv("NOTI_NAME")
	webhook := os.Getenv("NOTI_LARK_WEBHOOK")
	return NewHook(projectName, webhook)
}

// NewHook
func NewHook(projectName string, webhookURL string) (*ErrorHook, error) {
	bot := newLarkWebhook(webhookURL)
	return &ErrorHook{
		name: projectName,
		sendFunc: func(title string, contents []string) error {
			return bot.MDSend(title, contents)
		},
	}, nil
}

type ErrorHook struct {
	name     string
	sendFunc func(string, []string) error
}

func (h *ErrorHook) Levels() []logrus.Level {
	// fire only on ErrorLevel (.Error(), .Errorf(), etc.)
	return []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel}
}

func (h *ErrorHook) Fire(e *logrus.Entry) error {
	// e.Data is a map with all fields attached to entry
	arr := []string{
		"**Level**: " + e.Level.String(),
		"**Msg**: " + e.Message,
	}
	if e.Caller != nil {
		arr = append(arr,
			"**File**: "+e.Caller.File,
			"**Line**: "+fmt.Sprintf("%d", e.Caller.Line),
			"**Func**: "+e.Caller.Func.Name(),
		)
	}
	arr = append(arr, "**Time**: "+e.Time.Format("2006-01-02 15:04:05"))

	err := h.sendFunc("["+h.name+"] Error Log", arr)
	return err
}
