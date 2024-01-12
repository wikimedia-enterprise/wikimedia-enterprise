// Package httputil provides number of http helping methods.
package httputil

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// Available content types.
const (
	MIMENDJSON     = "application/x-ndjson"
	MIMEEVENTSTEAM = "text/event-stream"
)

// BindModel checks and compares model from context.
func BindModel(gcx *gin.Context, mdl interface{}) error {
	if err := gcx.ShouldBind(mdl); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NegotiateFormat returns an acceptable Accept format.
func NegotiateFormat(gcx *gin.Context) string {
	return gcx.NegotiateFormat(binding.MIMEJSON, MIMENDJSON)
}

// RecordDelimiter returns delimiter depending on format.
func RecordDelimiter(gcx *gin.Context) string {
	switch NegotiateFormat(gcx) {
	case binding.MIMEJSON:
		return ","
	case MIMENDJSON:
		return "\n"
	}

	return ""
}

// Render writes string to the request body.
func Render(gcx *gin.Context, status int, rct string) error {
	gcx.Status(status)
	gcx.Header("Content-type", NegotiateFormat(gcx))

	if _, err := gcx.Writer.WriteString(rct); err != nil {
		return err
	}

	return nil
}

// Format returns formatted string depending on the context format.
func Format(gcx *gin.Context, rct string) string {
	switch NegotiateFormat(gcx) {
	case binding.MIMEJSON:
		return fmt.Sprintf("[%s]", rct)
	default:
		return rct
	}
}

// HTTPError represents structure of an HTTP error.
type HTTPError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// UnprocessableEntity writes UnprocessableEntity error to the context.
func UnprocessableEntity(gcx *gin.Context, errs ...error) {
	RenderError(gcx, http.StatusUnprocessableEntity, errs...)
}

// InternalServerError writes InternalServerError error to the context.
func InternalServerError(gcx *gin.Context, errs ...error) {
	RenderError(gcx, http.StatusInternalServerError, errs...)
}

// AbortWithInternalServerError writes InternalServerError error to the context and aborts the chain.
func AbortWithInternalServerError(gcx *gin.Context, errs ...error) {
	RenderError(gcx, http.StatusInternalServerError, errs...)
	gcx.Abort()
}

// NotFound writes NotFound error to the context.
func NotFound(gcx *gin.Context, errs ...error) {
	RenderError(gcx, http.StatusNotFound, errs...)
}

// Unauthorized writes Unauthorized error to the context.
func Unauthorized(gcx *gin.Context, errs ...error) {
	RenderError(gcx, http.StatusUnauthorized, errs...)
}

// AbortWithUnauthorized writes Unauthorized error to the context and aborts the chain.
func AbortWithUnauthorized(gcx *gin.Context, errs ...error) {
	RenderError(gcx, http.StatusUnauthorized, errs...)
	gcx.Abort()
}

// BadRequest writes BadRequest error to the context.
func BadRequest(gcx *gin.Context, errs ...error) {
	RenderError(gcx, http.StatusBadRequest, errs...)
}

// Forbidden writes Forbidden error to the context.
func Forbidden(gcx *gin.Context, errs ...error) {
	RenderError(gcx, http.StatusForbidden, errs...)
}

// AbortWithForbidden writes Forbidden error to the context and aborts the chain.
func AbortWithForbidden(gcx *gin.Context, errs ...error) {
	RenderError(gcx, http.StatusForbidden, errs...)
	gcx.Abort()
}

// ToManyRequests writes TooManyRequests error to the context.
func ToManyRequests(gcx *gin.Context, errs ...error) {
	RenderError(gcx, http.StatusTooManyRequests, errs...)
}

// AbortWithToManyRequests writes TooManyRequests error to the context and aborts the chain.
func AbortWithToManyRequests(gcx *gin.Context, errs ...error) {
	RenderError(gcx, http.StatusTooManyRequests, errs...)
	gcx.Abort()
}

// RenderError writes an error to the context.
func RenderError(gcx *gin.Context, status int, errs ...error) {
	err := HTTPError{
		Status: status,
	}

	if len(errs) == 0 {
		err.Message = http.StatusText(status)
	}

	for _, mer := range errs {
		if len(err.Message) > 0 {
			err.Message = fmt.Sprintf("%s; %s", err.Message, mer.Error())
		} else {
			err.Message = mer.Error()
		}
	}

	gcx.Status(err.Status)
	gcx.Header("Content-type", NegotiateFormat(gcx))
	_ = json.NewEncoder(gcx.Writer).Encode(err)
}

// NoRoute is the middleware to initialize for not found endpoint.
func NoRoute() gin.HandlerFunc {
	return func(gcx *gin.Context) {
		NotFound(gcx)
	}
}

// CORS is the middleware with default CORS configuration for the API(s).
func CORS(ops ...func(cfg *cors.Config)) gin.HandlerFunc {
	cfg := cors.Config{
		AllowAllOrigins: true,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodOptions,
			http.MethodHead,
		},
		AllowHeaders: []string{"*"},
	}

	for _, opt := range ops {
		opt(&cfg)
	}

	return cors.New(cfg)
}
