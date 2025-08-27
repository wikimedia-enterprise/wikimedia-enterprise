package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type structuredTableRefTestSuite struct {
	suite.Suite
	tableRef *StructuredTableRef
}

func (s *structuredTableRefTestSuite) SetupTest() {
	s.tableRef = &StructuredTableRef{
		Identifier: "unique-table-identifier-123",
	}
}

func (s *structuredTableRefTestSuite) TestStructuredTableRefSchema() {
	sch, err := NewStructuredTablesRefSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.tableRef)
	s.Assert().NoError(err)

	unmarshaledRef := new(StructuredTableRef)
	err = avro.Unmarshal(sch, data, unmarshaledRef)
	s.Assert().NoError(err)

	s.Assert().Equal(s.tableRef.Identifier, unmarshaledRef.Identifier)
}

func TestStructuredTableRef(t *testing.T) {
	suite.Run(t, new(structuredTableRefTestSuite))
}
