// Package v2 provides to all API handlers the way
// to use DI container and all dependencies
package v2

import (
	"fmt"
	"wikimedia-enterprise/api/main/config/env"
	"wikimedia-enterprise/api/main/packages/proxy"
	"wikimedia-enterprise/general/httputil"
	"wikimedia-enterprise/general/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

// Params list of params for dependency injection.
type Params struct {
	dig.In
	Env    *env.Environment
	Proxy  proxy.Params
	Capper httputil.CapByRedis
	SOCK   proxy.SOCKModifier
}

// NewGroup creates new router group for the API.
func NewGroup(con *dig.Container, rtr *gin.Engine) (*gin.RouterGroup, error) {
	v2 := rtr.Group("/v2")

	for _, err := range []error{
		con.Invoke(func(pms Params) {
			for _, ent := range []string{
				"codes",
				"languages",
				"projects",
				"namespaces",
			} {

				lpt := fmt.Sprintf("/%s", ent)
				v2.GET(lpt, proxy.NewGetEntities(&pms.Proxy, proxy.NewEntitiesGetter(ent)))
				v2.POST(lpt, proxy.NewGetEntities(&pms.Proxy, proxy.NewEntitiesGetter(ent)))

				ipt := fmt.Sprintf("/%s/:identifier", ent)
				v2.GET(ipt, proxy.NewGetEntity(&pms.Proxy, proxy.NewEntityGetter(ent)))
				v2.POST(ipt, proxy.NewGetEntity(&pms.Proxy, proxy.NewEntityGetter(ent)))

			}

			// We will be caping group_1 usage for ondemand (articles, structured-contents) and snapshots download
			cmw := httputil.Cap(&pms.Capper, *pms.Env.CapConfig)

			for _, ent := range []string{
				"snapshots",
			} {
				lpt := fmt.Sprintf("/%s", ent)
				v2.GET(lpt, proxy.NewGetEntities(&pms.Proxy, proxy.NewByGroupEntitiesGetter(ent, pms.Env.FreeTierGroup)))
				v2.POST(lpt, proxy.NewGetEntities(&pms.Proxy, proxy.NewByGroupEntitiesGetter(ent, pms.Env.FreeTierGroup)))

				ipt := fmt.Sprintf("/%s/:identifier", ent)
				v2.GET(ipt, proxy.NewGetEntity(&pms.Proxy, proxy.NewByGroupEntityGetter(ent, pms.Env.FreeTierGroup)))
				v2.POST(ipt, proxy.NewGetEntity(&pms.Proxy, proxy.NewByGroupEntityGetter(ent, pms.Env.FreeTierGroup)))

				dpt := fmt.Sprintf("/%s/:identifier/download", ent)
				v2.GET(dpt, cmw, proxy.NewGetDownload(&pms.Proxy, proxy.NewByGroupEntityDownloader(ent, pms.Env.FreeTierGroup)))
				v2.HEAD(dpt, cmw, proxy.NewHeadDownload(&pms.Proxy, proxy.NewByGroupEntityDownloader(ent, pms.Env.FreeTierGroup)))

				cpt := fmt.Sprintf("/%s/:identifier/chunks", ent)
				bse := "chunks"
				v2.GET(cpt, proxy.NewGetEntities(&pms.Proxy, proxy.NewEntitiesGetter(bse)))
				v2.POST(cpt, proxy.NewGetEntities(&pms.Proxy, proxy.NewEntitiesGetter(bse)))

				cit := fmt.Sprintf("/%s/:identifier/chunks/:chunkIdentifier", ent)
				v2.GET(cit, proxy.NewGetEntity(&pms.Proxy, proxy.NewEntityGetter(bse)))
				v2.POST(cit, proxy.NewGetEntity(&pms.Proxy, proxy.NewEntityGetter(bse)))

				cdt := fmt.Sprintf("/%s/:identifier/chunks/:chunkIdentifier/download", ent)
				v2.HEAD(cdt, proxy.NewHeadDownload(&pms.Proxy, proxy.NewEntityDownloader(bse)))
				v2.GET(cdt, proxy.NewGetDownload(&pms.Proxy, proxy.NewEntityDownloader(bse)))
			}

			for _, ent := range []string{
				"batches",
			} {
				lpt := fmt.Sprintf("/%s/:date", ent)
				v2.GET(lpt, proxy.NewGetEntities(&pms.Proxy, proxy.NewDateEntitiesGetter(ent)))
				v2.POST(lpt, proxy.NewGetEntities(&pms.Proxy, proxy.NewDateEntitiesGetter(ent)))

				ipt := fmt.Sprintf("/%s/:date/:identifier", ent)
				v2.GET(ipt, proxy.NewGetEntity(&pms.Proxy, proxy.NewDateEntityGetter(ent)))
				v2.POST(ipt, proxy.NewGetEntity(&pms.Proxy, proxy.NewDateEntityGetter(ent)))

				dpt := fmt.Sprintf("/%s/:date/:identifier/download", ent)
				v2.GET(dpt, proxy.NewGetDownload(&pms.Proxy, proxy.NewDateEntityDownloader(ent)))
				v2.HEAD(dpt, proxy.NewHeadDownload(&pms.Proxy, proxy.NewDateEntityDownloader(ent)))
			}

			for _, ent := range []string{
				"articles",
			} {
				npt := fmt.Sprintf("/%s/*name", ent)

				if len(pms.Env.ArticleKeyTypeSuffix) > 0 {
					ent = fmt.Sprintf("%s_%s", ent, pms.Env.ArticleKeyTypeSuffix)
				}

				v2.GET(npt, cmw, proxy.NewGetLargeEntities(&pms.Proxy, ent, proxy.DefaultModifiers...))
				v2.POST(npt, cmw, proxy.NewGetLargeEntities(&pms.Proxy, ent, proxy.DefaultModifiers...))

				pth := "/structured-contents/*name"
				v2.GET(pth, cmw, proxy.NewGetLargeEntities(&pms.Proxy, ent, new(proxy.FilterModifier), &pms.SOCK))
				v2.POST(pth, cmw, proxy.NewGetLargeEntities(&pms.Proxy, ent, new(proxy.FilterModifier), &pms.SOCK))
			}

			// ************* Structured Contents Start ****************

			for _, ent := range []string{
				"structured-snapshots",
			} {
				lpt := "/snapshots/structured-contents"
				v2.GET(lpt, proxy.NewGetEntities(&pms.Proxy, proxy.NewEntitiesGetter(ent)))
				v2.POST(lpt, proxy.NewGetEntities(&pms.Proxy, proxy.NewEntitiesGetter(ent)))

				ipt := "/snapshots/structured-contents/:identifier"
				v2.GET(ipt, proxy.NewGetEntity(&pms.Proxy, proxy.NewEntityGetter(ent)))
				v2.POST(ipt, proxy.NewGetEntity(&pms.Proxy, proxy.NewEntityGetter(ent)))

				dpt := "/snapshots/structured-contents/:identifier/download"
				v2.GET(dpt, proxy.NewGetDownload(&pms.Proxy, proxy.NewEntityDownloader(ent)))
				v2.HEAD(dpt, proxy.NewHeadDownload(&pms.Proxy, proxy.NewEntityDownloader(ent)))
			}

			// ************* Structured Contents End ****************

			for _, ent := range []string{
				"files",
			} {
				fpt := fmt.Sprintf("/%s/:filename", ent)
				fgt := proxy.NewFileGetter()
				v2.GET(fpt, proxy.NewGetEntity(&pms.Proxy, fgt))
				v2.POST(fpt, proxy.NewGetEntity(&pms.Proxy, fgt))

				dpt := fmt.Sprintf("/%s/:filename/download", ent)
				v2.GET(dpt, proxy.NewGetDownload(&pms.Proxy, proxy.NewFileDownloader()))
			}

		}),
	} {
		if err != nil {
			log.Error(err, log.Tip("problem creating router group for the API"))
			return nil, err
		}
	}

	return v2, nil
}
