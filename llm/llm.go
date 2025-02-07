package llm

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/flosch/pongo2/v6"
	"github.com/openai/openai-go"
	openai2 "github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

// func Init() error {
// 	for _, v := range config.Get().LLM {
// 		client := openai.NewClient(
// 			option.WithAPIKey(v.ApiKey), // defaults to os.LookupEnv("OPENAI_API_KEY")
// 			option.WithBaseURL(v.BaseUrl),
// 		)
// 		clients[v.Name] = &Client{client: client}
// 	}
// 	return nil
// }

type ClientV1 struct {
	client *openai.Client
}

func (c *ClientV1) ChatOnce(promptTPL string, promptParams map[string]string, messages ...openai.ChatCompletionMessageParamUnion) (string, error) {
	pongoContext := pongo2.Context{}
	prompt := promptTPL
	if len(promptParams) > 0 {
		for k, v := range promptParams {
			if k != "" {
				pongoContext[k] = v
			}
		}

		tpl, err := pongo2.FromString(promptTPL)
		if err != nil {
			return "", fmt.Errorf("failed to parse prompt template, template: %s, error: %w", promptTPL, err)
		}
		prompt, err = tpl.Execute(pongoContext)
		if err != nil {
			return "", fmt.Errorf("failed to execute prompt template, template: %s, error: %w", promptTPL, err)
		}
	}

	msg := append([]openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(prompt),
	}, messages...)

	chatCompletion, err := c.client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F(msg),
		Model:    openai.F(openai.ChatModelGPT4o),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion, chatCompletion: %+v, error: %w", chatCompletion, err)
	}
	return chatCompletion.Choices[0].Message.Content, nil
}

func NewClient(opts ...ClientOption) *Client {
	conf := defaultOption()
	for _, opt := range opts {
		opt(conf)
	}
	if conf.apiKey == "" || conf.baseURL == "" {
		panic("apiKey or baseURL is not set")
	}
	c := openai2.DefaultConfig(conf.apiKey)
	c.BaseURL = conf.baseURL
	client := openai2.NewClientWithConfig(c)
	return &Client{client: client, opts: conf}
}

type Client struct {
	client *openai2.Client
	opts   *Option
}

func (c *Client) UpdateOption(opts ...ClientOption) *Client {
	o := c.opts
	for _, opt := range opts {
		opt(o)
	}
	return &Client{client: c.client, opts: o}
}

func (c *Client) ChatTextOnce(ctx context.Context, msg string) (<-chan string, error) {
	msgs := []openai2.ChatCompletionMessage{
		{
			Role:    openai2.ChatMessageRoleUser,
			Content: msg,
		},
	}
	rec, err := c.ChatBase(ctx, msgs)
	if err != nil {
		return nil, err
	}
	receiver := make(chan string, 10)
	go func() {
		for choice := range rec {
			receiver <- choice.Delta.Content
		}
		close(receiver)
	}()
	return receiver, nil
}

func (c *Client) ChatImageOnce(ctx context.Context, msg, imgURL string) (<-chan string, error) {
	content := []openai2.ChatMessagePart{
		{
			Type:     openai2.ChatMessagePartTypeImageURL,
			ImageURL: &openai2.ChatMessageImageURL{URL: imgURL},
		},
	}

	if msg != "" {
		content = append(content, openai2.ChatMessagePart{
			Type: openai2.ChatMessagePartTypeText,
			Text: msg,
		})
	}

	msgs := []openai2.ChatCompletionMessage{
		{
			Role:         openai2.ChatMessageRoleUser,
			MultiContent: content,
		},
	}
	rec, err := c.ChatBase(ctx, msgs)
	if err != nil {
		return nil, err
	}
	receiver := make(chan string, 10)
	go func() {
		for choice := range rec {
			receiver <- choice.Delta.Content
		}
		close(receiver)
	}()
	return receiver, nil
}

func (c *Client) ChatBase(ctx context.Context, msgs []openai2.ChatCompletionMessage) (<-chan openai2.ChatCompletionStreamChoice, error) {
	logrus.Debugf("base url: %+v, model: %+v", c.opts.baseURL, c.opts.model)
	newMsgs := []openai2.ChatCompletionMessage{
		{
			Role:    openai2.ChatMessageRoleSystem,
			Content: c.opts.prompt,
		},
	}
	newMsgs = append(newMsgs, msgs...)

	req := openai2.ChatCompletionRequest{
		Model:     c.opts.model,
		MaxTokens: c.opts.maxTokens,
		Messages:  newMsgs,
		Stream:    true,
	}
	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ChatCompletionStream error: %v\n", err)
	}

	receiver := make(chan openai2.ChatCompletionStreamChoice, 10)
	go func() {
		defer stream.Close()
		for {
			response, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				receiver <- openai2.ChatCompletionStreamChoice{
					FinishReason: openai2.FinishReason(fmt.Sprintf("request to llm error: %v\n", err)),
				}
				break
			}
			// logrus.Debugf("response: %+v", response.Choices[0])
			receiver <- response.Choices[0]
		}
		close(receiver)
	}()
	return receiver, nil
}
