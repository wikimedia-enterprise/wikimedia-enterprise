package schema

import "github.com/hamba/avro/v2"

// ReferenceDetails schema configuration for references in reference risk.
var ConfigReferenceDetails = &Config{
	Type: ConfigTypeValue,
	Name: "ReferenceDetails",
	Schema: `{
        "type": "record",
        "name": "ReferenceDetails",
        "namespace": "wikimedia_enterprise.general.schema",
        "fields": [
            {
                "name": "url",
                "type": "string"
            },
            {
                "name": "domain_name",
                "type": "string"
            },
            {
                "name": "domain_metadata",
                "type": ["null", "DomainMetadata"],
                "default": null
            }
        ]
    }`,
	References: []*Config{
		ConfigDomainMetadata,
	},
	Reflection: ReferenceDetails{},
}

// NewReferenceDetails creates new reference details avro schema.
func NewReferenceDetails() (avro.Schema, error) {
	return New(ConfigReferenceDetails)
}

// ReferenceDetails represents a reference, which includes its URL, associated domain name, and additional domain metadata.
type ReferenceDetails struct {
	URL            string          `json:"url" avro:"url"`
	DomainName     string          `json:"domain_name,omitempty" avro:"domain_name"`
	DomainMetadata *DomainMetadata `json:"domain_metadata,omitempty" avro:"domain_metadata"`
}
