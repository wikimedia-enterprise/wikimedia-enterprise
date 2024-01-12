package s3api_test

import (
	"testing"
	"wikimedia-enterprise/services/snapshots/config/env"
	"wikimedia-enterprise/services/snapshots/libraries/s3api"

	"github.com/stretchr/testify/suite"
)

type s3TestSuite struct {
	suite.Suite
	env *env.Environment
}

func (s *s3TestSuite) TestNewS3() {
	s3 := s3api.New(s.env)
	s.Assert().NotNil(s3)
}

func TestS3(t *testing.T) {
	for _, testCase := range []*s3TestSuite{
		{
			env: &env.Environment{},
		},
	} {
		suite.Run(t, testCase)
	}
}
