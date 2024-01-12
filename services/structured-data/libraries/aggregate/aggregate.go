// Package aggregate creates a wrapper for concurrent API calls to WMF API(s).
package aggregate

import (
	"context"
	"time"
	"wikimedia-enterprise/general/wmf"

	"go.uber.org/dig"
)

// Aggregation represent API calls response for the package.
type Aggregation struct {
	Page     *wmf.Page
	PageHTML *wmf.PageHTML
	Revision *wmf.Revision
	User     *wmf.User
	Score    *wmf.Score
}

// GetPage returns the page struct from aggregation.
func (a *Aggregation) GetPage() *wmf.Page {
	return a.Page
}

// GetPageMissing returns the page missing property from aggregation.
func (a *Aggregation) GetPageMissing() bool {
	if pge := a.GetPage(); pge != nil {
		return pge.Missing
	}

	return true
}

// GetPageTitle returns the page title property from aggregation.
func (a *Aggregation) GetPageTitle() string {
	if pge := a.GetPage(); pge != nil {
		return pge.Title
	}

	return ""
}

// GetPageLanguage returns the page language property fro aggregation.
func (a *Aggregation) GetPageLanguage() string {
	if pge := a.GetPage(); pge != nil {
		return pge.PageLanguage
	}

	return ""
}

// GetPageNs gets the page namespace property from aggregation.
func (a *Aggregation) GetPageNs() int {
	if pge := a.GetPage(); pge != nil {
		return pge.Ns
	}

	return 0
}

// GetPageWatchers returns the page watchers property from aggregation.
func (a *Aggregation) GetPageWatchers() int {
	if pge := a.GetPage(); pge != nil {
		return pge.Watchers
	}

	return 0
}

// GetPageID returns the page ID property from aggregation.
func (a *Aggregation) GetPageID() int {
	if pge := a.GetPage(); pge != nil {
		return pge.PageID
	}

	return 0
}

// GetPageLastRevID returns the page last revision ID property from aggregation.
func (a *Aggregation) GetPageLastRevID() int {
	if pge := a.GetPage(); pge != nil {
		return pge.LastRevID
	}

	return 0
}

// GetPageFlagged returns the page flagged property from aggregation.
func (a *Aggregation) GetPageFlagged() *wmf.Flagged {
	if pge := a.GetPage(); pge != nil && pge.Flagged != nil {
		return a.Page.Flagged
	}

	return nil
}

// GetPageFlaggedStableRevID returns the page flagged stable revision ID property from aggregation.
func (a *Aggregation) GetPageFlaggedStableRevID() int {
	if pgf := a.GetPageFlagged(); pgf != nil {
		return pgf.StableRevID
	}

	return 0
}

// GetPageCanonicalURL returns the page canonical url property from aggregation.
func (a *Aggregation) GetPageCanonicalURL() string {
	if pge := a.GetPage(); pge != nil {
		return pge.CanonicalURL
	}

	return ""
}

// GetPagePropsWikiBaseItem returns the page wiki base item property from aggregation.
func (a *Aggregation) GetPagePropsWikiBaseItem() string {
	if pge := a.GetPage(); pge != nil && pge.PageProps != nil {
		return pge.PageProps.WikiBaseItem
	}

	return ""
}

// GetPageWbEntityUsage returns the page web entity usage property from aggregation.
func (a *Aggregation) GetPageWbEntityUsage() map[string]*wmf.WbEntityUsage {
	if pge := a.GetPage(); pge != nil {
		return pge.WbEntityUsage
	}

	return nil
}

// GetPageProtection returns the page protection  property from aggregation.
func (a *Aggregation) GetPageProtection() []*wmf.Protection {
	if pge := a.GetPage(); pge != nil {
		return pge.Protection
	}

	return nil
}

// GetPageCategories returns the page categories  property from aggregation.
func (a *Aggregation) GetPageCategories() []*wmf.Category {
	if pge := a.GetPage(); pge != nil {
		return pge.Categories
	}

	return nil
}

// GetPageTemplates returns the page templates property from aggregation.
func (a *Aggregation) GetPageTemplates() []*wmf.Template {
	if pge := a.GetPage(); pge != nil {
		return pge.Templates
	}

	return nil
}

// GetPageRedirects returns the page redirects  property from aggregation.
func (a *Aggregation) GetPageRedirects() []*wmf.Redirect {
	if pge := a.GetPage(); pge != nil {
		return pge.Redirects
	}

	return nil
}

// GetPageHTML returns the page HTML from aggregation.
func (a *Aggregation) GetPageHTML() *wmf.PageHTML {
	return a.PageHTML
}

// GetPageHTMLContent returns the page HTML content property from aggregation.
func (a *Aggregation) GetPageHTMLContent() string {
	if phl := a.GetPageHTML(); phl != nil {
		return phl.Content
	}

	return ""
}

// GetPageHTMLError returns the page HTML error property from aggregation.
func (a *Aggregation) GetPageHTMLError() error {
	if phl := a.GetPageHTML(); phl != nil {
		return phl.Error
	}

	return nil
}

// GetUser returns the user from aggregation.
func (a *Aggregation) GetUser() *wmf.User {
	return a.User
}

// GetUserRegistration returns the user registration date property from aggregation.
func (a *Aggregation) GetUserRegistration() *time.Time {
	if usr := a.GetUser(); usr != nil {
		return usr.Registration
	}

	return nil
}

// GetUserEditCount returns the user edit count property from aggregation.
func (a *Aggregation) GetUserEditCount() int {
	if usr := a.GetUser(); usr != nil {
		return usr.EditCount
	}

	return 0
}

// GetUserGroups returns the user groups property from aggregation.
func (a *Aggregation) GetUserGroups() []string {
	if usr := a.GetUser(); usr != nil {
		return usr.Groups
	}

	return nil
}

// GetFirstRevision returns the first revision from aggregation.
func (a *Aggregation) GetFirstRevision() *wmf.Revision {
	return a.Revision
}

// GetFirstRevisionTimestamp returns the first revision timestamp property from aggregation.
func (a *Aggregation) GetFirstRevisionTimestamp() *time.Time {
	if frv := a.GetFirstRevision(); frv != nil {
		return frv.Timestamp
	}

	return nil
}

// GetFirstRevisionContent returns the first revision content property from aggregation.
func (a *Aggregation) GetFirstRevisionContent() string {
	if frv := a.GetFirstRevision(); frv != nil && frv.Slots != nil {
		return frv.Slots.Main.Content
	}

	return ""
}

// GetCurrentRevision returns the current revision from aggregation.
func (a *Aggregation) GetCurrentRevision() *wmf.Revision {
	if len(a.Page.Revisions) > 0 {
		return a.Page.Revisions[0]
	}

	return nil
}

// GetCurrentRevisionTags returns the current revision tags property from aggregation.
func (a *Aggregation) GetCurrentRevisionTags() []string {
	if crv := a.GetCurrentRevision(); crv != nil {
		return crv.Tags
	}

	return nil
}

// GetCurrentRevisionMinor returns the current revision minor property from aggregation.
func (a *Aggregation) GetCurrentRevisionMinor() bool {
	if crv := a.GetCurrentRevision(); crv != nil {
		return crv.Minor
	}

	return false
}

// GetCurrentRevisionComment returns the current revision comment property from aggregation.
func (a *Aggregation) GetCurrentRevisionComment() string {
	if crv := a.GetCurrentRevision(); crv != nil {
		return crv.Comment
	}

	return ""
}

// GetCurrentRevisionUser returns the current revision user property from aggregation.
func (a *Aggregation) GetCurrentRevisionUser() string {
	if crv := a.GetCurrentRevision(); crv != nil {
		return crv.User
	}

	return ""
}

// GetCurrentRevisionUserID returns the current revision user ID property from aggregation.
func (a *Aggregation) GetCurrentRevisionUserID() int {
	if crv := a.GetCurrentRevision(); crv != nil {
		return crv.UserID
	}

	return 0
}

// GetCurrentRevisionTimestamp returns the current revision timestamp property from aggregation.
func (a *Aggregation) GetCurrentRevisionTimestamp() *time.Time {
	if crv := a.GetCurrentRevision(); crv != nil {
		return crv.Timestamp
	}

	return nil
}

// GetCurrentRevisionContent returns the current revision content property from aggregation.
func (a *Aggregation) GetCurrentRevisionContent() string {
	if crv := a.GetCurrentRevision(); crv != nil && crv.Slots != nil {
		return crv.Slots.Main.Content
	}

	return ""
}

// GetCurrentRevisionContentModel returns the content model for current revision.
func (a *Aggregation) GetCurrentRevisionContentModel() string {
	if crv := a.GetCurrentRevision(); crv != nil && crv.Slots != nil {
		return crv.Slots.Main.ContentModel
	}

	return ""
}

// GetCurrentRevisionParentID returns the current revision parent ID property from aggregation.
func (a *Aggregation) GetCurrentRevisionParentID() int {
	if crv := a.GetCurrentRevision(); crv != nil {
		return crv.ParentID
	}

	return 0
}

// GetPreviousRevision returns the previous revision from aggregation.
func (a *Aggregation) GetPreviousRevision() *wmf.Revision {
	if len(a.Page.Revisions) > 1 {
		return a.Page.Revisions[1]
	}

	return nil
}

// GetPreviousRevisionID returns the previous revision ID property from aggregation.
func (a *Aggregation) GetPreviousRevisionID() int {
	if prv := a.GetPreviousRevision(); prv != nil {
		return prv.RevID
	}

	return 0
}

// GetPreviousRevisionContent returns the previous revision content property from aggregation.
func (a *Aggregation) GetPreviousRevisionContent() string {
	if prv := a.GetPreviousRevision(); prv != nil && prv.Slots != nil {
		return prv.Slots.Main.Content
	}

	return ""
}

// GetPreviousRevisionTimestamp returns the previous revision timestamp property from aggregation.
func (a *Aggregation) GetPreviousRevisionTimestamp() *time.Time {
	if prv := a.GetPreviousRevision(); prv != nil {
		return prv.Timestamp
	}

	return nil
}

// GetPageOriginalImage returns the original image for the page.
func (a *Aggregation) GetPageOriginalImage() *wmf.Image {
	if pge := a.GetPage(); pge != nil {
		return pge.Original
	}

	return nil
}

// GetPageThumbnailImage returns the thumbnail image for the page.
func (a *Aggregation) GetPageThumbnailImage() *wmf.Image {
	if pge := a.GetPage(); pge != nil {
		return pge.Thumbnail
	}

	return nil
}

// GetScore returns the score from aggregation.
func (a *Aggregation) GetScore() *wmf.Score {
	return a.Score
}

// GetRevertRiskScore returns the revertrisk score from aggregation.
func (a *Aggregation) GetRevertRiskScore() *wmf.LiftWingScore {
	if scr := a.GetScore(); scr != nil && scr.Output != nil {
		return scr.Output
	}

	return nil
}

// Getter an interface to allow injection of data getters (API calls) into aggregation helper.
type Getter interface {
	SetAPI(api wmf.API)
	GetData(ctx context.Context, dtb string, tls []string) (interface{}, error)
}

// AggregationGetter interface to hide GetAggregation method for dependency injection and unit testing.
type AggregationGetter interface {
	GetAggregation(ctx context.Context, dtb string, ttl string, val *Aggregation, grs ...Getter) error
}

// AggregationGetter interface to hide GetAggregations method for dependency injection and unit testing.
type AggregationsGetter interface {
	GetAggregations(ctx context.Context, dtb string, tls []string, val map[string]*Aggregation, grs ...Getter) error
}

// Aggregator interface to hide both GetAggregations and GetAggregation methods for unit testing and dependency injection.
type Aggregator interface {
	AggregationGetter
	AggregationsGetter
}

// New creates new instance of Aggregator for dependency injection.
func New(agg Aggregate) Aggregator {
	return &agg
}

// Aggregate holds implementation of aggregation methods that will combine and add concurrency to API calls.
type Aggregate struct {
	dig.In
	API wmf.API
}

// GetAggregation gets list of titles, database name and getters and provides single aggregation result for single title.
func (a *Aggregate) GetAggregation(ctx context.Context, dtb string, ttl string, agg *Aggregation, grs ...Getter) error {
	ags := map[string]*Aggregation{
		ttl: agg,
	}

	return a.GetAggregations(ctx, dtb, []string{ttl}, ags, grs...)
}

// GetAggregation gets list of titles, database name and getters and provides multiple aggregation results for different titles.
func (a *Aggregate) GetAggregations(ctx context.Context, dtb string, tls []string, ags map[string]*Aggregation, grs ...Getter) error {
	qln := len(grs)
	ers := make(chan error, qln)
	ars := make(chan interface{}, qln)

	for _, gtr := range grs {
		gtr.SetAPI(a.API)

		go func(gtr Getter) {
			agr, err := gtr.GetData(ctx, dtb, tls)
			ars <- agr
			ers <- err
		}(gtr)
	}

	for _, ttl := range tls {
		if _, ok := ags[ttl]; !ok {
			ags[ttl] = &Aggregation{}
		}
	}

	for i := 0; i < qln; i++ {
		if err := <-ers; err != nil {
			return err
		}

		switch dta := (<-ars).(type) {
		case map[string]*wmf.Page:
			for ttl, pge := range dta {
				if _, ok := ags[ttl]; !ok {
					ags[ttl] = &Aggregation{}
				}

				ags[ttl].Page = pge
			}
		case map[string]*wmf.PageHTML:
			for ttl, phl := range dta {
				if _, ok := ags[ttl]; !ok {
					ags[ttl] = &Aggregation{}
				}

				ags[ttl].PageHTML = phl
			}
		case map[string]*wmf.Revision:
			for ttl, rvn := range dta {
				if _, ok := ags[ttl]; !ok {
					ags[ttl] = &Aggregation{}
				}

				ags[ttl].Revision = rvn
			}
		case map[string]*wmf.User:
			for ttl, usr := range dta {
				if _, ok := ags[ttl]; !ok {
					ags[ttl] = &Aggregation{}
				}

				ags[ttl].User = usr
			}
		case map[string]*wmf.Score:
			for ttl, scr := range dta {
				if _, ok := ags[ttl]; !ok {
					ags[ttl] = &Aggregation{}
				}

				ags[ttl].Score = scr
			}
		}
	}

	return nil
}
