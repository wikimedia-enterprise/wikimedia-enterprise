package httputil

import (
	"encoding/json"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
)

// Limiter is a struct that holds the rate limit configuration.
type Limiter struct {
	limits map[string]*limiter.Limiter
}

// UnmarshalEnvironmentValue unmarshals the rate limit configuration from an environment variable.
func (l *Limiter) UnmarshalEnvironmentValue(dta string) error {
	lms := map[string]int{}

	if err := json.Unmarshal([]byte(dta), &lms); err != nil {
		return err
	}

	for grp, lmt := range lms {
		l.Set(grp, lmt)
	}

	return nil
}

// Set sets the rate limit for a specific group.
func (l *Limiter) Set(key string, lmt int) {
	if l.limits == nil {
		l.limits = make(map[string]*limiter.Limiter)
	}

	l.limits[key] = tollbooth.NewLimiter(float64(lmt), &limiter.ExpirableOptions{
		DefaultExpirationTTL: time.Second,
	})
}

// Get gets the rate limiter for a specific group.
func (l *Limiter) Get(key string) *limiter.Limiter {
	return l.limits[key]
}

// Limit is a middleware function that enforces rate limits for API requests.
func Limit(lmr *Limiter) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		// no limits will be applied if we have a empty limiter
		if lmr == nil {
			gcx.Next()
			return
		}

		iur, ok := gcx.Get("user")

		if iur == nil || !ok {
			AbortWithUnauthorized(gcx)
			return
		}

		usr, ok := iur.(*User)

		if usr == nil || !ok {
			AbortWithInternalServerError(gcx)
			return
		}

		grp := usr.GetGroup()

		lmt := lmr.Get(grp)

		if lmt == nil {
			gcx.Next()
			return
		}

		if err := tollbooth.LimitByKeys(lmt, []string{usr.Username}); err != nil {
			AbortWithToManyRequests(gcx)
		} else {
			gcx.Next()
		}
	}
}
