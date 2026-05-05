package aggregate_test

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"wikimedia-enterprise/services/structured-contents/libraries/aggregate"
	parser "wikimedia-enterprise/services/structured-contents/submodules/structured-content-parser"
	"wikimedia-enterprise/services/structured-contents/submodules/wmf"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type getterWmfMock struct {
	mock.Mock
	wmf.API
}

type getterParserMock struct {
	mock.Mock
	parser.API
}

func (m *getterWmfMock) GetPageSummary(_ context.Context, dtb string, ttl string, _ ...func(*url.Values)) (*wmf.PageSummary, error) {
	args := m.Called(dtb, ttl)
	return args.Get(0).(*wmf.PageSummary), args.Error(1)
}

func (m *getterParserMock) GetInfoBoxes(_ *goquery.Selection) []*parser.Part {
	args := m.Called()
	return args.Get(0).([]*parser.Part)
}

func (m *getterParserMock) GetSections(_ *goquery.Selection) []*parser.Part {
	args := m.Called()
	return args.Get(0).([]*parser.Part)
}

func (m *getterParserMock) GetReferences(_ *goquery.Selection) []*parser.Reference {
	args := m.Called()
	return args.Get(0).([]*parser.Reference)
}

type defaultGetterTestSuite struct {
	suite.Suite
	dgr *aggregate.DefaultGetter
	api wmf.API
	prs parser.API
}

func (s *defaultGetterTestSuite) SetupSuite() {
	s.dgr = new(aggregate.DefaultGetter)
}

func (s *defaultGetterTestSuite) TestSetAPIs() {
	s.dgr.SetAPIs(s.api, s.prs)
	s.Assert().Equal(s.api, s.dgr.API)
	s.Assert().Equal(s.prs, s.dgr.Parser)
}

func TestDefaultGetter(t *testing.T) {
	for _, testCase := range []*defaultGetterTestSuite{
		{
			api: wmf.NewAPI(),
			prs: parser.New(),
		},
		{
			api: nil,
			prs: nil,
		},
	} {
		suite.Run(t, testCase)
	}
}

type pageSummaryGetterTestSuite struct {
	suite.Suite
	ctx             context.Context
	psg             *aggregate.PageSummaryGetter
	wmc             *getterWmfMock
	dtb             string
	ttl             string
	html            string
	doc             *goquery.Document
	apiSummary      *wmf.PageSummary
	apiError        error
	expectedError   error
	expectedSummary string
}

func (s *pageSummaryGetterTestSuite) SetupSuite() {
	s.psg = &aggregate.PageSummaryGetter{}
	s.wmc = new(getterWmfMock)
	s.wmc.On("GetPageSummary", s.dtb, s.ttl).Return(s.apiSummary, s.apiError)
	s.psg.SetAPIs(s.wmc, nil)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(s.html))
	s.Require().NoError(err)
	s.doc = doc
}

func (s *pageSummaryGetterTestSuite) TestGetData() {
	data, err := s.psg.GetData(s.ctx, s.dtb, s.ttl, s.doc)
	if s.expectedError != nil {
		s.Assert().Contains(err.Error(), s.expectedError.Error())
	} else {
		s.Assert().NoError(err)
	}

	s.Assert().Equal(s.expectedSummary, data)
}

func (s *pageSummaryGetterTestSuite) TestGetField() {
	s.Assert().Equal(aggregate.PageSummaryFieldName, s.psg.GetField())
}

func TestPageSummaryGetter(t *testing.T) {
	for _, testCase := range []*pageSummaryGetterTestSuite{
		{
			dtb: "enwiki",
			ttl: "Earth",
			apiSummary: &wmf.PageSummary{
				Description: "Earth is the third planet from the Sun.",
			},
			expectedSummary: "Earth is the third planet from the Sun.",
			html:            `<html><head></head><body><div id="infobox">Infobox content</div></body></html>`,
		},
		{
			dtb:           "enwiki",
			ttl:           "Earth",
			apiError:      errors.New("page summary error"),
			expectedError: errors.New("page summary error"),
			html:          `<html><head></head><body><div id="infobox">Infobox content</div></body></html>`,
		},
		{
			dtb:             "enwiki",
			ttl:             "Earth",
			apiError:        errors.New("page summary error"),
			expectedSummary: "Short description",
			html:            `<html><head></head><body><span class="shortdescription">Short description</span></body></html>`,
		},
	} {
		suite.Run(t, testCase)
	}
}

type infoboxesGetterTestSuite struct {
	suite.Suite
	ctx context.Context
	ig  *aggregate.InfoboxesGetter
	prs *getterParserMock
	doc *goquery.Document
	inf []*parser.Part
	db  string
	ttl string
}

func (s *infoboxesGetterTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.ig = &aggregate.InfoboxesGetter{}
	s.prs = new(getterParserMock)
	s.prs.On("GetInfoBoxes", mock.Anything).Return(s.inf)
	s.ig.SetAPIs(nil, s.prs)

	html := `<html><head></head><body><div id="infobox">Infobox content</div></body></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	s.Require().NoError(err)
	s.doc = doc
}

func (s *infoboxesGetterTestSuite) TestGetData() {
	data, err := s.ig.GetData(s.ctx, s.db, s.ttl, s.doc)
	s.Assert().NoError(err)
	s.Assert().Equal(s.inf, data)
}

func (s *infoboxesGetterTestSuite) TestGetField() {
	s.Assert().Equal(aggregate.InfoboxesFieldName, s.ig.GetField())
}

func TestInfoboxesGetter(t *testing.T) {

	for _, testCase := range []*infoboxesGetterTestSuite{
		{
			inf: []*parser.Part{
				{Name: "Infobox-x"},
			},
			db:  "enwiki",
			ttl: "Earth",
		},
	} {
		suite.Run(t, testCase)
	}
}

type sectionsGetterTestSuite struct {
	suite.Suite
	ctx context.Context
	sg  *aggregate.SectionsGetter
	prs *getterParserMock
	doc *goquery.Document
	sec []*parser.Part
	db  string
	ttl string
}

func (s *sectionsGetterTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.sg = &aggregate.SectionsGetter{}
	s.prs = new(getterParserMock)
	s.prs.On("GetSections", mock.Anything).Return(s.sec)
	s.sg.SetAPIs(nil, s.prs)

	html := `<html><head></head><body><div id="infobox">Infobox content</div></body></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	s.Require().NoError(err)
	s.doc = doc
}

func (s *sectionsGetterTestSuite) TestGetData() {
	data, err := s.sg.GetData(s.ctx, s.db, s.ttl, s.doc)
	s.Assert().NoError(err)
	s.Assert().Equal(s.sec, data)
}

func (s *sectionsGetterTestSuite) TestGetField() {
	s.Assert().Equal(aggregate.SectionsFieldName, s.sg.GetField())
}

func TestSectionsGetter(t *testing.T) {
	for _, testCase := range []*sectionsGetterTestSuite{
		{
			db:  "enwiki",
			ttl: "Earth",
			sec: []*parser.Part{
				{Name: "Section 1"},
			},
		},
	} {
		suite.Run(t, testCase)
	}
}

type referencesGetterTestSuite struct {
	suite.Suite
	ctx context.Context
	sg  *aggregate.ReferencesGetter
	prs *getterParserMock
	doc *goquery.Document
	sec []*parser.Reference
	db  string
	ttl string
}

func (s *referencesGetterTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.sg = &aggregate.ReferencesGetter{}
	s.prs = new(getterParserMock)
	s.prs.On("GetReferences", mock.Anything).Return(s.sec)
	s.sg.SetAPIs(nil, s.prs)

	html := `<html><head></head><body><div id="infobox">Infobox content</div></body></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	s.Require().NoError(err)
	s.doc = doc
}

func (s *referencesGetterTestSuite) TestGetData() {
	data, err := s.sg.GetData(s.ctx, s.db, s.ttl, s.doc)
	s.Assert().NoError(err)
	s.Assert().Equal(s.sec, data)
}

func (s *referencesGetterTestSuite) TestGetField() {
	s.Assert().Equal(aggregate.ReferencesFieldName, s.sg.GetField())
}

func TestReferencesGetter(t *testing.T) {
	for _, testCase := range []*referencesGetterTestSuite{
		{
			db:  "enwiki",
			ttl: "Earth",
			sec: []*parser.Reference{
				{Identifier: "Section 1"},
			},
		},
	} {
		suite.Run(t, testCase)
	}
}
