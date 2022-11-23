package notifier

import (
	"testing"
)

func TestNewEmailNotifier(t *testing.T) {
	e1, _ := NewEmailNotifier("", "", "")
	if e1 != nil {
		t.Error("e1 is not nil")
	}
	e2, _ := NewEmailNotifier("host", "", "")
	if e2 != nil {
		t.Error("e2 is not nil")
	}
	e3, _ := NewEmailNotifier("host", "host", "")
	if e3 == nil {
		t.Error("e3 should ok")
	}
}

func TestEmailNotifier_SendMail(t *testing.T) {
	notifier, _ := NewEmailNotifier("smtp.exmail.qq.com:25", "notifier@laily.me", "")

	res := notifier.SendMail("", "", "")
	if res != nil {
		t.Log(res)
	} else {
		t.Error("error1")
	}

	res = notifier.SendMail("aa;aa@a.com", "", "")
	if res != nil {
		t.Log(res)
	} else {
		t.Error("error2")
	}

	res = notifier.SendMail("notifier@laily.me", "", "")
	if res != nil {
		t.Log(res)
	} else {
		t.Error("error3")
	}
}
