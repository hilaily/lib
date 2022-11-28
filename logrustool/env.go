package logrustool

import (
	"os"

	"github.com/sirupsen/logrus"
)

func FormatOnlyMsg() {
	logrus.SetFormatter(&onlyMsgFmt{})
	//logrus.SetFormatter(&logrus.TextFormatter{})
}

type onlyMsgFmt struct{}

func (o *onlyMsgFmt) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(entry.Message), nil
}

func InitLevel(l logrus.Level) {
	l1 := os.Getenv("LOGRUS_LEVEL")
	v, err := logrus.ParseLevel(l1)
	if err == nil {
		logrus.SetLevel(v)
	}
	logrus.SetLevel(l)
}

// can provide a function to set logrus level
// can init logrus level by environmeknt
func GetSetLevel() func(logrus.Level) {
	l := os.Getenv("LOGRUS_LEVEL")
	v, err := logrus.ParseLevel(l)
	if err == nil {
		logrus.SetLevel(v)
	}
	return func(l logrus.Level) {
		logrus.SetLevel(l)
	}
}
