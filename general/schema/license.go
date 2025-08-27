package schema

import "github.com/hamba/avro/v2"

// Content license, default values.
const (
	LicenseIdentifier = "CC-BY-SA-4.0"
	LicenseName       = "Creative Commons Attribution-ShareAlike License 4.0"
	LicenseURL        = "https://creativecommons.org/licenses/by-sa/4.0/"
)

// ConfigLicense schema configuration for License.
var ConfigLicense = &Config{
	Type: ConfigTypeValue,
	Name: "License",
	Schema: `{
		"type": "record",
		"name": "License",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "name",
				"type": "string"
			},
			{
				"name": "identifier",
				"type": "string"
			},
			{
				"name": "url",
				"type": "string"
			}
		]
	}`,
	Reflection: License{},
}

// NewLicenseSchema create new license avro schema.
func NewLicenseSchema() (avro.Schema, error) {
	return New(ConfigLicense)
}

// NewLicense create new license for an article.
func NewLicense() *License {
	return &License{
		Name:       LicenseName,
		Identifier: LicenseIdentifier,
		URL:        LicenseURL,
	}
}

// License representation according to https://schema.org/license.
type License struct {
	Name       string `json:"name,omitempty" avro:"name"`
	Identifier string `json:"identifier,omitempty" avro:"identifier"`
	URL        string `json:"url,omitempty" avro:"url"`
}
