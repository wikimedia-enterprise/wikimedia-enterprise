package config

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type configTestSuite struct {
	suite.Suite
	cfg API
}

func (s *configTestSuite) SetupSuite() {
	cfg, err := New()
	s.Assert().NoError(err)

	s.cfg = cfg
}

func (s *configTestSuite) TestGetProjects() {
	s.Assert().NotZero(len(s.cfg.GetProjects()))
}

func (s *configTestSuite) TestGetLanguage() {
	s.Assert().Equal("en", s.cfg.GetLanguage("enwiki"))
}

func (s *configTestSuite) TestGetNamespaces() {
	s.Assert().NotEmpty(s.cfg.GetNamespaces())
}

func (s *configTestSuite) TestGetPartitions() {
	s.Assert().Equal([]int{0}, s.cfg.GetPartitions("afwikibooks", 0))
}

func TestConfig(t *testing.T) {
	suite.Run(t, new(configTestSuite))
}
