package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/jonboulle/clockwork"
)

// WithClock is a middleware that adds a clock to the context
func WithClock(c *gin.Context) {
	ctx := clockwork.AddToContext(c, clockwork.NewRealClock())
	c.Request = c.Request.WithContext(ctx)
	c.Next()
}
