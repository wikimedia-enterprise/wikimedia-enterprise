// Package v1 holds a legacy implementation of the handlers to make sure we have
// backwards compatibility with the old APIs.
package v1

import (
	"context"
	"wikimedia-enterprise/api/realtime/handlers/v1/pages"
	"wikimedia-enterprise/api/realtime/packages/mware"
	"wikimedia-enterprise/general/schema"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

// NewGroup creates new router group for the API.
func NewGroup(ctx context.Context, cnt *dig.Container, router *gin.Engine) (*gin.RouterGroup, error) {
	v1 := router.Group("/v1")
	v1.Use(mware.Headers())

	err := cnt.Invoke(func(p pages.Parameters) {
		v1.GET("/page-update", pages.NewHandler(ctx, &p, schema.EventTypeUpdate))
		v1.GET("/page-delete", pages.NewHandler(ctx, &p, schema.EventTypeDelete))
		v1.GET("/page-visibility", pages.NewHandler(ctx, &p, schema.EventTypeVisibilityChange))
	})

	return v1, err
}
