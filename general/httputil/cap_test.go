package httputil

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type redisMock struct {
	redis.Cmdable
	mock.Mock
}

func (m *redisMock) Get(_ context.Context, key string) *redis.StringCmd {
	ags := m.Called(key)

	return redis.NewStringResult(
		strconv.Itoa(ags.Int(0)),
		ags.Error(1),
	)
}

func (m *redisMock) Set(_ context.Context, key string, val interface{}, exp time.Duration) *redis.StatusCmd {
	ags := m.Called(key, val, int(exp))

	return redis.NewStatusResult("", ags.Error(0))
}

func (m *redisMock) Incr(_ context.Context, key string) *redis.IntCmd {
	ags := m.Called(key)

	return redis.NewIntResult(0, ags.Error(0))
}

type capByRedisTestSuite struct {
	suite.Suite
	ctx context.Context
	cbr *CapByRedis
	idr string
	lmt int
	cnt int
	alw bool
	err error
	ger error
	ser error
	ier error
}

func (s *capByRedisTestSuite) SetupSuite() {
	key := fmt.Sprintf("%s:count", s.idr)

	cmd := new(redisMock)
	cmd.On("Get", key).Return(s.cnt, s.ger)
	cmd.On("Set", key, 1, 0).Return(s.ser)
	cmd.On("Incr", key).Return(s.ier)

	s.cbr = &CapByRedis{
		Redis: cmd,
	}
	s.ctx = context.Background()
}

func (s *capByRedisTestSuite) TestCheck() {
	alw, err := s.cbr.Check(s.ctx, s.idr, s.lmt)

	s.Assert().Equal(s.alw, alw)
	s.Assert().Equal(s.err, err)
}

func TestCapByRedis(t *testing.T) {
	ter := errors.New("error for testing")

	for _, testCase := range []*capByRedisTestSuite{
		{
			idr: "test",
			alw: true,
			cnt: 1,
			lmt: 2,
		},
		{
			idr: "test",
			alw: false,
			cnt: 10,
			lmt: 10,
		},
		{
			idr: "test",
			cnt: 0,
			lmt: 1,
			ger: redis.Nil,
			alw: true,
		},
		{
			idr: "test",
			cnt: 1,
			lmt: 2,
			ger: ter,
			err: ter,
		},
		{
			idr: "test",
			cnt: 1,
			lmt: 2,
			ier: ter,
			err: ter,
		},
		{
			idr: "cap:test:user:test",
			cnt: 1,
			lmt: 2,
			ier: ter,
			err: ter,
		},
		{
			idr: "cap:test:user:test2",
			alw: true,
			cnt: 2,
			lmt: 3,
		},
	} {
		suite.Run(t, testCase)
	}
}

type capMock struct {
	mock.Mock
}

func (m *capMock) Check(_ context.Context, _ string, _ int) (bool, error) {
	ags := m.Called()

	return ags.Bool(0), ags.Error(1)
}

type capTestSuite struct {
	suite.Suite
	sts       int
	idr       string
	lmt       int
	alw       bool
	err       error
	cfg       CapConfigWrapper
	usr       *User
	cmk       *capMock
	srv       *httptest.Server
	path      string
	expectCap bool
}

func (s *capTestSuite) createServer() http.Handler {
	gin.SetMode(gin.TestMode)
	rtr := gin.New()

	rtr.Use(func(gcx *gin.Context) {
		gcx.Set("user", s.usr)
	})

	if s.path == "" {
		s.path = "/test"
	}

	rtr.GET(s.path, Cap(s.cmk, s.cfg), func(gcx *gin.Context) {
		gcx.Status(http.StatusOK)
	})

	if strings.Contains(s.path, "*") {
		rtr.GET(s.path, Cap(s.cmk, s.cfg), func(gcx *gin.Context) {
			gcx.Status(http.StatusOK)
		})
	}

	return rtr
}

func (s *capTestSuite) SetupSuite() {
	s.cmk = new(capMock)
	if s.expectCap {
		s.cmk.On("Check").Return(s.alw, s.err)
	}
	s.srv = httptest.NewServer(s.createServer())
}

func (s *capTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *capTestSuite) TestCap() {

	requestPath := s.path
	requestPath = strings.Replace(requestPath, "*", "actual", 1)

	res, err := http.Get(fmt.Sprintf("%s%s", s.srv.URL, requestPath))
	s.Assert().NoError(err)
	defer res.Body.Close()

	s.Assert().Equal(s.sts, res.StatusCode)

	if s.err != nil {
		dta, err := io.ReadAll(res.Body)
		s.Assert().NoError(err)
		s.Assert().Contains(string(dta), s.err.Error())
	}

	if s.expectCap {
		s.cmk.AssertCalled(s.T(), "Check")
	} else {
		s.cmk.AssertNotCalled(s.T(), "Check")
	}
}

func TestCap(t *testing.T) {
	for _, testCase := range []*capTestSuite{
		{
			sts: http.StatusOK,
			usr: &User{
				Username: "test",
				Groups:   []string{"group_2"},
			},
			cfg: []*CapConfig{
				{
					Limit:  10,
					Groups: []string{"group_1"},
				},
			},
			idr:       "user:test",
			lmt:       10,
			expectCap: false,
		},
		{
			sts: http.StatusTooManyRequests,
			usr: &User{
				Username: "test",
				Groups:   []string{"group_2"},
			},
			cfg: []*CapConfig{
				{
					Limit:  10,
					Groups: []string{"group_2"},
				},
			},
			idr:       "user:test",
			lmt:       11,
			alw:       false,
			expectCap: true,
		},
		{
			sts: http.StatusUnauthorized,
			usr: nil,
			cfg: []*CapConfig{
				{
					Limit:  10,
					Groups: []string{"group_2"},
				},
			},
			idr:       "user:test",
			lmt:       10,
			alw:       true,
			expectCap: false,
		},
		{
			sts: http.StatusTooManyRequests,
			usr: &User{
				Username: "test",
				Groups:   []string{"group_2"},
			},
			cfg: []*CapConfig{
				{
					Limit:  10,
					Groups: []string{"group_2"},
				},
			},
			idr:       "user:test",
			lmt:       10,
			alw:       false,
			expectCap: true,
		},
		{
			sts: http.StatusInternalServerError,
			usr: &User{
				Username: "test",
				Groups:   []string{"group_2"},
			},
			cfg: []*CapConfig{
				{
					Limit:  10,
					Groups: []string{"group_2"},
				},
			},
			idr:       "user:test",
			lmt:       10,
			alw:       false,
			err:       errors.New("this is a test error"),
			expectCap: true,
		},
		{
			sts: http.StatusInternalServerError,
			usr: &User{
				Username: "test",
				Groups:   []string{"group_2"},
			},
			cfg: []*CapConfig{
				{
					Limit:  10,
					Groups: []string{"group_2"},
				},
			},
			idr:       "user:test",
			lmt:       10,
			alw:       false,
			err:       errors.New("this is a test error"),
			expectCap: true,
		},
		{
			sts: http.StatusOK,
			usr: &User{
				Username: "test1",
				Groups:   []string{"group_1"},
			},
			cfg: []*CapConfig{
				{
					Limit:       10,
					Groups:      []string{"group_1"},
					PrefixGroup: "cap:ondemand",
					Products:    []string{"test"},
				},
				{
					Limit:  19,
					Groups: []string{"group_2"},
				},
			},
			idr:       "cap:test:user:test",
			lmt:       5,
			alw:       true,
			err:       nil,
			expectCap: false,
		},
		{
			sts: http.StatusTooManyRequests,
			usr: &User{
				Username: "test",
				Groups:   []string{"group_1"},
			},
			cfg: []*CapConfig{
				{
					Limit:       10,
					Groups:      []string{"group_1"},
					PrefixGroup: "cap:ondemand",
					Products:    []string{"articles"},
				},
			},
			path:      "/v2/articles/Josephine_Baker",
			idr:       "cap:ondemand:user:test",
			lmt:       10,
			alw:       false,
			expectCap: true,
		},
		{
			sts: http.StatusOK,
			usr: &User{
				Username: "test2",
				Groups:   []string{"group_2"},
			},
			cfg: []*CapConfig{
				{
					Limit:  100,
					Groups: []string{"group_3"},
				},
				{
					Limit:  10,
					Groups: []string{"group_1"},
				},
			},
			idr:       "user:test2",
			lmt:       1000,
			alw:       true,
			err:       nil,
			expectCap: false,
		},

		{
			sts: http.StatusOK,
			usr: &User{
				Username: "test1",
				Groups:   []string{"group_1"},
			},
			path: "/v2/snapshots/enwiki/download",
			cfg: []*CapConfig{
				{
					PrefixGroup: "cap:snapshot",
					Products:    []string{"snapshots", "structured-snapshots"},
					Limit:       15,
					Groups:      []string{"group_1"},
				},
				{
					PrefixGroup: "cap:ondemand",
					Products:    []string{"articles", "structured-contents"},
					Limit:       5000,
					Groups:      []string{"group_1"},
				},
				{
					PrefixGroup: "cap:chunk",
					Products:    []string{"chunks"},
					Limit:       1500,
					Groups:      []string{"group_1"},
				},
			},
			idr:       "cap:snapshot:user:test1",
			lmt:       15,
			alw:       true,
			expectCap: true,
		},
		{
			sts: http.StatusOK,
			usr: &User{
				Username: "test1",
				Groups:   []string{"group_1"},
			},
			path: "/v2/articles/Josephine_Baker",
			cfg: []*CapConfig{
				{
					PrefixGroup: "cap:snapshot",
					Products:    []string{"snapshots", "structured-snapshots"},
					Limit:       15,
					Groups:      []string{"group_1"},
				},
				{
					PrefixGroup: "cap:ondemand",
					Products:    []string{"articles", "structured-contents"},
					Limit:       5000,
					Groups:      []string{"group_1"},
				},
				{
					PrefixGroup: "cap:chunk",
					Products:    []string{"chunks"},
					Limit:       1500,
					Groups:      []string{"group_1"},
				},
			},
			idr:       "cap:ondemand:user:test1",
			lmt:       5000,
			alw:       true,
			expectCap: true,
		},
		{
			sts: http.StatusOK,
			usr: &User{
				Username: "test1",
				Groups:   []string{"group_1"},
			},
			path: "/v2/structured-contents/Josephine_Baker",
			cfg: []*CapConfig{
				{
					PrefixGroup: "cap:snapshot",
					Products:    []string{"snapshots", "structured-snapshots"},
					Limit:       15,
					Groups:      []string{"group_1"},
				},
				{
					PrefixGroup: "cap:ondemand",
					Products:    []string{"articles", "structured-contents"},
					Limit:       5000,
					Groups:      []string{"group_1"},
				},
				{
					PrefixGroup: "cap:chunk",
					Products:    []string{"chunks"},
					Limit:       1500,
					Groups:      []string{"group_1"},
				},
			},
			idr:       "cap:ondemand:user:test1",
			lmt:       5000,
			alw:       true,
			expectCap: true,
		},
		{
			sts: http.StatusOK,
			usr: &User{
				Username: "test1",
				Groups:   []string{"group_1"},
			},
			path: "/v2/snapshots/enwiki/chunks/enwiki/download",
			cfg: []*CapConfig{
				{
					PrefixGroup: "cap:snapshot",
					Products:    []string{"snapshots", "structured-snapshots"},
					Limit:       15,
					Groups:      []string{"group_1"},
				},
				{
					PrefixGroup: "cap:ondemand",
					Products:    []string{"articles", "structured-contents"},
					Limit:       5000,
					Groups:      []string{"group_1"},
				},
				{
					PrefixGroup: "cap:chunk",
					Products:    []string{"chunks"},
					Limit:       1500,
					Groups:      []string{"group_1"},
				},
			},
			idr:       "cap:chunk:user:test1",
			lmt:       1500,
			alw:       true,
			expectCap: true,
		},
		{
			sts: http.StatusOK,
			usr: &User{
				Username: "test1",
				Groups:   []string{"group_3"},
			},
			path: "/v2/articles/Josephine_Baker",
			cfg: []*CapConfig{
				{
					PrefixGroup: "cap:ondemand",
					Products:    []string{"articles", "structured-contents"},
					Limit:       5000,
					Groups:      []string{"group_1"},
				},
			},
			expectCap: false,
		},
		{
			sts: http.StatusOK,
			usr: &User{
				Username: "test1",
				Groups:   []string{"group_1"},
			},
			path: "/v2/unmatched/path",
			cfg: []*CapConfig{
				{
					PrefixGroup: "cap:ondemand",
					Products:    []string{"articles", "structured-contents"},
					Limit:       5000,
					Groups:      []string{"group_1"},
				},
			},
			expectCap: false,
		},
		{
			sts: http.StatusOK,
			usr: &User{
				Username: "test1",
				Groups:   []string{"group_1"},
			},
			path: "/v2/any/path",
			cfg: []*CapConfig{
				{
					Limit:  10000,
					Groups: []string{"group_1"},
				},
			},
			idr:       "user:test1",
			lmt:       10000,
			alw:       true,
			expectCap: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

func TestCapConfigWrapperValidation(t *testing.T) {
	testCases := []struct {
		name          string
		jsonConfig    string
		expectedError error
		allowEmpty    bool
	}{
		// Allowed for local development.
		{
			name:          "Empty config",
			jsonConfig:    `[]`,
			expectedError: nil,
			allowEmpty:    true,
		},
		{
			name: "Missing groups",
			jsonConfig: `[{
                                "limit": 100,
                                "prefix_group": "cap:test"
                        }]`,
			expectedError: ErrMissingGroups,
		},
		{
			name: "Prefix group without products",
			jsonConfig: `[{
                                "limit": 100,
                                "prefix_group": "cap:test",
                                "groups": ["group_1"]
                        }]`,
			expectedError: ErrMissingProducts,
		},
		{
			name: "Valid config",
			jsonConfig: `[{
                                "limit": 100,
                                "prefix_group": "cap:test",
                                "products": ["snapshots"],
                                "groups": ["group_1"]
                        }]`,
			expectedError: nil,
		},
		{
			name: "Multiple configs with one invalid",
			jsonConfig: `[
                                {
                                        "limit": 100,
                                        "prefix_group": "cap:test",
                                        "products": ["snapshots"],
                                        "groups": ["group_1"]
                                },
                                {
                                        "limit": 200,
                                        "prefix_group": "cap:other",
                                        "groups": ["group_2"]
                                }
                        ]`,
			expectedError: ErrMissingProducts,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var cfg CapConfigWrapper
			err := cfg.UnmarshalEnvironmentValue(tc.jsonConfig)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tc.expectedError) ||
					strings.Contains(err.Error(), tc.expectedError.Error()),
					"Expected error containing %q, got %q", tc.expectedError, err)
			} else {
				assert.NoError(t, err)
				if !tc.allowEmpty {
					assert.NotEmpty(t, cfg)
				}
			}
		})
	}
}
