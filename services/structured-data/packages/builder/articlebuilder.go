package builder

import (
	"fmt"
	"strings" //nolint // using nolint to avoid false negative https://github.com/golangci/golangci-lint/discussions/2214
	"time"
	"wikimedia-enterprise/general/parser"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/general/wmf"
)

const (
	wikidataURL = "https://www.wikidata.org/entity/"
)

// ArticleBuilder follows the builder pattern for article schema.
type ArticleBuilder struct {
	article    *schema.Article
	projectURL string
}

// NewArticleBuilder creates an article builder.
func NewArticleBuilder(projectURL string) *ArticleBuilder {
	return &ArticleBuilder{
		article:    new(schema.Article),
		projectURL: projectURL,
	}
}

// Event adds the event to the article schema.
func (ab *ArticleBuilder) Event(event *schema.Event) *ArticleBuilder {
	ab.article.Event = event
	return ab
}

// Name sets the name to the article schema.
func (ab *ArticleBuilder) Name(name string) *ArticleBuilder {
	ab.article.Name = name
	return ab
}

// Abstract sets the abstract field to the article struct.
func (ab *ArticleBuilder) Abstract(abstract string) *ArticleBuilder {
	ab.article.Abstract = abstract
	return ab
}

// WatchersCount sets the watchers_count to the article schema.
func (ab *ArticleBuilder) WatchersCount(watchers int) *ArticleBuilder {
	ab.article.WatchersCount = watchers
	return ab
}

// Identifier adds an ID to the article schema.
func (ab *ArticleBuilder) Identifier(identifier int) *ArticleBuilder {
	ab.article.Identifier = identifier
	return ab
}

// DatePreviouslyModified adds the date previously modified timestamp to the article schema.
func (ab *ArticleBuilder) DatePreviouslyModified(datePreviouslyModified *time.Time) *ArticleBuilder {
	ab.article.DatePreviouslyModified = datePreviouslyModified
	return ab
}

// DateCreated adds the date created timestamp to the article schema.
func (ab *ArticleBuilder) DateCreated(dateCreated *time.Time) *ArticleBuilder {
	ab.article.DateCreated = dateCreated
	return ab
}

// DateModified adds the date modified timestamp to the article schema.
func (ab *ArticleBuilder) DateModified(dateModified *time.Time) *ArticleBuilder {
	ab.article.DateModified = dateModified
	return ab
}

// URL adds a URL to the article schema.
func (ab *ArticleBuilder) URL(url string) *ArticleBuilder {
	ab.article.URL = url
	return ab
}

// InLanguage adds the language to the article schema.
func (ab *ArticleBuilder) InLanguage(inLanguage *schema.Language) *ArticleBuilder {
	ab.article.InLanguage = inLanguage
	return ab
}

// IsPartOf adds the project to the article schema.
func (ab *ArticleBuilder) IsPartOf(isPartOf *schema.Project) *ArticleBuilder {
	ab.article.IsPartOf = isPartOf
	return ab
}

// Namespace adds the specified namespace to the article schema.
func (ab *ArticleBuilder) Namespace(namespace *schema.Namespace) *ArticleBuilder {
	ab.article.Namespace = namespace
	return ab
}

// Body creates the wikitext and HTML contents for the article.
func (ab *ArticleBuilder) Body(wiki, html string) *ArticleBuilder {
	ab.article.ArticleBody = &schema.ArticleBody{
		WikiText: wiki,
		HTML:     html,
	}

	return ab
}

// License adds the given licenses to the article.
func (ab *ArticleBuilder) License(license ...*schema.License) *ArticleBuilder {
	ab.article.License = license
	return ab
}

// Version adds version to the article schema.
func (ab *ArticleBuilder) Version(version *schema.Version) *ArticleBuilder {
	ab.article.Version = version
	return ab
}

// Version adds version to the article schema.
func (ab *ArticleBuilder) Image(orgImg *wmf.Image, thmb *wmf.Image) *ArticleBuilder {
	if orgImg == nil {
		return ab
	}

	ab.article.Image = &schema.Image{
		ContentUrl: orgImg.Source,
		Width:      orgImg.Width,
		Height:     orgImg.Height,
	}

	if thmb != nil {
		ab.article.Image.Thumbnail = &schema.Thumbnail{
			ContentUrl: thmb.Source,
			Width:      thmb.Width,
			Height:     thmb.Height,
		}
	}

	return ab
}

// PreviousVersion adds previous version to the article schema.
func (ab *ArticleBuilder) PreviousVersion(revID int, content string) *ArticleBuilder {
	ab.article.PreviousVersion = &schema.PreviousVersion{
		Identifier:         revID,
		NumberOfCharacters: len([]rune(content)),
	}

	return ab
}

// VersionIdentifier adds version key identifier to the article schema for future join queries.
func (ab *ArticleBuilder) VersionIdentifier(versionIdentifier string) *ArticleBuilder {
	ab.article.VersionIdentifier = versionIdentifier

	return ab
}

// MainEntity adds the main wikidata entity to the article.
func (ab *ArticleBuilder) MainEntity(identifier string) *ArticleBuilder {
	if len(identifier) != 0 {
		ab.article.MainEntity = &schema.Entity{
			Identifier: identifier,
			URL:        fmt.Sprintf("%s%s", wikidataURL, identifier),
		}
	}

	return ab
}

// AdditionalEntities adds any additional entities to the article.
func (ab *ArticleBuilder) AdditionalEntities(entities map[string]*wmf.WbEntityUsage) *ArticleBuilder {
	for id, entity := range entities { //nolint // using nolint to avoid false negative https://github.com/golangci/golangci-lint/discussions/2214
		ab.article.AdditionalEntities = append(ab.article.AdditionalEntities, &schema.Entity{
			Identifier: id,
			URL:        fmt.Sprintf("%s%s", wikidataURL, id),
			Aspects:    entity.Aspects,
		})
	}

	return ab
}

// Categories adds the non-hidden categories to the article.
func (ab *ArticleBuilder) Categories(categories []*parser.Category) *ArticleBuilder {
	for _, category := range categories {
		ab.article.Categories = append(ab.article.Categories, &schema.Category{
			Name: category.Name,
			URL:  category.URL,
		})
	}

	return ab
}

// Protection adds the given protection levels to the article.
func (ab *ArticleBuilder) Protection(protection []*wmf.Protection) *ArticleBuilder {
	for _, p := range protection { //nolint // using nolint to avoid false negative https://github.com/golangci/golangci-lint/discussions/2214
		ab.article.Protection = append(ab.article.Protection, &schema.Protection{
			Type:   p.Type,
			Level:  p.Level,
			Expiry: p.Expiry,
		})
	}

	return ab
}

// Templates adds the given templates to the article.
func (ab *ArticleBuilder) Templates(templates []*parser.Template) *ArticleBuilder {
	for _, template := range templates { //nolint // using nolint to avoid false negative https://github.com/golangci/golangci-lint/discussions/2214
		ab.article.Templates = append(ab.article.Templates, &schema.Template{
			Name: template.Name,
			URL:  template.URL,
		})
	}

	return ab
}

// Redirects adds the given redirects to the article.
func (ab *ArticleBuilder) Redirects(redirects []*wmf.Redirect) *ArticleBuilder {
	for _, redirect := range redirects { //nolint // using nolint to avoid false negative https://github.com/golangci/golangci-lint/discussions/2214
		ab.article.Redirects = append(ab.article.Redirects, &schema.Redirect{
			Name: redirect.Title,
			URL:  fmt.Sprintf("%s/wiki/%s", ab.projectURL, strings.ReplaceAll(redirect.Title, " ", "_")),
		})
	}

	return ab
}

// Build returns the created article instance.
func (ab *ArticleBuilder) Build() *schema.Article {
	return ab.article
}
