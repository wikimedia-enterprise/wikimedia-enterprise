package uploader_test

import (
	"testing"
	"wikimedia-enterprise/services/snapshots/config/env"
	"wikimedia-enterprise/services/snapshots/libraries/uploader"

	"github.com/stretchr/testify/suite"
)

type uploaderTestSuite struct {
	suite.Suite
}

func (s *uploaderTestSuite) TestNew_Default() {
	e := &env.Environment{
		AWSRegion: "us-east-1",
	}
	up, err := uploader.New(e)
	s.NoError(err)
	s.NotNil(up)
}

func (s *uploaderTestSuite) TestNew_WithCredentials() {
	e := &env.Environment{
		AWSRegion: "us-east-1",
		AWSID:     "test-id",
		AWSKey:    "test-key",
	}
	up, err := uploader.New(e)
	s.NoError(err)
	s.NotNil(up)
}

func (s *uploaderTestSuite) TestNew_WithEndpoint() {
	e := &env.Environment{
		AWSRegion: "us-east-1",
		AWSURL:    "https://s3.example.com",
	}
	up, err := uploader.New(e)
	s.NoError(err)
	s.NotNil(up)
}

func (s *uploaderTestSuite) TestNew_WithHTTP() {
	e := &env.Environment{
		AWSRegion: "us-east-1",
		AWSURL:    "http://localhost:9000",
	}
	up, err := uploader.New(e)
	s.NoError(err)
	s.NotNil(up)
}

func TestUploader(t *testing.T) {
	suite.Run(t, new(uploaderTestSuite))
}
