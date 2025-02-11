package infrastructures

import (
	"context"

	"github.com/junpayment/ollamaproxy/models"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// OpenAIClient is a client for OpenAI.
type OpenAIClient struct {
	client *openai.Client
}

// NewOpenAIClient creates a new OpenAI client.
func NewOpenAIClient(cfg models.Config) *OpenAIClient {
	client := openai.NewClient(
		option.WithAPIKey(cfg.APIKey),
	)
	return &OpenAIClient{
		client: client,
	}
}

// ListModels lists the available models.
func (c *OpenAIClient) ListModels(ctx context.Context) ([]openai.Model, error) {
	tmp, err := c.client.Models.List(ctx)
	if err != nil {
		return nil, err
	}
	var res []openai.Model
	for _, model := range tmp.Data {
		res = append(res, model)
	}
	return res, nil
}
