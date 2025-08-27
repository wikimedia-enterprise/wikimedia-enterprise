package proxy_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"wikimedia-enterprise/api/main/config/env"
	"wikimedia-enterprise/api/main/packages/proxy"
	parser "wikimedia-enterprise/api/main/submodules/structured-contents-parser"
	"wikimedia-enterprise/api/main/submodules/wmf"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
)

type filterModifierTestSuite struct {
	suite.Suite
	mdr *proxy.FilterModifier
	mdl *proxy.Model
	obj gjson.Result
	jsn string
	brk bool
}

func (s *filterModifierTestSuite) SetupTest() {
	s.mdl.BuildFieldsIndex()

	s.mdr = new(proxy.FilterModifier)
	s.obj = gjson.Parse(s.jsn)
}

func (s *filterModifierTestSuite) TestModifyError() {
	brk, err := s.mdr.Modify(context.Background(), s.mdl, &s.obj)

	s.Equal(s.brk, brk)
	s.NoError(err)
}

func TestFilterModifier(t *testing.T) {
	jsn := `{
		"in_language": {
			"identifier": "en"
		}
	}`

	for _, testCase := range []*filterModifierTestSuite{
		{
			mdl: &proxy.Model{
				Filters: []*proxy.Filter{
					{
						Field: "in_language.identifier",
						Value: "en",
					},
				},
			},
			jsn: jsn,
			brk: false,
		},
		{
			mdl: &proxy.Model{
				Filters: []*proxy.Filter{
					{
						Field: "in_language.identifier",
						Value: "fr",
					},
				},
			},
			jsn: jsn,
			brk: true,
		},
		{
			mdl: new(proxy.Model),
			jsn: jsn,
			brk: false,
		},
	} {
		suite.Run(t, testCase)
	}
}

type fieldModifierTestSuite struct {
	suite.Suite
	mdr *proxy.FieldModifier
	mdl *proxy.Model
	obj gjson.Result
	jsn string
	rjn string
}

func (s *fieldModifierTestSuite) SetupTest() {
	s.mdl.BuildFieldsIndex()

	s.mdr = new(proxy.FieldModifier)
	s.obj = gjson.Parse(s.jsn)
}

func (s *fieldModifierTestSuite) TestModify() {
	brk, err := s.mdr.Modify(context.Background(), s.mdl, &s.obj)

	s.False(brk)
	s.NoError(err)

	rjm := map[string]interface{}{}
	s.NoError(json.Unmarshal([]byte(s.rjn), &rjm))

	obm := map[string]interface{}{}
	s.NoError(json.Unmarshal([]byte(s.obj.Raw), &obm))

	s.Assert().Equal(rjm, obm)
}

func TestFieldsModifier(t *testing.T) {
	jsn := `{"identifier":100,"name":"Earth","version":{"identifier":999,"editor":{"identifier":1000,"name":"John Doe"}}}`

	for _, testCase := range []*fieldModifierTestSuite{
		{
			mdl: &proxy.Model{
				Fields: []string{
					"name",
				},
			},
			jsn: jsn,
			rjn: `{"name":"Earth"}`,
		},
		{
			mdl: &proxy.Model{
				Fields: []string{
					"version.identifier",
				},
			},
			jsn: jsn,
			rjn: `{"version":{"identifier":999}}`,
		},
		{
			mdl: &proxy.Model{
				Fields: []string{
					"version.editor",
					"version.editor.name",
					"version.editor.identifier",
				},
			},
			jsn: jsn,
			rjn: `{"version":{"editor":{"name":"John Doe","identifier":1000}}}`,
		},
		{
			mdl: &proxy.Model{
				Fields: []string{
					"version.*",
				},
			},
			jsn: jsn,
			rjn: `{"version":{"identifier":999,"editor":{"identifier":1000,"name":"John Doe"}}}`,
		},
	} {
		suite.Run(t, testCase)
	}
}

type baseModifierTestSuite struct {
	suite.Suite
	idx map[string]interface{}
	qry []string
	mdr *proxy.BaseFieldModifier
}

func (s *baseModifierTestSuite) TestBuildQuery() {
	sbr := new(strings.Builder)

	s.mdr.BuildQuery(s.idx, sbr)

	for _, fld := range s.qry {
		s.Contains(sbr.String(), fld)
	}
}

func TestBaseModifier(t *testing.T) {
	for _, testCase := range []*baseModifierTestSuite{
		{
			idx: map[string]interface{}{
				"identifier": "identifier",
				"in_language": map[string]interface{}{
					"identifier": "in_language.identifier",
				},
			},
			qry: []string{
				`"identifier":identifier`,
				`"in_language":{"identifier":in_language.identifier}`,
			},
		},
		{
			idx: map[string]interface{}{
				"identifier": "identifier",
				"version": map[string]interface{}{
					"editor": map[string]interface{}{
						"identifier": "version.editor.identifier",
					},
				},
			},
			qry: []string{
				`"identifier":identifier`,
				`"version":{"editor":{"identifier":version.editor.identifier}}`,
			},
		},
	} {
		suite.Run(t, testCase)
	}
}

type parserMock struct {
	mock.Mock
	parser.API
}

func (m *parserMock) GetInfoBoxes(sel *goquery.Selection) []*parser.Part {
	return m.Called(sel).Get(0).([]*parser.Part)
}

func (m *parserMock) GetSections(sel *goquery.Selection) []*parser.Part {
	return m.Called(sel).Get(0).([]*parser.Part)
}

func (m *parserMock) GetReferences(sel *goquery.Selection) []*parser.Reference {
	return m.Called(sel).Get(0).([]*parser.Reference)
}

// GetArticleParts mocks the parser's method to extract sections and tables from a goquery Selection.
func (m *parserMock) GetArticleParts(sel *goquery.Selection) ([]*parser.Part, []*parser.OrderedTable) {
	args := m.Called(sel)
	return args.Get(0).([]*parser.Part), args.Get(1).([]*parser.OrderedTable)
}

type sockModifierTestSuite struct {
	suite.Suite
	obj    gjson.Result
	mdl    *proxy.Model
	pmc    *parserMock
	smr    *proxy.SOCKModifier
	wmf    *wmf.APIFake
	ibs    []*parser.Part
	scs    []*parser.Part
	tables parser.Tables
	refs   []*parser.Reference
	dsc    string
}

func (s *sockModifierTestSuite) SetupTest() {
	s.obj = gjson.Parse(`{"identifier": 500, "date_modified": "2025-06-27T06:50:35Z"}`)
	s.ibs = []*parser.Part{
		{
			Name: `"Earth"`,
		},
	}
	s.scs = []*parser.Part{
		{
			Name: `Early "life"`,
		},
	}
	s.tables = parser.Tables{
		"Table 1": &parser.TableWithHeaders{
			Headers: [][]parser.Cell{
				{
					{Value: "Header1"},
					{Value: "Header2"},
				},
			},
			Rows: [][]parser.Cell{
				{
					{Value: "Row1Col1"},
					{Value: "Row1Col2"},
				},
			},
			ConfidenceScore: 0.9, // Optional, you can omit or set to any float64
		},
	}
	s.refs = []*parser.Reference{
		{Identifier: "id"},
	}
	s.dsc = `little "blue" planet`
	s.pmc = new(parserMock)
	s.wmf = wmf.NewAPIFake()
	s.smr = &proxy.SOCKModifier{
		Parser: s.pmc,
		WMF:    s.wmf,
		Env: &env.Environment{
			DescriptionEnabled: true,
			SectionsEnabled:    true,
			EnableAttribution:  true,
			EnableTables:       true,
		},
	}
}

func (s *sockModifierTestSuite) TestModify() {
	s.pmc.On("GetInfoBoxes", mock.Anything).Return(s.ibs)
	s.pmc.On("GetArticleParts", mock.Anything).Return(s.scs, []*parser.OrderedTable{
		{
			ID: "Table 1",
			Table: &parser.TableWithHeaders{
				Headers: [][]parser.Cell{},
				Rows:    [][]parser.Cell{},
			},
		},
	})
	s.pmc.On("GetReferences", mock.Anything).Return(s.refs)

	s.wmf.InsertPage(&wmf.FakePage{PageID: 500, Description: s.dsc, RegisteredContributors: 600, AnonymousContributors: 20})

	s.mdl.BuildFieldsIndex()

	brk, err := s.smr.Modify(context.Background(), s.mdl, &s.obj)

	s.Contains(s.obj.Raw, `"name":"\"Earth\""`)
	s.Contains(s.obj.Raw, `"infoboxes":[{"name":"\"Earth\""}]`)
	s.Contains(s.obj.Raw, `"sections":[{"name":"Early \"life\""}]`)
	s.Contains(s.obj.Raw, `"description":"little \"blue\" planet"`)
	// Note: \u003e100 = >100
	s.Contains(s.obj.Raw, `"attribution":{"editors_count":{"anonymous_count":"20","registered_count":"\u003e100"},"parsed_references_count":1,"last_revised":"06/25"}`)
	s.Contains(s.obj.Raw, `"Table 1"`)
	s.False(brk)
	s.NoError(err)
}

func TestSockModifier(t *testing.T) {
	for _, testCase := range []*sockModifierTestSuite{
		{
			mdl: new(proxy.Model),
		},
		{
			mdl: &proxy.Model{
				Fields: []string{
					"name",
					"infoboxes",
					"description",
					"sections",
					"tables",
					"attribution",
				},
			},
		},
	} {
		suite.Run(t, testCase)
	}
}

type attributionTestSuite struct {
	suite.Suite

	anon       int
	registered int

	expectedAnon       string
	expectedRegistered string
}

func (s *attributionTestSuite) TestAttributionCase() {
	api := wmf.NewAPIFake()
	mod := &proxy.SOCKModifier{
		WMF: api,
	}
	api.InsertPage(&wmf.FakePage{PageID: 0, AnonymousContributors: s.anon, RegisteredContributors: s.registered})

	res := gjson.Parse("{}")
	result, err := mod.GetAttribution(context.Background(), &res, nil)
	s.NoError(err)
	s.Equal(s.expectedAnon, result.EditorsCount.AnonymousCount)
	s.Equal(s.expectedRegistered, result.EditorsCount.RegisteredCount)
}

func TestAttribution(t *testing.T) {
	for _, testCase := range []*attributionTestSuite{
		{
			expectedAnon:       "0",
			expectedRegistered: "0",
		},
		{
			anon:               10,
			registered:         10,
			expectedAnon:       "10",
			expectedRegistered: "10",
		},
		{
			anon:               100,
			registered:         100,
			expectedAnon:       "100",
			expectedRegistered: "100",
		},
		{
			anon:               200,
			registered:         200,
			expectedAnon:       ">100",
			expectedRegistered: ">100",
		},
		// Technically a different case, but in practice it has the same result.
		{
			anon:               200,
			registered:         501,
			expectedAnon:       ">100",
			expectedRegistered: ">100",
		},
	} {
		suite.Run(t, testCase)
	}
}
