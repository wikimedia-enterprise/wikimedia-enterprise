package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"wikimedia-enterprise/api/main/config/env"
	"wikimedia-enterprise/api/main/submodules/log"
	"wikimedia-enterprise/api/main/submodules/schema"
	parser "wikimedia-enterprise/api/main/submodules/structured-contents-parser"
	"wikimedia-enterprise/api/main/submodules/wmf"

	"github.com/PuerkitoBio/goquery"
	"github.com/tidwall/gjson"
	"go.uber.org/dig"
	"golang.org/x/exp/slices"
)

// InfoboxFieldName name of the field that contains infobox of the page.
const InfoboxFieldName = "infoboxes"

// DescriptionFieldName name of the field that contains wikipedia/wikidata description.
const DescriptionFieldName = "description"

// SectionsFieldName name of the field that contains wikipedia sections.
const SectionsFieldName = "sections"

// TablesFieldName name of the field that contains wikipedia tables.
const TablesFieldName = "tables"

// ReferencesFieldName name of the field that contains wikipedia references.
const ReferencesFieldName = "references"

// AttributionFieldName is the name of the field that contains attribution derived data.
const AttributionFieldName = "attribution"

// DefaultModifiers list of default modifiers.
var DefaultModifiers = []Modifier{
	new(FilterModifier),
	new(FieldModifier),
}

type TableWithID struct {
	Identifier      string          `json:"identifier"`
	Headers         [][]parser.Cell `json:"headers,omitempty"`
	Footers         [][]parser.Cell `json:"footers,omitempty"`
	Rows            [][]parser.Cell `json:"rows,omitempty"`
	ConfidenceScore float64         `json:"confidence_score,omitempty"`
}

// SOCKFields list of SOCK fields.
var SOCKFields = []string{
	"name",
	"identifier",
	"abstract",
	"version",
	"url",
	"date_created",
	"date_modified",
	"main_entity",
	"is_part_of",
	"additional_entities",
	"in_language",
	"image",
	"license",
}

// FilterModifier modifier that help to apply filters to the entities.
type FilterModifier struct{}

// Modify method that apply filters to the entities.
func (f *FilterModifier) Modify(ctx context.Context, mdl *Model, obj *gjson.Result) (bool, error) {
	for _, ftr := range mdl.Filters {
		val := obj.Get(ftr.Field).Value()

		if ftr.Filter(val) {
			return true, nil
		}
	}

	return false, nil
}

// BaseFieldModifier modifier that help to select fields for the entities.
type BaseFieldModifier struct{}

// BuildQuery method that build json query for the entities.
func (f *BaseFieldModifier) BuildQuery(idx map[string]interface{}, sbr *strings.Builder) {
	cnt := 0
	lnt := len(idx)

	for nme, idi := range idx {
		cnt++

		switch val := idi.(type) {
		case map[string]interface{}:
			if _, err := fmt.Fprintf(sbr, `"%s":{`, nme); err != nil {
				log.Error(err)
			}

			f.BuildQuery(val, sbr)

			if _, err := sbr.WriteString("}"); err != nil {
				log.Error(err)
			}
		case string:
			if _, err := fmt.Fprintf(sbr, `"%s":%s`, nme, val); err != nil {
				log.Error(err)
			}
		}

		if cnt != lnt {
			if _, err := sbr.WriteString(","); err != nil {
				log.Error(err)
			}
		}
	}
}

// FieldModifier modifier that help to select fields for the entities.
type FieldModifier struct {
	BaseFieldModifier
}

// Modify method that select fields for the entities.
func (f *FieldModifier) Modify(ctx context.Context, mdl *Model, obj *gjson.Result) (bool, error) {
	sbr := new(strings.Builder)

	f.BuildQuery(mdl.GetFieldsIndex(), sbr)

	if sbr.Len() > 0 {
		stm := fmt.Sprintf("{%s}", sbr.String())
		*obj = obj.Get(stm)
	}

	return false, nil
}

// SOCKModifier modifier that helps to select fields for the entities.
type SOCKModifier struct {
	dig.In
	BaseFieldModifier `optional:"true"`
	Parser            parser.API
	Env               *env.Environment
	WMF               wmf.API
}

// addDescription method that fetches description from summary API.
func (f *SOCKModifier) addDescription(ctx context.Context, sbr *strings.Builder, obj *gjson.Result) error {
	if !f.Env.DescriptionEnabled {
		return nil
	}

	ttl := obj.Get("name").String()
	prj := obj.Get("is_part_of.identifier").String()
	sum, err := f.WMF.GetPageSummary(ctx, prj, ttl)

	if err != nil {
		return err
	}
	dsc, err := json.Marshal(sum.Description)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(sbr, `,"%s":!%s`, DescriptionFieldName, string(dsc))
	return err
}

// addInfoboxes method that adds infoboxes field.
func (f *SOCKModifier) addInfoboxes(sbr *strings.Builder, obj *gjson.Result) error {
	hml := obj.Get("article_body.html").String()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(hml))

	if err != nil {
		return err
	}

	ibs := f.Parser.GetInfoBoxes(doc.Selection)

	if len(ibs) == 0 {
		return nil
	}

	dta, err := json.Marshal(ibs)

	if err != nil {
		return err
	}

	hps := fmt.Sprintf(`,"%s":!%s`, InfoboxFieldName, string(dta))

	if _, err = sbr.WriteString(hps); err != nil {
		log.Error(err)
	}

	return nil
}

// addArticleParts method that adds tables and sections field.
func (f *SOCKModifier) addArticleParts(sbr *strings.Builder, obj *gjson.Result) error {
	if !f.Env.SectionsEnabled {
		return nil
	}

	html := obj.Get("article_body.html").String()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return err
	}

	sections, tables := f.Parser.GetArticleParts(doc.Selection)

	// Add sections
	if len(sections) > 0 {
		sectionData, err := json.Marshal(sections)
		if err != nil {
			return err
		}
		sectionPart := fmt.Sprintf(`,"%s":!%s`, SectionsFieldName, string(sectionData))
		if _, err := sbr.WriteString(sectionPart); err != nil {
			log.Error(err)
		}
	}

	// Add tables
	if f.Env.EnableTables && len(tables) > 0 {
		var tableList []TableWithID
		for _, tbl := range tables {
			tableList = append(tableList, TableWithID{
				Identifier:      tbl.ID,
				Headers:         tbl.Table.Headers,
				Rows:            tbl.Table.Rows,
				Footers:         tbl.Table.Footers,
				ConfidenceScore: tbl.Table.ConfidenceScore,
			})
		}
		tableData, err := json.Marshal(tableList)
		if err != nil {
			return err
		}
		tablePart := fmt.Sprintf(`,"%s":!%s`, TablesFieldName, string(tableData))
		if _, err := sbr.WriteString(tablePart); err != nil {
			log.Error(err)
		}
	}

	return nil
}

// getReferences parses and returns the references out of the article.
func (f *SOCKModifier) getReferences(obj *gjson.Result) ([]*parser.Reference, error) {
	hml := obj.Get("article_body.html").String()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(hml))
	if err != nil {
		return nil, err
	}

	return f.Parser.GetReferences(doc.Selection), nil
}

// addReferences adds the references field.
func (f *SOCKModifier) addReferences(sbr *strings.Builder, refs []*parser.Reference) error {
	if len(refs) == 0 {
		return nil
	}

	dta, err := json.Marshal(refs)
	if err != nil {
		return err
	}
	hps := fmt.Sprintf(`,"%s":!%s`, ReferencesFieldName, string(dta))

	if _, err = sbr.WriteString(hps); err != nil {
		log.Error(err)
	}

	return nil
}

func boundedString(count int, max int, alreadyBounded bool) string {
	if count > max {
		return fmt.Sprintf(">%d", max)
	}

	// count <= max, but we're told that this is an upper bound already, so the real number is >count.
	if alreadyBounded {
		return fmt.Sprintf(">%d", count)
	}

	return fmt.Sprintf("%d", count)
}

func getLastRevised(obj *gjson.Result) string {
	dateModified := obj.Get("date_modified").String()
	parsed, err := time.Parse("2006-01-02T15:04:05Z", dateModified)
	if err != nil {
		log.Warn("Failed to parse date_modified", log.Any("date_modified", dateModified), log.Any("error", err.Error()))
	}

	return parsed.Format("01/06")
}

func (f *SOCKModifier) GetAttribution(ctx context.Context, obj *gjson.Result, refs []*parser.Reference) (*schema.Attribution, error) {
	pageid := obj.Get("identifier").Int()
	prj := obj.Get("is_part_of.identifier").String()
	lastRevised := getLastRevised(obj)

	maxContrib := 100
	cnt, err := f.WMF.GetContributorsCount(ctx, prj, int(pageid), &maxContrib)
	if err != nil {
		return nil, fmt.Errorf("failed to get editors for %d: %w", pageid, err)
	}

	attr := &schema.Attribution{
		EditorsCount: &schema.EditorsCount{
			AnonymousCount:  boundedString(cnt.AnonymousCount, maxContrib, false),
			RegisteredCount: boundedString(cnt.RegisteredCount, maxContrib, !cnt.AllRegisteredIncluded),
		},
		ParsedReferencesCount: len(refs),
	}

	if lastRevised != "" {
		attr.LastRevised = lastRevised
	}

	return attr, nil
}

// addAttribution adds the attribution field.
func (f *SOCKModifier) addAttribution(ctx context.Context, sbr *strings.Builder, obj *gjson.Result, refs []*parser.Reference) error {
	attr, err := f.GetAttribution(ctx, obj, refs)
	if err != nil {
		return err
	}
	dta, err := json.Marshal(attr)
	if err != nil {
		return err
	}
	hps := fmt.Sprintf(`,"%s":!%s`, AttributionFieldName, string(dta))
	if _, err = sbr.WriteString(hps); err != nil {
		log.Error(err)
	}

	return nil
}

func (f *SOCKModifier) Modify(ctx context.Context, mdl *Model, obj *gjson.Result) (bool, error) {
	sbr := new(strings.Builder)
	fix := mdl.GetFieldsIndex()
	idx := map[string]interface{}{}

	// check if the fields correspond to the SOCK fields
	// if they are push them to new index map
	for fld := range fix {
		if slices.Contains(SOCKFields, fld) {
			idx[fld] = fix[fld]
		}
	}

	// build get query the json query
	f.BuildQuery(idx, sbr)

	// check if we need to compute has_parts field
	_, ibx := fix[InfoboxFieldName]
	_, dsc := fix[DescriptionFieldName]
	_, scs := fix[SectionsFieldName]
	_, refs := fix[ReferencesFieldName]
	_, attr := fix[AttributionFieldName]
	attr = attr && f.Env.EnableAttribution

	// if the query string is empty
	// then select all SOCK fields
	if sbr.Len() == 0 && !ibx && !dsc && !scs && !refs && !attr {
		fds := strings.Join(SOCKFields, ",")

		if _, err := sbr.WriteString(fds); err != nil {
			log.Error(err)
		}
	}

	// check if the index map contains the extra field(s)
	// or if the index map is empty
	// if not then add it and compute extra JSON fields
	if len(fix) == 0 || dsc {
		if err := f.addDescription(ctx, sbr, obj); err != nil {
			return false, err
		}
	}

	if len(fix) == 0 || ibx {
		if err := f.addInfoboxes(sbr, obj); err != nil {
			return false, err
		}
	}

	if len(fix) == 0 || scs {
		if err := f.addArticleParts(sbr, obj); err != nil {
			return false, err
		}
	}

	addAttribution := f.Env.EnableAttribution && (len(fix) == 0 || attr)

	var references []*parser.Reference
	if len(fix) == 0 || refs || addAttribution {
		var err error
		references, err = f.getReferences(obj)
		if err != nil {
			return false, err
		}
	}

	if len(fix) == 0 || refs {
		if err := f.addReferences(sbr, references); err != nil {
			return false, err
		}
	}

	if addAttribution {
		if err := f.addAttribution(ctx, sbr, obj, references); err != nil {
			return false, err
		}
	}

	*obj = obj.Get(fmt.Sprintf("{%s}", sbr.String()))
	return false, nil
}
