package getuser_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"wikimedia-enterprise/api/auth/config/env"
	"wikimedia-enterprise/api/auth/handlers/v1/getuser"
	"wikimedia-enterprise/api/auth/submodules/httputil"

	"github.com/alicebob/miniredis"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/suite"
)

type getuserHandlerTestSuite struct {
	suite.Suite
	srv                  *httptest.Server
	mr                   *miniredis.Miniredis
	cmd                  redis.Cmdable
	url                  string
	unm                  string
	ugs                  []string
	ondemandRequestCount int
	snapshotRequestCount int
	chunkRequestCount    int
	pth                  []env.AccessPath
	snapshotLimit        int
	ondemandLimit        int
	chunkLimit           int
}

func (s *getuserHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	s.url = "/get-user"
}

func (s *getuserHandlerTestSuite) SetupTest() {
	var err error
	s.mr, err = miniredis.Run()
	s.Assert().NoError(err)

	s.cmd = redis.NewClient(&redis.Options{
		Addr: s.mr.Addr(),
	})
	s.srv = httptest.NewServer(s.createServer())
}

func (s *getuserHandlerTestSuite) createServer() http.Handler {
	s.cmd.Set(
		context.Background(),
		fmt.Sprintf("cap:ondemand:user:%s:count", s.unm),
		s.ondemandRequestCount,
		0,
	)
	s.cmd.Set(
		context.Background(),
		fmt.Sprintf("cap:snapshot:user:%s:count", s.unm),
		s.snapshotRequestCount,
		0,
	)
	s.cmd.Set(
		context.Background(),
		fmt.Sprintf("cap:chunk:user:%s:count", s.unm),
		s.chunkRequestCount,
		0,
	)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user", &httputil.User{
			Username: s.unm,
			Groups:   s.ugs,
		})

		c.Next()
	})
	router.GET(s.url, getuser.NewHandler(&getuser.Parameters{
		Env: &env.Environment{
			OndemandLimit: strconv.Itoa(s.ondemandLimit),
			SnapshotLimit: strconv.Itoa(s.snapshotLimit),
			ChunkLimit:    strconv.Itoa(s.chunkLimit),
			AccessPolicy: &env.AccessPolicy{
				Map: map[string][]env.AccessPath{
					s.ugs[0]: s.pth,
				}},
		},
		Redis: s.cmd,
	}))

	return router
}

func (s *getuserHandlerTestSuite) TearDownTest() {
	s.srv.Close()
	s.mr.Close()
}

func (s *getuserHandlerTestSuite) TestGetUserHandler() {
	resp, err := http.Get(fmt.Sprintf("%s%s", s.srv.URL, s.url))
	s.Assert().NoError(err)
	s.Assert().Equal(resp.StatusCode, http.StatusOK)
	defer resp.Body.Close()

	tr := new(getuser.Response)
	s.Assert().NoError(json.NewDecoder(resp.Body).Decode(tr))
	s.Assert().Equal(http.StatusOK, resp.StatusCode)
	s.Assert().Equal(s.unm, tr.Username)
	s.Assert().Equal(s.ugs, tr.Groups)
	s.Assert().Equal(s.ondemandRequestCount, tr.OndemandRequestsCount)
	s.Assert().Equal(s.ondemandLimit, tr.OndemandLimit)
	s.Assert().Equal(s.snapshotRequestCount, tr.SnapshotRequestsCount)
	s.Assert().Equal(s.snapshotLimit, tr.SnapshotLimit)
	s.Assert().Equal(s.chunkRequestCount, tr.ChunkRequestsCount)
	s.Assert().Equal(s.chunkLimit, tr.ChunkLimit)
	s.Assert().Equal(s.pth, tr.Apis)
}

func TestCaptcha(t *testing.T) {
	for _, testCase := range []*getuserHandlerTestSuite{
		{
			unm:                  "username",
			ondemandRequestCount: 5,
			snapshotRequestCount: 10,
			chunkRequestCount:    15,
			ugs: []string{
				"group_1",
			},
			ondemandLimit: 10000,
			snapshotLimit: 10000,
			chunkLimit:    10000,
			pth: []env.AccessPath{
				{
					Path:   "/path-1",
					Method: "GET",
				},
				{
					Path:   "/path-2",
					Method: "POST",
				},
			},
		},
	} {
		suite.Run(t, testCase)
	}
}
