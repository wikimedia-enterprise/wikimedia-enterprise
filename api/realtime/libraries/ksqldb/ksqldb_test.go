package ksqldb_test

import (
	"testing"
	"wikimedia-enterprise/api/realtime/config/env"
	"wikimedia-enterprise/api/realtime/libraries/ksqldb"

	"github.com/stretchr/testify/suite"
)

type ksqldbTestSuite struct {
	suite.Suite
}

func (s *ksqldbTestSuite) TestNew() {
	clt := ksqldb.New(new(env.Environment))
	s.Assert().NotNil(clt)
}

func TestNew(t *testing.T) {
	suite.Run(t, new(ksqldbTestSuite))
}
