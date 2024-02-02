package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	card "github.com/go-lark/larkcard"
	"github.com/hilaily/kit/httpx"
)

func newLarkWebhook(u string) *LarkWebhook {
	return &LarkWebhook{
		u: u,
	}
}

type LarkWebhook struct {
	u string
}

func (l *LarkWebhook) MDSend(title string, contents []string) error {
	msg := MDMsg(title, contents)
	en, _ := json.Marshal(
		map[string]any{
			"msg_type": "interactive",
			"card":     json.RawMessage(msg),
		},
	)
	res, err := http.DefaultClient.Post(l.u, "application/json", bytes.NewBuffer(en))
	if err != nil {
		return fmt.Errorf("send msg fail, %w", err)
	}
	defer res.Body.Close()
	b, err := httpx.HandleResp(res, nil)
	if err != nil {
		return fmt.Errorf("body:%s, fail:%w", string(b), err)
	}
	return nil
}

// MDSend a msg with title and multiple line content.
func MDMsg(title string, contents []string) []byte {
	content := card.NewModContentMulti(contents, true, false)
	mods := make([]card.IModule, 0, 2)
	mods = append(mods, content)
	ca := card.New(nil, card.NewHeader(title).SetColor(card.HeaderYellow), mods...)
	return ca.Encode()
}
