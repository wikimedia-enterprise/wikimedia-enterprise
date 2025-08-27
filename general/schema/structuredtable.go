package schema

import (
	"encoding/json"
	"fmt"

	"github.com/hamba/avro/v2"
)

var ConfigAvroStructuredTableCell = &Config{
	Type: ConfigTypeValue,
	Name: "AvroStructuredTableCell",
	Schema: `{
		"type": "record",
		"name": "AvroStructuredTableCell",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{ "name": "value", "type": "string", "default": "" },
			{ "name": "nested_table", "type": "string", "default": "" }
		]
	}`,
	References: []*Config{},
	Reflection: AvroStructuredTableCell{},
}

// ConfigAvroStructuredTable is the main configuration for the Avro schema.
var ConfigAvroStructuredTable = &Config{
	Type: ConfigTypeValue,
	Name: "AvroStructuredTable",
	Schema: `{
        "type": "record",
        "name": "AvroStructuredTable",
        "namespace": "wikimedia_enterprise.general.schema",
        "fields": [
            {
                "name": "identifier",
                "type": "string",
                "default": ""
            },
            {
                "name": "headers",
                "type": {
                    "type": "array",
                    "items": {
                        "type": "array",
                        "items": "AvroStructuredTableCell"
                    }
                },
                "default": []
            },
            {
                "name": "rows",
                "type": {
                    "type": "array",
                    "items": {
                        "type": "array",
                        "items": "AvroStructuredTableCell"
                    }
                },
                "default": []
            },
            {
                "name": "footers",
                "type": [
					"null",
					{
						"type": "array",
						"items": {
							"type": "array",
							"items": "AvroStructuredTableCell"
                    	}
                    }

				],
                "default": null
            },
            {
                "name": "confidence_score",
                "type": [
                    "null",
                    "double"
                ],
                "default": null
            }
        ]
    }`,
	References: []*Config{ConfigAvroStructuredTableCell},
	Reflection: AvroStructuredTable{},
}

// NewStructuredTablesSchema creates the new Avro schema from the Avro definition.
func NewStructuredTablesSchema() (avro.Schema, error) {
	return New(ConfigAvroStructuredTable)
}

// AvroStructuredTable is the "flat" representation for Avro.
type AvroStructuredTable struct {
	Identifer       string                       `json:"identifier,omitempty" avro:"identifier"`
	Headers         [][]*AvroStructuredTableCell `json:"headers,omitempty" avro:"headers"`
	Rows            [][]*AvroStructuredTableCell `json:"rows,omitempty" avro:"rows"`
	Footers         [][]*AvroStructuredTableCell `json:"footers,omitempty" avro:"footers"`
	ConfidenceScore *float64                     `json:"confidence_score,omitempty" avro:"confidence_score"`
}

// AvroStructuredTableCell is the "flat" representation for Avro.
type AvroStructuredTableCell struct {
	Value       string `json:"value,omitempty" avro:"value"`
	NestedTable string `json:"nested_table,omitempty" avro:"nested_table"`
}

// StructuredTable is the rich, json representation with true recursion.
type StructuredTable struct {
	Identifier      string                   `json:"identifier,omitempty"`
	Headers         [][]*StructuredTableCell `json:"headers,omitempty"`
	Rows            [][]*StructuredTableCell `json:"rows,omitempty"`
	Footers         [][]*StructuredTableCell `json:"footers,omitempty"`
	ConfidenceScore *float64                 `json:"confidence_score,omitempty"`
}

// StructuredTableCell is the rich, json representation with true recursion.
type StructuredTableCell struct {
	Value       string           `json:"value,omitempty"`
	NestedTable *StructuredTable `json:"nested_table,omitempty"`
}

func (t *StructuredTable) ToAvroStruct() (*AvroStructuredTable, error) {
	avroTable := &AvroStructuredTable{
		Identifer:       t.Identifier,
		ConfidenceScore: t.ConfidenceScore,
	}

	// Convert Headers: Iterate over rows and then rows within each row
	avroHeaders := make([][]*AvroStructuredTableCell, len(t.Headers))
	for i, row := range t.Headers {
		avroRow := make([]*AvroStructuredTableCell, len(row))
		for j, h := range row {
			avroHeader, err := h.ToAvroStruct()
			if err != nil {
				return nil, fmt.Errorf("failed to convert header cell to Avro struct at row %d, col %d: %w", i, j, err)
			}
			avroRow[j] = avroHeader
		}
		avroHeaders[i] = avroRow
	}
	avroTable.Headers = avroHeaders

	avroData := make([][]*AvroStructuredTableCell, len(t.Rows))
	for i, row := range t.Rows {
		avroRow := make([]*AvroStructuredTableCell, len(row))
		for j, cell := range row {
			avroCell, err := cell.ToAvroStruct()
			if err != nil {
				return nil, fmt.Errorf("failed to convert cell to Avro struct at row %d, col %d: %w", i, j, err)
			}
			avroRow[j] = avroCell
		}
		avroData[i] = avroRow
	}
	avroTable.Rows = avroData

	avroFooters := make([][]*AvroStructuredTableCell, len(t.Footers))
	for i, row := range t.Footers {
		avroRow := make([]*AvroStructuredTableCell, len(row))
		for j, f := range row {
			avroFooter, err := f.ToAvroStruct()
			if err != nil {
				return nil, fmt.Errorf("failed to convert footer cell to Avro struct at row %d, col %d: %w", i, j, err)
			}
			avroRow[j] = avroFooter
		}
		avroFooters[i] = avroRow
	}
	avroTable.Footers = avroFooters

	return avroTable, nil
}

// ToJsonStruct converts the Avro-safe table to json version.
func (t *AvroStructuredTable) ToJsonStruct() (*StructuredTable, error) {
	richTable := &StructuredTable{
		Identifier:      t.Identifer,
		ConfidenceScore: t.ConfidenceScore,
	}

	richHeaders := make([][]*StructuredTableCell, len(t.Headers))
	for i, row := range t.Headers {
		richRow := make([]*StructuredTableCell, len(row))
		for j, h := range row {
			richHeader, err := h.ToJsonStruct()
			if err != nil {
				return nil, fmt.Errorf("failed to convert Avro header cell to JSON struct at row %d, col %d: %w", i, j, err)
			}
			richRow[j] = richHeader
		}
		richHeaders[i] = richRow
	}
	richTable.Headers = richHeaders

	richData := make([][]*StructuredTableCell, len(t.Rows))
	for i, row := range t.Rows {
		richRow := make([]*StructuredTableCell, len(row))
		for j, cell := range row {
			richCell, err := cell.ToJsonStruct()
			if err != nil {
				return nil, fmt.Errorf("failed to convert Avro cell to JSON struct at row %d, col %d: %w", i, j, err)
			}
			richRow[j] = richCell
		}
		richData[i] = richRow
	}
	richTable.Rows = richData

	richFooters := make([][]*StructuredTableCell, len(t.Footers))
	for i, row := range t.Footers {
		richRow := make([]*StructuredTableCell, len(row))
		for j, f := range row {
			richFooter, err := f.ToJsonStruct()
			if err != nil {
				return nil, fmt.Errorf("failed to convert Avro footer cell to JSON struct at row %d, col %d: %w", i, j, err)
			}
			richRow[j] = richFooter
		}
		richFooters[i] = richRow
	}
	richTable.Footers = richFooters

	return richTable, nil
}

// ToAvroStruct converts the table cell to the Avro-safe version.
func (c *StructuredTableCell) ToAvroStruct() (*AvroStructuredTableCell, error) {
	avroCell := &AvroStructuredTableCell{
		Value:       c.Value,
		NestedTable: "",
	}
	if c.NestedTable != nil {
		nested_tableData, err := json.Marshal(c.NestedTable)
		if err != nil {
			return nil, err
		}
		avroCell.NestedTable = string(nested_tableData)
	}
	return avroCell, nil
}

// ToJsonStruct converts the Avro-safe cell to the json version.
func (c *AvroStructuredTableCell) ToJsonStruct() (*StructuredTableCell, error) {
	richCell := &StructuredTableCell{
		Value: c.Value,
	}
	if c.NestedTable != "" {
		nestedTable := new(StructuredTable)
		if err := json.Unmarshal([]byte(c.NestedTable), nestedTable); err != nil {
			return nil, err
		}
		richCell.NestedTable = nestedTable
	}
	return richCell, nil
}
