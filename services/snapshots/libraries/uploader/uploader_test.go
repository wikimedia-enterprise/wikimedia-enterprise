package uploader_test

import (
	"testing"
	"wikimedia-enterprise/services/snapshots/config/env"
	"wikimedia-enterprise/services/snapshots/libraries/uploader"

	"github.com/stretchr/testify/suite"
)

type uploaderTestSuite struct {
	suite.Suite
	env *env.Environment
}

func (s *uploaderTestSuite) TestNew() {
	upr, err := uploader.New(s.env)
	s.Assert().NotNil(upr, err)
}

func TestUploader(t *testing.T) {
	for _, testCase := range []*uploaderTestSuite{
		{
			env: &env.Environment{},
		},
	} {
		suite.Run(t, testCase)
	}
}
