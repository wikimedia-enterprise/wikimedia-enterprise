package articles

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"slices"
	"strings"
	"time"
	"wikimedia-enterprise/api/realtime/libraries/resolver"

	"github.com/gin-gonic/gin"
)

// NewModel returns a model instance.
func NewModel(maxParts int, nPartitions int) *Model {
	m := new(Model)
	partitionsPerPart := nPartitions / maxParts

	for _, part := range m.Parts {
		for i := part * partitionsPerPart; i < (part+1)*partitionsPerPart; i++ {
			m.partitions = append(m.partitions, i)
		}
	}

	return m
}

// Model model for incoming requests. Time format is RFC 3339 (2019-10-12T07:20:50.52Z) for incoming requests for Since and SincePerPartition.
type Model struct {
	Since             time.Time                `form:"since" json:"since"`
	Fields            []string                 `form:"fields" json:"fields" binding:"max=255,dive,min=1,max=255"`
	Filters           []resolver.RequestFilter `form:"filters" json:"filters" binding:"max=255,dive,max=2"`
	Parts             []int                    `form:"parts" json:"parts"`
	Offsets           map[int]int64            `form:"offsets" json:"offsets"`
	SincePerPartition map[int]time.Time        `form:"since_per_partition" json:"since_per_partition"`
	partitions        []int
}

// SetPartitions sets the partitions for the model.
func (m *Model) SetPartitions(maxParts int, nPartitions int) {
	partitionsPerPart := nPartitions / maxParts

	for _, part := range m.Parts {
		for i := part * partitionsPerPart; i < (part+1)*partitionsPerPart; i++ {
			m.partitions = append(m.partitions, i)
		}
	}
}

// PartitionsContain checks if the partitions for the model contain the given partition.
func (m *Model) PartitionsContain(ptn int) bool {
	return slices.Contains(m.partitions, ptn)
}

// Parse validates the input request and sets the partitions.
// For example, if we have 10 parts and 50 partitions, each part corresponds to 5 partitions. If a query comes in for part 1,
// `mdl.partitions` will be set to partitions 5, 6, 7, 8 and 9.
func (mdl *Model) Parse(gcx *gin.Context, r *resolver.Resolver, maxParts int, numPartitions int) error {

	if err := gcx.ShouldBind(mdl); err != nil && err != io.EOF {
		return err
	}

	if err := gcx.ShouldBindJSON(mdl); err != nil && err != io.EOF {
		return err
	}

	for _, field := range mdl.Fields {
		name := strings.TrimSuffix(field, ".*")

		if !r.HasField(name) {
			return fmt.Errorf("field with name '%s' not found", field)
		}

		if strings.HasSuffix(field, ".*") {
			continue
		}

		if r.HasSlice(name) || r.HasStruct(name) {
			return fmt.Errorf("use wildcard '%s.*' to select all fields from the object", field)
		}
	}

	for _, filter := range mdl.Filters {
		if !r.HasField(filter.Field) {
			return fmt.Errorf("there is no such field '%s' to filter", filter.Field)
		}

		if r.HasSlice(filter.Field) || r.HasStruct(filter.Field) {
			return fmt.Errorf("can't filter on a '%s' object", filter.Field)
		}

		ffd := r.GetField(filter.Field)
		rfv := reflect.ValueOf(filter.Value)

		if !rfv.Type().ConvertibleTo(ffd.Value.Type()) {
			return fmt.Errorf("the filter '%s' should not be of type '%v'", filter.Field, rfv.Kind())
		}
	}

	// do not allow future `since`
	now := time.Now().UTC()
	if !mdl.Since.IsZero() && mdl.Since.After(now) {
		return fmt.Errorf("since time %s cannot be in the future", mdl.Since.Format(time.RFC3339))
	}

	// do not allow future `since_per_partition`
	for ptn, ts := range mdl.SincePerPartition {
		if ts.After(now) {
			return fmt.Errorf("since_per_partition for partition %d has future time %s", ptn, ts.Format(time.RFC3339))
		}
	}

	if len(mdl.Parts) > 0 {
		// Validate bounds for Parts
		if len(mdl.Parts) > maxParts {
			return fmt.Errorf("the query param parts has length %d, max allowed length %d", len(mdl.Parts), maxParts)
		}

		for _, prt := range mdl.Parts {
			if prt < 0 || prt > maxParts-1 {
				return fmt.Errorf("the part value is %d, allowed value should be in the range [0,%d]", prt, maxParts-1)
			}
		}

		if !mdl.Since.IsZero() {
			return errors.New("for parallel consumption, specify either offsets or time-offsets (since_per_partition) parameter")
		}
	}

	// Resolve partitions from Parts
	mdl.SetPartitions(maxParts, numPartitions)

	if len(mdl.Offsets) != 0 {
		// Validate bounds for Offsets
		if len(mdl.Offsets) > numPartitions {
			return fmt.Errorf("the query param offsets has length %d, max allowed length %d", len(mdl.Offsets), numPartitions)
		}

		// Pick relevant offsets related to the partitions resolved
		for ptn := range mdl.Offsets {
			if !mdl.PartitionsContain(ptn) {
				delete(mdl.Offsets, ptn)
			}
		}
	}

	if len(mdl.SincePerPartition) != 0 {
		// Validate bounds for SincePerPartition
		if len(mdl.SincePerPartition) > numPartitions {
			return fmt.Errorf("the query param since_per_partition has length %d, max allowed length %d", len(mdl.SincePerPartition), numPartitions)
		}

		// Pick relevant since_per_partition related to the partitions resolved
		for ptn := range mdl.SincePerPartition {
			if !mdl.PartitionsContain(ptn) {
				delete(mdl.SincePerPartition, ptn)
			}
		}
	}

	// Query with offset and time-offset for same partion not allowed
	if len(mdl.Offsets) != 0 && len(mdl.SincePerPartition) != 0 {
		for ptn := range mdl.Offsets {
			if _, ok := mdl.SincePerPartition[ptn]; ok {
				return errors.New("for a specific partion, specify either offset or time-offset (since_per_partition), not both")
			}
		}
	}

	if len(mdl.Parts) == 0 {
		// No parts requested, we will query all of them.
		for i := 0; i < numPartitions; i++ {
			mdl.partitions = append(mdl.partitions, i)
		}
	}

	return nil
}

// DistributePartitions distributes partitions evenly across num slices.
func (mdl *Model) DistributePartitions(num int) [][]int {
	numPartitions := len(mdl.partitions)
	div := numPartitions / num
	rem := numPartitions % num
	result := make([][]int, 0, num)

	start := 0

	for i := 0; i < num; i++ {
		end := start + div
		if i < rem {
			end++
		}
		result = append(result, mdl.partitions[start:end])
		start = end
	}

	return result
}
