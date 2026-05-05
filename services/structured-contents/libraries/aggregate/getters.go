package aggregate

import (
	"context"
	"errors"
	"fmt"
	parser "wikimedia-enterprise/services/structured-contents/submodules/structured-content-parser"
	"wikimedia-enterprise/services/structured-contents/submodules/wmf"

	"github.com/PuerkitoBio/goquery"
)

const (
	PageSummaryFieldName = "PageSummary"
	InfoboxesFieldName   = "Infoboxes"
	SectionsFieldName    = "Sections"
	ReferencesFieldName  = "References"
)

// ErrPageSummaryPrefix is the prefix for the PageSummaryGetter errors.
var ErrPageSummaryPrefix = errors.New("pagesummary: ")

// DefaultGetter implements the basic behaviour that will inject API clients.
type DefaultGetter struct {
	API    wmf.API
	Parser parser.API
}

// SetAPI sets API client as a dependency for the data getters.
func (d *DefaultGetter) SetAPIs(api wmf.API, prs parser.API) {
	d.API = api
	d.Parser = prs
}

// WithPageSummary initializes Page getter with required arguments.
func WithPageSummary() *PageSummaryGetter {
	return &PageSummaryGetter{}
}

// PageSummaryGetter allows to get page summary data in the aggregator.
type PageSummaryGetter struct {
	DefaultGetter
}

// GetData attempts to retrieve the contents of a short description template, and falls back on calling the PageSummary API.
func (p *PageSummaryGetter) GetData(ctx context.Context, dtb string, ttl string, doc *goquery.Document) (interface{}, error) {
	// https://en.wikipedia.org/wiki/Template:Short_description
	parsed := doc.Find(".shortdescription").First().Text()
	if parsed != "" {
		return parsed, nil
	}

	sum, err := p.API.GetPageSummary(ctx, dtb, ttl)

	if err != nil {
		return "", fmt.Errorf("%s%s", ErrPageSummaryPrefix, err)
	}

	return sum.Description, nil
}

// GetField returns the kind of of data the getter is getting.
func (p *PageSummaryGetter) GetField() string {
	return PageSummaryFieldName
}

// WithInfoboxes initializes Infobox getter with required arguments.
func WithInfoboxes() *InfoboxesGetter {
	return &InfoboxesGetter{}
}

// InfoboxesGetter allows to get page infoboxes in the aggregator.
type InfoboxesGetter struct {
	DefaultGetter
}

// GetData builds up the API query to get page metadata.
func (i *InfoboxesGetter) GetData(ctx context.Context, dtb string, ttl string, doc *goquery.Document) (interface{}, error) {
	return i.Parser.GetInfoBoxes(doc.Selection), nil
}

// GetField returns the kind of of data the getter is getting.
func (i *InfoboxesGetter) GetField() string {
	return InfoboxesFieldName
}

// WithSections initializes Section getter with required arguments.
func WithSections() *SectionsGetter {
	return &SectionsGetter{}
}

// SectionsGetter allows to get section data in the aggregator.
type SectionsGetter struct {
	DefaultGetter
}

// GetData builds up the API query to get page metadata.
func (s *SectionsGetter) GetData(ctx context.Context, dtb string, ttl string, doc *goquery.Document) (interface{}, error) {
	return s.Parser.GetSections(doc.Selection), nil
}

// GetField returns the kind of of data the getter is getting.
func (s *SectionsGetter) GetField() string {
	return SectionsFieldName
}

// WithReferences initializes References getter with required arguments.
func WithReferences() *ReferencesGetter {
	return &ReferencesGetter{}
}

// ReferencesGetter allows to get references data in the aggregator.
type ReferencesGetter struct {
	DefaultGetter
}

// GetData builds up the API query to get page metadata.
func (s *ReferencesGetter) GetData(ctx context.Context, dtb string, ttl string, doc *goquery.Document) (interface{}, error) {
	return s.Parser.GetReferences(doc.Selection), nil
}

// ReferencesGetter returns the kind of of data the getter is getting.
func (s *ReferencesGetter) GetField() string {
	return ReferencesFieldName
}
