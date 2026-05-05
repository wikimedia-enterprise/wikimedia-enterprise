package builder

import (
	"time"
	"wikimedia-enterprise/services/structured-contents/submodules/schema"
)

// StructuredBuilder follows the builder pattern for structured schema.
type StructuredBuilder struct {
	article *schema.AvroStructured
}

// NewStructuredBuilder creates an structured contents builder.
func NewStructuredBuilder() *StructuredBuilder {
	return &StructuredBuilder{
		article: new(schema.AvroStructured),
	}
}

// Name sets the name to the structured schema.
func (sb *StructuredBuilder) Name(name string) *StructuredBuilder {
	sb.article.Name = name
	return sb
}

// Identifier adds an ID to the structured schema.
func (sb *StructuredBuilder) Identifier(identifier int) *StructuredBuilder {
	sb.article.Identifier = identifier
	return sb
}

// Abstract sets the abstract field to the structured schema.
func (sb *StructuredBuilder) Abstract(abstract string) *StructuredBuilder {
	sb.article.Abstract = abstract
	return sb
}

// Version adds version to the structured schema.
func (sb *StructuredBuilder) Version(version *schema.Version) *StructuredBuilder {
	sb.article.Version = version
	return sb
}

// Event adds the event to the structured schema.
func (sb *StructuredBuilder) Event(event *schema.Event) *StructuredBuilder {
	sb.article.Event = event
	return sb
}

// URL adds a URL to the structured schema.
func (sb *StructuredBuilder) URL(url string) *StructuredBuilder {
	sb.article.URL = url
	return sb
}

// DateCreated adds the date created timestamp to the structured schema.
func (sb *StructuredBuilder) DateCreated(dateCreated *time.Time) *StructuredBuilder {
	sb.article.DateCreated = dateCreated
	return sb
}

// DateModified adds the date modified timestamp to the structured schema.
func (sb *StructuredBuilder) DateModified(dateModified *time.Time) *StructuredBuilder {
	sb.article.DateModified = dateModified

	return sb
}

// MainEntity adds the main wikidata entity to the structured schema.
func (sb *StructuredBuilder) MainEntity(entity *schema.Entity) *StructuredBuilder {
	sb.article.MainEntity = entity

	return sb
}

// AdditionalEntities adds any additional entities to the structured schema.
func (sb *StructuredBuilder) AdditionalEntities(entities []*schema.Entity) *StructuredBuilder {
	sb.article.AdditionalEntities = entities

	return sb
}

// IsPartOf adds the project to the structured schema.
func (sb *StructuredBuilder) IsPartOf(isPartOf *schema.Project) *StructuredBuilder {
	sb.article.IsPartOf = isPartOf
	return sb
}

// InLanguage adds the language to the structured schema.
func (sb *StructuredBuilder) InLanguage(inLanguage *schema.Language) *StructuredBuilder {
	sb.article.InLanguage = inLanguage
	return sb
}

// License adds the given licenses to the structured schema.
func (sb *StructuredBuilder) License(license []*schema.License) *StructuredBuilder {
	sb.article.License = license
	return sb
}

// Image adds image to the structured schema.
func (sb *StructuredBuilder) Image(img *schema.Image) *StructuredBuilder {
	sb.article.Image = img

	return sb
}

// Description sets the description field to the structured schema.
func (sb *StructuredBuilder) Description(description string) *StructuredBuilder {
	sb.article.Description = description
	return sb
}

// Infoboxes sets the infoboxes field to the structured schema.
func (sb *StructuredBuilder) Infoboxes(hasParts []*schema.AvroHasParts) *StructuredBuilder {
	sb.article.Infoboxes = hasParts
	return sb
}

// Section adds the sections to the structured content.
func (sb *StructuredBuilder) Sections(hasParts []*schema.AvroHasParts) *StructuredBuilder {
	sb.article.Sections = hasParts
	return sb
}

// References adds the references to the structured content.
func (sb *StructuredBuilder) References(refs []*schema.StructuredReference) *StructuredBuilder {
	sb.article.References = refs
	return sb
}

// Visibility adds the visibility object to the structured schema.
func (sb *StructuredBuilder) Visibility(visibility *schema.Visibility) *StructuredBuilder {
	sb.article.Visibility = visibility
	return sb
}

// Build returns the created structured instance.
func (sb *StructuredBuilder) Build() *schema.AvroStructured {
	return sb.article
}
