package port

import "github.com/gin-gonic/gin"

// Handlers is a struct that holds the handlers for the server
type Handlers interface {
	Tags(c *gin.Context)
	Version(c *gin.Context)
	Chat(c *gin.Context)
}
