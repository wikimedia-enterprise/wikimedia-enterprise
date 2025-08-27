package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type structuredTableTestSuite struct {
	suite.Suite
	table *StructuredTable
}

func (s *structuredTableTestSuite) SetupTest() {
	cs1 := 0.95
	cs2 := 0.88

	nestedTable := &StructuredTable{
		Identifier: "nested-table-002",
		Headers: [][]*StructuredTableCell{
			{
				{Value: "Nested Header 1"},
			},
		},
		Rows: [][]*StructuredTableCell{
			{
				{Value: "Nested Row1-Col1"},
			},
		},
		Footers: [][]*StructuredTableCell{
			{
				{Value: "Nested Footer 1"},
			},
		},
		ConfidenceScore: &cs1,
	}

	s.table = &StructuredTable{
		Identifier: "table-ref-001",
		Headers: [][]*StructuredTableCell{
			{
				{Value: "Header 1"},
				{Value: "Header 2"},
			},
		},
		Rows: [][]*StructuredTableCell{
			{
				{Value: "Row1-Col1"},
				{
					Value:       "This cell has a table",
					NestedTable: nestedTable,
				},
			},
			{
				{Value: "Row2-Col1"},
				{Value: "Row2-Col2"},
			},
		},
		Footers: [][]*StructuredTableCell{
			{
				{Value: "Footer 1"},
				{Value: "Footer 2"},
			},
		},
		ConfidenceScore: &cs2,
	}
}

func (s *structuredTableTestSuite) TestStructuredTableSchema() {
	avroTable, err := s.table.ToAvroStruct()
	s.Assert().NoError(err, "Conversion to Avro-safe struct should not fail")

	sch, err := NewStructuredTablesSchema()
	s.Assert().NoError(err, "Schema creation should not fail")

	data, err := avro.Marshal(sch, avroTable)
	s.Assert().NoError(err, "Avro marshaling should not fail")

	unmarshaledAvroTable := new(AvroStructuredTable)
	err = avro.Unmarshal(sch, data, unmarshaledAvroTable)
	s.Assert().NoError(err, "Avro unmarshaling should not fail")

	unmarshaledTable, err := unmarshaledAvroTable.ToJsonStruct()
	s.Assert().NoError(err, "Conversion to rich struct should not fail")

	// --- Assertions for the main table ---
	s.Assert().Equal(s.table.Identifier, unmarshaledTable.Identifier)
	s.Assert().Equal(s.table.ConfidenceScore, unmarshaledTable.ConfidenceScore)

	s.Assert().Len(unmarshaledTable.Headers, len(s.table.Headers))
	s.Assert().Len(unmarshaledTable.Headers[0], len(s.table.Headers[0]))
	s.Assert().Equal(s.table.Headers[0][0].Value, unmarshaledTable.Headers[0][0].Value)
	s.Assert().Equal(s.table.Headers[0][1].Value, unmarshaledTable.Headers[0][1].Value)

	s.Assert().Len(unmarshaledTable.Rows, 2)
	s.Assert().Len(unmarshaledTable.Rows[0], 2)
	s.Assert().Equal(s.table.Rows[0][0].Value, unmarshaledTable.Rows[0][0].Value)
	s.Assert().Equal(s.table.Rows[1][1].Value, unmarshaledTable.Rows[1][1].Value)

	s.Assert().Len(unmarshaledTable.Footers, len(s.table.Footers))
	s.Assert().Len(unmarshaledTable.Footers[0], len(s.table.Footers[0]))
	s.Assert().Equal(s.table.Footers[0][0].Value, unmarshaledTable.Footers[0][0].Value)
	s.Assert().Equal(s.table.Footers[0][1].Value, unmarshaledTable.Footers[0][1].Value)

	// --- Assertions for the nested table ---
	s.Assert().NotNil(unmarshaledTable.Rows[0][1].NestedTable)
	s.Assert().Equal(s.table.Rows[0][1].NestedTable.Identifier, unmarshaledTable.Rows[0][1].NestedTable.Identifier)
	s.Assert().NotNil(unmarshaledTable.Rows[0][1].NestedTable.ConfidenceScore)
	s.Assert().Equal(*s.table.Rows[0][1].NestedTable.ConfidenceScore, *unmarshaledTable.Rows[0][1].NestedTable.ConfidenceScore)

	s.Assert().Len(unmarshaledTable.Rows[0][1].NestedTable.Headers, 1)
	s.Assert().Len(unmarshaledTable.Rows[0][1].NestedTable.Headers[0], 1)
	s.Assert().Equal(s.table.Rows[0][1].NestedTable.Headers[0][0].Value, unmarshaledTable.Rows[0][1].NestedTable.Headers[0][0].Value)

	s.Assert().Len(unmarshaledTable.Rows[0][1].NestedTable.Rows, 1)
	s.Assert().Len(unmarshaledTable.Rows[0][1].NestedTable.Rows[0], 1)
	s.Assert().Equal(s.table.Rows[0][1].NestedTable.Rows[0][0].Value, unmarshaledTable.Rows[0][1].NestedTable.Rows[0][0].Value)

	s.Assert().Len(unmarshaledTable.Rows[0][1].NestedTable.Footers, 1)
	s.Assert().Len(unmarshaledTable.Rows[0][1].NestedTable.Footers[0], 1)
	s.Assert().Equal(s.table.Rows[0][1].NestedTable.Footers[0][0].Value, unmarshaledTable.Rows[0][1].NestedTable.Footers[0][0].Value)
}

func TestStructuredTable(t *testing.T) {
	suite.Run(t, new(structuredTableTestSuite))
}
