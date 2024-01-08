package builder

import (
	"wikimedia-enterprise/general/schema"
)

// DeltaBuilder follows the builder pattern for the Delta schema.
type DeltaBuilder struct {
	delta *schema.Delta
	prev  FrequencyTable
	cur   FrequencyTable
}

// NewDeltaBuilder creates a delta builder.
func NewDeltaBuilder() *DeltaBuilder {
	return &DeltaBuilder{
		delta: new(schema.Delta),
	}
}

// Previous adds parent revision tokens as frequency table.
func (d *DeltaBuilder) Previous(tokens []string) *DeltaBuilder {
	d.prev = NewFrequencyTableBuilder().
		Tokens(tokens).
		Build()
	return d
}

// Current adds current revision tokens as frequency table.
func (d *DeltaBuilder) Current(tokens []string) *DeltaBuilder {
	d.cur = NewFrequencyTableBuilder().
		Tokens(tokens).
		Build()
	return d
}

// Build returns the created Delta instance according to schema.
func (d *DeltaBuilder) Build() *schema.Delta {
	tbl := map[string]int{}
	propDelta := map[string]float64{}

	for item, newCount := range d.cur {
		oldCount, exists := d.prev[item]

		if !exists {
			oldCount = 0
		}

		if newCount != oldCount {
			tbl[item] = newCount - oldCount
		}
	}

	for item := range d.prev {
		_, exists := d.cur[item]

		if !exists {
			tbl[item] = d.prev[item] * -1
		}
	}

	for token := range tbl {
		oldDletaValue, exists := d.prev[token]

		if !exists {
			oldDletaValue = 0
		}

		if tbl[token] > 0 {
			propDelta[token] = float64(tbl[token]) / float64(oldDletaValue+1)
			d.delta.Increase += tbl[token]
		} else {
			propDelta[token] = float64(tbl[token]) / float64(oldDletaValue)
		}

		if tbl[token] < 0 {
			d.delta.Decrease += tbl[token]
		}

		d.delta.Sum += tbl[token]
	}

	for item := range propDelta {
		if propDelta[item] > 0 {
			d.delta.ProportionalIncrease += propDelta[item]
		}

		if propDelta[item] < 0 {
			d.delta.ProportionalDecrease += propDelta[item]
		}
	}

	return d.delta
}
