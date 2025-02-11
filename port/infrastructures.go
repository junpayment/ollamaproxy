package port

import (
	"context"

	"github.com/ollama/ollama/api"
	"github.com/openai/openai-go"
)

// Client is an interface that defines the methods that the OpenAI client should implement
type Client interface {
	ListModels(ctx context.Context) ([]openai.Model, error)
	Chat(ctx context.Context, req api.ChatRequest) (api.ChatResponse, error)
}
