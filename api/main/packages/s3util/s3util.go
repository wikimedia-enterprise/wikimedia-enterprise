// Package s3util provides s3 helper functions.
package s3util

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
)

// EventContent returns string from event stream without special characters.
func EventContent(est *s3.SelectObjectContentEventStream) string {
	bdr := strings.Builder{}

	for evt := range est.Reader.Events() {
		switch evt := evt.(type) {
		case *s3.RecordsEvent:
			_, _ = bdr.WriteString(string(evt.Payload))
		}
	}

	return strings.TrimSuffix(
		strings.TrimSuffix(bdr.String(), "\n"),
		",")
}

// WriteHeaders function that helps to extract headers from the s3 HEAD object response
// and write them to gin context.
func WriteHeaders(gcx *gin.Context, out *s3.HeadObjectOutput) {
	if out.AcceptRanges != nil {
		gcx.Header("Accept-Ranges", *out.AcceptRanges)
	}

	if out.LastModified != nil {
		gcx.Header("Last-Modified", out.LastModified.Format(time.RFC1123))
	}

	if out.ContentLength != nil {
		gcx.Header("Content-Length", fmt.Sprint(*out.ContentLength))
	}

	if out.ETag != nil {
		gcx.Header("ETag", *out.ETag)
	}

	if out.CacheControl != nil {
		gcx.Header("Cache-Control", *out.CacheControl)
	}

	if out.ContentDisposition != nil {
		gcx.Header("Content-Disposition", *out.ContentDisposition)
	}

	if out.ContentEncoding != nil {
		gcx.Header("Content-Encoding", *out.ContentEncoding)
	}

	if out.ContentType != nil {
		gcx.Header("Content-Type", *out.ContentType)
	}

	if out.Expires != nil {
		gcx.Header("Expires", *out.Expires)
	}

	gcx.Header("Access-Control-Expose-Headers", "Accept-Ranges, Last-Modified, Content-Length, ETag, Cache-Control, Content-Disposition, Content-Encoding, Content-Type, Expires")
}
