package resolver_test

import (
	"testing"
	"wikimedia-enterprise/api/realtime/libraries/resolver"
	"wikimedia-enterprise/api/realtime/submodules/schema"

	"github.com/stretchr/testify/suite"
)

type filterTestSuite struct {
	suite.Suite
	res     *resolver.Resolver
	filters []resolver.RequestFilter
	art     *schema.Article
	result  bool
}

func (s *filterTestSuite) SetupTest() {
	var err error
	s.res, err = resolver.New(new(schema.Article))

	s.Assert().NoError(err)
	s.Assert().NotNil(s.res)
}

func (s *filterTestSuite) TestNewValueFilter() {
	s.Assert().NotNil(s.res.NewValueFilter(s.filters))
}

func (s *filterTestSuite) TestShouldKeep() {
	vf := s.res.NewValueFilter(s.filters)

	s.Equal(s.result, vf(s.art))
}

func TestFilter(t *testing.T) {
	for _, testCase := range []*filterTestSuite{
		{
			filters: []resolver.RequestFilter{
				{Field: "is_part_of.identifier", Value: "enwiki"},
				{Field: "event.type", Value: "update"},
				{Field: "version.noindex", Value: false},
				{Field: "namespace.identifier", Value: 0},
			},
			art: &schema.Article{
				IsPartOf:  &schema.Project{Identifier: "enwiki"},
				Event:     &schema.Event{Type: "update"},
				Version:   &schema.Version{Noindex: false},
				Namespace: &schema.Namespace{Identifier: 0},
			},
			result: true,
		},
		{
			filters: []resolver.RequestFilter{
				{Field: "is_part_of.identifier", Value: "enwiki"},
				{Field: "event.type", Value: "update"},
				{Field: "version.noindex", Value: false},
				{Field: "namespace.identifier", Value: 0},
			},
			art: &schema.Article{
				IsPartOf:  &schema.Project{Identifier: "frwiki"},
				Event:     &schema.Event{Type: "update"},
				Version:   &schema.Version{Noindex: false},
				Namespace: &schema.Namespace{Identifier: 0},
			},
			result: false,
		},
		{
			filters: []resolver.RequestFilter{
				{Field: "is_part_of.identifier", Value: "enwiki"},
				{Field: "event.type", Value: "update"},
				{Field: "version.noindex", Value: false},
				{Field: "namespace.identifier", Value: 0},
			},
			art: &schema.Article{
				IsPartOf:  &schema.Project{Identifier: "frwiki"},
				Event:     &schema.Event{Type: "delete"},
				Version:   &schema.Version{Noindex: true},
				Namespace: &schema.Namespace{Identifier: 6},
			},
			result: false,
		},
		{
			filters: []resolver.RequestFilter{
				{Field: "is_part_of.identifier", Value: "enwiki"},
				{Field: "event.type", Value: "update"},
				{Field: "version.noindex", Value: false},
				{Field: "namespace.identifier", Value: 0},
			},
			art: &schema.Article{
				IsPartOf: &schema.Project{Identifier: "enwiki"},
				Event:    &schema.Event{Type: "update"},
				Version:  &schema.Version{Noindex: false},
			},
			result: false,
		},
		{
			filters: []resolver.RequestFilter{},
			art: &schema.Article{
				IsPartOf: &schema.Project{Identifier: "enwiki"},
				Event:    &schema.Event{Type: "update"},
				Version:  &schema.Version{Noindex: false},
			},
			result: true,
		},
		{
			filters: []resolver.RequestFilter{
				{Field: "is_part_of.identifier", Value: "enwiki"},
				{Field: "event.type", Value: "update"},
				{Field: "version.noindex", Value: false},
				{Field: "namespace.identifier", Value: 0},
			},
			art:    &schema.Article{},
			result: false,
		},
	} {
		suite.Run(t, testCase)
	}
}
