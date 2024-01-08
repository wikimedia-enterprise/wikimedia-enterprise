package schema

import (
	"github.com/hamba/avro/v2"
)

// ConfigMaintenanceTag is the config for maintenance tag avro schema.
var ConfigMaintenanceTags = &Config{
	Type: ConfigTypeValue,
	Name: "MaintenanceTags",
	Schema: `{
		"type": "record",
		"name": "MaintenanceTags",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "citation_needed_count",
				"type": "int",
				"default": 0
			},
			{
				"name": "pov_count",
				"type": "int",
				"default": 0
			},
			{
				"name": "clarification_needed_count",
				"type": "int",
				"default": 0
			},
			{
				"name": "update_count",
				"type": "int",
				"default": 0
			}
		]
	}`,
	Reflection: MaintenanceTags{},
}

// NewMaintenanceTag creates new maintenance tag avro schema.
func NewMaintenanceTagsSchema() (avro.Schema, error) {
	return New(ConfigMaintenanceTags)
}

// MaintenanceTags represents the maintenance tags in the wiki text.
type MaintenanceTags struct {
	CitationNeededCount      int `json:"citation_needed_count,omitempty" avro:"citation_needed_count"`
	PovCount                 int `json:"pov_count,omitempty" avro:"pov_count"`
	ClarificationNeededCount int `json:"clarification_needed_count,omitempty" avro:"clarification_needed_count"`
	UpdateCount              int `json:"update_count,omitempty" avro:"update_count"`
}
