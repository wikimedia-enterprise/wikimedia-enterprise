// Package integrity exposes API to different content integrity
// algorithms and information.
package integrity

import (
	"context"
	"strings"
	"time"
	"wikimedia-enterprise/services/content-integrity/libraries/collector"

	"github.com/redis/go-redis/v9"
	"golang.org/x/exp/slices"
)

// Article structure to track values of Article in content integrity.
type Article struct {
	Identifier         int
	Name               string
	Project            string
	VersionIdentifier  int
	EditsCount         int
	UniqueEditorsCount int
	IsBreakingNews     bool
	DateCreated        *time.Time
	DateNamespaceMoved *time.Time
	Versions           collector.Versions
}

// SetIdentifier sets the article identifier.
func (a *Article) SetIdentifier(idr int) {
	a.Identifier = idr
}

// SetProject sets the article project.
func (a *Article) SetProject(prj string) {
	a.Project = prj
}

// SetVersionIdentifier sets the article version identifier.
func (a *Article) SetVersionIdentifier(idr int) {
	a.VersionIdentifier = idr
}

// SetEditsCount sets the article edits count.
func (a *Article) SetEditsCount(ect int) {
	a.EditsCount = ect
}

// SetDateCreated sets the article date created.
func (a *Article) SetDateCreated(dct *time.Time) {
	a.DateCreated = dct
}

// SetDateNamespaceMoved sets the article date when namespace was moved.
func (a *Article) SetDateNamespaceMoved(dmt *time.Time) {
	a.DateNamespaceMoved = dmt
}

// SetUniqueEditorsCount sets the article unique editors count.
func (a *Article) SetUniqueEditorsCount(uec int) {
	a.UniqueEditorsCount = uec
}

// SetIsBreakingNews sets the article flag for breaking news.
func (a *Article) SetIsBreakingNews(ibn bool) {
	a.IsBreakingNews = ibn
}

// SetVersions sets the article versions.
func (a *Article) SetVersions(vrs collector.Versions) {
	a.Versions = vrs
}

// SetName sets the article name.
func (a *Article) SetName(nme string) {
	a.Name = nme
}

// GetIdentifier gets the article identifier.
func (a *Article) GetIdentifier() int {
	if a == nil {
		return 0
	}

	return a.Identifier
}

// GetProject gets the article project.
func (a *Article) GetProject() string {
	if a == nil {
		return ""
	}

	return a.Project
}

// GetIsBreakingNews gets the article flag for breaking news.
func (a *Article) GetIsBreakingNews() bool {
	if a == nil {
		return false
	}

	return a.IsBreakingNews
}

// GetVersionIdentifier gets the article version identifier.
func (a *Article) GetVersionIdentifier() int {
	if a == nil {
		return 0
	}

	return a.VersionIdentifier
}

// GetEditsCount gets the article edits count.
func (a *Article) GetEditsCount() int {
	if a == nil {
		return 0
	}

	return a.EditsCount
}

// GetUniqueEditorsCount gets the article unique editors count.
func (a *Article) GetUniqueEditorsCount() int {
	if a == nil {
		return 0
	}

	return a.UniqueEditorsCount
}

// GetDateCreated gets the article date created.
func (a *Article) GetDateCreated() *time.Time {
	if a == nil {
		return nil
	}

	return a.DateCreated
}

// GetDateNamespaceMoved gets the article date when namespace was moved.
func (a *Article) GetDateNamespaceMoved() *time.Time {
	if a == nil {
		return nil
	}

	return a.DateNamespaceMoved
}

// GetVersions gets the article versions.
func (a *Article) GetVersions() collector.Versions {
	if a == nil {
		return nil
	}

	return a.Versions
}

// GetName gets the article name.
func (a *Article) GetName() string {
	if a == nil {
		return ""
	}

	return a.Name
}

// ArticleParams main structure for content integrity parameters holder.
type ArticleParams struct {
	Project           string
	Identifier        int
	VersionIdentifier int
	DateModified      time.Time
	Templates         []string
	Categories        []string
}

// GetDateModified gets the article date modified
func (a *ArticleParams) GetDateModified() time.Time {
	if a == nil {
		return time.Now()
	}

	return a.DateModified
}

// GetIdentifier gets the article identifier.
func (a *ArticleParams) GetIdentifier() int {
	if a == nil {
		return 0
	}

	return a.Identifier
}

// GetProject gets the article project.
func (a *ArticleParams) GetProject() string {
	if a == nil {
		return ""
	}

	return a.Project
}

// GetVersionIdentifier gets the article version (revision ID).
func (a *ArticleParams) GetVersionIdentifier() int {
	if a == nil {
		return 0
	}

	return a.VersionIdentifier
}

// GetTemplates gets all Templates for given article parameters.
func (a *ArticleParams) GetTemplates() []string {
	return a.Templates
}

// GetMatchingTemplates returns all templates that match input list.
func (a *ArticleParams) GetMatchingTemplates(tps ...string) []string {
	if a == nil {
		return nil
	}

	if len(tps) == 0 {
		return []string{}
	}

	rts := []string{}

	for _, tpl := range a.Templates {
		if slices.ContainsFunc(tps, func(a string) bool { return strings.EqualFold(a, tpl) }) {
			rts = append(rts, tpl)
		}
	}

	return rts
}

// GetTemplatesByPrefix gets all templates that have prefixes in input list.
func (a *ArticleParams) GetTemplatesByPrefix(pfs ...string) []string {
	if a == nil {
		return nil
	}

	if len(pfs) == 0 {
		return make([]string, 0)
	}

	rts := []string{}

	for _, tpl := range a.Templates {
		if slices.ContainsFunc(pfs, func(a string) bool { return strings.HasPrefix(tpl, a) }) {
			rts = append(rts, tpl)
		}
	}

	return rts
}

// ArticleChecker interface exposes a method to check the article based on parameters.
type ArticleChecker interface {
	CheckArticle(ctx context.Context, art *Article, aps *ArticleParams) error
}

// ArticleGetter exposes method to get article content integrity information.
type ArticleGetter interface {
	GetArticle(ctx context.Context, aps *ArticleParams, chs ...ArticleChecker) (*Article, error)
}

// API the content integrity main API.
type API interface {
	ArticleGetter
}

// New creates new instance of the Integrity.
func New(col collector.ArticleAPI) API {
	return &Integrity{
		Collector: col,
	}
}

// Integrity the main integrity struct for content integrity.
type Integrity struct {
	Collector collector.ArticleAPI
}

// GetArticle method to get article content integrity information.
func (h *Integrity) GetArticle(ctx context.Context, aps *ArticleParams, chs ...ArticleChecker) (*Article, error) {
	prj := aps.GetProject()
	idr := aps.GetIdentifier()

	ibt, err := h.Collector.GetIsBeingTracked(ctx, prj, idr)

	if err != nil {
		return nil, err
	}

	if !ibt {
		return nil, nil
	}

	dtc, err := h.Collector.GetDateCreated(ctx, prj, idr)

	if err != nil && err != redis.Nil {
		return nil, err
	}

	dnm, err := h.Collector.GetDateNamespaceMoved(ctx, prj, idr)

	if err != nil && err != redis.Nil {
		return nil, err
	}

	vrs, err := h.Collector.GetVersions(ctx, prj, idr)

	if err != nil {
		return nil, err
	}

	atn, err := h.Collector.GetName(ctx, prj, idr)

	if err != nil {
		return nil, err
	}

	art := new(Article)
	art.SetIdentifier(idr)
	art.SetProject(prj)
	art.SetVersionIdentifier(vrs.GetCurrentIdentifier())
	art.SetUniqueEditorsCount(vrs.GetUniqueEditorsCount())
	art.SetEditsCount(len(vrs))
	art.SetDateCreated(dtc)
	art.SetDateNamespaceMoved(dnm)
	art.SetVersions(vrs)
	art.SetName(atn)

	for _, chr := range chs {
		if err := chr.CheckArticle(ctx, art, aps); err != nil {
			return nil, err
		}
	}

	return art, nil
}
