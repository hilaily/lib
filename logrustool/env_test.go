package logrustool

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSetLevel(t *testing.T) {
	assert.Equal(t, logrus.InfoLevel, logrus.GetLevel())
	SetLevel(logrus.ErrorLevel)
	assert.Equal(t, logrus.ErrorLevel, logrus.GetLevel())

	ll := logrus.New()
	assert.Equal(t, logrus.InfoLevel, ll.GetLevel())
	SetLevel2(ll, logrus.DebugLevel)
	assert.Equal(t, logrus.DebugLevel, ll.GetLevel())
	assert.Equal(t, logrus.ErrorLevel, logrus.GetLevel())
}

func TestFromat(t *testing.T) {
	FormatOnlyMsg()
	logrus.Infoln("this is a test")
	time.Sleep(200 * time.Millisecond)
}
