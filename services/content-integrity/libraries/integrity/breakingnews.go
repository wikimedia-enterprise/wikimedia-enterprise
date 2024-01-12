package integrity

import (
	"context"
	"errors"
	"strings"
	"time"
	"wikimedia-enterprise/general/log"
	"wikimedia-enterprise/services/content-integrity/config/env"
	"wikimedia-enterprise/services/content-integrity/libraries/collector"

	"go.uber.org/dig"
	"golang.org/x/exp/slices"
)

// ErrEmptyDate error for empty date created.
var ErrEmptyDate = errors.New("empty date created and date namespace moved")

// BreakingNews Holder for CheckArticle for Breaking News.
type BreakingNews struct {
	dig.In
	Env       *env.Environment
	Collector collector.ArticleAPI
}

// CheckArticle determines if an article is breaking news or not.
func (c *BreakingNews) CheckArticle(ctx context.Context, art *Article, aps *ArticleParams) error {
	if art.DateCreated == nil && art.DateNamespaceMoved == nil {
		art.SetIsBreakingNews(false)

		return ErrEmptyDate
	}

	dir := func(dtc *time.Time, dur int) bool {
		if dtc == nil {
			return false
		}

		dtm := aps.GetDateModified()
		dlt := dtc.Add(time.Hour * time.Duration(dur))

		return (dtm.Equal(*dtc) || dtm.After(*dtc)) && (dtm.Equal(dlt) || dtm.Before(dlt))
	}

	mts := aps.GetMatchingTemplates(c.Env.BreakingNewsMandatoryTemplates...)
	pts := c.getTemplatesByPrefixFiltered(aps, c.Env.BreakingNewsTemplatesPrefixIgnore...)
	dtr := art.GetDateCreated()

	if dtr == nil {
		dtr = art.GetDateNamespaceMoved()
	}

	une := c.getUniqueEditorsCount(art.GetVersions(), aps.GetVersionIdentifier(), dtr)
	edt := c.getUniqueEditsCount(art.GetVersions(), aps.GetVersionIdentifier(), dtr)
	ibn := false

	if dir(art.GetDateCreated(), c.Env.BreakingNewsCreatedHours) ||
		dir(art.GetDateNamespaceMoved(), c.Env.BreakingNewsMovedHours) {
		if len(mts) > 0 {
			if len(pts) > 0 {
				ibn = true
			} else if une >= c.Env.BreakingNewsUniqueEditors {
				ibn = true
			} else if edt >= c.Env.BreakingNewsEdits {
				ibn = true
			}
		}
	}

	dif := aps.DateModified.Sub(*dtr)

	wbn, err := c.Collector.GetIsBreakingNews(ctx, art.Project, art.Identifier)
	itk := false

	if err != nil {
		log.Warn("failed to determine if revision is first marked as breaking news",
			log.Any("project", aps.GetProject()),
			log.Any("identifier", aps.GetIdentifier()),
			log.Any("revision", aps.GetVersionIdentifier()))

	} else if !wbn && ibn {
		err := c.Collector.SetIsBreakingNews(ctx, art.Project, art.Identifier)

		if err != nil {
			log.Warn("failed to mark revision as first breaking news",
				log.Any("project", aps.GetProject()),
				log.Any("identifier", aps.GetIdentifier()),
				log.Any("revision", aps.GetVersionIdentifier()))
		}

		itk = true
	}

	log.Info("breaking news event",
		log.Any("is_breaking_news", ibn),
		log.Any("date_created", art.GetDateCreated()),
		log.Any("created_date_check", dir(art.GetDateCreated(), c.Env.BreakingNewsCreatedHours)),
		log.Any("date_namespace_moved", art.GetDateNamespaceMoved()),
		log.Any("date_namespace_moved_check", dir(art.GetDateNamespaceMoved(), c.Env.BreakingNewsCreatedHours)),
		log.Any("date_tracked", dtr),
		log.Any("mins_since_tracked", dif.Minutes()),
		log.Any("revision_time", aps.DateModified),
		log.Any("templates", aps.GetTemplates()),
		log.Any("matched_templates", mts),
		log.Any("unique_editors", une >= c.Env.BreakingNewsUniqueEditors),
		log.Any("unique_editors_count", une),
		log.Any("edits", edt >= c.Env.BreakingNewsEdits),
		log.Any("edits_count", edt),
		log.Any("project", aps.GetProject()),
		log.Any("identifier", aps.GetIdentifier()),
		log.Any("revision", aps.GetVersionIdentifier()),
		log.Any("first", itk),
		log.Any("article_name", art.GetName()),
		log.Any("revision_distance", len(art.GetVersions())),
	)

	art.SetIsBreakingNews(ibn)

	return nil
}

func (c *BreakingNews) getTemplatesByPrefixFiltered(aps *ArticleParams, ign ...string) []string {
	pts := aps.GetTemplatesByPrefix(c.Env.BreakingNewsTemplatesPrefix...)
	rst := make([]string, 0)

	for _, elm := range pts {
		if !slices.ContainsFunc(c.Env.BreakingNewsTemplatesPrefixIgnore,
			func(val string) bool { return strings.EqualFold(val, elm) }) {
			rst = append(rst, elm)
		}
	}

	return rst
}

func (c *BreakingNews) getUniqueEditorsCount(vrs collector.Versions, ver int, dtr *time.Time) int {
	dlt := dtr.Add(time.Hour * time.Duration(c.Env.BreakingNewsUniqueEditorsHours))
	eds := make([]string, 0)
	trk := false

	for _, val := range vrs {
		if val.Identifier == ver {
			trk = true
		}

		if trk {
			if (dlt.Equal(*val.DateCreated) || dlt.After(*val.DateCreated)) &&
				(dtr.Equal(*val.DateCreated) || dtr.Before(*val.DateCreated)) &&
				!slices.ContainsFunc(eds, func(elm string) bool { return strings.EqualFold(elm, val.Editor) }) {
				eds = append(eds, val.Editor)
			}

		}
	}

	return len(eds)
}

func (c *BreakingNews) getUniqueEditsCount(vrs collector.Versions, ver int, dtr *time.Time) int {
	dlt := dtr.Add(time.Hour * time.Duration(c.Env.BreakingNewsEditsHours))
	edt := 0
	trk := false

	for _, val := range vrs {
		if val.Identifier == ver {
			trk = true
		}

		if trk {
			if (dlt.Equal(*val.DateCreated) || dlt.After(*val.DateCreated)) &&
				(dtr.Equal(*val.DateCreated) || dtr.Before(*val.DateCreated)) {
				edt++
			}
		}
	}

	return edt
}
