package handlers_test

import (
	"context"
	"os"
	"testing"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/general/wmf"
	"wikimedia-enterprise/services/bulk-ingestion/handlers"
	"wikimedia-enterprise/services/bulk-ingestion/handlers/articles"
	"wikimedia-enterprise/services/bulk-ingestion/handlers/namespaces"
	"wikimedia-enterprise/services/bulk-ingestion/handlers/projects"
	pb "wikimedia-enterprise/services/bulk-ingestion/handlers/protos"
	"wikimedia-enterprise/services/bulk-ingestion/packages/container"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/dig"
)

type unmarshalProducerMock struct {
	schema.UnmarshalProducer
	mock.Mock
}

type wmfMock struct {
	wmf.API
	mock.Mock
}

type handlersTestSuite struct {
	suite.Suite
	cont *dig.Container
	nReq *pb.NamespacesRequest
	pReq *pb.ProjectsRequest
	aReq *pb.ArticlesRequest
	srv  *handlers.Server
	ctx  context.Context
}

// SetupTest intializes an instance of handlersTestSuite.
func (s *handlersTestSuite) SetupTest() {
	os.Setenv("KAFKA_BOOTSTRAP_SERVERS", "broker:29092")
	os.Setenv("SERVER_PORT", "50051")
	os.Setenv("MEDIAWIKI_CLIENT_URL", "http://localhost:3000")
	os.Setenv("SCHEMA_REGISTRY_URL", "http://shema:8080")

	var err error
	s.cont, err = container.New()
	s.Assert().NoError(err)
	s.Assert().NoError(err)

	s.nReq = new(pb.NamespacesRequest)
	s.pReq = new(pb.ProjectsRequest)
	s.aReq = new(pb.ArticlesRequest)
	s.ctx = context.Background()

	s.srv = &handlers.Server{
		PParams: projects.Parameters{Stream: new(unmarshalProducerMock)},
		NParams: namespaces.Parameters{Stream: new(unmarshalProducerMock)},
		AParams: articles.Parameters{
			Stream: new(unmarshalProducerMock),
			Client: new(wmfMock),
		},
	}
}

// TestNewServer verifies NewServer constructor's return type and values.
func (s *handlersTestSuite) TestNewServer() {
	var err error
	s.srv, err = handlers.NewServer(s.cont)
	s.Assert().NoError(err)
	s.Assert().NotNil(s.srv)
	s.Assert().IsType(new(handlers.Server), s.srv)
}

// TestNamespaces verifies Namespaces method's return objects.
func (s *handlersTestSuite) TestNamespaces() {
	srv, err := handlers.NewServer(s.cont)
	s.Assert().NoError(err)

	res, err := srv.Namespaces(s.ctx, s.nReq)
	s.Assert().Error(err)
	s.Assert().IsType(new(pb.NamespacesResponse), res)
}

// TestProjects verifies Projects method's return objects.
func (s *handlersTestSuite) TestProjects() {
	srv, err := handlers.NewServer(s.cont)
	s.Assert().NoError(err)

	res, err := srv.Projects(s.ctx, s.pReq)
	s.Assert().Error(err)
	s.Assert().IsType(new(pb.ProjectsResponse), res)
}

func (s *handlersTestSuite) TestArticles() {
	srv, err := handlers.NewServer(s.cont)
	s.Assert().NoError(err)

	res, err := srv.Articles(s.ctx, s.aReq)
	s.Assert().Error(err)
	s.Assert().IsType(new(pb.ArticlesResponse), res)
}

// TestHandlers runs the test suite for handlers package.
func TestHandlers(t *testing.T) {
	suite.Run(t, new(handlersTestSuite))
}
