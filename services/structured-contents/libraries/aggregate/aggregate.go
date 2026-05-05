// Package aggregate creates a wrapper for concurrent API calls to WMF APIs and parser.
package aggregate

import (
	"context"
	"wikimedia-enterprise/services/structured-contents/submodules/log"
	parser "wikimedia-enterprise/services/structured-contents/submodules/structured-content-parser"
	"wikimedia-enterprise/services/structured-contents/submodules/wmf"

	"github.com/PuerkitoBio/goquery"
)

// Aggregation represent API calls response for the package.
type Aggregation struct {
	PageSummary string
	Infoboxes   []*parser.Part
	Sections    []*parser.Part
	References  []*parser.Reference
}

// GetPageSummary get page summary
func (a *Aggregation) GetPageSummary() string {
	return a.PageSummary
}

// GetInfoboxes get infoboxes
func (a *Aggregation) GetInfoboxes() []*parser.Part {
	return a.Infoboxes
}

// GetSections get page sections
func (a *Aggregation) GetSections() []*parser.Part {
	return a.Sections
}

// GetReferences get page references
func (a *Aggregation) GetReferences() []*parser.Reference {
	return a.References
}

// Getter an interface to allow injection of data getters (API calls) into aggregation helper.
type Getter interface {
	SetAPIs(api wmf.API, prs parser.API)
	GetData(ctx context.Context, dtb string, ttl string, doc *goquery.Document) (interface{}, error)
	GetField() string
}

// AggregationGetter interface to hide GetAggregation method for dependency injection and unit testing.
type AggregationGetter interface {
	GetAggregation(ctx context.Context, dtb string, ttl string, doc *goquery.Document, val *Aggregation, grs ...Getter) error
}

// Aggregator interface to hide GetAggregation methods for unit testing and dependency injection.
type Aggregator interface {
	AggregationGetter
}

// New creates new instance of Aggregator for dependency injection.
func New(wmf wmf.API, prs parser.API) Aggregator {
	return &Aggregate{
		API:    wmf,
		Parser: prs,
	}
}

// Aggregate implements GetAggregation to concurrently get data from parser and API.
type Aggregate struct {
	API    wmf.API
	Parser parser.API
}

type result struct {
	data  interface{}
	err   error
	field string
}

// GetAggregation gets list of titles, database name and getters and provides single aggregation result for single title.
func (a *Aggregate) GetAggregation(ctx context.Context, dtb string, ttl string, doc *goquery.Document, val *Aggregation, grs ...Getter) error {
	qln := len(grs)
	res := make(chan result, qln)

	for _, gtr := range grs {
		gtr.SetAPIs(a.API, a.Parser)

		go func(gtr Getter) {
			agr, err := gtr.GetData(ctx, dtb, ttl, doc)
			res <- result{data: agr, err: err, field: gtr.GetField()}
		}(gtr)
	}

	for i := 0; i < qln; i++ {
		rst := <-res

		if rst.err != nil {
			if rst.field == PageSummaryFieldName {
				log.Debug("pagesummary error, ignoring", log.Any("title", ttl), log.Any("database", dtb), log.Any("error", rst.err))
				continue
			}
			return rst.err
		}

		switch rst.field {
		case PageSummaryFieldName:
			if summary, ok := rst.data.(string); ok && summary != "" {
				val.PageSummary = summary
			}

		case InfoboxesFieldName:
			val.Infoboxes = rst.data.([]*parser.Part)

		case SectionsFieldName:
			val.Sections = rst.data.([]*parser.Part)

		case ReferencesFieldName:
			val.References = rst.data.([]*parser.Reference)
		}

	}

	return nil
}
