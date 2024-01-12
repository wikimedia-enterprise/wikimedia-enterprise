// Package v1 creates a set of legacy handlers that are identical to what we had in v1.
package v1

import (
	"fmt"
	"wikimedia-enterprise/api/main/config/env"
	"wikimedia-enterprise/api/main/packages/legacy"
	"wikimedia-enterprise/api/main/packages/proxy"
	"wikimedia-enterprise/general/httputil"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

// Params list of params for dependency injection.
type Params struct {
	dig.In
	Env    *env.Environment
	Proxy  proxy.Params
	Legacy legacy.Params
	Capper httputil.CapByRedis
}

// NewGroup creates new router group for the legacy v1 API.
func NewGroup(con *dig.Container, rtr *gin.Engine) (*gin.RouterGroup, error) {
	v1 := rtr.Group("/v1")

	return v1, con.Invoke(func(pms Params) {
		for _, ent := range []string{
			"projects",
			"namespaces",
		} {
			v1.GET(fmt.Sprintf("/%s", ent), proxy.NewGetEntities(&pms.Proxy, proxy.NewEntitiesGetter(ent)))
		}

		cmw := httputil.Cap(&pms.Capper, pms.Env.CapConfig)
		v1.GET("/pages/meta/:project/*name", cmw, legacy.NewGetPageHandler(&pms.Legacy))

		v1.GET("/diffs/meta/:date/:namespace", legacy.NewListDiffsHandler(&pms.Legacy))
		v1.GET("/diffs/meta/:date/:namespace/:project", legacy.NewGetDiffHandler(&pms.Legacy))
		v1.HEAD("/diffs/download/:date/:namespace/:project", legacy.NewHeadDiffHandler(&pms.Legacy))
		v1.GET("/diffs/download/:date/:namespace/:project", legacy.NewDownloadDiffHandler(&pms.Legacy))

		v1.GET("/exports/meta/:namespace", legacy.NewListExportsHandler(&pms.Legacy))
		v1.GET("/exports/meta/:namespace/:project", legacy.NewGetExportHandler(&pms.Legacy))
		v1.HEAD("/exports/download/:namespace/:project", legacy.NewHeadExportHandler(&pms.Legacy))
		v1.GET("/exports/download/:namespace/:project", legacy.NewDownloadExportHandler(&pms.Legacy))
	})
}
