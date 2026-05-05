package aggregate_test

import (
	"context"
	"errors"
	"testing"
	"wikimedia-enterprise/services/structured-contents/libraries/aggregate"
	parser "wikimedia-enterprise/services/structured-contents/submodules/structured-content-parser"
	"wikimedia-enterprise/services/structured-contents/submodules/wmf"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type getterMock struct {
	mock.Mock
}

func (m *getterMock) SetAPIs(api wmf.API, prs parser.API) {
	m.Called(api, prs)
}

func (m *getterMock) GetData(_ context.Context, dtb, ttl string, doc *goquery.Document) (interface{}, error) {
	args := m.Called(dtb, ttl, doc)
	return args.Get(0), args.Error(1)
}

func (m *getterMock) GetField() string {
	args := m.Called()
	return args.String(0)
}

type aggregateTestSuite struct {
	suite.Suite
	ctx    context.Context
	agg    *aggregate.Aggregate
	wmf    wmf.API
	parser parser.API
	gtr    *getterMock
	dtb    string
	ttl    string
	doc    *goquery.Document
	val    *aggregate.Aggregation
	rsp    interface{}
	err    error
}

func (s *aggregateTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.wmf = wmf.NewAPI()
	s.parser = parser.New()
	s.gtr = new(getterMock)

	aggtr := aggregate.New(s.wmf, s.parser)
	s.Assert().NotNil(aggtr)

	s.agg = aggtr.(*aggregate.Aggregate)
	s.val = new(aggregate.Aggregation)

	s.gtr.On("SetAPIs", s.wmf, s.parser).Return()
}

func (s *aggregateTestSuite) Test() {
	s.gtr.On("GetData", s.dtb, s.ttl, s.doc).Return(s.rsp, s.err)
	s.gtr.On("GetField").Return("Infoboxes")

	err := s.agg.GetAggregation(s.ctx, s.dtb, s.ttl, s.doc, s.val, s.gtr)

	if s.err != nil {
		s.Assert().Equal(s.err, err)
		s.Assert().Empty(s.val.PageSummary)
		s.Assert().Nil(s.val.Infoboxes)
		s.Assert().Nil(s.val.Sections)
	} else {
		s.Assert().NoError(err)
		s.Assert().NotNil(s.val.Infoboxes)
	}
}

func TestPageSummaryError(t *testing.T) {
	ctx := context.Background()
	api := wmf.NewAPI()
	parser := parser.New()
	gtr := new(getterMock)
	dtb := "enwiki"
	ttl := "Earth"
	doc := new(goquery.Document)
	rsp := &wmf.PageSummary{}

	aggtr := aggregate.New(api, parser)
	assert.NotNil(t, aggtr)

	agg := aggtr.(*aggregate.Aggregate)
	val := new(aggregate.Aggregation)

	gtr.On("SetAPIs", api, parser).Return()

	gtr.On("GetData", dtb, ttl, doc).Return(rsp, errors.New("test error"))
	gtr.On("GetField").Return("PageSummary")

	err := agg.GetAggregation(ctx, dtb, ttl, doc, val, gtr)

	assert.NoError(t, err)
	assert.Nil(t, val.Infoboxes)
}

func TestAggregate(t *testing.T) {
	for _, testCase := range []*aggregateTestSuite{
		{
			dtb: "enwiki",
			ttl: "Earth",
			doc: new(goquery.Document),
			rsp: []*parser.Part{{
				Name: "part",
			}},
		},
		{
			dtb: "enwiki",
			ttl: "Earth",
			rsp: []*parser.Part{},
			err: errors.New("test error"),
		},
	} {
		suite.Run(t, testCase)
	}
}

type aggregationTestSuite struct {
	suite.Suite
	val *aggregate.Aggregation
}

func (s *aggregationTestSuite) SetupTest() {
	s.val = &aggregate.Aggregation{
		PageSummary: "something",
		Infoboxes:   []*parser.Part{},
		Sections:    []*parser.Part{},
	}
}

func (s *aggregationTestSuite) TestGetInfoboxes() {
	s.Assert().Equal(s.val.Infoboxes, s.val.GetInfoboxes())
}

func (s *aggregationTestSuite) TestGetSections() {
	s.Assert().Equal(s.val.Sections, s.val.GetSections())
}

func (s *aggregationTestSuite) TestGetPageSummary() {
	s.Assert().Equal(s.val.PageSummary, s.val.GetPageSummary())
}

func TestAggregation(t *testing.T) {
	suite.Run(t, new(aggregationTestSuite))
}
