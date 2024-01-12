package proxy_test

import (
	"testing"
	"wikimedia-enterprise/api/main/packages/proxy"
	"wikimedia-enterprise/general/httputil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type newEntitiesGetterTestSuite struct {
	suite.Suite
	url string
}

func (s *newEntitiesGetterTestSuite) TestNewEntitiesGetter() {
	gtr := proxy.NewEntitiesGetter(s.url)

	s.Assert().NotNil(gtr)
	s.Assert().Equal(s.url, gtr.URL)
}

func TestNewEntitiesGetter(t *testing.T) {
	for _, testCase := range []*newEntitiesGetterTestSuite{
		{
			url: "entity",
		},
	} {
		suite.Run(t, testCase)
	}
}

type entitiesGetterTestSuite struct {
	suite.Suite
	url string
	gtr *proxy.EntitiesGetter
}

func (s *entitiesGetterTestSuite) SetupSuite() {
	s.gtr = &proxy.EntitiesGetter{
		URL: s.url,
	}
}

func (s *entitiesGetterTestSuite) TestGetPath() {
	pth, err := s.gtr.GetPath(new(gin.Context))

	s.Assert().NotEmpty(pth)
	s.Assert().NoError(err)
}

func TestEntitiesGetter(t *testing.T) {
	for _, testCase := range []*entitiesGetterTestSuite{
		{
			url: "entity",
		},
	} {
		suite.Run(t, testCase)
	}
}

type newEntityGetterTestSuite struct {
	suite.Suite
	url string
}

func (s *newEntityGetterTestSuite) TestNewEntityGetter() {
	gtr := proxy.NewEntityGetter(s.url)

	s.Assert().NotNil(gtr)
	s.Assert().Equal(s.url, gtr.URL)
}

func TestNewEntityGetter(t *testing.T) {
	for _, testCase := range []*newEntityGetterTestSuite{
		{
			url: "entity",
		},
	} {
		suite.Run(t, testCase)
	}
}

type entityGetterTestSuite struct {
	suite.Suite
	gcx *gin.Context
	gtr *proxy.EntityGetter
	url string
	idn string
	err error
}

func (s *entityGetterTestSuite) SetupSuite() {
	s.gcx = &gin.Context{
		Params: []gin.Param{
			{
				Key:   "identifier",
				Value: s.idn,
			},
		},
	}
	s.gtr = &proxy.EntityGetter{
		URL: s.url,
	}
}

func (s *entityGetterTestSuite) TestGetPath() {
	pth, err := s.gtr.GetPath(s.gcx)

	if s.err != nil {
		s.Assert().Empty(pth)
		s.Assert().Equal(s.err, err)
	} else {
		s.Assert().Contains(pth, s.url)
		s.Assert().Contains(pth, s.idn)
		s.Assert().NoError(err)
	}
}

func TestEntityGetter(t *testing.T) {
	for _, testCase := range []*entityGetterTestSuite{
		{
			url: "entity",
			idn: "Earth",
		},
		{
			err: proxy.ErrEmptyIdentifier,
		},
	} {
		suite.Run(t, testCase)
	}
}

type newEntityDownloaderTestSuite struct {
	suite.Suite
	url string
}

func (s *newEntityDownloaderTestSuite) TestNewEntityDownloader() {
	gtr := proxy.NewEntityDownloader(s.url)

	s.Assert().NotNil(gtr)
	s.Assert().Equal(s.url, gtr.URL)
}

func TestNewEntityDownloader(t *testing.T) {
	for _, testCase := range []*newEntityDownloaderTestSuite{
		{
			url: "entity",
		},
	} {
		suite.Run(t, testCase)
	}
}

type entityDownloaderTestSuite struct {
	suite.Suite
	gcx *gin.Context
	gtr *proxy.EntityDownloader
	url string
	idn string
	err error
}

func (s *entityDownloaderTestSuite) SetupSuite() {
	s.gcx = &gin.Context{
		Params: []gin.Param{
			{
				Key:   "identifier",
				Value: s.idn,
			},
		},
	}
	s.gtr = &proxy.EntityDownloader{
		URL: s.url,
	}
}

func (s *entityDownloaderTestSuite) TestGetPath() {
	pth, err := s.gtr.GetPath(s.gcx)

	if s.err != nil {
		s.Assert().Empty(pth)
		s.Assert().Equal(s.err, err)
	} else {
		s.Assert().Contains(pth, s.url)
		s.Assert().Contains(pth, s.idn)
		s.Assert().NoError(err)
	}
}

func TestEntityDownloader(t *testing.T) {
	for _, testCase := range []*entityDownloaderTestSuite{
		{
			url: "entity",
			idn: "Earth",
		},
		{
			err: proxy.ErrEmptyIdentifier,
		},
	} {
		suite.Run(t, testCase)
	}
}

type newDateEntitiesGetterTestSuite struct {
	suite.Suite
	url string
}

func (s *newDateEntitiesGetterTestSuite) TestNewDateEntitiesGetter() {
	gtr := proxy.NewDateEntitiesGetter(s.url)

	s.Assert().NotNil(gtr)
	s.Assert().Equal(s.url, gtr.URL)
}

func TestNewDateEntitiesGetter(t *testing.T) {
	for _, testCase := range []*newDateEntitiesGetterTestSuite{
		{
			url: "entity",
		},
	} {
		suite.Run(t, testCase)
	}
}

type dateEntitiesGetterTestSuite struct {
	suite.Suite
	gcx *gin.Context
	gtr *proxy.DateEntitiesGetter
	url string
	dte string
	err error
}

func (s *dateEntitiesGetterTestSuite) SetupSuite() {
	s.gcx = &gin.Context{
		Params: []gin.Param{
			{
				Key:   "date",
				Value: s.dte,
			},
		},
	}
	s.gtr = &proxy.DateEntitiesGetter{
		URL: s.url,
	}
}

func (s *dateEntitiesGetterTestSuite) TestGetPath() {
	pth, err := s.gtr.GetPath(s.gcx)

	if s.err != nil {
		s.Assert().Empty(pth)
		s.Assert().Equal(s.err, err)
	} else {
		s.Assert().Contains(pth, s.url)
		s.Assert().Contains(pth, s.dte)
		s.Assert().NoError(err)
	}
}

func TestDateEntitiesGetter(t *testing.T) {
	for _, testCase := range []*dateEntitiesGetterTestSuite{
		{
			url: "entity",
			dte: "2022-07-14",
		},
		{
			err: proxy.ErrEmptyDate,
		},
	} {
		suite.Run(t, testCase)
	}
}

type newDateEntityGetterTestSuite struct {
	suite.Suite
	url string
}

func (s *newDateEntityGetterTestSuite) TestNewDateEntityGetter() {
	gtr := proxy.NewDateEntityGetter(s.url)

	s.Assert().NotNil(gtr)
	s.Assert().Equal(s.url, gtr.URL)
}

func TestNewDateEntityGetter(t *testing.T) {
	for _, testCase := range []*newDateEntityGetterTestSuite{
		{
			url: "entity",
		},
	} {
		suite.Run(t, testCase)
	}
}

type dateEntityGetterTestSuite struct {
	suite.Suite
	gcx *gin.Context
	gtr *proxy.DateEntityGetter
	url string
	idn string
	dte string
	err error
}

func (s *dateEntityGetterTestSuite) SetupSuite() {
	s.gcx = &gin.Context{
		Params: []gin.Param{
			{
				Key:   "identifier",
				Value: s.idn,
			},
			{
				Key:   "date",
				Value: s.dte,
			},
		},
	}
	s.gtr = &proxy.DateEntityGetter{
		URL: s.url,
	}
}

func (s *dateEntityGetterTestSuite) TestGetPath() {
	pth, err := s.gtr.GetPath(s.gcx)

	if s.err != nil {
		s.Assert().Empty(pth)
		s.Assert().Equal(s.err, err)
	} else {
		s.Assert().Contains(pth, s.url)
		s.Assert().Contains(pth, s.dte)
		s.Assert().Contains(pth, s.idn)
		s.Assert().NoError(err)
	}
}

func TestDateEntityGetter(t *testing.T) {
	for _, testCase := range []*dateEntityGetterTestSuite{
		{
			url: "entity",
			err: proxy.ErrEmptyDate,
		},
		{
			url: "entity",
			dte: "2022-07-14",
			err: proxy.ErrEmptyIdentifier,
		},
		{
			url: "entity",
			dte: "2022-07-14",
			idn: "Earth",
		},
	} {
		suite.Run(t, testCase)
	}
}

type newDateEntityDownloaderTestSuite struct {
	suite.Suite
	url string
}

func (s *newDateEntityDownloaderTestSuite) TestNewDateEntityDownloader() {
	gtr := proxy.NewDateEntityDownloader(s.url)

	s.Assert().NotNil(gtr)
	s.Assert().Equal(s.url, gtr.URL)
}

func TestNewDateEntityDownloader(t *testing.T) {
	for _, testCase := range []*newDateEntityDownloaderTestSuite{
		{
			url: "entity",
		},
	} {
		suite.Run(t, testCase)
	}
}

type dateEntityDownloaderTestSuite struct {
	suite.Suite
	gcx *gin.Context
	gtr *proxy.DateEntityDownloader
	url string
	idn string
	dte string
	err error
}

func (s *dateEntityDownloaderTestSuite) SetupSuite() {
	s.gcx = &gin.Context{
		Params: []gin.Param{
			{
				Key:   "identifier",
				Value: s.idn,
			},
			{
				Key:   "date",
				Value: s.dte,
			},
		},
	}
	s.gtr = &proxy.DateEntityDownloader{
		URL: s.url,
	}
}

func (s *dateEntityDownloaderTestSuite) TestGetPath() {
	pth, err := s.gtr.GetPath(s.gcx)

	if s.err != nil {
		s.Assert().Empty(pth)
		s.Assert().Equal(s.err, err)
	} else {
		s.Assert().Contains(pth, s.url)
		s.Assert().Contains(pth, s.dte)
		s.Assert().Contains(pth, s.idn)
		s.Assert().NoError(err)
	}
}

func TestDateEntityDownloader(t *testing.T) {
	for _, testCase := range []*dateEntityDownloaderTestSuite{
		{
			url: "entity",
			err: proxy.ErrEmptyDate,
		},
		{
			url: "entity",
			dte: "2022-07-14",
			err: proxy.ErrEmptyIdentifier,
		},
		{
			url: "entity",
			dte: "2022-07-14",
			idn: "Earth",
		},
	} {
		suite.Run(t, testCase)
	}
}

type byGroupGetterBaseTestSuite struct {
	suite.Suite
	ctx *gin.Context
	err error
	gtr *proxy.ByGroupGetterBase
}

func (s *byGroupGetterBaseTestSuite) TestGetUser() {
	usr, err := s.gtr.GetUser(s.ctx)

	s.Assert().ErrorIs(err, s.err)

	if s.err != nil {
		s.Assert().Nil(usr)
	} else {
		s.Assert().Equal(s.ctx.Keys["user"], usr)
	}
}

func TestByGroupGetterBase(t *testing.T) {
	for _, testCase := range []*byGroupGetterBaseTestSuite{
		{
			ctx: &gin.Context{},
			err: proxy.ErrUnauthorized,
		},
		{
			ctx: &gin.Context{
				Keys: map[string]any{"user": "wrong user type"},
			},
			err: proxy.ErrWrongUserType,
		},
		{
			ctx: &gin.Context{
				Keys: map[string]any{"user": &httputil.User{
					Username: "new_user",
					Groups:   []string{"group_1"},
				},
				},
			},
		},
	} {
		suite.Run(t, testCase)
	}
}

type newByGroupEntitiesGetterTestSuite struct {
	suite.Suite
	url   string
	group string
}

func (s *newByGroupEntitiesGetterTestSuite) TestNewByGroupEntitiesGetter() {
	gtr := proxy.NewByGroupEntitiesGetter(s.url, s.group)

	s.Assert().NotNil(gtr)
	s.Assert().Equal(s.url, gtr.URL)
	s.Assert().Equal(s.group, gtr.Group)
}

func TestNewByGroupEntitiesGetter(t *testing.T) {
	for _, testCase := range []*newByGroupEntitiesGetterTestSuite{
		{
			url:   "entity",
			group: "group_1",
		},
	} {
		suite.Run(t, testCase)
	}
}

type byGroupEntitiesGetterTestSuite struct {
	suite.Suite
	url  string
	path string
	ctx  *gin.Context
	gtr  *proxy.ByGroupEntitiesGetter
}

func (s *byGroupEntitiesGetterTestSuite) SetupSuite() {
	s.gtr = &proxy.ByGroupEntitiesGetter{
		URL:   s.url,
		Group: "group_1",
	}
}

func (s *byGroupEntitiesGetterTestSuite) TestGetPath() {
	pth, err := s.gtr.GetPath(s.ctx)

	s.Assert().Equal(s.path, pth)
	s.Assert().NoError(err)
}

func TestByGroupEntitiesGetter(t *testing.T) {
	for _, testCase := range []*byGroupEntitiesGetterTestSuite{
		{
			url: "snapshots",
			ctx: &gin.Context{
				Keys: map[string]any{"user": &httputil.User{
					Username: "new_user",
					Groups:   []string{"group_1"},
				},
				},
			},
			path: "aggregations/snapshots/snapshots_group_1.ndjson",
		},
		{
			url: "snapshots",
			ctx: &gin.Context{
				Keys: map[string]any{"user": &httputil.User{
					Username: "new_user2",
					Groups:   []string{"group_2"},
				},
				},
			},
			path: "aggregations/snapshots/snapshots.ndjson",
		},
	} {
		suite.Run(t, testCase)
	}
}

type newByGroupEntityGetterTestSuite struct {
	suite.Suite
	url   string
	group string
}

func (s *newByGroupEntitiesGetterTestSuite) TestNewByGroupEntityGetter() {
	gtr := proxy.NewByGroupEntityGetter(s.url, s.group)

	s.Assert().NotNil(gtr)
	s.Assert().Equal(s.url, gtr.URL)
	s.Assert().Equal(s.group, gtr.Group)
}

func TestNewByGroupEntityGetter(t *testing.T) {
	for _, testCase := range []*newByGroupEntityGetterTestSuite{
		{
			url:   "entity",
			group: "group_1",
		},
	} {
		suite.Run(t, testCase)
	}
}

type byGroupEntityGetterTestSuite struct {
	suite.Suite
	url  string
	path string
	idn  string
	ctx  *gin.Context
	err  error
	gtr  *proxy.ByGroupEntityGetter
}

func (s *byGroupEntityGetterTestSuite) SetupSuite() {
	s.gtr = &proxy.ByGroupEntityGetter{
		URL:   s.url,
		Group: "group_1",
	}
}

func (s *byGroupEntityGetterTestSuite) TestGetPath() {
	pth, err := s.gtr.GetPath(s.ctx)

	s.Assert().Equal(s.path, pth)
	s.Assert().Equal(s.err, err)
}

func TestByGroupEntityGetter(t *testing.T) {
	for _, testCase := range []*byGroupEntityGetterTestSuite{
		{
			url: "snapshots",
			idn: "enwiki_namespace_0",
			ctx: &gin.Context{
				Params: []gin.Param{
					{Key: "identifier", Value: "enwiki_namespace_0"}},
				Keys: map[string]any{"user": &httputil.User{
					Username: "new_user",
					Groups:   []string{"group_1"},
				},
				},
			},
			path: "snapshots/enwiki_namespace_0_group_1.json",
		},
		{
			url: "snapshots",
			idn: "enwiki_namespace_0",
			ctx: &gin.Context{
				Params: []gin.Param{
					{Key: "identifier", Value: "enwiki_namespace_0"}},
				Keys: map[string]any{"user": &httputil.User{
					Username: "new_user2",
					Groups:   []string{"group_2"},
				},
				},
			},
			path: "snapshots/enwiki_namespace_0.json",
		},
		{
			url: "snapshots",
			idn: "",
			ctx: &gin.Context{
				Params: []gin.Param{
					{Key: "identifier", Value: ""}},
				Keys: map[string]any{"user": &httputil.User{
					Username: "new_user2",
					Groups:   []string{"group_2"},
				},
				},
			},
			path: "",
			err:  proxy.ErrEmptyIdentifier,
		},
	} {
		suite.Run(t, testCase)
	}
}

type newByGroupEntityDownloaderTestSuite struct {
	suite.Suite
	url   string
	group string
}

func (s *newByGroupEntityDownloaderTestSuite) TestNewByGroupEntityDownloader() {
	gtr := proxy.NewByGroupEntityDownloader(s.url, s.group)

	s.Assert().NotNil(gtr)
	s.Assert().Equal(s.url, gtr.URL)
	s.Assert().Equal(s.group, gtr.Group)
}

func TestNewByGroupEntityDownloader(t *testing.T) {
	for _, testCase := range []*newByGroupEntityDownloaderTestSuite{
		{
			url:   "entity",
			group: "group_1",
		},
	} {
		suite.Run(t, testCase)
	}
}

type byGroupEntityDownloaderTestSuite struct {
	suite.Suite
	url  string
	path string
	idn  string
	ctx  *gin.Context
	err  error
	gtr  *proxy.ByGroupEntityDownloader
}

func (s *byGroupEntityDownloaderTestSuite) SetupSuite() {
	s.gtr = &proxy.ByGroupEntityDownloader{
		URL:   s.url,
		Group: "group_1",
	}
}

func (s *byGroupEntityDownloaderTestSuite) TestGetPath() {
	pth, err := s.gtr.GetPath(s.ctx)

	s.Assert().Equal(s.path, pth)
	s.Assert().Equal(s.err, err)
}

func TestByGroupEntityDownloader(t *testing.T) {
	for _, testCase := range []*byGroupEntityDownloaderTestSuite{
		{
			url: "snapshots",
			idn: "enwiki_namespace_0",
			ctx: &gin.Context{
				Params: []gin.Param{
					{Key: "identifier", Value: "enwiki_namespace_0"}},
				Keys: map[string]any{"user": &httputil.User{
					Username: "new_user",
					Groups:   []string{"group_1"},
				},
				},
			},
			path: "snapshots/enwiki_namespace_0_group_1.tar.gz",
		},
		{
			url: "snapshots",
			idn: "enwiki_namespace_0",
			ctx: &gin.Context{
				Params: []gin.Param{
					{Key: "identifier", Value: "enwiki_namespace_0"}},
				Keys: map[string]any{"user": &httputil.User{
					Username: "new_user2",
					Groups:   []string{"group_2"},
				},
				},
			},
			path: "snapshots/enwiki_namespace_0.tar.gz",
		},
		{
			url: "snapshots",
			idn: "",
			ctx: &gin.Context{
				Params: []gin.Param{
					{Key: "identifier", Value: ""}},
				Keys: map[string]any{"user": &httputil.User{
					Username: "new_user2",
					Groups:   []string{"group_2"},
				},
				},
			},
			path: "",
			err:  proxy.ErrEmptyIdentifier,
		},
	} {
		suite.Run(t, testCase)
	}
}
