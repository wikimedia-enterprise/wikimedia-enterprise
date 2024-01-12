package builder

type FrequencyTable map[string]int

// FrequencyTableBuilder defines a struct for a builder.
type FrequencyTableBuilder struct {
	tbl FrequencyTable
}

// NewFrequencyTableBuilder creates an instance of FrequencyTableBuilder with an empty frequency table.
func NewFrequencyTableBuilder() *FrequencyTableBuilder {
	return &FrequencyTableBuilder{
		tbl: make(map[string]int),
	}
}

// Tokens creates a frequency table from a slice of word tokens.
func (t *FrequencyTableBuilder) Tokens(tokens []string) *FrequencyTableBuilder {
	for _, name := range tokens {
		t.tbl[name] += 1
	}

	return t
}

// Build creates a frequency table instance.
func (t *FrequencyTableBuilder) Build() FrequencyTable {
	return t.tbl
}
