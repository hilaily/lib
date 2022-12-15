package logrustool

import (
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func FormatOnlyMsg() {
	logrus.SetFormatter(&onlyMsgFmt{})
	//logrus.SetFormatter(&logrus.TextFormatter{})
}

type onlyMsgFmt struct{}

func (o *onlyMsgFmt) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(entry.Message), nil
}

// Param: _filepath can be relative path
func SetRotateSimple(_filepath string) {
	lumberjackLogger := &lumberjack.Logger{
		// Log file abbsolute path, os agnostic
		Filename:   filepath.ToSlash(_filepath),
		MaxSize:    10, // MB
		MaxBackups: 5,
		MaxAge:     30, // days
		//Compress:   true, // disabled by default
	}
	SetRotate(lumberjackLogger)
}

func SetRotate(l *lumberjack.Logger) {
	// Fork writing into two outputs
	multiWriter := io.MultiWriter(os.Stderr, l)
	logrus.SetOutput(multiWriter)
}

func SetLevel(l logrus.Level) {
	l1 := os.Getenv("LOGRUS_LEVEL")
	v, err := logrus.ParseLevel(l1)
	if err == nil {
		logrus.SetLevel(v)
	}
	logrus.SetLevel(l)
}

// Deprecated use SetLevel instead
func InitLevel(l logrus.Level) {
	l1 := os.Getenv("LOGRUS_LEVEL")
	v, err := logrus.ParseLevel(l1)
	if err == nil {
		logrus.SetLevel(v)
	}
	logrus.SetLevel(l)
}

// Deprecated use SetLevel instead
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
