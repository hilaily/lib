package logrustool

import (
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
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

// Param: _filepath can be relative path
func SetRotateSimple(_filepath string) {
	SetRotateSimple2(logrus.StandardLogger(), _filepath)
}

func SetRotateSimple2(ll *logrus.Logger, _filepath string) {
	lumberjackLogger := &lumberjack.Logger{
		// Log file abbsolute path, os agnostic
		Filename:   filepath.ToSlash(_filepath),
		MaxSize:    10, // MB
		MaxBackups: 5,
		MaxAge:     30, // days
		//Compress:   true, // disabled by default
	}
	SetRotate2(ll, lumberjackLogger)
}

func SetRotate(l *lumberjack.Logger) {
	SetRotate2(logrus.StandardLogger(), l)
}

func SetRotate2(ll *logrus.Logger, l *lumberjack.Logger) {
	// Fork writing into two outputs
	multiWriter := io.MultiWriter(os.Stderr, l)
	ll.SetOutput(multiWriter)
}

func SetLevel(l ...logrus.Level) {
	SetLevel2(logrus.StandardLogger(), l...)
}

func SetLevel2(ll *logrus.Logger, l ...logrus.Level) {
	l1 := os.Getenv("LOGRUS_LEVEL")
	v, err := logrus.ParseLevel(l1)
	if err == nil {
		ll.SetLevel(v)
		return
	}
	if len(l) > 0 {
		ll.SetLevel(l[0])
	}
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
