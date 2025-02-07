package llm

import (
	"context"
	"os"
	"testing"

	"github.com/hilaily/lib/env"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestChatText(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	err := env.LoadEnv(".env.test")
	if err != nil {
		t.Fatal(err)
	}

	client := NewClient()
	client.UpdateOption(
		WithModel(os.Getenv("LLM_MODEL")),
		WithPrompt("You are a helpful assistant."),
	)
	rec, err := client.ChatTextOnce(context.TODO(), "你是谁")
	if err != nil {
		assert.NoError(t, err)
	}
	for msg := range rec {
		t.Log(msg)
	}
}
