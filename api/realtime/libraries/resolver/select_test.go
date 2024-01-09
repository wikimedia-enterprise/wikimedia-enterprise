package resolver_test

import (
	"testing"
	"wikimedia-enterprise/api/realtime/libraries/resolver"

	"github.com/stretchr/testify/suite"
)

type selectSubModelMock struct {
	Identifier int    `json:"identifier,omitempty" avro:"identifier"`
	InLanguage string `json:"in_language,omitempty" avro:"in_language"`
	Event      string `json:"event,omitempty" avro:"event"`
}

type selectModelMock struct {
	TableName        struct{}            `json:"-" ksql:"model"`
	Identifier       int                 `json:"identifier,omitempty" avro:"identifier"`
	ParentIdentifier int                 `json:"parent_identifier,omitempty" avro:"parent_identifier"`
	Comment          string              `json:"comment,omitempty" avro:"comment"`
	SubModel         *selectSubModelMock `json:"sub_model,omitempty" avro:"sub_model"`
}

type selectTestSuite struct {
	suite.Suite
	rvr         *resolver.Resolver
	fields      []string
	table       []string
	model       interface{}
	contains    []string
	notContains []string
}

func (s *selectTestSuite) SetupSuite() {
	var err error
	s.rvr, err = resolver.New(s.model, func(r *resolver.Resolver) {
		r.HasJoins = true
	})
	s.Assert().NoError(err)
}

func (s *selectTestSuite) TestNewSelect() {
	sel := resolver.NewSelect(s.rvr, s.fields, s.table...)
	s.Assert().NotNil(sel)
}

func (s *selectTestSuite) TestGetSql() {
	sql := resolver.NewSelect(s.rvr, s.fields, s.table...).GetSql()

	for _, statement := range s.contains {
		s.Assert().Contains(sql, statement)
	}

	for _, statement := range s.notContains {
		s.Assert().NotContains(sql, statement)
	}
}

func TestSelect(t *testing.T) {
	for _, testCase := range []*selectTestSuite{
		{
			model:  new(selectModelMock),
			table:  []string{"sub_model"},
			fields: []string{"sub_model.*"},
			contains: []string{
				"model_identifier,",
				"model_parent_identifier,",
				"model_comment,",
				"STRUCT(identifier := model_sub_model->identifier) as model_sub_model",
			},
			notContains: []string{
				"in_language := model_sub_model->in_language",
				"event := model_sub_model->event",
			},
		},
		{
			model:  new(selectModelMock),
			table:  []string{},
			fields: []string{"identifier", "sub_model.identifier"},
			contains: []string{
				"model_identifier, STRUCT(identifier := model_sub_model->identifier) as model_sub_model",
			},
			notContains: []string{
				"in_language := model_sub_model->in_language",
				"event := model_sub_model->event",
				"model_parent_identifier",
				"model_comment",
			},
		},
	} {
		suite.Run(t, testCase)
	}
}
