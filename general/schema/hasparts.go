package schema

import (
	"encoding/json"
	"strings"

	"github.com/hamba/avro/v2"
)

// ConfigHasParts schema configuration for HasParts.
var ConfigAvroHasParts = &Config{
	Type: ConfigTypeValue,
	Name: "AvroHasParts",
	Schema: `{
  "type": "record",
  "name": "AvroHasParts",
  "namespace": "wikimedia_enterprise.general.schema",
  "fields": [
    {
      "name": "type",
      "type": "string",
      "default": ""
    },
    {
      "name": "name",
      "type": "string",
      "default": ""
    },
    {
      "name": "value",
      "type": "string",
      "default": ""
    },
    {
      "name": "values",
      "type": {
        "type": "array",
        "items": {
          "type": "string"
        }
      }
    },
    {
      "name": "has_parts",
      "type": "string",
      "default": ""
	},
    {
      "name": "images",
      "type":{
            "type": "array",
            "items": {
              "type": "Image"
            }
          
        }
    },
    {
      "name": "links",
      "type":{
            "type": "array",
            "items": {
              "type": "Link"
            }
          
        }
    },
	{
      "name": "citations",
      "type":[ "null",
	  		{
				"type": "array",
				"items": {
				"type": "StructuredCitation"
				}
			}
        ],
		"default": null
    },
	{
		"name": "table_references",
		"type": [
			"null",
			{
				"type": "array",
				"items": {
					"type": "StructuredTableRef"
				}
			}
		],
		"default": null
	}
  ]
}`,
	References: []*Config{
		ConfigLink,
		ConfigImage,
		ConfigStructuredCitation,
		ConfigStructuredTableRef,
	},
	Reflection: AvroHasParts{},
}

// NewHasPartsSchema creates new article image avro schema.
func NewHasPartsSchema() (avro.Schema, error) {
	return New(ConfigAvroHasParts)
}

// AvroHasParts json representation of HasParts structure.
type AvroHasParts struct {
	Type            string                `json:"type,omitempty" avro:"type"`
	Name            string                `json:"name,omitempty" avro:"name"`
	Value           string                `json:"value,omitempty" avro:"value"`
	Values          []string              `json:"values,omitempty" avro:"values"`
	HasParts        string                `json:"has_parts,omitempty" avro:"has_parts"`
	Images          []*Image              `json:"images,omitempty" avro:"images"`
	Links           []*Link               `json:"links,omitempty" avro:"links"`
	Citations       []*StructuredCitation `json:"citations,omitempty" avro:"citations"`
	TableReferences []*StructuredTableRef `json:"table_references,omitempty" avro:"table_references"`
}

// HasParts json representation of HasParts structure.
type HasParts struct {
	Type            string                `json:"type,omitempty"`
	Name            string                `json:"name,omitempty"`
	Value           string                `json:"value,omitempty"`
	Values          []string              `json:"values,omitempty"`
	HasParts        []*HasParts           `json:"has_parts,omitempty"`
	Images          []*Image              `json:"images,omitempty"`
	Links           []*Link               `json:"links,omitempty"`
	Citations       []*StructuredCitation `json:"citations,omitempty"`
	TableReferences []*StructuredTableRef `json:"table_references,omitempty"`
}

func (h *AvroHasParts) ToJsonStruct() (*HasParts, error) {
	data, err := json.Marshal(h)

	if err != nil {
		return nil, err
	}

	px := new(HasParts)

	err = json.Unmarshal(data, px)

	if err != nil && !strings.Contains(err.Error(), "HasParts") {
		return nil, err
	}

	px2 := new([]*HasParts)

	err = json.Unmarshal([]byte(h.HasParts), px2)

	if err != nil {
		return nil, err
	}
	px.HasParts = *px2

	return px, nil
}

func (h *HasParts) ToAvroStruct() (*AvroHasParts, error) {
	data, err := json.Marshal(h)

	if err != nil {
		return nil, err
	}
	px := new(AvroHasParts)

	err = json.Unmarshal(data, px)

	if err != nil && !strings.Contains(err.Error(), "HasParts") {
		return nil, err
	}

	hp, err := json.Marshal(h.HasParts)

	if err != nil {
		return nil, err
	}

	px.HasParts = string(hp)

	return px, nil
}
