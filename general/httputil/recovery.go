package httputil

import (
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery returns a Gin middleware that recovers from panics and logs the panic details.
func Recovery(lgr *zap.Logger) gin.HandlerFunc {
	return ginzap.RecoveryWithZap(lgr, true)
}
