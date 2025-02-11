package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/junpayment/ollamaproxy/port"
	"github.com/ollama/ollama/api"
	"github.com/openai/openai-go"
	"github.com/samber/lo"
)

// Handlers is a struct that holds the handlers for the server
type Handlers struct {
	client port.Client
}

// NewHandlers returns a new Handlers instance
func NewHandlers(client port.Client) *Handlers {
	return &Handlers{
		client: client,
	}
}

// Tags returns a list of models
func (r *Handlers) Tags(c *gin.Context) {
	tmp, err := r.client.ListModels(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var models []api.ListModelResponse
	models = lo.Map(tmp, func(item openai.Model, _ int) api.ListModelResponse {
		return api.ListModelResponse{
			Name:       item.ID,
			Model:      item.ID,
			ModifiedAt: time.Unix(item.Created, 0),
			Size:       0,
			Digest:     "dummy",
			Details:    api.ModelDetails{},
		}
	})
	c.JSON(http.StatusOK, api.ListResponse{Models: models})
}

func (r *Handlers) Version(c *gin.Context) {}

func (r *Handlers) Chat(c *gin.Context) {
	var req api.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := r.client.Chat(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
