package notifier

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
)

// Alertover contain a source
// Alertover website: https://www.alertover.com/
type Alertover struct {
	source string
}

// NewAlertover create a Alertover instance
func NewAlertover(source string) *Alertover {
	alert := &Alertover{}
	alert.source = source
	return alert
}

// SendAlert to send notification
// param are receiver alertover id, and message title and content
func (self *Alertover) SendAlert(receiver, title, content string) string {
	data := url.Values{
		"source":   {self.source},
		"receiver": {receiver},
		"title":    {title},
		"content":  {content},
	}
	resp, err := http.PostForm("https://api.alertover.com/v1/alert", data)
	if err != nil {
		log.Error("alertover send massage error: ", err)
	}
	defer resp.Body.Close()
	log.Info("alertover send message success")
	return fmt.Sprintf("%s", resp)
}
