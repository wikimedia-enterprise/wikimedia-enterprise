package v2_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"wikimedia-enterprise/api/main/config/env"
	v2 "wikimedia-enterprise/api/main/handlers/v2"
	"wikimedia-enterprise/api/main/packages/container"
	"wikimedia-enterprise/api/main/submodules/config"
	"wikimedia-enterprise/api/main/submodules/log"
	parser "wikimedia-enterprise/api/main/submodules/structured-contents-parser"
	"wikimedia-enterprise/api/main/submodules/wmf"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/dig"
)

type v2TestSuite struct {
	suite.Suite
	cnt *dig.Container
	rtr *gin.Engine
}

func (s *v2TestSuite) SetupTest() {
	os.Setenv("AWS_REGION", "us-east-2")
	os.Setenv("AWS_ID", "foo")
	os.Setenv("AWS_KEY", "bar")
	os.Setenv("AWS_BUCKET", "bar")
	os.Setenv("AWS_BUCKET_COMMONS", "commons")
	os.Setenv("AWS_URL", "bar")
	os.Setenv("REDIS_ADDR", "some-addr")
	os.Setenv("REDIS_PASSWORD", "strong-password")
	os.Setenv("COGNITO_CLIENT_ID", "id")
	os.Setenv("COGNITO_CLIENT_SECRET", "secret")
	os.Setenv("ACCESS_MODEL", "id")
	os.Setenv("ACCESS_POLICY", "secret")
	os.Setenv("FREE_TIER_GROUP", "group_1")
	os.Setenv("KEY_TYPE_SUFFIX", "v2")
	os.Setenv("CAP_CONFIGURATION", `[{"groups":["test_group"],"limit":100,"products":["test_product"],"prefix_group":"cap:test"}]
`)

	gin.SetMode(gin.TestMode)
	s.rtr = gin.New()

	var err error
	s.cnt, err = container.New()
	s.NoError(err)
}

func (s *v2TestSuite) TestNewGroup() {
	grp, err := v2.NewGroup(s.cnt, s.rtr)
	s.NoError(err)
	s.NotNil(grp)
}

func TestV2(t *testing.T) {
	suite.Run(t, new(v2TestSuite))
}

type v2IntegratedTestSuite struct {
	suite.Suite
	server *httptest.Server
	s3     *s3APIMock

	url            string
	s3Method       string
	s3ExpectedPath string
	s3Error        error
}

type s3APIMock struct {
	mock.Mock
	s3iface.S3API
}

func (s *s3APIMock) SelectObjectContentWithContext(_ aws.Context, inp *s3.SelectObjectContentInput, _ ...request.Option) (*s3.SelectObjectContentOutput, error) {
	return s.
		Called(*inp.Bucket, *inp.Key).
		Get(0).(*s3.SelectObjectContentOutput), nil
}

func (s *s3APIMock) GetObjectRequest(inp *s3.GetObjectInput) (*request.Request, *s3.GetObjectOutput) {
	return s.
		Called(*inp.Bucket, *inp.Key).
		Get(0).(*request.Request), nil
}

func (s *s3APIMock) HeadObjectWithContext(_ aws.Context, inp *s3.HeadObjectInput, _ ...request.Option) (*s3.HeadObjectOutput, error) {
	return nil, s.
		Called(*inp.Bucket, *inp.Key).
		Error(0)
}

type RecordsEvent struct {
	Payload []byte
}

type EndEvent struct{}
type fakeEventStreamReader struct {
	events chan s3.SelectObjectContentEventStreamEvent
	err    error
}

func newFakeEventStreamReader(payload []byte) *fakeEventStreamReader {
	ch := make(chan s3.SelectObjectContentEventStreamEvent, 2)
	ch <- &s3.RecordsEvent{Payload: payload}
	ch <- &s3.EndEvent{}
	close(ch)

	return &fakeEventStreamReader{
		events: ch,
	}
}

func (f *fakeEventStreamReader) Events() <-chan s3.SelectObjectContentEventStreamEvent {
	return f.events
}

func (f *fakeEventStreamReader) Err() error {
	return f.err
}

func (f *fakeEventStreamReader) Close() error {
	// Always closed.
	return nil
}

func (s *v2IntegratedTestSuite) createContainer() (*dig.Container, error) {
	client := redis.NewClient(&redis.Options{})

	cnt := dig.New()
	for _, err := range []error{
		cnt.Provide(env.New),
		cnt.Provide(wmf.NewAPIFake().AsAPI),
		cnt.Provide(config.New),
		cnt.Provide(parser.New),
		cnt.Provide(func() s3iface.S3API { return s.s3 }),
		cnt.Provide(func() redis.Cmdable { return client }),
	} {
		if err != nil {
			s.T().Logf("problem creating container for dependency injection: %s", err.Error())
			return nil, err
		}
	}

	return cnt, nil
}

func (s *v2IntegratedTestSuite) SetupTest() {
	os.Setenv("AWS_BUCKET", "bucket")

	s.s3 = &s3APIMock{}
	cnt, err := s.createContainer()

	if err != nil {
		log.Fatal(err, log.Tip("problem creating container for dependency injection"))
	}

	gin.SetMode(gin.TestMode)
	rtr := gin.New()

	_, err = v2.NewGroup(cnt, rtr)
	s.NoError(err)

	s.server = httptest.NewServer(rtr)
}

func (s *v2IntegratedTestSuite) TearDownTest() {
	s.server.Close()
}

func (s *v2IntegratedTestSuite) mockS3() {
	switch s.s3Method {
	case "SelectObjectContentWithContext":
		reader := newFakeEventStreamReader(nil)
		es := s3.NewSelectObjectContentEventStream(func(o *s3.SelectObjectContentEventStream) {
			o.Reader = reader
			o.StreamCloser = reader
		})
		s.s3.On("SelectObjectContentWithContext", "bucket", s.s3ExpectedPath).Return(&s3.SelectObjectContentOutput{
			EventStream: es,
		}, s.s3Error)
	case "GetObjectRequest":
		req, err := http.NewRequest("GET", "", strings.NewReader(""))
		if err != nil {
			panic(err)
		}
		s.s3.On("GetObjectRequest", "bucket", s.s3ExpectedPath).Return(&request.Request{
			HTTPRequest: req,
			Operation:   &request.Operation{},
		}, &s3.GetObjectOutput{})
	}
}

func (s *v2IntegratedTestSuite) TestOriginalBatches() {
	s.mockS3()

	req, err := http.NewRequest("GET", s.server.URL+s.url, strings.NewReader(""))
	s.NoError(err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/x-ndjson")

	// Don't follow redirects from /download.
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}
	res, err := client.Do(req)
	s.NoError(err)
	defer res.Body.Close()

	s.s3.AssertExpectations(s.T())
}

func TestIntegrated(t *testing.T) {
	for _, testcase := range []*v2IntegratedTestSuite{
		{
			url:            "/v2/batches/2025-06-26/23",
			s3Method:       "SelectObjectContentWithContext",
			s3ExpectedPath: "aggregations/batches/2025-06-26/23/batches.ndjson",
			s3Error:        nil,
		},
		{
			url:            "/v2/batches/2025-06-25/23/enwiki_namespace_0",
			s3Method:       "SelectObjectContentWithContext",
			s3ExpectedPath: "batches/2025-06-25/23/enwiki_namespace_0.json",
			s3Error:        nil,
		},
		{
			url:            "/v2/batches/2025-06-25/23/enwiki_namespace_0/download",
			s3Method:       "GetObjectRequest",
			s3ExpectedPath: "batches/2025-06-25/23/enwiki_namespace_0.tar.gz",
			s3Error:        nil,
		},
	} {
		suite.Run(t, testcase)
	}
}
