package builder

import (
	"wikimedia-enterprise/services/eventstream-listener/submodules/schema"
)

// VersionBuilder implements the builder pattern for the version schema.
type VersionBuilder struct {
	version *schema.Version
}

// NewVersionBuilder creates a new builder for the version schema.
func NewVersionBuilder() *VersionBuilder {
	return &VersionBuilder{version: new(schema.Version)}
}

// Identifier adds the specified ID to the version schema.
func (vb *VersionBuilder) Identifier(identifier int) *VersionBuilder {
	vb.version.Identifier = identifier
	return vb
}

// NumberOfCharacters adds number of character in current version to the version schema.
func (vb *VersionBuilder) NumberOfCharacters(content string) *VersionBuilder {
	vb.version.NumberOfCharacters = len([]rune(content))
	return vb
}

// Comment adds the specified comment to the version schema.
func (vb *VersionBuilder) Comment(comment string) *VersionBuilder {
	vb.version.Comment = comment
	return vb
}

// Tags adds a list of tags to the version schema.
func (vb *VersionBuilder) Tags(tags []string) *VersionBuilder {
	vb.version.Tags = tags
	return vb
}

// IsMinorEdit sets the IsMinorEdit boolean flag for the version schema.
func (vb *VersionBuilder) IsMinorEdit(isMinorEdit bool) *VersionBuilder {
	vb.version.IsMinorEdit = isMinorEdit
	return vb
}

// IsFlaggedStable sets the IsFlaggedStable boolean flag for the version schema.
func (vb *VersionBuilder) IsFlaggedStable(isFlaggedStable bool) *VersionBuilder {
	vb.version.IsFlaggedStable = isFlaggedStable
	return vb
}

// Scores sets the given scores to the version schema.
func (vb *VersionBuilder) Scores(scores *schema.Scores) *VersionBuilder {
	vb.version.Scores = scores
	return vb
}

// Editor sets the given editor to the version schema.
func (vb *VersionBuilder) Editor(editor *schema.Editor) *VersionBuilder {
	vb.version.Editor = editor
	return vb
}

// Size adds the version size (in bytes) to the version schema.
func (vb *VersionBuilder) Size(size int) *VersionBuilder {
	vb.version.Size = &schema.Size{
		UnitText: "B",
		Value:    float64(size),
	}

	return vb
}

// Build returns the version schema.
func (vb *VersionBuilder) Build() *schema.Version {
	return vb.version
}
