package logrustool

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestFromat(t *testing.T) {
	FormatOnlyMsg()
	logrus.Infoln("this is a test")
	time.Sleep(200 * time.Millisecond)
}
