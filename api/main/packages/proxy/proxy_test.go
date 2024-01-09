package proxy_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
	"wikimedia-enterprise/api/main/config/env"
	"wikimedia-enterprise/api/main/packages/proxy"
	"wikimedia-enterprise/general/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client/metadata"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type s3Mock struct {
	mock.Mock
	s3iface.S3API
}

func (m *s3Mock) SelectObjectContentWithContext(_ aws.Context, sip *s3.SelectObjectContentInput, _ ...request.Option) (*s3.SelectObjectContentOutput, error) {
	arg := m.Called(*sip.Bucket, *sip.Key, *sip.Expression)
	return nil, arg.Error(0)
}

func (m *s3Mock) GetObjectRequest(sip *s3.GetObjectInput) (*request.Request, *s3.GetObjectOutput) {
	m.Called(*sip.Bucket, *sip.Key)

	req := request.New(aws.Config{}, metadata.ClientInfo{}, request.Handlers{}, nil, &request.Operation{}, nil, nil)

	return req, nil
}

func (m *s3Mock) GetObjectWithContext(_ aws.Context, sip *s3.GetObjectInput, _ ...request.Option) (*s3.GetObjectOutput, error) {
	arg := m.Called(*sip.Bucket, *sip.Key)

	if arg.Get(0) == nil {
		return nil, arg.Error(1)
	}

	return arg.Get(0).(*s3.GetObjectOutput), arg.Error(1)
}

func (m *s3Mock) HeadObjectWithContext(_ aws.Context, sip *s3.HeadObjectInput, _ ...request.Option) (*s3.HeadObjectOutput, error) {
	arg := m.Called(*sip.Bucket, *sip.Key)

	if arg.Get(0) == nil {
		return nil, arg.Error(1)
	}

	return arg.Get(0).(*s3.HeadObjectOutput), arg.Error(1)
}

type pgMock struct {
	mock.Mock
	proxy.PathGetter
}

func (m *pgMock) GetPath(_ *gin.Context) (string, error) {
	arg := m.Called()
	return arg.String(0), arg.Error(1)
}

type testFilterTestSuite struct {
	suite.Suite
	ftr *proxy.Filter
}

func (s *testFilterTestSuite) SetupTest() {
	s.ftr = &proxy.Filter{
		Field: "a",
		Value: "b",
	}
}

func (s *testFilterTestSuite) TestFilter() {
	s.False(s.ftr.Filter("b"))
	s.True(s.ftr.Filter("a"))
}

func TestFilter(t *testing.T) {
	suite.Run(t, new(testFilterTestSuite))
}

type testModelTestSuite struct {
	suite.Suite
	mdl *proxy.Model
}

func (s *testModelTestSuite) SetupTest() {
	s.mdl = &proxy.Model{
		Fields: []string{"a", "b", "c", "d.*", "e.v.*"},
		Filters: []*proxy.Filter{
			{
				Field: "a",
				Value: "b",
			},
			{
				Field: "b",
				Value: "a",
			},
		},
	}
}

func (s *testModelTestSuite) TestGetBuildFieldsIndex() {
	s.Nil(s.mdl.GetFieldsIndex())

	s.mdl.BuildFieldsIndex()
	idx := s.mdl.GetFieldsIndex()

	s.Equal("d", idx["d"])
	s.Equal("e.v", idx["e"].(map[string]interface{})["v"])
	s.NotNil(idx)
	s.Equal(len(s.mdl.Fields), len(idx))
}

func (s *testModelTestSuite) TestGetFilterByField() {
	s.Assert().Equal(s.mdl.Filters[0], s.mdl.GetFilterByField("a"))
	s.Assert().Equal(s.mdl.Filters[1], s.mdl.GetFilterByField("b"))
}

func TestModel(t *testing.T) {
	suite.Run(t, new(testModelTestSuite))
}

type testProxyGetSuite struct {
	suite.Suite
	hdl func(*proxy.Params, proxy.PathGetter) gin.HandlerFunc
	par *proxy.Params
	srv *httptest.Server
	env *env.Environment
	str *s3Mock
	pgt *pgMock
	ctx context.Context
	exp string
	bkt string
	key string
	err error
}

func (s *testProxyGetSuite) createServer() http.Handler {
	gin.SetMode(gin.TestMode)

	rtr := gin.New()
	rtr.GET("/hourlys", s.hdl(s.par, s.pgt))

	return rtr
}

func (s *testProxyGetSuite) SetupSuite() {
	s.err = errors.New("test suite error")
	s.ctx = context.Background()
	s.env = &env.Environment{
		AWSBucket: s.bkt,
	}
}

func (s *testProxyGetSuite) SetupTest() {
	s.pgt = new(pgMock)
	s.str = new(s3Mock)
	s.par = &proxy.Params{
		S3:  s.str,
		Env: s.env,
	}
	s.srv = httptest.NewServer(s.createServer())
}

func (s *testProxyGetSuite) TearDownTest() {
	s.srv.Close()
}

func (s *testProxyGetSuite) TestProxySelectObjectErr() {
	s.pgt.On("GetPath").Return(s.key, nil)
	s.str.On("SelectObjectContentWithContext", s.bkt, s.key, s.exp).Return(s.err)

	res, err := http.Get(fmt.Sprintf("%s%s", s.srv.URL, "/hourlys"))
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusInternalServerError, res.StatusCode)

	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)
	s.Assert().Contains(string(data), s.err.Error())
}

func (s *testProxyGetSuite) TestProxyPathErr() {
	s.pgt.On("GetPath").Return("", s.err)
	res, err := http.Get(fmt.Sprintf("%s%s", s.srv.URL, "/hourlys"))

	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusInternalServerError, res.StatusCode)

	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)
	s.Assert().Contains(string(data), s.err.Error())
}

func TestProxyGet(t *testing.T) {
	for _, testCase := range []*testProxyGetSuite{
		{
			hdl: proxy.NewGetEntities,
			bkt: "test",
			exp: "SELECT * FROM S3Object as main",
			key: "/agregations/hourlys/2022-08-14.ndjson",
		},
		{
			hdl: proxy.NewGetEntity,
			bkt: "test",
			exp: "SELECT * FROM S3Object as main",
			key: "/agregations/hourlys/2022-08-14.ndjson",
		},
	} {
		suite.Run(t, testCase)
	}
}

type testProxyDownloadSuite struct {
	suite.Suite
	par *proxy.Params
	srv *httptest.Server
	env *env.Environment
	str *s3Mock
	pgt *pgMock
	clt *http.Client
	ctx context.Context
	pth string
	bkt string
	key string
	err error
}

func (s *testProxyDownloadSuite) SetupSuite() {
	s.err = errors.New("test error")
	s.ctx = context.Background()
	s.env = &env.Environment{
		AWSBucket: s.bkt,
	}
}

func (s *testProxyDownloadSuite) SetupTest() {
	s.pgt = new(pgMock)
	s.str = new(s3Mock)
	s.pth = "/hourlys"
	s.par = &proxy.Params{
		S3:  s.str,
		Env: s.env,
	}
	s.srv = httptest.NewServer(s.createServer())
	s.clt = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			req.Response.StatusCode = http.StatusTemporaryRedirect
			return http.ErrUseLastResponse
		},
	}
}

func (s *testProxyDownloadSuite) createServer() http.Handler {
	gin.SetMode(gin.TestMode)

	rtr := gin.New()
	rtr.GET(s.pth, proxy.NewGetDownload(s.par, s.pgt))

	return rtr
}

func (s *testProxyDownloadSuite) TestDownloadPathErr() {
	s.pgt.On("GetPath").Return("", s.err)

	res, err := http.Get(fmt.Sprintf("%s%s", s.srv.URL, s.pth))
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusInternalServerError, res.StatusCode)

	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)
	s.Assert().Contains(string(data), s.err.Error())
}

func (s *testProxyDownloadSuite) TestProxyGetDownloadRedirect() {
	s.pgt.On("GetPath").Return(s.key, nil)
	s.str.On("GetObjectRequest", s.bkt, s.key).Return()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", s.srv.URL, s.pth), nil)
	s.Assert().NoError(err)

	res, err := s.clt.Do(req)
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusTemporaryRedirect, res.StatusCode)
}

func TestProxyDownload(t *testing.T) {
	for _, testCase := range []*testProxyDownloadSuite{
		{
			bkt: "test",
			key: "/hourlys/2022-08-14/dewiki_namespace_0.tar.gz",
		},
	} {
		suite.Run(t, testCase)
	}
}

type testProxyHeadSuite struct {
	suite.Suite
	par *proxy.Params
	srv *httptest.Server
	env *env.Environment
	str *s3Mock
	pgt *pgMock
	ctx context.Context
	bkt string
	key string
	hoo *s3.HeadObjectOutput
	err error
	pth string
}

func (s *testProxyHeadSuite) SetupSuite() {
	s.err = errors.New("test suite error")
	s.ctx = context.Background()
	s.env = &env.Environment{
		AWSBucket: s.bkt,
	}
}

func (s *testProxyHeadSuite) SetupTest() {
	s.pgt = new(pgMock)
	s.str = new(s3Mock)
	s.pth = "/hourlys"
	s.par = &proxy.Params{
		S3:  s.str,
		Env: s.env,
	}
	s.srv = httptest.NewServer(s.createServer())
}

func (s *testProxyHeadSuite) createServer() http.Handler {
	gin.SetMode(gin.TestMode)

	rtr := gin.New()
	rtr.HEAD(s.pth, proxy.NewHeadDownload(s.par, s.pgt))

	return rtr
}

func (s *testProxyHeadSuite) TestProxyHeadPathErr() {
	s.pgt.On("GetPath").Return("", s.err)

	res, err := http.Head(fmt.Sprintf("%s%s", s.srv.URL, s.pth))
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusInternalServerError, res.StatusCode)
}

func (s *testProxyHeadSuite) TestProxyHeadObjectErr() {
	s.pgt.On("GetPath").Return(s.key, nil)
	s.str.On("HeadObjectWithContext", s.bkt, s.key).Return(nil, s.err)

	res, err := http.Head(fmt.Sprintf("%s%s", s.srv.URL, s.pth))
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusNotFound, res.StatusCode)
}

func (s *testProxyHeadSuite) TestProxyHead() {
	s.pgt.On("GetPath").Return(s.key, nil)
	s.str.On("HeadObjectWithContext", s.bkt, s.key).Return(s.hoo, nil)

	res, err := http.Head(fmt.Sprintf("%s%s", s.srv.URL, s.pth))
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	s.Assert().Equal(*s.hoo.ContentType, res.Header.Get("Content-Type"))
	s.Assert().Equal(*s.hoo.Expires, res.Header.Get("Expires"))
	s.Assert().Equal(strconv.FormatInt(*s.hoo.ContentLength, 10), res.Header.Get("Content-Length"))
	s.Assert().Equal(s.hoo.LastModified.Format(time.RFC1123), res.Header.Get("Last-Modified"))
	s.Assert().Equal(*s.hoo.AcceptRanges, res.Header.Get("Accept-Ranges"))
	s.Assert().Equal(*s.hoo.ETag, res.Header.Get("ETag"))
	s.Assert().Equal(*s.hoo.CacheControl, res.Header.Get("Cache-Control"))
	s.Assert().Equal(*s.hoo.ContentDisposition, res.Header.Get("Content-Disposition"))
	s.Assert().Equal(*s.hoo.ContentEncoding, res.Header.Get("Content-Encoding"))

	s.Assert().NoError(err)
}

func TestProxyHead(t *testing.T) {
	for _, testCase := range []*testProxyHeadSuite{
		{
			bkt: "test",
			key: "/hourlys/2022-08-14/dewiki_namespace_0.tar.gz",
			hoo: &s3.HeadObjectOutput{
				ContentLength:      aws.Int64(123),
				ContentType:        aws.String("application/json"),
				Expires:            aws.String("2022-08-15"),
				LastModified:       aws.Time(time.Now()),
				AcceptRanges:       aws.String("Accept-Ranges"),
				ETag:               aws.String("test tag"),
				CacheControl:       aws.String("Cache-Control"),
				ContentDisposition: aws.String("Content-Disposition"),
				ContentEncoding:    aws.String("Content-Encoding"),
			},
		},
	} {
		suite.Run(t, testCase)
	}
}

type pltMock struct {
	mock.Mock
	config.API
}

func (m *pltMock) GetProjects() []string {
	arg := m.Called()

	if arg.Get(0) == nil {
		return []string{}
	}

	return arg.Get(0).([]string)
}

func (m *pltMock) GetLanguage(dtb string) string {
	ags := m.Called(dtb)

	return ags.String(0)
}

type getLargeEntitiesTestSuite struct {
	suite.Suite
	sts int
	pth string
	ent string
	nme string
	prs []string
	lgs []string
	ets []string
	bkt string
	pts *proxy.Params
	srv *httptest.Server
	erg error
}

func (s *getLargeEntitiesTestSuite) createServer() http.Handler {
	gin.SetMode(gin.TestMode)

	rtr := gin.New()
	rtr.GET(fmt.Sprintf("%s/:name", s.pth), proxy.NewGetLargeEntities(s.pts, s.ent))

	return rtr
}

func (s *getLargeEntitiesTestSuite) SetupSuite() {
	cfm := new(pltMock)
	cfm.On("GetProjects").Return(s.prs)

	s3m := new(s3Mock)

	for i, prg := range s.prs {
		cfm.On("GetLanguage", prg).Return(s.lgs[i])

		for _, ent := range s.ets {
			s3m.On("GetObjectWithContext", s.bkt, fmt.Sprintf("%s/%s/%s.json", s.ent, prg, s.nme)).
				Return(&s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader([]byte(ent)))}, s.erg)
		}
	}

	s.pts = &proxy.Params{
		Cfg: cfm,
		S3:  s3m,
		Env: &env.Environment{
			AWSBucket: s.bkt,
		},
	}

	s.srv = httptest.NewServer(s.createServer())
}

func (s *getLargeEntitiesTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getLargeEntitiesTestSuite) TestNewGetLargeEntities() {
	res, err := http.Get(fmt.Sprintf("%s%s/%s", s.srv.URL, s.pth, s.nme))
	s.Assert().NoError(err)
	s.Assert().Equal(s.sts, res.StatusCode)

	dta, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)

	for _, ent := range s.ets {
		s.Assert().Contains(string(dta), ent)
	}
}

func TestNewGetLargeEntities(t *testing.T) {
	for _, testCase := range []*getLargeEntitiesTestSuite{
		{
			sts: http.StatusOK,
			pth: "/articles",
			ent: "articles",
			nme: "Earth",
			bkt: "wme-data",
			ets: []string{`{"name":"Earth"}`, `{"name":"Earth"}`},
			prs: []string{"enwiki", "plwiki"},
			lgs: []string{"en", "pl"},
		},
		{
			sts: http.StatusNotFound,
			pth: "/articles",
			ent: "articles",
			nme: "Earth",
			bkt: "wme-data",
			ets: []string{`{"status":404,"message":"Not Found"}`},
			prs: []string{"enwiki", "plwiki"},
			lgs: []string{"en", "pl"},
			erg: errors.New("not found"),
		},
	} {
		suite.Run(t, testCase)
	}
}
