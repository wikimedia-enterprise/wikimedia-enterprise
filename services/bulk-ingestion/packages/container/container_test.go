package container_test

import (
	"os"
	"testing"

	"wikimedia-enterprise/services/bulk-ingestion/packages/container"
	"wikimedia-enterprise/services/bulk-ingestion/submodules/schema"

	"github.com/stretchr/testify/suite"
)

type containerTestSuite struct {
	suite.Suite
}

func (s *containerTestSuite) SetupSuite() {
	os.Setenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:8001")
	os.Setenv("SCHEMA_REGISTRY_URL", "localhost:8085")
}

func (s *containerTestSuite) TestNew() {
	cont, err := container.New()
	s.Assert().NoError(err)
	s.Assert().NotNil(cont)
	s.Assert().NotPanics(func() {
		s.Assert().NoError(cont.Invoke(func(stream schema.UnmarshalProducer) {
			s.Assert().NotNil(stream)
		}))
	})
}

func TestContainer(t *testing.T) {
	suite.Run(t, new(containerTestSuite))
}
