package articles_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"wikimedia-enterprise/api/realtime/config/env"
	"wikimedia-enterprise/api/realtime/handlers/v2/articles"
	"wikimedia-enterprise/api/realtime/libraries/resolver"
	"wikimedia-enterprise/general/ksqldb"
	"wikimedia-enterprise/general/schema"

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

type modelTestSuite struct {
	suite.Suite
	mdl *articles.Model
}

func (s *modelTestSuite) SetupSuite() {
	s.mdl = new(articles.Model)
	s.mdl.Parts = []int{0, 1}
	s.mdl.SetPartitions(50, 10)
}

func (s *modelTestSuite) TestGetPartitions() {
	s.NotEmpty(s.mdl.GetPartitions())
	s.Equal(
		[]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		s.mdl.GetPartitions())
}

func (s *modelTestSuite) TestPartitionsContain() {
	s.True(s.mdl.PartitionsContain(0))
	s.True(s.mdl.PartitionsContain(1))
	s.True(s.mdl.PartitionsContain(2))
	s.True(s.mdl.PartitionsContain(4))
	s.True(s.mdl.PartitionsContain(9))
}

func (s *modelTestSuite) TestGetFormattedPartitions() {
	s.Equal("0,1,2,3,4,5,6,7,8,9", s.mdl.GerFormattedPartitions())
}

func TestModel(t *testing.T) {
	suite.Run(t, new(modelTestSuite))
}

type articlesTestSuite struct {
	suite.Suite
	pms *articles.Parameters
	ctx context.Context
	srv *httptest.Server
	pld string
	sts int
	ker error
}

func (s *articlesTestSuite) createServer() http.Handler {
	gin.SetMode(gin.TestMode)
	rtr := gin.New()

	rtr.POST("/articles", articles.NewHandler(s.ctx, s.pms))

	return rtr
}

func (s *articlesTestSuite) SetupSuite() {
	rvr, err := resolver.New(new(schema.Article))
	s.Assert().NoError(err)

	s.ctx = context.Background()

	kdb := new(ksqldbMock)
	kdb.On("Push").Return(s.ker)

	s.pms = &articles.Parameters{
		Resolver: rvr,
		KSQLDB:   kdb,
		Env: &env.Environment{
			MaxParts:   2,
			Partitions: 4,
		},
	}

	s.srv = httptest.NewServer(s.createServer())
}

func (s *articlesTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *articlesTestSuite) TestNewHandler() {
	res, err := http.Post(fmt.Sprintf("%s/articles", s.srv.URL), "application/json", strings.NewReader(s.pld))
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
	for _, testCase := range []*articlesTestSuite{
		{
			sts: http.StatusOK,
		},
		{
			sts: http.StatusInternalServerError,
			ker: errors.New("ksqldb not available"),
		},
		{
			pld: `{"since":"string"}`,
			sts: http.StatusUnprocessableEntity,
		},
		{
			pld: `{"fields":["string"]}`,
			sts: http.StatusUnprocessableEntity,
		},
		{
			pld: `{"parts":[0,1], "offsets":{"0":100, "1":101, "2":100, "3":50}}`,
			sts: http.StatusOK,
		},
		{
			pld: `{"parts":[1], "since_per_partition":{"0":"2023-06-19T19:43:44.52Z"}}`,
			sts: http.StatusOK,
		},
		{
			pld: `{"parts":[1], "since":"2023-06-19T19:43:44.52Z"}`,
			sts: http.StatusUnprocessableEntity,
		},
		{
			pld: `{"parts":[0], "since_per_partition":{"0":"2023-06-19T19:43:44.52Z"}, "offsets":{"0":100, "1":101, "2":100, "3":50}}`,
			sts: http.StatusUnprocessableEntity,
		},
		{
			pld: `{"parts":[0,2], "offsets":{"0":100, "1":101, "2":100, "3":50}}`,
			sts: http.StatusUnprocessableEntity,
		},
		{
			pld: `{"parts":[0,1], "offsets":{"0":100, "1":101, "2":100, "3":50, "4":50}}`,
			sts: http.StatusUnprocessableEntity,
		},
	} {
		suite.Run(t, testCase)
	}
}
