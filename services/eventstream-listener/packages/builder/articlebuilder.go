package builder

import (
	"fmt" //nolint // using nolint to avoid false negative https://github.com/golangci/golangci-lint/discussions/2214
	"time"
	"wikimedia-enterprise/services/eventstream-listener/submodules/schema"
)

const (
	wikidataURL = "https://www.wikidata.org/entity/"
)

// ArticleBuilder follows the builder pattern for article schema.
type ArticleBuilder struct {
	article *schema.Article
}

// NewArticleBuilder creates an article builder.
func NewArticleBuilder() *ArticleBuilder {
	return &ArticleBuilder{
		article: new(schema.Article),
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

// PreviousVersion adds previous version to the article schema.
func (ab *ArticleBuilder) PreviousVersion(revID int) *ArticleBuilder {
	ab.article.PreviousVersion = &schema.PreviousVersion{
		Identifier: revID,
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

func (ab *ArticleBuilder) Visibility(visibility *schema.Visibility) *ArticleBuilder {
	ab.article.Visibility = visibility

	return ab
}

// Build returns the created article instance.
func (ab *ArticleBuilder) Build() *schema.Article {
	return ab.article
}
