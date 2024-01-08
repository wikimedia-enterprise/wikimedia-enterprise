package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type keyTestSuite struct {
	suite.Suite
	key *Key
}

func (s *keyTestSuite) SetupTest() {
	s.key = &Key{
		Identifier: "test",
		Type:       "article",
	}
}

func (s *keyTestSuite) TestNewKeySchema() {
	sch, err := NewKeySchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.key)
	s.Assert().NoError(err)

	key := new(Key)
	s.Assert().NoError(avro.Unmarshal(sch, data, key))
	s.Assert().Equal(s.key.Identifier, key.Identifier)
	s.Assert().Equal(s.key.Type, key.Type)
}

func (s *keyTestSuite) TestNewKey() {
	key := NewKey(s.key.Identifier, s.key.Type)
	s.Assert().Equal(s.key.Identifier, key.Identifier)
	s.Assert().Equal(s.key.Type, key.Type)
}

func TestKey(t *testing.T) {
	suite.Run(t, new(keyTestSuite))
}
