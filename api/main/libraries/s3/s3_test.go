package s3_test

import (
	"testing"
	"wikimedia-enterprise/api/main/config/env"
	"wikimedia-enterprise/api/main/libraries/s3"

	"github.com/stretchr/testify/suite"
)

type s3TestSuite struct {
	suite.Suite
	env *env.Environment
}

func (s *s3TestSuite) SetupSuite() {
	s.env = new(env.Environment)
	s.env.AWSID = "AWSID"
	s.env.AWSKey = "AWSKey"
	s.env.AWSRegion = "AWSRegion"
}

func (s *s3TestSuite) TestNewS3() {
	s3i := s3.New(s.env)
	s.Assert().NotNil(s3i)
}

func TestS3(t *testing.T) {
	suite.Run(t, new(s3TestSuite))
}
