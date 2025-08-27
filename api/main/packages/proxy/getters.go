package proxy

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"strings"
	"wikimedia-enterprise/api/main/submodules/httputil"

	"github.com/gin-gonic/gin"
)

// Errors for the getters.
var (
	ErrEmptyIdentifier      = errors.New("identifier is empty")
	ErrEmptyChunkIdentifier = errors.New("chunk identifier is empty")
	ErrEmptyFilename        = errors.New("filename is empty")
	ErrEmptyDate            = errors.New("date is empty")
	ErrWrongUserType        = errors.New("user is of a wrong type")
	ErrUnauthorized         = errors.New("user not found in the context")

	perHourAggregationPath = "/%s/:date/:hour"
	perHourMetadataPath    = "/%s/:date/:hour/:database"
	perHourDownloadPath    = "/%s/:date/:hour/:database/download"

	HourlyPaths = struct {
		PerHourAggregationPath string
		PerHourAggregationS3   *template.Template
		PerHourMetadataPath    string
		PerHourMetadataS3      *template.Template
		PerHourDownloadPath    string
		PerHourDownloadS3      *template.Template
	}{
		PerHourAggregationPath: perHourAggregationPath,
		PerHourAggregationS3:   template.Must(template.New("per-hour-aggregation").Parse("aggregations/{{.root}}/{{.date}}/{{.hour}}/batches.ndjson")),
		PerHourMetadataPath:    perHourMetadataPath,
		PerHourMetadataS3:      template.Must(template.New("per-hour-metadata").Parse("{{.root}}/{{.date}}/{{.hour}}/{{.database}}.json")),
		PerHourDownloadPath:    perHourDownloadPath,
		PerHourDownloadS3:      template.Must(template.New("per-hour-download").Parse("{{.root}}/{{.date}}/{{.hour}}/{{.database}}.tar.gz")),
	}
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

// NewFileGetter creates new file getter instance.
func NewFileGetter() *FileGetter {
	return &FileGetter{}
}

// FileGetter this is a path getter for a single file metadata.
type FileGetter struct {
}

// GetPath returns path for single file metadata by filename.
func (e *FileGetter) GetPath(gcx *gin.Context) (string, error) {
	fln := gcx.Param("filename")

	if len(fln) == 0 {
		return "", ErrEmptyFilename
	}

	return fmt.Sprintf("commons/pages/%s.json", strings.ReplaceAll(fln, " ", "_")), nil
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

// NewFileDownloader creates new instance of a file downloader.
func NewFileDownloader() *FileDownloader {
	return &FileDownloader{}
}

// FileDownloader gives the ability to provide download path fo single file.
type FileDownloader struct {
}

// GetPath returns a s3 bucket location for a file or error.
func (e *FileDownloader) GetPath(gcx *gin.Context) (string, error) {
	fln := gcx.Param("filename")

	if len(fln) == 0 {
		return "", ErrEmptyFilename
	}

	return fmt.Sprintf("commons/files/%s", strings.ReplaceAll(fln, " ", "_")), nil
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
	// Resolve s3 key for chunk metadata
	if strings.Contains(g.URL, "chunks") {
		idn := gcx.Param("identifier")
		groupSuffix := ""
		if usr.IsInGroup(g.Group) {
			groupSuffix = fmt.Sprintf("_%s", g.Group)
		}
		return fmt.Sprintf("aggregations/%s/%s/%s%s.ndjson", g.URL, idn, g.URL, groupSuffix), nil
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
	if strings.Contains(g.URL, "chunk") {
		groupSuffix := ""
		cdn := gcx.Param("chunkIdentifier")

		if len(cdn) == 0 {
			return "", ErrEmptyChunkIdentifier
		}

		sps := strings.Split(cdn, "_")
		if usr.IsInGroup(g.Group) {
			groupSuffix = fmt.Sprintf("_%s", g.Group)
		}

		return fmt.Sprintf("%s/%s/%s%s.json", g.URL, idn, fmt.Sprintf("chunk_%s", sps[len(sps)-1]), groupSuffix), nil
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

	// Resolve s3 key for chunk tar
	if strings.Contains(g.URL, "chunk") {
		cdn := gcx.Param("chunkIdentifier")

		if len(cdn) == 0 {
			return "", ErrEmptyChunkIdentifier
		}
		groupSuffix := ""
		if usr.IsInGroup(g.Group) {
			groupSuffix = fmt.Sprintf("_%s", g.Group)
		}

		sps := strings.Split(cdn, "_")

		return fmt.Sprintf("%s/%s/%s%s.tar.gz", g.URL, idn, fmt.Sprintf("chunk_%s", sps[len(sps)-1]), groupSuffix), nil
	}

	if usr.IsInGroup(g.Group) {
		return fmt.Sprintf("%s/%s_%s.tar.gz", g.URL, idn, g.Group), nil
	}

	return fmt.Sprintf("%s/%s.tar.gz", g.URL, idn), nil
}

func getParamMap(gcx *gin.Context, params ...string) (map[string]string, error) {
	dict := map[string]string{}
	for _, param := range params {
		dict[param] = gcx.Param(param)
		if len(dict[param]) == 0 {
			return nil, fmt.Errorf("expected parameter missing: %s", param)
		}
	}

	return dict, nil
}

func runTemplate(tmpl *template.Template, params map[string]string) (string, error) {
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, params)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// HourlyEntityAggregationGetter allows to download aggregations keyed by date and hour.
type HourlyEntityAggregationGetter struct {
	Root string
}

// GetPath returns the appropriate path for the requested file.
func (d *HourlyEntityAggregationGetter) GetPath(gcx *gin.Context) (string, error) {
	params, err := getParamMap(gcx, "date", "hour")
	if err != nil {
		return "", err
	}

	params["root"] = d.Root
	return runTemplate(HourlyPaths.PerHourAggregationS3, params)
}

// HourlyEntityMetadataGetter allows to download entity metadata keyed by date, hour and database.
type HourlyEntityMetadataGetter struct {
	Root string
}

// GetPath returns the appropriate path for the requested file.
func (d *HourlyEntityMetadataGetter) GetPath(gcx *gin.Context) (string, error) {
	params, err := getParamMap(gcx, "date", "hour", "database")
	if err != nil {
		return "", err
	}

	params["root"] = d.Root
	return runTemplate(HourlyPaths.PerHourMetadataS3, params)
}

// HourlyEntityDownloader allows to download entities keyed by date, hour and database.
type HourlyEntityDownloader struct {
	Root string
}

// GetPath returns the appropriate path for the requested file.
func (d *HourlyEntityDownloader) GetPath(gcx *gin.Context) (string, error) {
	params, err := getParamMap(gcx, "date", "hour", "database")
	if err != nil {
		return "", err
	}

	params["root"] = d.Root
	return runTemplate(HourlyPaths.PerHourDownloadS3, params)
}
