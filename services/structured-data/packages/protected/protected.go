package protected

import (
	"context"
	"errors"
	"strings"
	"time"
	"wikimedia-enterprise/services/structured-data/submodules/log"
	"wikimedia-enterprise/services/structured-data/submodules/wmf"

	"golang.org/x/exp/slices"
)

var (
	retries         = 3
	refreshInterval = time.Hour
)

type project struct {
	dtb string
	tls []string
}

// Protected helper struct to check whether a page is protected.
type Protected struct {
	projects  []project
	cache     map[string][]string
	refreshed time.Time
	wmf       wmf.API
	retries   int
}

// New create protected helper instance.
func New(wmf wmf.API) *Protected {

	return &Protected{
		projects: []project{
			{dtb: "mywiki", tls: []string{"mypage"}},
		},

		wmf:     wmf,
		retries: retries,
	}
}

// IsProtectedPage checks if a page is protected.
func (ptd *Protected) IsProtectedPage(prj string, ttl string) bool {
	ptd.refresh()

	val, ok := ptd.cache[prj]

	if !ok {
		return false
	}

	return slices.Contains(val, ttl)
}

func (ptd *Protected) refresh() {

	if time.Since(ptd.refreshed) >= refreshInterval {

		for n := 0; n < retries; n++ {
			err := ptd.load()

			if err == nil {
				ptd.refreshed = time.Now()
				break
			}
		}
	}
}

func (ptd *Protected) load() error {

	m := make(map[string][]string)
	ctx := context.Background()

	for _, prj := range ptd.projects {
		lst := make([]string, 0)
		log.Info("handling project", log.Any("project", prj.dtb))

		for _, ttl := range prj.tls {
			pge, err := ptd.wmf.GetPage(ctx, prj.dtb, ttl)
			log.Info("fecthing holder page", log.Any("page", ttl))
			if err != nil {
				return errors.New(ttl)
			}

			if len(pge.Revisions) > 0 {
				rev := pge.Revisions[0]

				if rev != nil && rev.Slots != nil && rev.Slots.Main != nil {
					txt := rev.Slots.Main.Content

					pgs := strings.Split(txt, "\n")

					lst = append(lst, pgs...)

				}
			}
			if len(lst) > 0 {
				m[prj.dtb] = lst
			}
		}
	}

	if len(m) > 0 {
		ptd.cache = m
	} else {
		log.Error("Failed to get configuration", log.Any("config", ptd.projects))
		return errors.New("failed to get pages")
	}

	return nil
}
