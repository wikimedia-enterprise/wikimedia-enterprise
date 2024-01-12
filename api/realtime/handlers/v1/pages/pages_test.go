package pages_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"wikimedia-enterprise/api/realtime/config/env"
	"wikimedia-enterprise/api/realtime/handlers/v1/pages"
	"wikimedia-enterprise/general/ksqldb"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ksqldbMock struct {
	mock.Mock
	ksqldb.PushPuller
}

func (s *ksqldbMock) Push(_ context.Context, _ *ksqldb.QueryRequest, _ func(qrw *ksqldb.HeaderRow, row ksqldb.Row) error) error {
	ags := s.Called()
	return ags.Error(0)
}

type pagesTestSuite struct {
	suite.Suite
	pms *pages.Parameters
	ctx context.Context
	srv *httptest.Server
	snc string
	sts int
	ker error
}

func (s *pagesTestSuite) createServer() http.Handler {
	gin.SetMode(gin.TestMode)
	rtr := gin.New()

	rtr.GET("/pages", pages.NewHandler(s.ctx, s.pms, "test"))

	return rtr
}

func (s *pagesTestSuite) SetupSuite() {
	s.ctx = context.Background()

	kdb := new(ksqldbMock)
	kdb.On("Push").Return(s.ker)

	s.pms = &pages.Parameters{
		KSQLDB: kdb,
		Env:    new(env.Environment),
	}

	s.srv = httptest.NewServer(s.createServer())
}

func (s *pagesTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *pagesTestSuite) TestNewHandler() {
	ulr := fmt.Sprintf("%s/pages", s.srv.URL)

	if len(s.snc) > 0 {
		ulr += fmt.Sprintf("?since=%s", s.snc)
	}

	res, err := http.Get(ulr)
	s.Assert().NoError(err)
	defer res.Body.Close()

	dta, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)

	if s.ker != nil {
		s.Assert().Equal(s.sts, res.StatusCode)
		s.Assert().Contains(string(dta), s.ker.Error())
	} else {
		s.Assert().Equal(s.sts, res.StatusCode)
	}
}

func TestNewHandler(t *testing.T) {
	for _, testCase := range []*pagesTestSuite{
		{
			sts: http.StatusOK,
		},
		{
			snc: "100",
			sts: http.StatusOK,
		},
		{
			snc: "2019-10-12T07:20:50.52Z",
			sts: http.StatusOK,
		},
		{
			sts: http.StatusInternalServerError,
			ker: errors.New("ksqldb not available"),
		},
		{
			snc: "not a valid date",
			sts: http.StatusBadRequest,
		},
	} {
		suite.Run(t, testCase)
	}
}
