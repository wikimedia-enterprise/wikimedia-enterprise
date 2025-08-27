package s3api_test

import (
	"testing"
	"wikimedia-enterprise/services/snapshots/config/env"
	"wikimedia-enterprise/services/snapshots/libraries/s3api"

	"github.com/stretchr/testify/suite"
)

type s3TestSuite struct {
	suite.Suite
}

func (s *s3TestSuite) TestNewS3_Default() {
	e := &env.Environment{
		AWSRegion: "us-east-1",
	}
	s3 := s3api.New(e)
	s.Assert().NotNil(s3)
}

func (s *s3TestSuite) TestNewS3_WithCredentials() {
	e := &env.Environment{
		AWSRegion: "us-east-1",
		AWSID:     "test-id",
		AWSKey:    "test-key",
	}
	s3 := s3api.New(e)
	s.Assert().NotNil(s3)
}

func (s *s3TestSuite) TestNewS3_WithEndpoint() {
	e := &env.Environment{
		AWSRegion: "us-east-1",
		AWSURL:    "https://s3.example.com",
	}
	s3 := s3api.New(e)
	s.Assert().NotNil(s3)
}

func (s *s3TestSuite) TestNewS3_WithHTTP() {
	e := &env.Environment{
		AWSRegion: "us-east-1",
		AWSURL:    "http://localhost:9000",
	}
	s3 := s3api.New(e)
	s.Assert().NotNil(s3)
}

func TestS3(t *testing.T) {
	suite.Run(t, new(s3TestSuite))
}
