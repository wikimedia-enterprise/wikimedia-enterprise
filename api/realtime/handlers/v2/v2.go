// Package v2 holds the routing and handlers (in subsequent packages)
// for the v2 API endpoints.
package v2

import (
	"context"
	"wikimedia-enterprise/api/realtime/handlers/v2/articles"
	ksql "wikimedia-enterprise/api/realtime/handlers/v2/ksqldbarticles"
	"wikimedia-enterprise/api/realtime/packages/mware"
	"wikimedia-enterprise/api/realtime/submodules/schema"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

// NewGroup creates new router group for the API.
func NewGroup(ctx context.Context, cnt *dig.Container, router *gin.Engine) (*gin.RouterGroup, error) {
	v2 := router.Group("/v2")
	v2.Use(mware.Headers())

	err := cnt.Invoke(func(p articles.Parameters, pksql ksql.Parameters) {
		if !p.Env.UseKsqldb {
			v2.GET("/articles", articles.NewHandler(ctx, &p, p.Env.ArticlesTopic, schema.KeyTypeArticle))
			v2.POST("/articles", articles.NewHandler(ctx, &p, p.Env.ArticlesTopic, schema.KeyTypeArticle))
		} else {
			// To be deprecated
			v2.GET("/articles", ksql.NewHandler(ctx, &pksql))
			v2.POST("/articles", ksql.NewHandler(ctx, &pksql))
		}
	})

	return v2, err
}
