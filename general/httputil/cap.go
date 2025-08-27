package httputil

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/dig"
)

// Errors for cap config.
var (
	ErrNoCapsConfigured = errors.New("no caps configured, this is only OK in local development")
	ErrMissingProducts  = errors.New("prefix group provided but missing Products in cap config")
	ErrMissingGroups    = errors.New("user groups missing in cap config")
)

// Compile the regex patterns for different Products (snapshots, articles, chunks)
var ProductsPatterns = map[string]*regexp.Regexp{
	"snapshots":            regexp.MustCompile(`/v2/snapshots/([^/]+)/download`),
	"structured-snapshots": regexp.MustCompile(`/v2/snapshots/structured-contents/([^/]+)/download`),
	"articles":             regexp.MustCompile(`/v2/articles/([^/]+)`),
	"structured-contents":  regexp.MustCompile(`/v2/structured-contents/([^/]+)`),
	"chunks":               regexp.MustCompile(`/v2/snapshots/([^/]+)/chunks/([^/]+)/download`),
}

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
// The purpose of PrefixGroup and Products is to provide limits per set of Products.
// Some Products together have a limit. For example, on-demand APIs (/articles, /structured-contents) have a combined limit of 10K.
// So, for PrefixGroup `ondemand`, the relevant Products are `articles` and `structured-contents`.
// The cap config can be used without PrefixGroup and Products. In this case, the cap is applied universally to all Products
// if one of the config group matches with the user group.
// More valid examples of CapConfig:
// {PrefixGroup: "cap:ondemand", Products: []string{"articles", "structured-contents"}, Limit: 100000, Groups: []string{"group-a", "group-b"}}
// {Limit: 10000000000, Groups: []string{ "group-c"}}
type CapConfig struct {
	Paths       []string `json:"paths,omitempty"`
	Products    []string `json:"products,omitempty"`
	PrefixGroup string   `json:"prefix_group,omitempty"`
	Limit       int      `json:"limit"`
	Groups      []string `json:"groups"`
}

// CapConfigWrapper is a type that wraps a slice of CapConfig structs
type CapConfigWrapper []*CapConfig

func New(cap []*CapConfig) CapConfigWrapper {
	return CapConfigWrapper(cap)
}

// Validate performs validation checks on the CapConfigWrapper
func (c *CapConfigWrapper) Validate() error {
	for i, cfg := range *c {
		// Check for missing groups
		if len(cfg.Groups) == 0 {
			return fmt.Errorf("config at index %d: %w", i, ErrMissingGroups)
		}

		// Check for missing products
		if len(cfg.PrefixGroup) > 0 && len(cfg.Products) == 0 {
			return fmt.Errorf("config at index %d: %w", i, ErrMissingProducts)
		}
	}

	return nil
}

// UnmarshalEnvironmentValue is a method of CapConfigWrapper that unmarshals the
// configuration data from a string and validates it.
func (c *CapConfigWrapper) UnmarshalEnvironmentValue(dta string) error {
	if err := json.Unmarshal([]byte(dta), &c); err != nil {
		return err
	}

	// Validate the configuration after unmarshalling
	return c.Validate()
}

// Cap is a function that returns a Gin middleware that limits the number of
// requests that can be made by a user. The middleware uses the given Capper
// implementation and configuration to check whether a request can be processed
// or not.
func Cap(cpr Capper, cfg CapConfigWrapper) gin.HandlerFunc {
	// Log warning if no caps are configured (should only occur in development)
	if len(cfg) == 0 {
		log.Println(ErrNoCapsConfigured)
	}

	return func(gcx *gin.Context) {
		if cpr == nil {
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
		var requestConfig *CapConfig
		var lmt int

		for _, c := range cfg {
			if !slices.Contains(c.Groups, grp) {
				continue
			}

			switch lpg := len(c.PrefixGroup); {
			case lpg > 0: // Products specific limiting
			OUTER:
				for _, prd := range c.Products {
					for products, pattern := range ProductsPatterns {
						if !strings.Contains(prd, products) || !pattern.MatchString(gcx.Request.URL.Path) {
							continue
						}

						requestConfig = c
						lmt = c.Limit
						idr = fmt.Sprintf("%s:%s", c.PrefixGroup, idr) // If PrefixGroup is provided, the identifier will be prefix group-specific
						break OUTER
					}
				}

			default: // Products agnostic limiting
				requestConfig = c
				lmt = c.Limit
			}
		}

		// If no rules match then allow the request
		if requestConfig == nil {
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
