package httputil

import (
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger returns a Gin middleware that logs requests and responses.
func Logger(lgr *zap.Logger, ops ...func(cfg *ginzap.Config)) gin.HandlerFunc {
	cfg := &ginzap.Config{
		TimeFormat: time.RFC3339,
		UTC:        true,
		Context: func(gcx *gin.Context) []zapcore.Field {
			unm := NewUser(gcx).GetUsername()

			if len(unm) == 0 {
				unm = "anonymous"
			}

			return []zapcore.Field{
				zap.Any("username", unm),
			}
		},
	}

	for _, opt := range ops {
		opt(cfg)
	}

	return ginzap.GinzapWithConfig(lgr, cfg)
}
