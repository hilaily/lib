package logrustool

import (
	"os"

	"github.com/sirupsen/logrus"
)

// SetLevel
// @param l is a logrus.Level
func SetLevel(l ...logrus.Level) {
	SetLevel2(logrus.StandardLogger(), l...)
}

// SetLevel2
// @param ll is a logrus.Logger, you can use yourself logrus.Logger
// @param l is a logrus.Level
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
