package mware

import "github.com/gin-gonic/gin"

// Headers set streaming headers.
func Headers() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Next()
	}
}
