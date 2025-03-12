package di

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/junpayment/ollamaproxy/handlers"
	"github.com/junpayment/ollamaproxy/port"
)

// API is the API server
type API struct {
	handlers port.Handlers
}

// NewAPI returns a new API instance
func newAPI(handlers port.Handlers) *API {
	return &API{
		handlers: handlers,
	}
}

// Run starts the API server
func (r *API) Run() error {
	router := gin.Default()
	router.Use(handlers.WithClock)

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Ollama is running")
	})
	router.HEAD("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Ollama is running")
	})

	api := router.Group("/api")
	api.GET("/tags", r.handlers.Tags)
	api.GET("/version", r.handlers.Version)
	api.POST("/chat", r.handlers.Chat)

	return router.Run(":11434")
}
