package schema

import (
	"encoding/binary"
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type marshalTestSuite struct {
	suite.Suite
	id  int
	sch avro.Schema
	v   interface{}
}

func (s *marshalTestSuite) TestMarshal() {
	data, err := Marshal(s.id, s.sch, s.v)

	s.Assert().NoError(err)
	s.Assert().Equal(byte(0), data[0])
	s.Assert().Equal(uint32(s.id), binary.BigEndian.Uint32(data[1:5]))

	expect, err := avro.Marshal(s.sch, s.v)
	s.Assert().NoError(err)
	s.Assert().Equal(expect, data[5:])
}

func TestMarshal(t *testing.T) {
	assert := assert.New(t)

	aSch, err := NewArticleSchema()
	assert.NoError(err)

	vSch, err := NewVersionSchema()
	assert.NoError(err)

	for _, testCase := range []*marshalTestSuite{
		{
			id:  15,
			sch: aSch,
			v: Article{
				Name:       "Main",
				Identifier: 100,
			},
		},
		{
			id:  99,
			sch: aSch,
			v: Article{
				Name:       "New",
				Identifier: 100,
			},
		},
		{
			id:  11,
			sch: vSch,
			v: Version{
				Identifier: 999,
				Comment:    "Super cool comment!",
			},
		},
	} {
		suite.Run(t, testCase)
	}
}
