package proxy

import (
	"errors"
	"fmt"
	"wikimedia-enterprise/general/httputil"

	"github.com/gin-gonic/gin"
)

// Errors for the getters.
var (
	ErrEmptyIdentifier = errors.New("identifier is empty")
	ErrEmptyDate       = errors.New("date is empty")
	ErrWrongUserType   = errors.New("user is of a wrong type")
	ErrUnauthorized    = errors.New("user not found in the context")
)

// ByGroupGetterBase allows to get user from request context.
type ByGroupGetterBase struct{}

// GetUser returns user from the context if present.
func (g *ByGroupGetterBase) GetUser(gcx *gin.Context) (*httputil.User, error) {
	mdl, ok := gcx.Get("user")

	if !ok {
		return nil, ErrUnauthorized
	}

	usr, ok := mdl.(*httputil.User)

	if !ok {
		return nil, ErrWrongUserType
	}

	return usr, nil
}

// NewEntitiesGetter creates new entities getter instance.
func NewEntitiesGetter(url string) *EntitiesGetter {
	return &EntitiesGetter{
		URL: url,
	}
}

// EntitiesGetter allows to list entities in bulk.
type EntitiesGetter struct {
	URL string
}

// GetPath returns path for multiple entities.
func (e *EntitiesGetter) GetPath(gcx *gin.Context) (string, error) {
	return fmt.Sprintf("aggregations/%[1]s/%[1]s.ndjson", e.URL), nil
}

// NewEntityGetter creates new entity getter instance.
func NewEntityGetter(url string) *EntityGetter {
	return &EntityGetter{
		url,
	}
}

// EntityGetter this is a path getter for a single entity.
type EntityGetter struct {
	URL string
}

// GetPath returns path for single entity by identifier.
func (e *EntityGetter) GetPath(gcx *gin.Context) (string, error) {
	idn := gcx.Param("identifier")

	if len(idn) == 0 {
		return "", ErrEmptyIdentifier
	}

	return fmt.Sprintf("%s/%s.json", e.URL, idn), nil
}

// NewEntityDownloader creates new instance of entity downloader.
func NewEntityDownloader(url string) *EntityDownloader {
	return &EntityDownloader{
		url,
	}
}

// EntityDownloader gives the ability to provide download path fo single entity.
type EntityDownloader struct {
	URL string
}

// GetPath returns a s3 bucket location for entity or error.
func (e *EntityDownloader) GetPath(gcx *gin.Context) (string, error) {
	idn := gcx.Param("identifier")

	if len(idn) == 0 {
		return "", ErrEmptyIdentifier
	}

	return fmt.Sprintf("%s/%s.tar.gz", e.URL, idn), nil
}

// NewDateEntitiesGetter returns DateEntitiesGetter structure.
func NewDateEntitiesGetter(url string) *DateEntitiesGetter {
	return &DateEntitiesGetter{
		url,
	}
}

// DateEntitiesGetter allows to list date entities in bulk.
type DateEntitiesGetter struct {
	URL string
}

// GetPath returns path for multiple entities by date.
func (d *DateEntitiesGetter) GetPath(gcx *gin.Context) (string, error) {
	dte := gcx.Param("date")

	if len(dte) == 0 {
		return "", ErrEmptyDate
	}

	return fmt.Sprintf("aggregations/%[1]s/%[2]s/%[1]s.ndjson", d.URL, dte), nil
}

// NewDateEntityGetter creates new instance of date entity getter.
func NewDateEntityGetter(url string) *DateEntityGetter {
	return &DateEntityGetter{
		url,
	}
}

// DateEntityGetter allows to get a single date based entity.
type DateEntityGetter struct {
	URL string
}

// GetPath returns path for single entity by identifier and date.
func (d *DateEntityGetter) GetPath(gcx *gin.Context) (string, error) {
	dte := gcx.Param("date")

	if len(dte) == 0 {
		return "", ErrEmptyDate
	}

	idn := gcx.Param("identifier")

	if len(idn) == 0 {
		return "", ErrEmptyIdentifier
	}

	return fmt.Sprintf("%s/%s/%s.json", d.URL, dte, idn), nil
}

// NewDateEntityDownloader creates new instance of date entity downloader.
func NewDateEntityDownloader(url string) *DateEntityDownloader {
	return &DateEntityDownloader{
		url,
	}
}

// DateEntityDownloader allows to download date based entity.
type DateEntityDownloader struct {
	URL string
}

// GetPath returns path for downloadable entity by date and identifier.
func (d *DateEntityDownloader) GetPath(gcx *gin.Context) (string, error) {
	dte := gcx.Param("date")

	if len(dte) == 0 {
		return "", ErrEmptyDate
	}

	idn := gcx.Param("identifier")

	if len(idn) == 0 {
		return "", ErrEmptyIdentifier
	}

	return fmt.Sprintf("%s/%s/%s.tar.gz", d.URL, dte, idn), nil
}

// NewByGroupEntitiesGetter creates newentities getter by user group instance.
func NewByGroupEntitiesGetter(url string, grp string) *ByGroupEntitiesGetter {
	return &ByGroupEntitiesGetter{
		URL:   url,
		Group: grp,
	}
}

// ByGroupEntitiesGetter allows to get path of aggregates entities by user group.
type ByGroupEntitiesGetter struct {
	ByGroupGetterBase
	URL   string
	Group string
}

// GetPath returns path of the aggregated metadata by user group.
func (g *ByGroupEntitiesGetter) GetPath(gcx *gin.Context) (string, error) {
	usr, err := g.GetUser(gcx)

	if err != nil {
		return "", err
	}

	if usr.IsInGroup(g.Group) {
		return fmt.Sprintf("aggregations/%[1]s/%[1]s_%[2]s.ndjson", g.URL, g.Group), nil
	}

	return fmt.Sprintf("aggregations/%[1]s/%[1]s.ndjson", g.URL), nil
}

// NewByGroupEntityGetter creates newentity getter by user group instance.
func NewByGroupEntityGetter(url string, grp string) *ByGroupEntityGetter {
	return &ByGroupEntityGetter{
		URL:   url,
		Group: grp,
	}
}

// ByGroupEntityGetter allows to get path of entity by user group.
type ByGroupEntityGetter struct {
	ByGroupGetterBase
	URL   string
	Group string
}

// GetPath returns path of the entity by user group.
func (g *ByGroupEntityGetter) GetPath(gcx *gin.Context) (string, error) {
	idn := gcx.Param("identifier")

	if len(idn) == 0 {
		return "", ErrEmptyIdentifier
	}

	usr, err := g.GetUser(gcx)

	if err != nil {
		return "", err
	}

	if usr.IsInGroup(g.Group) {
		return fmt.Sprintf("%s/%s_%s.json", g.URL, idn, g.Group), nil
	}

	return fmt.Sprintf("%s/%s.json", g.URL, idn), nil
}

// NewByGroupEntityDownloader creates new instance of entity downloader by user group.
func NewByGroupEntityDownloader(url string, grp string) *ByGroupEntityDownloader {
	return &ByGroupEntityDownloader{
		URL:   url,
		Group: grp,
	}
}

// ByGroupEntityDownloader gives the ability to provide download path of a single entity by user group.
type ByGroupEntityDownloader struct {
	ByGroupGetterBase
	URL   string
	Group string
}

// GetPath returns a s3 bucket location for entity for a certain user group.
func (g *ByGroupEntityDownloader) GetPath(gcx *gin.Context) (string, error) {
	idn := gcx.Param("identifier")

	if len(idn) == 0 {
		return "", ErrEmptyIdentifier
	}

	usr, err := g.GetUser(gcx)

	if err != nil {
		return "", err
	}

	if usr.IsInGroup(g.Group) {
		return fmt.Sprintf("%s/%s_%s.tar.gz", g.URL, idn, g.Group), nil
	}

	return fmt.Sprintf("%s/%s.tar.gz", g.URL, idn), nil
}
