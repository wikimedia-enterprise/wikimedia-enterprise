package handlers_test

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"wikimedia-enterprise/services/bulk-ingestion/handlers"
	pb "wikimedia-enterprise/services/bulk-ingestion/handlers/protos"
	"wikimedia-enterprise/services/bulk-ingestion/packages/container"
	"wikimedia-enterprise/services/bulk-ingestion/submodules/schema"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/dig"
)

type unmarshalProducerMock struct {
	schema.UnmarshalProducer
	mock.Mock
}

func (m *unmarshalProducerMock) Produce(_ context.Context, msgs ...*schema.Message) error {
	msg := msgs[0]
	return m.Called(msg).Error(0)
}

func (m *unmarshalProducerMock) Flush(time int) int {
	return 0
}

type handlersTestSuite struct {
	suite.Suite
	cont *dig.Container
	nReq *pb.NamespacesRequest
	pReq *pb.ProjectsRequest
	aReq *pb.ArticlesRequest
	ctx  context.Context
	s3c  *s3ClientMock
}

type s3ClientMock struct {
	s3iface.S3API
	mock.Mock
}

func (c *s3ClientMock) PutObjectWithContext(ctx aws.Context, in *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, error) {
	args := c.Called(in)
	return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
}

// Allows the caller to provide a validation callback for objects sent to S3.
func (c *s3ClientMock) MockPutObject(cb func(args mock.Arguments)) {
	call := c.On("PutObjectWithContext", mock.Anything).Return(&s3.PutObjectOutput{}, nil)

	if cb != nil {
		call.Run(cb)
	}
}

// SetupTest intializes an instance of handlersTestSuite.
func (s *handlersTestSuite) SetupTest() {
	os.Setenv("KAFKA_BOOTSTRAP_SERVERS", "broker:29092")
	os.Setenv("SERVER_PORT", "50051")
	os.Setenv("MEDIAWIKI_CLIENT_URL", "http://localhost:3000")
	os.Setenv("SCHEMA_REGISTRY_URL", "http://shema:8080")

	s.s3c = &s3ClientMock{}

	upm := &unmarshalProducerMock{}
	upm.On("Produce", mock.Anything).Return(nil)

	var err error
	s.cont, err = container.New()
	_ = s.cont.Decorate(func() s3iface.S3API { return s.s3c })
	_ = s.cont.Decorate(func() schema.UnmarshalProducer { return upm })

	s.Assert().NoError(err)
	s.Assert().NoError(err)

	s.nReq = new(pb.NamespacesRequest)
	s.pReq = new(pb.ProjectsRequest)
	s.aReq = new(pb.ArticlesRequest)
	s.ctx = context.Background()
}

// TestNewServer verifies NewServer constructor's return type and values.
func (s *handlersTestSuite) TestNewServer() {
	srv, err := handlers.NewServer(s.cont)
	s.Assert().NoError(err)
	s.Assert().NotNil(srv)
	s.Assert().IsType(new(handlers.Server), srv)
}

// TestNamespaces verifies Namespaces method's return objects and S3 integration.
func (s *handlersTestSuite) TestNamespaces() {
	srv, err := handlers.NewServer(s.cont)
	s.Assert().NoError(err)

	calls := 0
	s.s3c.MockPutObject(func(args mock.Arguments) {
		in := args.Get(0).(*s3.PutObjectInput)

		calls++
		if strings.Contains(*in.Key, ".ndjson") {
			return
		}

		snp := &schema.Namespace{}
		bod, _ := io.ReadAll(in.Body)
		err = json.Unmarshal(bod, snp)
		s.Assert().NoError(err)
		s.Assert().NotEmpty(snp.Description)
	})

	res, err := srv.Namespaces(s.ctx, s.nReq)
	s.Assert().NoError(err)
	s.Assert().IsType(new(pb.NamespacesResponse), res)

	// One per namespace, plus an aggregated file.
	s.Assert().Equal(5, calls)
}

// TestProjects verifies Projects method's return objects.
func (s *handlersTestSuite) TestProjects() {
	srv, err := handlers.NewServer(s.cont)
	s.Assert().NoError(err)
	s.s3c.MockPutObject(nil)

	res, err := srv.Projects(s.ctx, s.pReq)
	s.Assert().NoError(err)
	s.Assert().IsType(new(pb.ProjectsResponse), res)
}

func (s *handlersTestSuite) TestArticles() {
	srv, err := handlers.NewServer(s.cont)
	s.Assert().NoError(err)
	s.s3c.MockPutObject(nil)

	res, err := srv.Articles(s.ctx, s.aReq)
	s.Assert().Error(err)
	s.Assert().IsType(new(pb.ArticlesResponse), res)
}

// TestHandlers runs the test suite for handlers package.
func TestHandlers(t *testing.T) {
	suite.Run(t, new(handlersTestSuite))
}
