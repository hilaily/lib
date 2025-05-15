package logrustool

import (
	"github.com/sirupsen/logrus"
)

var InitLevel = SetLevel

func FormatOnlyMsg() {
	logrus.SetFormatter(&onlyMsgFmt{})
	//logrus.SetFormatter(&logrus.TextFormatter{})
}

type onlyMsgFmt struct{}

func (o *onlyMsgFmt) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(entry.Message), nil
}
