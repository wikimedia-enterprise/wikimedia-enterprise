package s3api_test

import (
	"testing"
	"wikimedia-enterprise/services/bulk-ingestion/config/env"
	"wikimedia-enterprise/services/bulk-ingestion/libraries/s3api"

	"github.com/stretchr/testify/suite"
)

type kafkaTestSuite struct {
	suite.Suite
	env *env.Environment
}

func (s *kafkaTestSuite) SetupSuite() {
	s.env = new(env.Environment)
	// s.env.AWSID = "AWSID"
	// s.env.AWSKey = "AWSKey"
	s.env.AWSRegion = "AWSRegion"
}

func (s *kafkaTestSuite) TestNewS3() {
	s3 := s3api.New(s.env)
	s.Assert().NotNil(s3)
}

func TestS3(t *testing.T) {
	suite.Run(t, new(kafkaTestSuite))
}
