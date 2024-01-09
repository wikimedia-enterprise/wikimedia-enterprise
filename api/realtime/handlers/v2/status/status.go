// Package status creates HTTP handler for example endpoint. And serves and status health check.
package status

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewHandler creates new example HTTP handler.
func NewHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Status(http.StatusOK)
	}
}
