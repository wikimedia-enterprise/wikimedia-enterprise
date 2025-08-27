package schema

import (
	"encoding/json"
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type haspartsTestSuite struct {
	suite.Suite
	hasparts *AvroHasParts
}

func to_struct_ptr(obj []*HasParts) string {
	out, _ := json.Marshal(obj)

	return string(out)
}

func (s *haspartsTestSuite) SetupTest() {

	citation := &StructuredCitation{
		Identifier: "id",
		Group:      "group",
		Text:       "text",
	}

	tableRef := &StructuredTableRef{
		Identifier: "table-ref-id-001",
	}

	hasParts := &HasParts{
		Type:      "update2",
		Name:      "test2",
		Value:     "value2",
		HasParts:  []*HasParts{},
		Citations: []*StructuredCitation{citation},
	}

	hasParts2 := &HasParts{
		Type:      "update3",
		Name:      "test3",
		Value:     "value3",
		HasParts:  []*HasParts{hasParts},
		Citations: []*StructuredCitation{citation},
	}
	s.hasparts = &AvroHasParts{
		Type:     "update",
		Name:     "test",
		Value:    "value",
		Values:   []string{"val1"},
		HasParts: to_struct_ptr([]*HasParts{hasParts2}),
		Images: []*Image{{
			ContentUrl: "https://localhost",
			Width:      100,
			Height:     100,
			Thumbnail: &Thumbnail{
				ContentUrl: "https://localhost",
				Width:      10,
				Height:     10,
			},
			AlternativeText: "altText",
			Caption:         "Caption",
		}},
		Links: []*Link{{
			URL:  "url",
			Text: "text",
			Images: []*Image{{
				ContentUrl: "https://localhost",
				Width:      100,
				Height:     100,
				Thumbnail: &Thumbnail{
					ContentUrl: "https://localhost",
					Width:      10,
					Height:     10,
				},
				AlternativeText: "altText",
				Caption:         "Caption",
			}},
		}},
		Citations:       []*StructuredCitation{citation},
		TableReferences: []*StructuredTableRef{tableRef},
	}
}

func (s *haspartsTestSuite) TestHasPartsSchema() {
	sch, err := NewHasPartsSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.hasparts)
	s.Assert().NoError(err)

	hasparts := new(AvroHasParts)
	s.Assert().NoError(avro.Unmarshal(sch, data, hasparts))

	php, _ := hasparts.ToJsonStruct()
	ohp, _ := s.hasparts.ToJsonStruct()

	s.Assert().Equal(ohp.Type, php.Type)
	s.Assert().Equal(ohp.Name, php.Name)
	s.Assert().Equal(ohp.Value, php.Value)

	s.Assert().Equal(ohp.Values, php.Values)

	s.Assert().Equal(ohp.HasParts[0].Type, php.HasParts[0].Type)
	s.Assert().Equal(ohp.HasParts[0].Name, php.HasParts[0].Name)
	s.Assert().Equal(ohp.HasParts[0].Value, php.HasParts[0].Value)

	s.Assert().Equal(ohp.HasParts[0].Citations[0].Identifier, php.HasParts[0].Citations[0].Identifier)
	s.Assert().Equal(ohp.HasParts[0].Citations[0].Group, php.HasParts[0].Citations[0].Group)
	s.Assert().Equal(ohp.HasParts[0].Citations[0].Text, php.HasParts[0].Citations[0].Text)

	s.Assert().Equal(ohp.HasParts[0].HasParts[0].Type, php.HasParts[0].HasParts[0].Type)
	s.Assert().Equal(ohp.HasParts[0].HasParts[0].Name, php.HasParts[0].HasParts[0].Name)
	s.Assert().Equal(ohp.HasParts[0].HasParts[0].Value, php.HasParts[0].HasParts[0].Value)

	s.Assert().Equal(ohp.HasParts[0].HasParts[0].Citations[0].Identifier, php.HasParts[0].HasParts[0].Citations[0].Identifier)
	s.Assert().Equal(ohp.HasParts[0].HasParts[0].Citations[0].Group, php.HasParts[0].HasParts[0].Citations[0].Group)
	s.Assert().Equal(ohp.HasParts[0].HasParts[0].Citations[0].Text, php.HasParts[0].HasParts[0].Citations[0].Text)

	s.Assert().Equal(ohp.Images[0].ContentUrl, php.Images[0].ContentUrl)
	s.Assert().Equal(ohp.Images[0].Thumbnail, php.Images[0].Thumbnail)
	s.Assert().Equal(ohp.Images[0].Width, php.Images[0].Width)
	s.Assert().Equal(ohp.Images[0].Height, php.Images[0].Height)
	s.Assert().Equal(ohp.Images[0].AlternativeText, php.Images[0].AlternativeText)
	s.Assert().Equal(ohp.Images[0].Caption, php.Images[0].Caption)

	s.Assert().Equal(ohp.Links[0].URL, php.Links[0].URL)
	s.Assert().Equal(ohp.Links[0].Text, php.Links[0].Text)
	s.Assert().Equal(ohp.Links[0].Images[0].ContentUrl, php.Links[0].Images[0].ContentUrl)
	s.Assert().Equal(ohp.Links[0].Images[0].Thumbnail, php.Links[0].Images[0].Thumbnail)
	s.Assert().Equal(ohp.Links[0].Images[0].Width, php.Links[0].Images[0].Width)
	s.Assert().Equal(ohp.Links[0].Images[0].Height, php.Links[0].Images[0].Height)
	s.Assert().Equal(ohp.Links[0].Images[0].AlternativeText, php.Links[0].Images[0].AlternativeText)
	s.Assert().Equal(ohp.Links[0].Images[0].Caption, php.Links[0].Images[0].Caption)

	s.Assert().Equal(ohp.Citations[0].Identifier, php.Citations[0].Identifier)
	s.Assert().Equal(ohp.Citations[0].Group, php.Citations[0].Group)
	s.Assert().Equal(ohp.Citations[0].Text, php.Citations[0].Text)

	s.Assert().Equal(ohp.TableReferences[0].Identifier, php.TableReferences[0].Identifier)
}

func (s *haspartsTestSuite) TestAvroHasPartsSchema() {
	sch, err := NewHasPartsSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.hasparts)
	s.Assert().NoError(err)

	hasparts := new(AvroHasParts)
	s.Assert().NoError(avro.Unmarshal(sch, data, hasparts))

	p, _ := hasparts.ToJsonStruct()

	php, _ := p.ToAvroStruct()

	o, _ := s.hasparts.ToJsonStruct()

	ohp, _ := o.ToAvroStruct()

	s.Assert().Equal(ohp.Type, php.Type)
	s.Assert().Equal(ohp.Name, php.Name)
	s.Assert().Equal(ohp.Value, php.Value)

	s.Assert().Equal(ohp.Values, php.Values)

	s.Assert().Equal(ohp.HasParts, php.HasParts)

	s.Assert().Equal(ohp.Images[0].ContentUrl, php.Images[0].ContentUrl)
	s.Assert().Equal(ohp.Images[0].Thumbnail, php.Images[0].Thumbnail)
	s.Assert().Equal(ohp.Images[0].Width, php.Images[0].Width)
	s.Assert().Equal(ohp.Images[0].Height, php.Images[0].Height)
	s.Assert().Equal(ohp.Images[0].AlternativeText, php.Images[0].AlternativeText)
	s.Assert().Equal(ohp.Images[0].Caption, php.Images[0].Caption)

	s.Assert().Equal(ohp.Links[0].URL, php.Links[0].URL)
	s.Assert().Equal(ohp.Links[0].Text, php.Links[0].Text)
	s.Assert().Equal(ohp.Links[0].Images[0].ContentUrl, php.Links[0].Images[0].ContentUrl)
	s.Assert().Equal(ohp.Links[0].Images[0].Thumbnail, php.Links[0].Images[0].Thumbnail)
	s.Assert().Equal(ohp.Links[0].Images[0].Width, php.Links[0].Images[0].Width)
	s.Assert().Equal(ohp.Links[0].Images[0].Height, php.Links[0].Images[0].Height)
	s.Assert().Equal(ohp.Links[0].Images[0].AlternativeText, php.Links[0].Images[0].AlternativeText)
	s.Assert().Equal(ohp.Links[0].Images[0].Caption, php.Links[0].Images[0].Caption)

	s.Assert().Equal(ohp.Citations[0].Identifier, php.Citations[0].Identifier)
	s.Assert().Equal(ohp.Citations[0].Group, php.Citations[0].Group)
	s.Assert().Equal(ohp.Citations[0].Text, php.Citations[0].Text)

	s.Assert().Equal(ohp.TableReferences[0].Identifier, php.TableReferences[0].Identifier)
}
func TestHasParts(t *testing.T) {
	suite.Run(t, new(haspartsTestSuite))
}
