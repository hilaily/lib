package logrustool

import (
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// SetRotateSimple
// it is for you just set a filepath to save log, other options are default
// @param _filepath can be relative path
func SetRotateSimple(_filepath string) {
	SetRotateSimple2(logrus.StandardLogger(), _filepath)
}

// SetRotateSimple2
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

// SetRotate
// @param l is a lumberjack.Logger
func SetRotate(l *lumberjack.Logger) {
	SetRotate2(logrus.StandardLogger(), l)
}

// SetRotate2
// @param ll is a logrus.Logger, you can use any logrus.Logger
// @param l is a lumberjack.Logger
func SetRotate2(ll *logrus.Logger, l *lumberjack.Logger) {
	// Fork writing into two outputs
	multiWriter := io.MultiWriter(os.Stderr, l)
	ll.SetOutput(multiWriter)
}
