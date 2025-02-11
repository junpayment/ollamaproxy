package config

import (
	"os"

	"github.com/junpayment/ollamaproxy/models"
)

// NewConfig is a factory function that returns a new Config instance.
func NewConfig() models.Config {
	return models.Config{
		APIKey: os.Getenv("OPENAI_API_KEY"),
	}
}
