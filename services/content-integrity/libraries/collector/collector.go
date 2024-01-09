// Package collector exposes a set of interfaces,
// that will encapsulate collection of data per article.
package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"wikimedia-enterprise/services/content-integrity/config/env"

	redis "github.com/redis/go-redis/v9"
	"go.uber.org/dig"
	"golang.org/x/exp/slices"
)

// Versions a defined type of an Version array.
type Versions []*Version

// GetCurrentIdentifier returns the last version identifier being tracked.
func (v Versions) GetCurrentIdentifier() int {
	if len(v) == 0 {
		return 0
	}

	return v[0].Identifier
}

// GetUniqueEditorsCount returns unique editors count.
func (v Versions) GetUniqueEditorsCount() int {
	if len(v) == 0 {
		return 0
	}

	eds := make([]string, 0)

	for _, val := range v {
		if !slices.ContainsFunc(eds, func(elm string) bool { return strings.EqualFold(elm, val.Editor) }) {
			eds = append(eds, val.Editor)
		}
	}

	return len(eds)
}

// ArticleVersionsGetter defines an interface for retrieving versions of an article.
type ArticleVersionsGetter interface {
	GetVersions(ctx context.Context, prj string, idr int) (Versions, error)
}

// ArticleDateCreatedGetter defines an interface for retrieving the creation date of an article.
type ArticleDateCreatedGetter interface {
	GetDateCreated(ctx context.Context, prj string, idr int) (*time.Time, error)
}

// ArticleDateNamespaceMovedGetter defines an interface for retrieving the date an article was moved to a different namespace.
type ArticleDateNamespaceMovedGetter interface {
	GetDateNamespaceMoved(ctx context.Context, prj string, idr int) (*time.Time, error)
}

// ArticleVersionsPrepender defines an interface for setting versions of an article.
type ArticleVersionsPrepender interface {
	PrependVersion(ctx context.Context, prj string, idr int, vrns *Version) (Versions, error)
}

// ArticleDateCreatedSetter defines an interface for setting the creation date of an article.
type ArticleDateCreatedSetter interface {
	SetDateCreated(ctx context.Context, prj string, idr int, dct *time.Time) (*time.Time, error)
}

// ArticleDateNamespaceMovedSetter defines an interface for setting the date an article was moved into the main namespace.
type ArticleDateNamespaceMovedSetter interface {
	SetDateNamespaceMoved(ctx context.Context, prj string, idr int, dnm *time.Time) (*time.Time, error)
}

// ArticleDateCreatedGetter defines an interface for determining is an article is being tracked, has information stored.
type ArticleIsBeingTrackedGetter interface {
	GetIsBeingTracked(ctx context.Context, prj string, idr int) (bool, error)
}

// ArticleIsBreakingNewsGetter defines interface to for marking an article as Breaking News.
type ArticleIsBreakingNewsSetter interface {
	SetIsBreakingNews(ctx context.Context, prj string, idr int) error
}

// ArticleIsBreakingNewsGetter defines interface for determining if an article has been marked as Breaking News already.
type ArticleIsBreakingNewsGetter interface {
	GetIsBreakingNews(ctx context.Context, prj string, idr int) (bool, error)
}

// ArticleNameSetter defines interface for setting an article name.
type ArticleNameSetter interface {
	SetName(ctx context.Context, prj string, idr int, name string) error
}

// ArticleNameGetter defines interface for getting an article name.
type ArticleNameGetter interface {
	GetName(ctx context.Context, prj string, idr int) (string, error)
}

// ArticleAPI combines all of the article interfaces into a single interface.
type ArticleAPI interface {
	ArticleVersionsGetter
	ArticleDateCreatedGetter
	ArticleDateNamespaceMovedGetter
	ArticleVersionsPrepender
	ArticleDateCreatedSetter
	ArticleDateNamespaceMovedSetter
	ArticleIsBeingTrackedGetter
	ArticleIsBreakingNewsSetter
	ArticleIsBreakingNewsGetter
	ArticleNameSetter
	ArticleNameGetter
}

// NewArticle creates new instance of the Article collector.
func NewArticle(cmd redis.Cmdable, env *env.Environment) ArticleAPI {
	return &Article{
		Redis: cmd,
		Env:   env,
	}
}

// Article represents an article struct for dependency injection.
type Article struct {
	dig.In
	Redis redis.Cmdable
	Env   *env.Environment
}

// Version represents a version struct.
type Version struct {
	Identifier  int        `json:"identifier,omitempty"`
	Editor      string     `json:"editor,omitempty"`
	DateCreated *time.Time `json:"date_created,omitempty"`
}

// GetVersions retrieves the versions of an article.
func (c *Article) GetVersions(ctx context.Context, prj string, idr int) (Versions, error) {
	cnt, err := c.Redis.LRange(ctx, c.getVersionsKey(prj, idr), 0, -1).Result()

	if err != nil {
		return nil, err
	}

	vrs := []*Version{}

	// Unmarshal list of versions.
	for _, vrm := range cnt {
		vrn := new(Version)

		if err := json.Unmarshal([]byte(vrm), vrn); err != nil {
			return nil, err
		}

		vrs = append(vrs, vrn)
	}

	return vrs, nil
}

// GetDateCreated retrieves the creation date of an article.
func (c *Article) GetDateCreated(ctx context.Context, prj string, idr int) (*time.Time, error) {
	cnt, err := c.Redis.Get(ctx, c.getDateCreatedKey(prj, idr)).Time()

	if err != nil {
		return nil, err
	}

	return &cnt, nil
}

// GetDateNamespaceMoved retrieves the date an article was moved into the main namespace.
func (c *Article) GetDateNamespaceMoved(ctx context.Context, prj string, idr int) (*time.Time, error) {
	cnt, err := c.Redis.Get(ctx, c.getDateNamespaceMovedKey(prj, idr)).Time()

	if err != nil {
		return nil, err
	}

	return &cnt, nil
}

// PrependVersion prepends to the versions list, adds the most recent version ID to the head.
func (c *Article) PrependVersion(ctx context.Context, prj string, idr int, vrn *Version) (Versions, error) {
	// Check whether to start the collection.
	ext, err := c.GetIsBeingTracked(ctx, prj, idr)

	if err != nil {
		return nil, err
	}

	if !ext {
		return nil, nil
	}

	mrv, err := json.Marshal(vrn)

	if err != nil {
		return nil, err
	}

	if err := c.Redis.LPush(ctx, c.getVersionsKey(prj, idr), mrv).Err(); err != nil {
		return nil, err
	}

	if err := c.Redis.ExpireNX(ctx, c.getVersionsKey(prj, idr), c.getDefaultKeyExpiration()).Err(); err != nil {
		return nil, err
	}

	// Retrieve full list to return back.
	cnt, err := c.GetVersions(ctx, prj, idr)

	if err != nil {
		return nil, err
	}

	return cnt, nil
}

// SetDateCreated sets the creation date for an article.
func (c *Article) SetDateCreated(ctx context.Context, prj string, idr int, dct *time.Time) (*time.Time, error) {
	if err := c.Redis.Set(ctx, c.getDateCreatedKey(prj, idr), dct.Format(time.RFC3339Nano), c.getDefaultKeyExpiration()).Err(); err != nil {
		return nil, err
	}

	if err := c.setIsBeingTracked(ctx, prj, idr); err != nil {
		return nil, err
	}

	return dct, nil
}

// SetDateNamespaceMoved sets the date an article was moved into the main namespace.
func (c *Article) SetDateNamespaceMoved(ctx context.Context, prj string, idr int, dnm *time.Time) (*time.Time, error) {
	if err := c.Redis.Set(ctx, c.getDateNamespaceMovedKey(prj, idr), dnm.Format(time.RFC3339Nano), c.getDefaultKeyExpiration()).Err(); err != nil {
		return nil, err
	}

	if err := c.setIsBeingTracked(ctx, prj, idr); err != nil {
		return nil, err
	}

	return dnm, nil
}

// GetIsBeingTracked method for determining is an article is being tracked, has information stored.
func (c *Article) GetIsBeingTracked(ctx context.Context, prj string, idr int) (bool, error) {
	ext, err := c.Redis.Get(ctx, c.getIsBeingTrackedKey(prj, idr)).Bool()

	if err == redis.Nil {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return ext, nil
}

// SetIsBreakingNews method for setting if an article is already breaking news.
func (c *Article) SetIsBreakingNews(ctx context.Context, prj string, idr int) error {
	if err := c.Redis.Set(ctx, c.getIsBreakingNewsKey(prj, idr), true, c.getDefaultKeyExpiration()).Err(); err != nil {
		return err
	}

	return nil
}

// GetIsBreakingNews method for getting info about article if is already breaking news.
func (c *Article) GetIsBreakingNews(ctx context.Context, prj string, idr int) (bool, error) {
	ext, err := c.Redis.Get(ctx, c.getIsBreakingNewsKey(prj, idr)).Bool()

	if err == redis.Nil {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return ext, nil
}

// SetName method for setting an article name.
func (c *Article) SetName(ctx context.Context, prj string, idr int, name string) error {
	err := c.Redis.Set(ctx, c.getNameKey(prj, idr), name, c.getDefaultKeyExpiration()).Err()

	return err
}

// GetName method for getting an article name.
func (c *Article) GetName(ctx context.Context, prj string, idr int) (string, error) {
	nme, err := c.Redis.Get(ctx, c.getNameKey(prj, idr)).Result()
	return nme, err
}

// getIsBreakingNewsKey returns Redis key for is_breaking_news collection.
func (c *Article) getIsBreakingNewsKey(prj string, idr int) string {
	return c.getFormattedKey(prj, idr, "is_breaking_news")
}

// getNameKey returns Redis key for name collection.
func (c *Article) getNameKey(prj string, idr int) string {
	return c.getFormattedKey(prj, idr, "name")
}

// getDateCreatedKey returns Redis key for date_created collection.
func (c *Article) getDateCreatedKey(prj string, idr int) string {
	return c.getFormattedKey(prj, idr, "date_created")
}

// getDateNamespaceMovedKey returns Redis key for date_namespace_moved collection.
func (c *Article) getDateNamespaceMovedKey(prj string, idr int) string {
	return c.getFormattedKey(prj, idr, "date_namespace_moved")
}

// getVersionsKey returns Redis key for versions collection.
func (c *Article) getVersionsKey(prj string, idr int) string {
	return c.getFormattedKey(prj, idr, "versions")
}

// getIsBeingTrackedKey returns Redis key for is_being_tracked flag in Redis.
func (c *Article) getIsBeingTrackedKey(prj string, idr int) string {
	return c.getFormattedKey(prj, idr, "is_being_tracked")
}

// setIsBeingTracked sets the is_being_tracked flag in Redis.
func (c *Article) setIsBeingTracked(ctx context.Context, prj string, idr int) error {
	return c.Redis.Set(ctx, c.getIsBeingTrackedKey(prj, idr), true, c.getDefaultKeyExpiration()).Err()
}

// getFormatterKey returns a formatter Redis key.
func (c *Article) getFormattedKey(prj string, idr int, sfx string) string {
	return fmt.Sprintf("article:%s:%d:%s", prj, idr, sfx)
}

// getDefaultKeyExpiration returns default key expiration from environment variable.
func (c *Article) getDefaultKeyExpiration() time.Duration {
	return time.Hour * time.Duration(c.Env.BreakingNewsKeysExpiration)
}
