package infrastructures

import (
	"context"
	"encoding/base64"
	"log"
	"time"

	"github.com/ollama/ollama/api"
	"github.com/openai/openai-go"
	"github.com/samber/lo"
)

// Chat sends a chat request to OpenAI.
func (c *OpenAIClient) Chat(ctx context.Context, req api.ChatRequest) (api.ChatResponse, error) {
	tmp, err := c.client.Chat.Completions.New(ctx, toChatCompletionRequest(req))
	if err != nil {
		return api.ChatResponse{}, err
	}
	log.Println(tmp.Choices[0].Message.Content)
	return api.ChatResponse{
		Model:     req.Model,
		CreatedAt: time.Now().UTC(),
		Message: api.Message{
			Content: tmp.Choices[0].Message.Content,
			Role:    string(openai.MessageRoleAssistant),
		},
		DoneReason: "unload",
		Done:       true,
	}, nil
}

func toChatCompletionRequest(r api.ChatRequest) openai.ChatCompletionNewParams {
	messages := lo.Map(r.Messages, func(msg api.Message, _ int) openai.ChatCompletionMessageParamUnion {
		if openai.MessageRole(msg.Role) == openai.MessageRoleUser {
			parts := lo.Map(msg.Images, func(v api.ImageData, _ int) openai.ChatCompletionContentPartUnionParam {
				b := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(v)
				return openai.ChatCompletionContentPartImageParam{
					Type: openai.F(openai.ChatCompletionContentPartImageTypeImageURL),
					ImageURL: openai.F(openai.ChatCompletionContentPartImageImageURLParam{
						URL: openai.F(b),
					}),
				}
			})
			if len(parts) == 0 {
				parts = append(parts, openai.TextPart(msg.Content))
			}
			return openai.UserMessageParts(parts...)
		}
		return openai.AssistantMessage(msg.Content)
	})
	return openai.ChatCompletionNewParams{
		Messages: openai.F(messages),
		Model:    openai.F(r.Model),
	}
}
