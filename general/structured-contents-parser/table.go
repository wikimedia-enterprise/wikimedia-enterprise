package parser

import (
	"fmt"
)

type SpanCell struct {
	RowSpanRemaining int    `json:"rows,omitempty"`  // Tracks the remaining span for the cell
	ColSpanRemaining int    `json:"cols,omitempty"`  // Tracks the column span for the cell (needed when we have both rowspan and colspan)
	Value            string `json:"value,omitempty"` // The value to insert as a placeholder, if needed
}

type TableWithHeaders struct {
	Headers         [][]Cell `json:"headers"`
	Footers         [][]Cell `json:"footers"` //ToDo, ignored by POC table parser
	Rows            [][]Cell `json:"rows"`
	ConfidenceScore float64  `json:"confidence_score,omitempty"`
}

type TablesWithHeaders []TableWithHeaders

// Cell represents a table cell with a textual value and optionally nested inner tables.
type Cell struct {
	Value string            `json:"value"`           // TODO: Should we make this an interface and allow different types? Maybe put inner tables in here and drop this struct completely? Then remove inner field
	Inner TablesWithHeaders `json:"inner,omitempty"` // Inner tables
}

// RowGetter interface for accessing a specific row.
type RowGetter interface {
	GetRow(isHeader bool, rowIndex int) ([]Cell, error)
}

// CellValueGetter interface for accessing the value of a specific cell.
type CellValueGetter interface {
	Get(rowIndex, columnIndex int) (string, error)
	GetHeader(rowIndex, columnIndex int) (string, error)
}

// InnerTableGetter interface for accessing inner tables.
type InnerTableGetter interface {
	GetInnerTable(rowIndex, columnIndex int) ([]TableWithHeaders, error)
	GetHeaderInnerTable(rowIndex, columnIndex int) ([]TableWithHeaders, error)
}

// BalanceChecker interface for checking if the table is balanced.
type BalanceChecker interface {
	IsBalanced() bool
}

// InnerTablePresenceChecker interface for checking for inner tables.
type InnerTablePresenceChecker interface {
	HasInnerTablesInHeaders() bool
	HasInnerTables() bool
}

// DimensionGetter interface for accessing table dimensions.
type DimensionGetter interface {
	Rows() int
	Columns() int
	HeaderRows() int
	HeaderColumns() int
}

// TableAPI interface combining all the functionalities.
type TableAPI interface {
	RowGetter
	CellValueGetter
	InnerTableGetter
	BalanceChecker
	InnerTablePresenceChecker
	DimensionGetter
}

// GetRow accesses a specific row from Headers or Data, returning Cell slices and an error if out of range.
func (t *TableWithHeaders) GetRow(isHeader bool, rowIndex int) ([]Cell, error) {
	//ToDo: Need footer support?
	if isHeader {
		if rowIndex >= 0 && rowIndex < len(t.Headers) {
			return t.Headers[rowIndex], nil
		}
	} else {
		if rowIndex >= 0 && rowIndex < len(t.Rows) {
			return t.Rows[rowIndex], nil
		}
	}
	return nil, fmt.Errorf("row index out of range")
}

// GetCell accesses a specific cell from Headers or Data, returning a Cell and an error if out of range.
func (t *TableWithHeaders) getCell(isHeader bool, rowIndex, columnIndex int) (*Cell, error) {
	var row []Cell
	var err error
	if row, err = t.GetRow(isHeader, rowIndex); err != nil {
		return nil, err
	}
	if columnIndex >= 0 && columnIndex < len(row) {
		return &row[columnIndex], nil
	}
	return nil, fmt.Errorf("column index out of range")
}

// Get accesses the value of a specific cell in the data section.
func (t *TableWithHeaders) Get(rowIndex, columnIndex int) (string, error) {
	cell, err := t.getCell(false, rowIndex, columnIndex)
	if err != nil {
		return "", err
	}
	return cell.Value, nil
}

// GetHeader accesses the value of a specific cell in the header section.
func (t *TableWithHeaders) GetHeader(rowIndex, columnIndex int) (string, error) {
	cell, err := t.getCell(true, rowIndex, columnIndex)
	if err != nil {
		return "", err
	}

	return cell.Value, nil
}

// getInner Access the inner table of a specific cell
func (t *TableWithHeaders) getInner(isHeader bool, rowIndex, columnIndex int) ([]TableWithHeaders, error) {
	cell, err := t.getCell(isHeader, rowIndex, columnIndex)
	if err != nil {
		return nil, err
	}

	if len(cell.Inner) == 0 {
		return nil, fmt.Errorf("no inner table found")
	}

	return cell.Inner, nil
}

// GetInnerTable Access the inner table of a specific cell in the data object
func (t *TableWithHeaders) GetInnerTable(rowIndex, columnIndex int) ([]TableWithHeaders, error) {
	return t.getInner(false, rowIndex, columnIndex)
}

// GetHeaderInnerTable Access the inner table of a specific cell in the header object
func (t *TableWithHeaders) GetHeaderInnerTable(rowIndex, columnIndex int) ([]TableWithHeaders, error) {
	return t.getInner(true, rowIndex, columnIndex)
}

// HasInnerTablesInHeaders Check for inner tables in Header cells
func (t *TableWithHeaders) HasInnerTablesInHeaders() bool {
	for _, row := range t.Headers {
		for _, cell := range row {
			if len(cell.Inner) > 0 {
				return true
			}
		}
	}
	return false
}

// HasInnerTables Check for inner tables in Data cells
func (t *TableWithHeaders) HasInnerTables() bool {
	for _, row := range t.Rows {
		for _, cell := range row {
			if len(cell.Inner) > 0 {
				return true
			}
		}
	}
	return false
}

// Get the number of rows in the table
func (t *TableWithHeaders) RowCount() int {
	return len(t.Rows)
}

// Get the number of columns in the table
func (t *TableWithHeaders) ColumnCount() int {
	return len(t.Rows[0])
}

// Get the number of header rows in the table
func (t *TableWithHeaders) HeaderRowCount() int {
	return len(t.Headers)
}

// Get the number of header columns in the table
func (t *TableWithHeaders) HeaderColumnCount() int {
	return len(t.Headers[0])
}
