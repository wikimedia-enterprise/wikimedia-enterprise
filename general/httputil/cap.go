package httputil

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/dig"

	"golang.org/x/exp/slices"
)

// Capper is an interface that defines the Check method, which is used to check
// whether a request can be processed or not.
type Capper interface {
	Check(context.Context, string, int) (bool, error)
}

// CapByRedis is a struct that implements the Capper interface using Redis as
// the storage backend.
type CapByRedis struct {
	dig.In
	Redis redis.Cmdable
}

// Check is a method of CapByRedis that implements the Check method of the
// Capper interface. It checks the number of requests made by a user and returns
// true if the limit has not been reached, or false otherwise.
func (c *CapByRedis) Check(ctx context.Context, idr string, lmt int) (bool, error) {
	key := fmt.Sprintf("%s:count", idr)
	cnt, err := c.Redis.Get(ctx, key).Int()

	if err != nil && err != redis.Nil {
		return false, err
	}

	if err == redis.Nil {
		if err := c.Redis.Set(ctx, key, 1, 0).Err(); err != nil {
			return false, err
		}

		return true, nil
	}

	if cnt > lmt {
		return false, nil
	}

	if err := c.Redis.Incr(ctx, key).Err(); err != nil {
		return false, err
	}

	return true, nil
}

// CapConfig is a struct that holds the configuration for the request capper.
type CapConfig struct {
	Prefix string   `json:"prefix,omitempty"`
	Limit  int      `json:"limit"`
	Groups []string `json:"groups"`
}

// UnmarshalEnvironmentValue is a method of CapConfig that unmarshals the
// configuration data from a string.
func (c *CapConfig) UnmarshalEnvironmentValue(dta string) error {
	if c == nil {
		c = new(CapConfig)
	}

	return json.Unmarshal([]byte(dta), c)
}

// Cap is a function that returns a Gin middleware that limits the number of
// requests that can be made by a user. The middleware uses the given Capper
// implementation and configuration to check whether a request can be processed
// or not.
func Cap(cpr Capper, cfg *CapConfig) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		if cpr == nil || cfg == nil {
			gcx.Next()
			return
		}

		var usr *User

		if mdl, ok := gcx.Get("user"); ok {
			switch mdl := mdl.(type) {
			case *User:
				usr = mdl
			}
		}

		if usr == nil {
			AbortWithUnauthorized(gcx)
			return
		}

		grp := usr.GetGroup()

		if len(cfg.Groups) > 0 && !slices.Contains(cfg.Groups, grp) {
			gcx.Next()
			return
		}

		idr := fmt.Sprintf("user:%s", usr.GetUsername())

		if len(cfg.Prefix) > 0 {
			idr = fmt.Sprintf("%s:%s", cfg.Prefix, idr)
		}

		isa, err := cpr.Check(gcx.Request.Context(), idr, cfg.Limit)

		if err != nil {
			AbortWithInternalServerError(gcx, err)
			return
		}

		if !isa {
			AbortWithToManyRequests(gcx)
			return
		}

		gcx.Next()
	}
}
