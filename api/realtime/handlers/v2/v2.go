// Package v2 holds the routing and handlers (in subsequent packages)
// for the v2 API endpoints.
package v2

import (
	"context"
	"wikimedia-enterprise/api/realtime/handlers/v2/articles"
	"wikimedia-enterprise/api/realtime/packages/mware"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

// NewGroup creates new router group for the API.
func NewGroup(ctx context.Context, cnt *dig.Container, router *gin.Engine) (*gin.RouterGroup, error) {
	v2 := router.Group("/v2")
	v2.Use(mware.Headers())

	err := cnt.Invoke(func(p articles.Parameters) {
		v2.GET("/articles", articles.NewHandler(ctx, &p))
		v2.POST("/articles", articles.NewHandler(ctx, &p))
	})

	return v2, err
}
