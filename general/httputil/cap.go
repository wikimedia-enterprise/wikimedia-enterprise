package httputil

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/dig"
)

// Errors for cap config.
var (
	ErrMissingPaths  = errors.New("prefix group provided but missing paths in cap config")
	ErrMissingGroups = errors.New("user groups missing in cap config")
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

	if cnt >= lmt {
		return false, nil
	}

	if err := c.Redis.Incr(ctx, key).Err(); err != nil {
		return false, err
	}

	return true, nil
}

// CapConfig is a struct that holds the configuration for the request capper.
// The purpose of PrefixGroup and Paths is to provide limits per set of paths.
// Some paths together have a limit. For example, on-demand APIs (/articles, /structured-contents) have a combined limit of 10K.
// So, for PrefixGroup `ondemand`, the relevant paths are `articles` and `structured-contents`.
// The cap config can be used without PrefixGroup and Paths. In this case, the cap is applied universally to all paths
// if one of the config group matches with the user group.
// More valid examples of CapConfig:
// {PrefixGroup: "cap:ondemand", Paths: []string{"articles", "structured-contents"}, Limit: 100000, Groups: []string{"group-a", "group-b"}}
// {Limit: 10000000000, Groups: []string{ "group-c"}}
type CapConfig struct {
	Paths       []string `json:"paths,omitempty"`
	PrefixGroup string   `json:"prefix_group,omitempty"`
	Limit       int      `json:"limit"`
	Groups      []string `json:"groups"`
}

// CapConfigWrapper is a type that wraps a slice of CapConfig structs and
type CapConfigWrapper []*CapConfig

func New(cap []*CapConfig) CapConfigWrapper {
	return CapConfigWrapper(cap)
}

// UnmarshalEnvironmentValue is a method of CapConfig that unmarshals the
// configuration data from a string.
func (c *CapConfigWrapper) UnmarshalEnvironmentValue(dta string) error {
	return json.Unmarshal([]byte(dta), &c)
}

// Cap is a function that returns a Gin middleware that limits the number of
// requests that can be made by a user. The middleware uses the given Capper
// implementation and configuration to check whether a request can be processed
// or not.
func Cap(cpr Capper, cfg CapConfigWrapper) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		if cpr == nil || len(cfg) == 0 {
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

		idr := fmt.Sprintf("user:%s", usr.GetUsername()) // If no PrefixGroup is provided, the identifier will be prefix group-agnostic
		// for example user:userx
		var cnt, lmt int

		for _, c := range cfg {
			if len(c.Groups) == 0 {
				AbortWithInternalServerError(gcx, ErrMissingGroups)
				return
			}

			if !slices.Contains(c.Groups, grp) {
				cnt++
				continue
			}

			switch lpg := len(c.PrefixGroup); {

			case lpg > 0: // Path specific limiting
				if len(c.Paths) == 0 {
					AbortWithInternalServerError(gcx, ErrMissingPaths)
					return
				}

				for _, pth := range c.Paths {
					if strings.Contains(gcx.Request.URL.Path, pth) {
						idr = fmt.Sprintf("%s:%s", c.PrefixGroup, idr) // If PrefixGroup is provided, the identifier will be prefix group-specific
						lmt = c.Limit                                  // for example cap:ondemand:user:userx
						break
					}
				}

			default: // Path-agnostic limiting
				lmt = c.Limit
			}
		}

		// If no rules match then the allow request
		if cnt == len(cfg) {
			gcx.Next()
			return
		}

		isa, err := cpr.Check(gcx.Request.Context(), idr, lmt)

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
