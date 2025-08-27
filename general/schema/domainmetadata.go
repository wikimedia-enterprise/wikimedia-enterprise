package schema

import "github.com/hamba/avro/v2"

// DomainMetadata schema configuration for domain metadata in reference risk.
var ConfigDomainMetadata = &Config{
	Type: ConfigTypeValue,
	Name: "DomainMetadata",
	Schema: `{
        "namespace": "wikimedia_enterprise.general.schema",
        "name": "DomainMetadata",
        "type": "record",
        "fields": [
            {
                "name": "ps_label_local",
                "type": ["null", "string"],
                "default": null
            },
            {
                "name": "ps_label_enwiki",
                "type": ["null", "string"],
                "default": null
            },
            {
                "name": "survival_ratio",
                "type": "double"
            },
            {
                "name": "page_count",
                "type": "int"
            },
            {
                "name": "editors_count",
                "type": "int"
            }
        ]
    }`,
	Reflection: DomainMetadata{},
}

// NewDomainMetadata creates new domain metadata avro schema.
func NewDomainMetadata() (avro.Schema, error) {
	return New(ConfigDomainMetadata)
}

// DomainMetadata represents additional metadata about a domain.
type DomainMetadata struct {
	PsLabelLocal  *string `json:"ps_label_local,omitempty" avro:"ps_label_local"`
	PsLabelEnwiki *string `json:"ps_label_enwiki,omitempty" avro:"ps_label_enwiki"`
	SurvivalRatio float64 `json:"survival_ratio,omitempty" avro:"survival_ratio"`
	PageCount     int     `json:"page_count,omitempty" avro:"page_count"`
	EditorsCount  int     `json:"editors_count,omitempty" avro:"editors_count"`
}
