package httputil

import (
	"bytes"
	"io"

	"github.com/gin-gonic/gin"
)

// RestoreBodyMiddleware reads the request body, stores it in context, and restores it for future handlers.
func RestoreBodyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.Next()
			}

			// Store the body in context for future use
			c.Set("request_body", bodyBytes)

			// Restore the body so other handlers can read it
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		c.Next()
	}
}
