package schema

import (
	"encoding/binary"
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

const unmarshalSchema = `{
	"type": "record",
	"name": "Key",
	"namespace": "wikimedia_enterprise.general.schema",
	"fields": [
		{
			"name": "identifier",
			"type": "string"
		},
		{
			"name": "type",
			"type": "string"
		}
	]
}`

type unmarshalValue struct {
	Identifier string `json:"identifier" avro:"identifier"`
	Type       string `json:"entity" avro:"type"`
}

type unmarshalTestSuite struct {
	suite.Suite
	data []byte
	sch  avro.Schema
	v    *unmarshalValue
}

func (s *unmarshalTestSuite) SetupSuite() {
	var err error
	s.sch, err = avro.Parse(unmarshalSchema)
	s.Assert().NoError(err)
}

func (s *unmarshalTestSuite) SetupTest() {
	data, err := avro.Marshal(s.sch, s.v)
	s.Assert().NoError(err)

	s.data = append(append(append([]byte{}, byte(0)), make([]byte, 4)...), data...)
}

func (s *unmarshalTestSuite) TestUnmarshal() {
	v := new(unmarshalValue)

	s.Assert().NoError(Unmarshal(s.sch, s.data, v, nil))
	s.Assert().Equal(*s.v, *v)
}

func (s *unmarshalTestSuite) TestUnmarshalMagicByteErr() {
	s.data[0] = byte(1)
	v := new(unmarshalValue)

	s.Assert().Equal(ErrUnknownMagicByte, Unmarshal(s.sch, s.data, v, nil))
}

func (s *unmarshalTestSuite) TestUnmarshalEmptyDataErr() {
	v := new(unmarshalValue)

	s.Assert().Equal(ErrEmptyData, Unmarshal(s.sch, []byte{}, v, nil))
}

func TestUnmarshal(t *testing.T) {
	for _, testCase := range []*unmarshalTestSuite{
		{
			v: &unmarshalValue{
				Identifier: "unique",
				Type:       "new",
			},
		},
		{
			v: &unmarshalValue{
				Identifier: "",
				Type:       "",
			},
		},
	} {
		suite.Run(t, testCase)
	}
}

type getIDTestSuite struct {
	suite.Suite
	id  int
	idb []byte
}

func (s *getIDTestSuite) SetupSuite() {
	s.idb = make([]byte, 4)
	binary.BigEndian.PutUint32(s.idb, uint32(s.id))
}

func (s *getIDTestSuite) TestGetID() {
	data := append(append(append([]byte{}, byte(0)), s.idb...), make([]byte, 5)...)
	s.Assert().Equal(s.id, GetID(data))
}

func TestGetID(t *testing.T) {
	for _, testCase := range []*getIDTestSuite{
		{
			id: 10,
		},
		{
			id: 99,
		},
		{
			id: 0,
		},
	} {
		suite.Run(t, testCase)
	}
}
