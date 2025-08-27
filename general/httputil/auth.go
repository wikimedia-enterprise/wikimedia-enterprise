package httputil

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"
)

// AuthProvider is an interface that wraps GetUser function
type AuthProvider interface {
	GetUser(string) (*User, error)
}

// NewAuthParams returns params for auth middleware
func NewAuthParams(clientID string, cmd redis.Cmdable, provider AuthProvider) *AuthParams {
	return &AuthParams{
		Provider: provider,
		Cache:    cmd,
		JWK:      new(JWKWellKnown),
		ClientID: clientID,
		Expire:   time.Hour * 24,
	}
}

// AuthParams structure represents middleware parameters.
type AuthParams struct {
	Provider AuthProvider
	Cache    redis.Cmdable
	JWK      JWKProvider
	ClientID string
	Expire   time.Duration
}

// AuthClaims claims object from JWT token.
type AuthClaims struct {
	jwt.StandardClaims
	ClientID string   `json:"client_id"`
	ISS      string   `json:"iss"`
	Groups   []string `json:"cognito:groups"`
	Sub      string   `json:"sub"`
}

// Auth middleware for cognito authentication through Authorization Bearer Token
// Note:
// If the expiration duration is less than one, the items in the cache never expire (by default), and must be deleted manually.
// If the cleanup interval is less than one, expired items are not deleted from the cache.
func Auth(p *AuthParams) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		usr, ok := gcx.Get("user")

		if ok && usr != nil {
			gcx.Next()
			return
		}

		token := strings.Replace(gcx.GetHeader("Authorization"), "Bearer ", "", 1)

		if len(token) <= 0 {
			Unauthorized(gcx)
			gcx.Abort()
			return
		}

		user := new(User)

		_, err := jwt.ParseWithClaims(token, new(AuthClaims), func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			kid, ok := token.Header["kid"].(string)

			if !ok {
				return nil, errors.New("kid header not found")
			}

			claims, ok := token.Claims.(*AuthClaims)

			if !ok {
				return nil, errors.New("couldn't resolve claims")
			}

			if claims.ClientID != p.ClientID {
				return nil, errors.New("incorrect client id")
			}

			if err := p.JWK.Fetch(claims.ISS); err != nil {
				return nil, err
			}

			key, err := p.JWK.Find(kid)

			if err != nil {
				return nil, err
			}

			user.SetGroups(claims.Groups)
			user.Sub = claims.Sub

			return key.RSA256()
		})

		if err != nil {
			Unauthorized(gcx, err)
			gcx.Abort()
			return
		}

		key := fmt.Sprintf("access_token:%s", token)
		data, err := p.Cache.Get(gcx, key).Bytes()

		if err != nil && err != redis.Nil {
			InternalServerError(gcx, err)
			gcx.Abort()
			return
		}

		if err == nil {
			if err := json.Unmarshal(data, user); err != nil {
				InternalServerError(gcx, err)
				gcx.Abort()
				return
			}
		}

		if err == redis.Nil {
			usr, err := p.Provider.GetUser(token)

			if err != nil {
				Unauthorized(gcx, err)
				gcx.Abort()
				return
			}

			user.SetUsername(usr.Username)
			// Important: `usr` only has `Username` set, don't override the other data in `user` that we obtained from parsing the token.
			data, err := json.Marshal(user)

			if err != nil {
				InternalServerError(gcx, err)
				gcx.Abort()
				return
			}

			if err := p.Cache.Set(gcx, key, data, p.Expire).Err(); err != nil {
				InternalServerError(gcx, err)
				gcx.Abort()
				return
			}
		}

		gcx.Set("user", user)
		gcx.Next()
	}
}
