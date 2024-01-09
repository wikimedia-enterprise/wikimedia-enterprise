package legacy_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"wikimedia-enterprise/api/main/config/env"
	"wikimedia-enterprise/api/main/packages/legacy"
	"wikimedia-enterprise/general/config"
	"wikimedia-enterprise/general/httputil"

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
	arg := m.Called(*sip.Bucket, *sip.Key)

	return arg.Get(0).(*s3.SelectObjectContentOutput), arg.Error(1)
}

func (m *s3Mock) GetObjectRequest(gip *s3.GetObjectInput) (*request.Request, *s3.GetObjectOutput) {
	m.Called(*gip.Bucket, *gip.Key)

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

func (m *s3Mock) HeadObjectWithContext(_ aws.Context, hip *s3.HeadObjectInput, _ ...request.Option) (*s3.HeadObjectOutput, error) {
	arg := m.Called(*hip.Bucket, *hip.Key)

	return arg.Get(0).(*s3.HeadObjectOutput), arg.Error(1)
}

type cfgMock struct {
	mock.Mock
	config.API
}

func (m *cfgMock) GetNamespaces() []int {
	return m.Called().Get(0).([]int)
}

func setUserMW(group string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := new(httputil.User)
		user.SetUsername("user")
		user.SetGroups([]string{group})

		c.Set("user", user)
	}
}

func createServer(mtd string, pth string, hdr gin.HandlerFunc, mw gin.HandlerFunc) http.Handler {
	gin.SetMode(gin.TestMode)

	rtr := gin.New()
	rtr.Use(mw)
	rtr.Handle(mtd, pth, hdr)

	return rtr
}

type newGetPageHandlerTestSuite struct {
	suite.Suite
	srv *httptest.Server
	nme string
	dtb string
	bdy string
	err error
}

func (s *newGetPageHandlerTestSuite) SetupSuite() {
	env := &env.Environment{
		AWSBucket: "wme-data",
	}

	s3m := new(s3Mock)
	oup := &s3.GetObjectOutput{Body: io.NopCloser(strings.NewReader(s.bdy))}
	s3m.On("GetObjectWithContext", env.AWSBucket, fmt.Sprintf("articles/%s/%s.json", s.dtb, s.nme)).Return(oup, s.err)

	pms := &legacy.Params{
		S3:  s3m,
		Env: env,
	}

	s.srv = httptest.NewServer(createServer(http.MethodGet, "/pages/meta/:project/*name", legacy.NewGetPageHandler(pms), func(_ *gin.Context) {}))
}

func (s *newGetPageHandlerTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *newGetPageHandlerTestSuite) TestHandler() {
	res, err := http.Get(fmt.Sprintf("%s/pages/meta/%s/%s", s.srv.URL, s.dtb, s.nme))
	s.Assert().NoError(err)
	defer res.Body.Close()

	dta, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)

	if s.err != nil {
		s.Assert().Equal(http.StatusNotFound, res.StatusCode)
		s.Assert().Contains(string(dta), s.err.Error())
	} else {
		s.Assert().Equal(http.StatusOK, res.StatusCode)
		s.Assert().Equal(s.bdy, string(dta))
	}
}

func TestNewGetPageHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, testCase := range []*newGetPageHandlerTestSuite{
		{
			nme: "Earth",
			dtb: "enwiki",
			bdy: `{"name":"Eearth"}`,
		},
		{
			nme: "Earth",
			dtb: "enwiki",
			bdy: `{"name":"Eearth"}`,
			err: errors.New("specific key not found"),
		},
	} {
		suite.Run(t, testCase)
	}
}

type diffsListTestSuite struct {
	suite.Suite
	nsp int
	dte string
	nss []int
	err error
	srv *httptest.Server
}

func (s *diffsListTestSuite) SetupSuite() {
	env := &env.Environment{
		AWSBucket: "wme-data",
	}

	s3m := new(s3Mock)
	sop := new(s3.SelectObjectContentOutput)
	s3m.On("SelectObjectContentWithContext", env.AWSBucket, fmt.Sprintf("aggregations/batches/%s/batches.ndjson", s.dte)).Return(sop, s.err)

	cfm := new(cfgMock)
	cfm.On("GetNamespaces").Return(s.nss)

	pms := &legacy.Params{
		S3:  s3m,
		Env: env,
		Cfg: cfm,
	}

	s.srv = httptest.NewServer(createServer(http.MethodGet, "/diffs/meta/:date/:namespace", legacy.NewListDiffsHandler(pms), func(_ *gin.Context) {}))
}

func (s *diffsListTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *diffsListTestSuite) TestHandler() {
	res, err := http.Get(fmt.Sprintf("%s/diffs/meta/%s/%d", s.srv.URL, s.dte, s.nsp))
	s.Assert().NoError(err)
	defer res.Body.Close()

	dta, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusInternalServerError, res.StatusCode)
	s.Assert().Contains(string(dta), s.err.Error())
}

func TestNewListDiffsHandler(t *testing.T) {
	for _, testCase := range []*diffsListTestSuite{
		{
			dte: "2023-03-3",
			nss: []int{0, 2, 3},
			nsp: 0,
			err: errors.New("internal error"),
		},
	} {
		suite.Run(t, testCase)
	}
}

type getDiffTestSuite struct {
	suite.Suite
	nsp int
	dte string
	nss []int
	dbn string
	err error
	srv *httptest.Server
}

func (s *getDiffTestSuite) SetupSuite() {
	env := &env.Environment{
		AWSBucket: "wme-data",
	}

	s3m := new(s3Mock)
	sop := new(s3.SelectObjectContentOutput)
	s3m.On("SelectObjectContentWithContext", env.AWSBucket, fmt.Sprintf("batches/%s/%s_namespace_%d.json", s.dte, s.dbn, s.nsp)).Return(sop, s.err)

	cfm := new(cfgMock)
	cfm.On("GetNamespaces").Return(s.nss)

	pms := &legacy.Params{
		S3:  s3m,
		Env: env,
		Cfg: cfm,
	}

	s.srv = httptest.NewServer(createServer(http.MethodGet, "/diffs/meta/:date/:namespace/:project", legacy.NewGetDiffHandler(pms), func(_ *gin.Context) {}))
}

func (s *getDiffTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getDiffTestSuite) TestHandler() {
	res, err := http.Get(fmt.Sprintf("%s/diffs/meta/%s/%d/%s", s.srv.URL, s.dte, s.nsp, s.dbn))
	s.Assert().NoError(err)
	defer res.Body.Close()

	dta, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusInternalServerError, res.StatusCode)
	s.Assert().Contains(string(dta), s.err.Error())
}

func TestNewGetDiffHandler(t *testing.T) {
	for _, testCase := range []*getDiffTestSuite{
		{
			dbn: "enwiki",
			dte: "2023-03-03",
			nss: []int{0, 2, 3},
			nsp: 0,
			err: errors.New("internal error"),
		},
	} {
		suite.Run(t, testCase)
	}
}

type downloadDiffTestSuite struct {
	suite.Suite
	nsp int
	dte string
	nss []int
	dbn string
	err error
	srv *httptest.Server
}

func (s *downloadDiffTestSuite) SetupSuite() {
	env := &env.Environment{
		AWSBucket: "wme-data",
	}

	key := fmt.Sprintf("batches/%s/%s_namespace_%d.tar.gz", s.dte, s.dbn, s.nsp)
	s3m := new(s3Mock)
	s3m.On("HeadObjectWithContext", env.AWSBucket, key).Return(new(s3.HeadObjectOutput), s.err)
	s3m.On("GetObjectRequest", env.AWSBucket, key).Return()

	cfm := new(cfgMock)
	cfm.On("GetNamespaces").Return(s.nss)

	pms := &legacy.Params{
		S3:  s3m,
		Env: env,
		Cfg: cfm,
	}

	s.srv = httptest.NewServer(createServer(http.MethodGet, "/diffs/download/:date/:namespace/:project", legacy.NewDownloadDiffHandler(pms), func(_ *gin.Context) {}))
}

func (s *downloadDiffTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *downloadDiffTestSuite) TestHandler() {
	clt := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	res, err := clt.Get(fmt.Sprintf("%s/diffs/download/%s/%d/%s", s.srv.URL, s.dte, s.nsp, s.dbn))
	s.Assert().NoError(err)
	defer res.Body.Close()

	if s.err != nil {
		s.Assert().Equal(http.StatusNotFound, res.StatusCode)
	} else {
		s.Assert().Equal(http.StatusTemporaryRedirect, res.StatusCode)
	}
}

func TestNewDownloadDiffHandler(t *testing.T) {
	for _, testCase := range []*downloadDiffTestSuite{
		{
			dbn: "enwiki",
			dte: "2023-03-03",
			nss: []int{0, 2, 3},
			nsp: 0,
			err: errors.New("not found"),
		},
		{
			dbn: "enwiki",
			dte: "2023-03-03",
			nss: []int{0, 2, 3},
			nsp: 0,
		},
	} {
		suite.Run(t, testCase)
	}
}

type headDiffTestSuite struct {
	suite.Suite
	nsp int
	dte string
	nss []int
	dbn string
	err error
	srv *httptest.Server
}

func (s *headDiffTestSuite) SetupSuite() {
	env := &env.Environment{
		AWSBucket: "wme-data",
	}

	key := fmt.Sprintf("batches/%s/%s_namespace_%d.tar.gz", s.dte, s.dbn, s.nsp)
	s3m := new(s3Mock)
	hop := new(s3.HeadObjectOutput)
	s3m.On("HeadObjectWithContext", env.AWSBucket, key).Return(hop, s.err)

	cfm := new(cfgMock)
	cfm.On("GetNamespaces").Return(s.nss)

	pms := &legacy.Params{
		S3:  s3m,
		Env: env,
		Cfg: cfm,
	}

	s.srv = httptest.NewServer(createServer(http.MethodHead, "/diffs/download/:date/:namespace/:project", legacy.NewHeadDiffHandler(pms), func(_ *gin.Context) {}))
}

func (s *headDiffTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *headDiffTestSuite) TestHandler() {
	res, err := http.Head(fmt.Sprintf("%s/diffs/download/%s/%d/%s", s.srv.URL, s.dte, s.nsp, s.dbn))
	s.Assert().NoError(err)
	defer res.Body.Close()

	if s.err != nil {
		s.Assert().NoError(err)
		s.Assert().Equal(http.StatusNotFound, res.StatusCode)
	} else {
		s.Assert().Equal(http.StatusOK, res.StatusCode)
	}
}

func TestNewHeadDiffHandler(t *testing.T) {
	for _, testCase := range []*headDiffTestSuite{
		{
			dbn: "enwiki",
			dte: "2023-03-03",
			nss: []int{0, 2, 3},
			nsp: 0,
			err: errors.New("not found"),
		},
		{
			dbn: "enwiki",
			dte: "2023-03-03",
			nss: []int{0, 2, 3},
			nsp: 0,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getExportTestSuite struct {
	suite.Suite
	nsp     int
	nss     []int
	dbn     string
	err     error
	grp     string
	url     string
	setUser bool
	srv     *httptest.Server
	env     *env.Environment
}

func (s *getExportTestSuite) SetupSuite() {
	s.env = &env.Environment{
		AWSBucket:     "wme-data",
		FreeTierGroup: "group_1",
	}
}

func (s *getExportTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getExportTestSuite) TestHandler() {
	s3m := new(s3Mock)
	s3m.On("SelectObjectContentWithContext", s.env.AWSBucket, s.url).Return(&s3.SelectObjectContentOutput{}, s.err)

	cfm := new(cfgMock)
	cfm.On("GetNamespaces").Return(s.nss)

	pms := &legacy.Params{
		S3:  s3m,
		Env: s.env,
		Cfg: cfm,
	}

	mw := func(_ *gin.Context) {}
	if s.setUser {
		mw = setUserMW(s.grp)
	}

	s.srv = httptest.NewServer(createServer(http.MethodGet, "/exports/meta/:namespace/:project", legacy.NewGetExportHandler(pms), mw))

	res, err := http.Get(fmt.Sprintf("%s/exports/meta/%d/%s", s.srv.URL, s.nsp, s.dbn))

	s.Assert().NoError(err)
	defer res.Body.Close()

	dta, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)

	s.Assert().Contains(string(dta), s.err.Error())
	if s.setUser {
		s.Assert().Equal(http.StatusInternalServerError, res.StatusCode)
	} else {
		s.Assert().Equal(http.StatusUnauthorized, res.StatusCode)
	}
}

func TestNewGetExportHandler(t *testing.T) {
	for _, testCase := range []*getExportTestSuite{
		{
			dbn:     "enwiki",
			nss:     []int{0, 2, 3},
			nsp:     0,
			setUser: true,
			grp:     "group_1",
			url:     "snapshots/enwiki_namespace_0_group_1.json",
			err:     errors.New("internal error"),
		},
		{
			dbn:     "dewiki",
			nss:     []int{0, 2, 3},
			nsp:     0,
			setUser: true,
			grp:     "group_2",
			url:     "snapshots/dewiki_namespace_0.json",
			err:     errors.New("internal error"),
		},
		{
			dbn:     "dewiki",
			nss:     []int{0, 2, 3},
			nsp:     0,
			setUser: false,
			url:     "snapshots/dewiki_namespace_0.json",
			err:     errors.New("user not found"),
		},
	} {
		suite.Run(t, testCase)
	}
}

type exportsListTestSuite struct {
	suite.Suite
	nsp     int
	nss     []int
	err     error
	url     string
	grp     string
	setUser bool
	srv     *httptest.Server
	env     *env.Environment
}

func (s *exportsListTestSuite) SetupSuite() {
	s.env = &env.Environment{
		AWSBucket:     "wme-data",
		FreeTierGroup: "group_1",
	}
}

func (s *exportsListTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *exportsListTestSuite) TestHandler() {
	s3m := new(s3Mock)
	s3m.On("SelectObjectContentWithContext", s.env.AWSBucket, s.url).Return(&s3.SelectObjectContentOutput{}, s.err)

	cfm := new(cfgMock)
	cfm.On("GetNamespaces").Return(s.nss)

	pms := &legacy.Params{
		S3:  s3m,
		Env: s.env,
		Cfg: cfm,
	}

	mw := func(c *gin.Context) {
		c.Set("user", "Wrong user type")
	}

	if s.setUser {
		mw = setUserMW(s.grp)
	}
	s.srv = httptest.NewServer(createServer(http.MethodGet, "/exports/meta/:namespace", legacy.NewListExportsHandler(pms), mw))

	res, err := http.Get(fmt.Sprintf("%s/exports/meta/%d", s.srv.URL, s.nsp))
	s.Assert().NoError(err)
	defer res.Body.Close()

	dta, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusInternalServerError, res.StatusCode)
	s.Assert().Contains(string(dta), s.err.Error())
}

func TestNewListExportsHandler(t *testing.T) {
	for _, testCase := range []*exportsListTestSuite{
		{
			nss:     []int{0, 2, 3},
			nsp:     0,
			grp:     "group_1",
			setUser: true,
			err:     errors.New("internal error"),
			url:     "aggregations/snapshots/snapshots_group_1.ndjson",
		},
		{
			nss:     []int{0, 2, 3},
			nsp:     0,
			grp:     "group_2",
			setUser: true,
			err:     errors.New("internal error"),
			url:     "aggregations/snapshots/snapshots.ndjson",
		},
		{
			nss:     []int{0, 2, 3},
			nsp:     0,
			grp:     "group_2",
			setUser: false,
			err:     errors.New("unknown user type"),
			url:     "aggregations/snapshots/snapshots.ndjson",
		},
	} {
		suite.Run(t, testCase)
	}
}

type headExportTestSuite struct {
	suite.Suite
	nsp    int
	nss    []int
	dbn    string
	url    string
	err    error
	srv    *httptest.Server
	env    *env.Environment
	mw     func(c *gin.Context)
	status int
}

func (s *headExportTestSuite) SetupSuite() {
	s.env = &env.Environment{
		AWSBucket:     "wme-data",
		FreeTierGroup: "group_1",
	}
}

func (s *headExportTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *headExportTestSuite) TestHandler() {
	s3m := new(s3Mock)
	hop := new(s3.HeadObjectOutput)
	s3m.On("HeadObjectWithContext", s.env.AWSBucket, s.url).Return(hop, s.err)

	cfm := new(cfgMock)
	cfm.On("GetNamespaces").Return(s.nss)

	pms := &legacy.Params{
		S3:  s3m,
		Env: s.env,
		Cfg: cfm,
	}

	s.srv = httptest.NewServer(createServer(http.MethodHead, "/exports/download/:namespace/:project", legacy.NewHeadExportHandler(pms), s.mw))

	res, err := http.Head(fmt.Sprintf("%s/exports/download/%d/%s", s.srv.URL, s.nsp, s.dbn))
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(s.status, res.StatusCode)
}

func TestNewHeadExportHandler(t *testing.T) {
	for _, testCase := range []*headExportTestSuite{
		{
			dbn:    "enwiki",
			nss:    []int{0, 2, 3},
			nsp:    0,
			url:    "snapshots/enwiki_namespace_0_group_1.tar.gz",
			status: http.StatusUnauthorized,
			mw: func(_ *gin.Context) {
			},
		},
		{
			dbn:    "enwiki",
			nss:    []int{0, 2, 3},
			nsp:    0,
			url:    "snapshots/enwiki_namespace_0_group_1.tar.gz",
			status: http.StatusInternalServerError,
			mw: func(c *gin.Context) {
				c.Set("user", "wrong user type")
			},
		},
		{
			dbn:    "enwiki",
			nss:    []int{0, 2, 3},
			nsp:    10,
			url:    "snapshots/enwiki_namespace_0_group_1.tar.gz",
			status: http.StatusBadRequest,
			mw: func(c *gin.Context) {
				user := new(httputil.User)
				user.SetUsername("user")
				user.SetGroups([]string{"group_1"})
				c.Set("user", user)
			},
		},
		{
			dbn: "enwiki",
			nss: []int{0, 2, 3},
			nsp: 0,
			url: "snapshots/enwiki_namespace_0_group_1.tar.gz",
			mw: func(c *gin.Context) {
				user := new(httputil.User)
				user.SetUsername("user")
				user.SetGroups([]string{"group_1"})
				c.Set("user", user)
			},
			err:    errors.New("not found"),
			status: http.StatusNotFound,
		},
		{
			dbn: "enwiki",
			nss: []int{0, 2, 3},
			nsp: 0,
			url: "snapshots/enwiki_namespace_0_group_1.tar.gz",
			mw: func(c *gin.Context) {
				user := new(httputil.User)
				user.SetUsername("user")
				user.SetGroups([]string{"group_1"})
				c.Set("user", user)
			},
			status: http.StatusOK,
		},
		{
			dbn: "enwiki",
			nss: []int{0, 2, 3},
			nsp: 0,
			url: "snapshots/enwiki_namespace_0.tar.gz",
			mw: func(c *gin.Context) {
				user := new(httputil.User)
				user.SetUsername("user")
				user.SetGroups([]string{"group_2"})
				c.Set("user", user)
			},
			status: http.StatusOK,
		},
	} {
		suite.Run(t, testCase)
	}
}

type downloadExportTestSuite struct {
	suite.Suite
	nsp    int
	nss    []int
	dbn    string
	err    error
	url    string
	env    *env.Environment
	srv    *httptest.Server
	mw     func(c *gin.Context)
	status int
}

func (s *downloadExportTestSuite) SetupSuite() {
	s.env = &env.Environment{
		AWSBucket:     "wme-data",
		FreeTierGroup: "group_1",
	}
}

func (s *downloadExportTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *downloadExportTestSuite) TestHandler() {
	s3m := new(s3Mock)
	s3m.On("HeadObjectWithContext", s.env.AWSBucket, s.url).Return(new(s3.HeadObjectOutput), s.err)
	s3m.On("GetObjectRequest", s.env.AWSBucket, s.url).Return()

	cfm := new(cfgMock)
	cfm.On("GetNamespaces").Return(s.nss)

	pms := &legacy.Params{
		S3:  s3m,
		Env: s.env,
		Cfg: cfm,
	}

	s.srv = httptest.NewServer(createServer(http.MethodGet, "/exports/download/:namespace/:project", legacy.NewDownloadExportHandler(pms), s.mw))

	clt := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	res, err := clt.Get(fmt.Sprintf("%s/exports/download/%d/%s", s.srv.URL, s.nsp, s.dbn))
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(s.status, res.StatusCode)
}

func TestNewDownloadExportHandler(t *testing.T) {
	for _, testCase := range []*downloadExportTestSuite{
		{
			dbn:    "enwiki",
			nss:    []int{0, 2, 3},
			nsp:    0,
			url:    "snapshots/enwiki_namespace_0_group_1.tar.gz",
			status: http.StatusUnauthorized,
			mw: func(_ *gin.Context) {
			},
		},
		{
			dbn:    "enwiki",
			nss:    []int{0, 2, 3},
			nsp:    0,
			url:    "snapshots/enwiki_namespace_0_group_1.tar.gz",
			status: http.StatusInternalServerError,
			mw: func(c *gin.Context) {
				c.Set("user", "wrong user type")
			},
		},
		{
			dbn:    "enwiki",
			nss:    []int{0, 2, 3},
			nsp:    10,
			url:    "snapshots/enwiki_namespace_0_group_1.tar.gz",
			status: http.StatusBadRequest,
			mw: func(c *gin.Context) {
				user := new(httputil.User)
				user.SetUsername("user")
				user.SetGroups([]string{"group_1"})
				c.Set("user", user)
			},
		},
		{
			dbn: "enwiki",
			nss: []int{0, 2, 3},
			nsp: 0,
			url: "snapshots/enwiki_namespace_0_group_1.tar.gz",
			mw: func(c *gin.Context) {
				user := new(httputil.User)
				user.SetUsername("user")
				user.SetGroups([]string{"group_1"})
				c.Set("user", user)
			},
			err:    errors.New("not found"),
			status: http.StatusNotFound,
		},
		{
			dbn: "enwiki",
			nss: []int{0, 2, 3},
			nsp: 0,
			url: "snapshots/enwiki_namespace_0_group_1.tar.gz",
			mw: func(c *gin.Context) {
				user := new(httputil.User)
				user.SetUsername("user")
				user.SetGroups([]string{"group_1"})
				c.Set("user", user)
			},
			status: http.StatusTemporaryRedirect,
		},
		{
			dbn: "enwiki",
			nss: []int{0, 2, 3},
			nsp: 0,
			url: "snapshots/enwiki_namespace_0.tar.gz",
			mw: func(c *gin.Context) {
				user := new(httputil.User)
				user.SetUsername("user")
				user.SetGroups([]string{"group_2"})
				c.Set("user", user)
			},
			status: http.StatusTemporaryRedirect,
		},
	} {
		suite.Run(t, testCase)
	}
}
