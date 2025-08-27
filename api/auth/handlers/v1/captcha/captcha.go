// Package captcha creates a CAPTCHA image.
package captcha

import (
	"errors"
	"net/http"
	"strconv"
	"wikimedia-enterprise/api/auth/submodules/httputil"
	"wikimedia-enterprise/api/auth/submodules/log"

	"github.com/dchest/captcha"
	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/dig"
)

// Parameters dependency injection for the handler.
type Parameters struct {
	dig.In
	Redis redis.Cmdable
}

// Response structure represents response data format.
type Response struct {
	Identifier string `json:"identifier,omitempty"`
}

// NewGetHandler creates new captcha.
func NewGetHandler(p *Parameters) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		res := Response{
			captcha.New(),
		}

		if err := p.Redis.Set(gcx.Request.Context(), res.Identifier, uniuri.NewLenChars(6, []byte("0123456789")), captcha.Expiration).Err(); err != nil {
			log.Error(err, log.Tip("problem setting redis captcha"))
			httputil.InternalServerError(gcx, err)
			return
		}

		gcx.JSON(http.StatusOK, res)
	}
}

// NewShowHandler generates captcha image.
func NewShowHandler(p *Parameters) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		identifier := gcx.Param("identifier")
		solution, err := p.Redis.Get(gcx.Request.Context(), identifier).Result()

		if err == redis.Nil {
			log.Error(err, log.Tip("problem with redis nil in captcha image"))
			httputil.BadRequest(gcx, errors.New("Captcha not found!"))
			return
		}

		if err != nil {
			log.Error(err, log.Tip("problem with results for captcha image"))
			httputil.InternalServerError(gcx, err)
			return
		}

		digits := make([]byte, len(solution))

		for index, item := range solution {
			val, err := strconv.Atoi(string(item))

			if err != nil {
				log.Error(err, log.Tip("problem converting string to number in captcha image"))
				httputil.InternalServerError(gcx, err)
				return
			}

			digits[index] = byte(val)
		}

		gcx.Writer.Header().Set("Content-Type", "image/png")
		gcx.Writer.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		gcx.Writer.Header().Set("Pragma", "no-cache")
		gcx.Writer.Header().Set("Expires", "0")

		if _, err = captcha.NewImage(identifier, digits, captcha.StdWidth, captcha.StdHeight).WriteTo(gcx.Writer); err != nil {
			log.Error(err, log.Tip("problem with creating image writer in captcha image"))
			httputil.InternalServerError(gcx, err)
			return
		}
	}
}
