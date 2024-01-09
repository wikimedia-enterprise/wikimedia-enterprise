package proxy_test

import (
	"encoding/json"
	"strings"
	"testing"
	"wikimedia-enterprise/api/main/packages/proxy"

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
	brk, err := s.mdr.Modify(s.mdl, &s.obj)

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
	brk, err := s.mdr.Modify(s.mdl, &s.obj)

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
