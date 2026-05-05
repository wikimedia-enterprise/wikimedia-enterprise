package parser

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite struct
type TableWithHeadersTestSuite struct {
	suite.Suite
	table TableWithHeaders
}

// SetupTest is run before each test in the suite
func (suite *TableWithHeadersTestSuite) SetupTest() {
	// Initialize your table with headers, data, and the balanced flag here
	suite.table = TableWithHeaders{
		Headers: [][]Cell{
			{{Value: "Header1"}, {Value: "Header2"}},
		},
		Rows: [][]Cell{
			{{Value: "Data1"}, {Value: "Data2"}},
			{{Value: "Data3", Inner: []TableWithHeaders{{ConfidenceScore: 1}}}, {Value: "Data4"}},
		},
	}
}

// TestRowGetter tests the GetRow method
func (suite *TableWithHeadersTestSuite) TestRowGetter() {
	row, err := suite.table.GetRow(false, 0)
	suite.NoError(err)
	suite.Len(row, 2, "there should be two cells in the first data row")

	_, err = suite.table.GetRow(false, 10)
	suite.Error(err, "should return an error for out of range row index")
}

// TestCellValueGetter tests the Get and GetHeader methods
func (suite *TableWithHeadersTestSuite) TestCellValueGetter() {
	value, err := suite.table.Get(0, 1)
	suite.NoError(err)
	suite.Equal("Data2", value, "the value should match the cell value")

	value, err = suite.table.GetHeader(0, 0)
	suite.NoError(err)
	suite.Equal("Header1", value, "the header value should match the cell value")
}

// TestInnerTableGetter tests the GetInnerTable and GetHeaderInnerTable methods
func (suite *TableWithHeadersTestSuite) TestInnerTableGetter() {
	inner, err := suite.table.GetInnerTable(1, 0)
	suite.NoError(err)
	suite.Len(inner, 1, "there should be one inner table in the specified cell")

	_, err = suite.table.GetHeaderInnerTable(0, 0)
	suite.Error(err, "should return an error for header cells without inner tables")
}

// TestInnerTablePresenceChecker tests the HasInnerTablesInHeaders and HasInnerTables methods
func (suite *TableWithHeadersTestSuite) TestInnerTablePresenceChecker() {
	suite.False(suite.table.HasInnerTablesInHeaders(), "there should be no inner tables in headers")
	suite.True(suite.table.HasInnerTables(), "there should be inner tables in data")
}

// TestDimensionGetter tests the Rows, Columns, HeaderRows, and HeaderColumns methods
func (suite *TableWithHeadersTestSuite) TestDimensionGetter() {
	suite.Equal(2, suite.table.RowCount(), "there should be two rows in data")
	suite.Equal(2, suite.table.ColumnCount(), "there should be two columns in data")
	suite.Equal(1, suite.table.HeaderRowCount(), "there should be one header row")
	suite.Equal(2, suite.table.HeaderColumnCount(), "there should be two header columns")
}

// Setup the suite and run the tests
func TestTableWithHeadersTestSuite(t *testing.T) {
	suite.Run(t, new(TableWithHeadersTestSuite))
}
