package notifier

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/smtp"
	"strings"
)

// EmailNotifier has param SMTPHost, SMTP account
type EmailNotifier struct {
	sMTPHost string
	from     string
	password string
}

// NewEmailNotifier create a EmailNotifier instance
// host 邮箱 stmp 地址
// from 邮箱地址
// password 邮箱密码
func NewEmailNotifier(host, from, password string) (*EmailNotifier, error) {
	if host == "" {
		msg := "host is empty"
		log.Error("new emailNotifier error: ", msg)
		return nil, errors.New(msg)
	}
	if from == "" {
		msg := "from value is empty"
		log.Error("new emailNotifier error: ", msg)
		return nil, errors.New(msg)
	}
	email := &EmailNotifier{}
	email.sMTPHost = host
	email.from = from
	email.password = password
	return email, nil
}

// SendMail to send notification to others,
// Params are receiver emails, separate with ";"; email subject and content; content can be html text
func (m *EmailNotifier) SendMail(receiver string, title string, body string) error {
	if receiver == "" {
		msg := "receiver is empty"
		log.Error(msg)
		return errors.New(msg)
	}
	addrInfo := strings.Split(m.sMTPHost, ":")
	if len(addrInfo) != 2 {
		msg := "smtp_host wrong, eg: host_name:25"
		log.Error(msg)
		return errors.New(msg)
	}
	auth := smtp.PlainAuth("", m.from, m.password, addrInfo[0])

	_sendTo := strings.Split(receiver, ";")
	var sendTo []string
	for _, _to := range _sendTo {
		_to = strings.TrimSpace(_to)
		if _to != "" && strings.Contains(_to, "@") {
			sendTo = append(sendTo, _to)
		} else {
			msg := "receiver email address is wrong"
			log.Error("send error: ", msg)
			return errors.New(msg)
		}
	}

	if len(sendTo) < 1 {
		msg := "mail receiver is empty"
		log.Info(msg)
		return errors.New(msg)
	}

	msgBody := fmt.Sprintf("To: %s\r\nFrom: %s\r\nContent-Type: text/html;charset=utf-8\r\nSubject: %s\r\n\r\n%s", strings.Join(sendTo, ";"), m.from, title, body)
	err := smtp.SendMail(m.sMTPHost, auth, m.from, sendTo, []byte(msgBody))
	if err == nil {
		log.Info("send mail success")
		return nil
	} else {
		log.Error("send mail failed: ", err)
		return err
	}
}
