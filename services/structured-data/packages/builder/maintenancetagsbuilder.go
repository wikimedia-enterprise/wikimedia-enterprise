package builder

import (
	"regexp"
	"wikimedia-enterprise/general/schema"
)

var (
	citationNeededRegex = regexp.MustCompile(`(?i){{Citation needed(\|*).*?}}`)
	povRegex            = regexp.MustCompile(`(?i){{POV(\|*).*?}}`)
	needsClarification  = regexp.MustCompile(`(?i){{clarify(\|*).*?}}`)
	updateRegex         = regexp.MustCompile(`(?i){{update(\|*).*?}}`)
)

// MaintenanceTagsBuilder follows the builder pattern for the MaintenanceTags schema
type MaintenanceTagsBuilder struct {
	maintenanceTags *schema.MaintenanceTags
}

// NewMaintenanceTagsBuilder creates a new builder for the MaintenanceTags schema.
func NewMaintenanceTagsBuilder() *MaintenanceTagsBuilder {
	return &MaintenanceTagsBuilder{new(schema.MaintenanceTags)}
}

// CitationNeeded sets the citation count for the MaintenanceTags schema
func (mb *MaintenanceTagsBuilder) CitationNeeded(wiki string) *MaintenanceTagsBuilder {
	mb.maintenanceTags.CitationNeededCount = len(citationNeededRegex.FindAllString(wiki, -1))
	return mb
}

// PovIdentified sets the pov count for the MaintenanceTags schema
func (mb *MaintenanceTagsBuilder) PovIdentified(wiki string) *MaintenanceTagsBuilder {
	mb.maintenanceTags.PovCount = len(povRegex.FindAllString(wiki, -1))
	return mb
}

// ClarificationNeeded sets the clarification count for the MaintenanceTags schema
func (mb *MaintenanceTagsBuilder) ClarificationNeeded(wiki string) *MaintenanceTagsBuilder {
	mb.maintenanceTags.ClarificationNeededCount = len(needsClarification.FindAllString(wiki, -1))
	return mb
}

// Update sets the update count for the MaintenanceTags schema
func (mb *MaintenanceTagsBuilder) Update(wiki string) *MaintenanceTagsBuilder {
	mb.maintenanceTags.UpdateCount = len(updateRegex.FindAllString(wiki, -1))
	return mb
}

// Build returns the built MaintenanceTags schema
func (mb *MaintenanceTagsBuilder) Build() *schema.MaintenanceTags {
	return mb.maintenanceTags
}
