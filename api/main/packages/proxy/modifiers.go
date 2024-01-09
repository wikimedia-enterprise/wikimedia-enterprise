package proxy

import (
	"fmt"
	"strings"
	"wikimedia-enterprise/general/log"

	"github.com/tidwall/gjson"
)

// InfoboxFieldName name of the field that contains infobox of the page.
const InfoboxFieldName = "infobox"

// DescriptionFieldName name of the field that contains wikipedia/wikidata description.
const DescriptionFieldName = "description"

// SectionsFieldName name of the field that contains wikipedia sections.
const SectionsFieldName = "article_sections"

// DefaultModifiers list of default modifiers.
var DefaultModifiers = []Modifier{
	new(FilterModifier),
	new(FieldModifier),
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
}

// FilterModifier modifier that help to apply filters to the entities.
type FilterModifier struct{}

// Modify method that apply filters to the entities.
func (f *FilterModifier) Modify(mdl *Model, obj *gjson.Result) (bool, error) {
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
			if _, err := sbr.WriteString(fmt.Sprintf(`"%s":{`, nme)); err != nil {
				log.Error(err)
			}

			f.BuildQuery(val, sbr)

			if _, err := sbr.WriteString("}"); err != nil {
				log.Error(err)
			}
		case string:
			if _, err := sbr.WriteString(fmt.Sprintf(`"%s":%s`, nme, val)); err != nil {
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
func (f *FieldModifier) Modify(mdl *Model, obj *gjson.Result) (bool, error) {
	sbr := new(strings.Builder)

	f.BuildQuery(mdl.GetFieldsIndex(), sbr)

	if sbr.Len() > 0 {
		stm := fmt.Sprintf("{%s}", sbr.String())
		*obj = obj.Get(stm)
	}

	return false, nil
}
