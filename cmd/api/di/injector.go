//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/junpayment/ollamaproxy/config"
	"github.com/junpayment/ollamaproxy/handlers"
	"github.com/junpayment/ollamaproxy/infrastructures"
	"github.com/junpayment/ollamaproxy/port"
)

func InitAPI() *API {
	wire.Build(
		newAPI,
		wire.Bind(new(port.Handlers), new(*handlers.Handlers)),
		handlers.NewHandlers,
		wire.Bind(new(port.Client), new(*infrastructures.OpenAIClient)),
		infrastructures.NewOpenAIClient,
		config.NewConfig,
	)
	return nil
}
