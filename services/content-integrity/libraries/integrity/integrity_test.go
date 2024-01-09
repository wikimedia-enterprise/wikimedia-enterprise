package integrity_test

import (
	"context"
	"errors"
	"testing"
	"time"
	"wikimedia-enterprise/services/content-integrity/libraries/collector"
	"wikimedia-enterprise/services/content-integrity/libraries/integrity"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type articleTestSuite struct {
	suite.Suite
	art *integrity.Article
}

func (s *articleTestSuite) SetupTest() {
	s.art = new(integrity.Article)
}

func (s *articleTestSuite) TestSetIdentifier() {
	s.art.SetIdentifier(1)

	s.Equal(1, s.art.Identifier)
}

func (s *articleTestSuite) TestSetProject() {
	s.art.SetProject("test")

	s.Equal("test", s.art.Project)
}

func (a *articleTestSuite) TestSetVersionIdentifier() {
	a.art.SetVersionIdentifier(1)

	a.Equal(1, a.art.VersionIdentifier)
}

func (a *articleTestSuite) TestSetEditsCount() {
	a.art.SetEditsCount(1)

	a.Equal(1, a.art.EditsCount)
}

func (a *articleTestSuite) TestSetDateCreated() {
	dct := new(time.Time)
	a.art.SetDateCreated(dct)

	a.Equal(dct, a.art.DateCreated)
}

func (a *articleTestSuite) TestSetDateNamespaceMoved() {
	dnm := new(time.Time)
	a.art.SetDateNamespaceMoved(dnm)

	a.Equal(dnm, a.art.DateNamespaceMoved)
}

func (a *articleTestSuite) TestSetUniqueEditorsCount() {
	a.art.SetUniqueEditorsCount(1)

	a.Equal(1, a.art.UniqueEditorsCount)
}

func (a *articleTestSuite) TestSetIsBreakingNews() {
	a.art.SetIsBreakingNews(true)

	a.True(a.art.IsBreakingNews)
}

func (a *articleTestSuite) TestSetVersions() {
	vrs := make(collector.Versions, 0)
	a.art.SetVersions(vrs)

	a.Equal(vrs, a.art.Versions)
}

func (a *articleTestSuite) TestGetIdentifier() {
	a.art.Identifier = 1

	a.Equal(1, a.art.GetIdentifier())
}

func (a *articleTestSuite) TestGetProject() {
	a.art.Project = "test"

	a.Equal("test", a.art.GetProject())
}

func (a *articleTestSuite) TestGetVersionIdentifier() {
	a.art.VersionIdentifier = 1

	a.Equal(1, a.art.GetVersionIdentifier())
}

func (a *articleTestSuite) TestGetEditsCount() {
	a.art.EditsCount = 1

	a.Equal(1, a.art.GetEditsCount())
}

func (a *articleTestSuite) TestGetDateCreated() {
	dct := new(time.Time)
	a.art.DateCreated = dct

	a.Equal(dct, a.art.GetDateCreated())
}

func (a *articleTestSuite) TestGetDateNamespaceMoved() {
	dnm := new(time.Time)
	a.art.DateNamespaceMoved = dnm

	a.Equal(dnm, a.art.GetDateNamespaceMoved())
}

func (a *articleTestSuite) TestGetUniqueEditorsCount() {
	a.art.UniqueEditorsCount = 1

	a.Equal(1, a.art.GetUniqueEditorsCount())
}

func (a *articleTestSuite) TestGetIsBreakingNews() {
	a.art.IsBreakingNews = true

	a.True(a.art.GetIsBreakingNews())
}

func (a *articleTestSuite) TestGetVersions() {
	vrs := make(collector.Versions, 0)
	a.art.Versions = vrs

	a.Equal(vrs, a.art.GetVersions())
}

func TestArticle(t *testing.T) {
	suite.Run(t, new(articleTestSuite))
}

type articleParamsTestSuite struct {
	suite.Suite
	aps *integrity.ArticleParams
}

func (s *articleParamsTestSuite) SetupTest() {
	s.aps = new(integrity.ArticleParams)
}

func (s *articleParamsTestSuite) TestGetProject() {
	s.aps.Project = "test"

	s.Equal("test", s.aps.GetProject())
}

func (s *articleParamsTestSuite) TestGetIdentifier() {
	s.aps.Identifier = 1

	s.Equal(1, s.aps.GetIdentifier())
}

func (s *articleParamsTestSuite) TestGetName() {
	s.aps.Project = "test"

	s.Equal("test", s.aps.GetProject())
}

func (s *articleParamsTestSuite) TestGetVersionIdentifier() {
	s.aps.VersionIdentifier = 1

	s.Equal(1, s.aps.GetVersionIdentifier())
}

func (s *articleParamsTestSuite) TestGetTemplates() {
	s.aps.Templates = []string{"test"}

	s.Equal([]string{"test"}, s.aps.GetTemplates())
}

func (s *articleParamsTestSuite) TestGetMatchingTemplates() {
	s.aps.Templates = []string{"test", "test2", "test3", "test4"}

	s.Equal([]string{"test", "test2"}, s.aps.GetMatchingTemplates("test", "test2"))
}

func (s *articleParamsTestSuite) TestGetTemplatesByPrefix() {
	s.aps.Templates = []string{"test", "test2", "another1", "another2"}

	s.Equal([]string{"test", "test2"}, s.aps.GetTemplatesByPrefix("test"))
}

func TestArticleParams(t *testing.T) {
	suite.Run(t, new(articleParamsTestSuite))
}

type collectorMock struct {
	mock.Mock
	collector.ArticleAPI
}

func (m *collectorMock) GetIsBeingTracked(_ context.Context, prj string, idr int) (bool, error) {
	ags := m.Called(prj, idr)
	return ags.Bool(0), ags.Error(1)
}

func (m *collectorMock) GetDateCreated(_ context.Context, prj string, idr int) (*time.Time, error) {
	ags := m.Called(prj, idr)

	if tme := ags.Get(0); tme != nil {
		return tme.(*time.Time), ags.Error(1)
	}

	return nil, ags.Error(1)
}

func (m *collectorMock) GetDateNamespaceMoved(ctx context.Context, prj string, idr int) (*time.Time, error) {
	ags := m.Called(prj, idr)

	if tme := ags.Get(0); tme != nil {
		return tme.(*time.Time), ags.Error(1)
	}

	return nil, ags.Error(1)
}

func (m *collectorMock) GetVersions(ctx context.Context, prj string, idr int) (collector.Versions, error) {
	ags := m.Called(prj, idr)

	if vrs := ags.Get(0); vrs != nil {
		return vrs.(collector.Versions), ags.Error(1)
	}

	return nil, ags.Error(1)
}

func (m *collectorMock) GetName(ctx context.Context, prj string, idr int) (string, error) {
	ags := m.Called(prj, idr)
	return ags.String(0), ags.Error(1)
}

type checkerMock struct {
	mock.Mock
}

func (m *checkerMock) CheckArticle(ctx context.Context, art *integrity.Article, aps *integrity.ArticleParams) error {
	ags := m.Called(aps)
	return ags.Error(0)
}

type newTestSuite struct {
	suite.Suite
}

func (s *newTestSuite) TestNew() {
	s.NotNil(integrity.New(new(collectorMock)))
}

func TestNew(t *testing.T) {
	suite.Run(t, new(newTestSuite))
}

type integrityTestSuite struct {
	suite.Suite
	ctx context.Context
	igt *integrity.Integrity
	clm *collectorMock
	ckm *checkerMock
	aps *integrity.ArticleParams
	err error
}

func (s *integrityTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.clm = new(collectorMock)
	s.ckm = new(checkerMock)
	s.err = errors.New("test")
	s.aps = &integrity.ArticleParams{
		Project:           "test",
		Identifier:        100,
		VersionIdentifier: 10,
	}
	s.igt = new(integrity.Integrity)
	s.igt.Collector = s.clm
}

func (s *integrityTestSuite) TestGetArticleIsBeingTrackedError() {
	s.clm.On("GetIsBeingTracked", s.aps.GetProject(), s.aps.GetIdentifier()).Return(false, s.err)

	art, err := s.igt.GetArticle(s.ctx, s.aps, s.ckm)
	s.Nil(art)
	s.Equal(s.err, err)
}

func (s *integrityTestSuite) TestGetArticleIsBeingTrackedFalse() {
	s.clm.On("GetIsBeingTracked", s.aps.GetProject(), s.aps.GetIdentifier()).Return(false, nil)

	art, err := s.igt.GetArticle(s.ctx, s.aps, s.ckm)
	s.Nil(art)
	s.NoError(err)
}

func (s *integrityTestSuite) TestGetArticleDateCreatedError() {
	s.clm.On("GetIsBeingTracked", s.aps.GetProject(), s.aps.GetIdentifier()).Return(true, nil)
	s.clm.On("GetDateCreated", s.aps.GetProject(), s.aps.GetIdentifier()).Return(nil, s.err)

	art, err := s.igt.GetArticle(s.ctx, s.aps, s.ckm)
	s.Nil(art)
	s.Equal(s.err, err)
}

func (s *integrityTestSuite) TestGetArticleDateNamespaceMovedError() {
	s.clm.On("GetIsBeingTracked", s.aps.GetProject(), s.aps.GetIdentifier()).Return(true, nil)
	s.clm.On("GetDateCreated", s.aps.GetProject(), s.aps.GetIdentifier()).Return(new(time.Time), nil)
	s.clm.On("GetDateNamespaceMoved", s.aps.GetProject(), s.aps.GetIdentifier()).Return(nil, s.err)

	art, err := s.igt.GetArticle(s.ctx, s.aps, s.ckm)
	s.Nil(art)
	s.Equal(s.err, err)
}

func (s *integrityTestSuite) TestGetArticleVersionsError() {
	s.clm.On("GetIsBeingTracked", s.aps.GetProject(), s.aps.GetIdentifier()).Return(true, nil)
	s.clm.On("GetDateCreated", s.aps.GetProject(), s.aps.GetIdentifier()).Return(new(time.Time), nil)
	s.clm.On("GetDateNamespaceMoved", s.aps.GetProject(), s.aps.GetIdentifier()).Return(new(time.Time), nil)
	s.clm.On("GetVersions", s.aps.GetProject(), s.aps.GetIdentifier()).Return(nil, s.err)

	art, err := s.igt.GetArticle(s.ctx, s.aps, s.ckm)
	s.Nil(art)
	s.Equal(s.err, err)
}

func (s *integrityTestSuite) TestGetArticleCheckArticleError() {
	s.clm.On("GetIsBeingTracked", s.aps.GetProject(), s.aps.GetIdentifier()).Return(true, nil)
	s.clm.On("GetDateCreated", s.aps.GetProject(), s.aps.GetIdentifier()).Return(new(time.Time), nil)
	s.clm.On("GetDateNamespaceMoved", s.aps.GetProject(), s.aps.GetIdentifier()).Return(new(time.Time), nil)
	s.clm.On("GetVersions", s.aps.GetProject(), s.aps.GetIdentifier()).Return(collector.Versions{}, nil)
	s.clm.On("GetName", s.aps.GetProject(), s.aps.GetIdentifier()).Return("test", nil)
	s.ckm.On("CheckArticle", s.aps).Return(s.err)

	art, err := s.igt.GetArticle(s.ctx, s.aps, s.ckm)
	s.Nil(art)
	s.Equal(s.err, err)
}

func (s *integrityTestSuite) TestGetArticle() {
	s.clm.On("GetIsBeingTracked", s.aps.GetProject(), s.aps.GetIdentifier()).Return(true, nil)
	s.clm.On("GetDateCreated", s.aps.GetProject(), s.aps.GetIdentifier()).Return(new(time.Time), nil)
	s.clm.On("GetDateNamespaceMoved", s.aps.GetProject(), s.aps.GetIdentifier()).Return(new(time.Time), nil)
	s.clm.On("GetVersions", s.aps.GetProject(), s.aps.GetIdentifier()).Return(collector.Versions{}, nil)
	s.clm.On("GetName", s.aps.GetProject(), s.aps.GetIdentifier()).Return("test", nil)
	s.ckm.On("CheckArticle", s.aps).Return(nil)

	art, err := s.igt.GetArticle(s.ctx, s.aps, s.ckm)
	s.NotNil(art)
	s.NoError(err)
}

func TestIntegrity(t *testing.T) {
	suite.Run(t, new(integrityTestSuite))
}
