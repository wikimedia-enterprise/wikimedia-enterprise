package schema

import "github.com/hamba/avro/v2"

// ConfigDiff schema configuration for Diff.
var ConfigDiff = &Config{
	Type: ConfigTypeValue,
	Name: "Diff",
	Schema: `{
		"type": "record",
		"name": "Diff",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "longest_new_repeated_character",
				"type": "int"
			},
			{
				"name": "dictionary_words",
				"type": [
					"null",
					"Delta"
				]
			},
			{
				"name": "non_dictionary_words",
				"type": [
					"null",
					"Delta"
				]
			},
			{
				"name": "non_safe_words",
				"type": [
					"null",
					"Delta"
				]
			},
			{
				"name": "informal_words",
				"type": [
					"null",
					"Delta"
				]
			},
			{
				"name": "uppercase_words",
				"type": [
					"null",
					"Delta"
				]
			},
			{
				"name": "sizes",
				"type": [
					"null",
					"Size"
				],
				"default": null
			}
		]
	}`,
	References: []*Config{
		ConfigDelta,
		ConfigSize,
	},
	Reflection: Diff{},
}

// NewDiffSchema creates new diff avro schema.
func NewDiffSchema() (avro.Schema, error) {
	return New(ConfigDiff)
}

// Diff representats the difference between current and previous version.
type Diff struct {
	LongestNewRepeatedCharacter int    `json:"longest_new_repeated_character,omitempty" avro:"longest_new_repeated_character"`
	DictionaryWords             *Delta `json:"dictionary_words,omitempty" avro:"dictionary_words"`
	NonDictionaryWords          *Delta `json:"non_dictionary_words,omitempty" avro:"non_dictionary_words"`
	NonSafeWords                *Delta `json:"non_safe_words,omitempty" avro:"non_safe_words"`
	InformalWords               *Delta `json:"informal_words,omitempty" avro:"informal_words"`
	UppercaseWords              *Delta `json:"uppercase_words,omitempty" avro:"uppercase_words"`
	Size                        *Size  `json:"size,omitempty" avro:"sizes"` // note that there's intentional `sizes` instead of `size` because size is ksqldb keyword
}
