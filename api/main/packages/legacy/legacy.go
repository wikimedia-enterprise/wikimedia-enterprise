// Package legacy provides legacy HTTP handlers.
// This is mirroring of endpoints that exist in v1 APIs
// for backward compatibility.
package legacy

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"wikimedia-enterprise/api/main/config/env"
	"wikimedia-enterprise/api/main/packages/s3util"
	"wikimedia-enterprise/general/config"
	"wikimedia-enterprise/general/httputil"

	"wikimedia-enterprise/general/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

const dateFormat = "2006-01-02"
const sql = `SELECT
o.is_part_of.identifier as identifier,
o.url as url,
o.version as version,
o.date_modified as date_modified,
o.in_language as in_language,
o."size" as "size"
FROM S3Object o WHERE o.namespace.identifier = %[1]s`

// Params holds the dependencies required by the legacy http handlers.
type Params struct {
	dig.In
	S3  s3iface.S3API
	Env *env.Environment
	Cfg config.API
}

// NewGetPageHandler creates a new version of the legacy pages handler.
func NewGetPageHandler(p *Params) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		dbn := gcx.Param("project")

		if len(dbn) <= 1 || len(dbn) > 255 {
			log.Error("problem with new get page handler, dbn to short or too long")
			httputil.BadRequest(gcx)
			return
		}

		nme := strings.ReplaceAll(gcx.Param("name"), " ", "_")

		if len(nme) <= 1 || len(nme) > 1000 {
			log.Error("problem with new get page handler, nme to short or too long")
			httputil.BadRequest(gcx)
			return
		}

		res, err := p.S3.GetObjectWithContext(gcx.Request.Context(), &s3.GetObjectInput{
			Bucket: aws.String(p.Env.AWSBucket),
			Key:    aws.String(fmt.Sprintf("articles/%s%s.json", dbn, nme)),
		})

		if err != nil {
			log.Error(err, log.Tip("problem with new get page handler, no s3 object found"))
			httputil.NotFound(gcx, err)
			return
		}

		defer res.Body.Close()

		if err != nil {
			log.Error(err, log.Tip("problem with new get page handler, unknown problem"))
			httputil.InternalServerError(gcx, err)
			return
		}

		gcx.DataFromReader(http.StatusOK, -1, fmt.Sprintf("%s; charset=UTF-8", gin.MIMEJSON), res.Body, nil)
	}
}

// NewListDiffsHandler creates a new version of HTTP handler that will list diffs by namespace.
func NewListDiffsHandler(p *Params) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		nsp := gcx.Param("namespace")
		dte := gcx.Param("date")
		nss := p.Cfg.GetNamespaces()

		if len(nsp) > 0 && !isStringInIntSlice(nsp, nss) {
			log.Error("problem with new list diff unknown namespace")
			httputil.BadRequest(gcx, fmt.Errorf("namespace '%s' not supported", nsp))
			return
		}

		sip := &s3.SelectObjectContentInput{
			Bucket:         aws.String(p.Env.AWSBucket),
			Key:            aws.String(fmt.Sprintf("aggregations/batches/%s/batches.ndjson", dte)),
			ExpressionType: aws.String(s3.ExpressionTypeSql),
			Expression:     aws.String(fmt.Sprintf(sql, nsp)),
			InputSerialization: &s3.InputSerialization{
				JSON: &s3.JSONInput{
					Type: aws.String(s3.JSONTypeLines),
				},
			},
			OutputSerialization: &s3.OutputSerialization{
				JSON: &s3.JSONOutput{},
			},
		}

		res, err := p.S3.SelectObjectContentWithContext(gcx.Request.Context(), sip)

		if err != nil {
			log.Error(err, log.Tip("problem with new list diff, error with s3 select"))
			httputil.InternalServerError(gcx, err)
			return
		}

		defer res.EventStream.Close()

		if err := res.EventStream.Err(); err != nil {
			log.Error(err, log.Tip("problem with new list diff, error with s3 event stream"))
			httputil.InternalServerError(gcx, err)
			return
		}

		err = httputil.Render(gcx, http.StatusOK, httputil.Format(gcx, strings.ReplaceAll(s3util.EventContent(res.EventStream), "\n", ",")))

		if err != nil {
			log.Error(err, log.Tip("problem with new list diff, error rendering s3 content"))
			httputil.InternalServerError(gcx, err)
			return
		}
	}
}

// NewGetDIffHandler is a HTTP handler that is responsible for getting single diff metadata.
func NewGetDiffHandler(p *Params) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		nsp := gcx.Param("namespace")
		dte := gcx.Param("date")
		nss := p.Cfg.GetNamespaces()

		if len(nsp) > 0 && !isStringInIntSlice(nsp, nss) {
			err := fmt.Errorf("namespace '%s' not supported", nsp)
			log.Error(err, log.Tip("problem with new get diff, error in namespace name"))
			httputil.BadRequest(gcx, err)
			return
		}

		dbn := gcx.Param("project")

		if len(dbn) <= 1 || len(dbn) > 255 {
			log.Error("problem with new get diff, error in project name")
			httputil.BadRequest(gcx)
			return
		}

		sip := &s3.SelectObjectContentInput{
			Bucket:         aws.String(p.Env.AWSBucket),
			Key:            aws.String(fmt.Sprintf("batches/%s/%s_namespace_%s.json", dte, dbn, nsp)),
			ExpressionType: aws.String(s3.ExpressionTypeSql),
			Expression:     aws.String(fmt.Sprintf(sql, nsp)),
			InputSerialization: &s3.InputSerialization{
				JSON: &s3.JSONInput{
					Type: aws.String(s3.JSONTypeLines),
				},
			},
			OutputSerialization: &s3.OutputSerialization{
				JSON: &s3.JSONOutput{},
			},
		}

		res, err := p.S3.SelectObjectContentWithContext(gcx.Request.Context(), sip)

		if err != nil {
			log.Error(err, log.Tip("problem with new get diff, error with s3 select"))
			httputil.InternalServerError(gcx, err)
			return
		}

		defer res.EventStream.Close()

		if err := res.EventStream.Err(); err != nil {
			log.Error(err, log.Tip("problem with new get diff, error with s3 event stream"))
			httputil.InternalServerError(gcx, err)
			return
		}

		err = httputil.Render(gcx, http.StatusOK, s3util.EventContent(res.EventStream))

		if err != nil {
			log.Error(err, log.Tip("problem with new get diff, error rendering s3 content"))
			httputil.InternalServerError(gcx, err)
			return
		}
	}
}

// NewDownloadDiffHandler creates new HTTP handler for dif downloads.
func NewDownloadDiffHandler(p *Params) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		dte := gcx.Param("date")

		if _, err := time.Parse(dateFormat, dte); len(dte) == 0 || err != nil {
			log.Error(err, log.Tip("problem with new download diff, error in date format"))
			httputil.BadRequest(gcx)
			return
		}

		nsp := gcx.Param("namespace")
		nss := p.Cfg.GetNamespaces()

		if len(nsp) > 0 && !isStringInIntSlice(nsp, nss) {
			err := fmt.Errorf("namespace '%s' not supported", nsp)
			log.Error(err, log.Tip("problem with new download diff, error in namespace name"))
			httputil.BadRequest(gcx, err)
			return
		}

		dbn := gcx.Param("project")

		if len(dbn) <= 1 || len(dbn) > 255 {
			log.Error("problem with new download diff, empty project name")
			httputil.BadRequest(gcx)
			return
		}

		pth := fmt.Sprintf("batches/%s/%s_namespace_%s.tar.gz", dte, dbn, nsp)

		_, err := p.S3.HeadObjectWithContext(gcx.Request.Context(), &s3.HeadObjectInput{
			Bucket: aws.String(p.Env.AWSBucket),
			Key:    aws.String(pth),
		})

		if err != nil {
			log.Error(err, log.Tip("problem with new download diff, error with head object"))
			httputil.NotFound(gcx, fmt.Errorf("diff from '%s' for '%s' not found", dte, dbn))
			return
		}

		req, _ := p.S3.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(p.Env.AWSBucket),
			Key:    aws.String(pth),
		})
		url, err := req.Presign(time.Second * 60)

		if err != nil {
			log.Error(err, log.Tip("problem with new download diff, error with presign"))
			httputil.InternalServerError(gcx, err)
			return
		}

		gcx.Redirect(http.StatusTemporaryRedirect, url)
	}
}

// NewHeadDiffHandler creates new HTTP handler for making HEAD request to diffs.
func NewHeadDiffHandler(p *Params) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		nsp := gcx.Param("namespace")
		dte := gcx.Param("date")
		nss := p.Cfg.GetNamespaces()

		if len(nsp) > 0 && !isStringInIntSlice(nsp, nss) {
			err := fmt.Errorf("namespace '%s' not supported", nsp)
			log.Error(err, log.Tip("problem with new head diff, error in namespace name"))
			httputil.BadRequest(gcx, err)
			return
		}

		dbn := gcx.Param("project")

		if len(dbn) <= 1 || len(dbn) > 255 {
			log.Error("problem with new head diff, empty project name")
			httputil.BadRequest(gcx)
			return
		}

		out, err := p.S3.HeadObjectWithContext(gcx.Request.Context(), &s3.HeadObjectInput{
			Bucket: aws.String(p.Env.AWSBucket),
			Key:    aws.String(fmt.Sprintf("batches/%s/%s_namespace_%s.tar.gz", dte, dbn, nsp)),
		})

		if err != nil {
			log.Error(err, log.Tip("problem with new download diff, error"))
			httputil.NotFound(gcx, err)
			return
		}

		s3util.WriteHeaders(gcx, out)
	}
}

// NewGetExportHandler creates new HTTP handler to get a single export by project and namespace.
func NewGetExportHandler(p *Params) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		uRaw, ok := gcx.Get("user")

		if !ok {
			log.Error("problem with new get export, user not found")
			httputil.Unauthorized(gcx, errors.New("user not found"))
			return
		}

		user, ok := uRaw.(*httputil.User)

		if !ok {
			log.Error("problem with new get export, user not correct type")
			httputil.InternalServerError(gcx, errors.New("unknown user type"))
			return
		}

		nsp := gcx.Param("namespace")
		nss := p.Cfg.GetNamespaces()

		if len(nsp) > 0 && !isStringInIntSlice(nsp, nss) {
			log.Error("problem with new get export, error in namespace name")
			httputil.BadRequest(gcx, fmt.Errorf("namespace '%s' not supported", nsp))
			return
		}

		dbn := gcx.Param("project")

		if len(dbn) <= 1 || len(dbn) > 255 {
			log.Error("problem with new get export, error in project name")
			httputil.BadRequest(gcx)
			return
		}

		pth := fmt.Sprintf("snapshots/%s_namespace_%s.json", dbn, nsp)

		if user.IsInGroup(p.Env.FreeTierGroup) {
			pth = fmt.Sprintf("snapshots/%s_namespace_%s_%s.json", dbn, nsp, p.Env.FreeTierGroup)
		}

		sip := &s3.SelectObjectContentInput{
			Bucket:         aws.String(p.Env.AWSBucket),
			Key:            aws.String(pth),
			ExpressionType: aws.String(s3.ExpressionTypeSql),
			Expression:     aws.String(fmt.Sprintf(sql, nsp)),
			InputSerialization: &s3.InputSerialization{
				JSON: &s3.JSONInput{
					Type: aws.String(s3.JSONTypeLines),
				},
			},
			OutputSerialization: &s3.OutputSerialization{
				JSON: &s3.JSONOutput{},
			},
		}

		res, err := p.S3.SelectObjectContentWithContext(gcx.Request.Context(), sip)

		if err != nil {
			log.Error(err, log.Tip("problem with new get export, error in select object content"))
			httputil.InternalServerError(gcx, err)
			return
		}

		defer res.EventStream.Close()

		if err := res.EventStream.Err(); err != nil {
			log.Error(err, log.Tip("problem with new get export, error in event stream"))
			httputil.InternalServerError(gcx, err)
			return
		}

		err = httputil.Render(gcx, http.StatusOK, s3util.EventContent(res.EventStream))

		if err != nil {
			log.Error(err, log.Tip("problem with new get export, error in render"))
			httputil.InternalServerError(gcx, err)
			return
		}
	}
}

// NewListExportsHandler creates a new HTTP handler that will list all available exports by namespace.
func NewListExportsHandler(p *Params) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		uRaw, ok := gcx.Get("user")

		if !ok {
			log.Error("problem with new list export, error in user")
			httputil.Unauthorized(gcx, errors.New("user not found"))
			return
		}

		user, ok := uRaw.(*httputil.User)

		if !ok {
			log.Error("problem with new list export, error in user type")
			httputil.InternalServerError(gcx, errors.New("unknown user type"))
			return
		}

		nsp := gcx.Param("namespace")
		nss := p.Cfg.GetNamespaces()

		if len(nsp) > 0 && !isStringInIntSlice(nsp, nss) {
			log.Error("problem with new list export, error in namespace name")
			httputil.BadRequest(gcx, fmt.Errorf("namespace '%s' not supported", nsp))
			return
		}

		pth := "aggregations/snapshots/snapshots.ndjson"

		if user.IsInGroup(p.Env.FreeTierGroup) {
			pth = fmt.Sprintf("aggregations/snapshots/snapshots_%s.ndjson", p.Env.FreeTierGroup)
		}

		sip := &s3.SelectObjectContentInput{
			Bucket:         aws.String(p.Env.AWSBucket),
			Key:            aws.String(pth),
			ExpressionType: aws.String(s3.ExpressionTypeSql),
			Expression:     aws.String(fmt.Sprintf(sql, nsp)),
			InputSerialization: &s3.InputSerialization{
				JSON: &s3.JSONInput{
					Type: aws.String(s3.JSONTypeLines),
				},
			},
			OutputSerialization: &s3.OutputSerialization{
				JSON: &s3.JSONOutput{},
			},
		}

		res, err := p.S3.SelectObjectContentWithContext(gcx.Request.Context(), sip)

		if err != nil {
			log.Error(err, log.Tip("problem with new list export, error in select object content"))
			httputil.InternalServerError(gcx, err)
			return
		}

		defer res.EventStream.Close()

		if err := res.EventStream.Err(); err != nil {
			log.Error(err, log.Tip("problem with new list export, error in event stream"))
			httputil.InternalServerError(gcx, err)
			return
		}

		err = httputil.Render(gcx, http.StatusOK, httputil.Format(gcx, strings.ReplaceAll(s3util.EventContent(res.EventStream), "\n", ",")))

		if err != nil {
			log.Error(err, log.Tip("problem with new list export, error in render"))
			httputil.InternalServerError(gcx, err)
			return
		}
	}
}

// NewHeadExportHandler returns a HTTP handler that will show return export headers by identifier.
func NewHeadExportHandler(p *Params) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		uRaw, ok := gcx.Get("user")

		if !ok {
			log.Error("problem with new head export, error in user")
			httputil.Unauthorized(gcx, errors.New("user not found"))
			return
		}

		user, ok := uRaw.(*httputil.User)

		if !ok {
			log.Error("problem with new head export, error in user type")
			httputil.InternalServerError(gcx, errors.New("unknown user type"))
			return
		}

		nsp := gcx.Param("namespace")
		nss := p.Cfg.GetNamespaces()

		if len(nsp) > 0 && !isStringInIntSlice(nsp, nss) {
			log.Error("problem with new head export, error in namespace name")
			httputil.BadRequest(gcx, fmt.Errorf("namespace '%s' not supported", nsp))
			return
		}

		dbn := gcx.Param("project")

		if len(dbn) <= 1 || len(dbn) > 255 {
			log.Error("problem with new head export, error in project name")
			httputil.BadRequest(gcx)
			return
		}

		pth := fmt.Sprintf("snapshots/%s_namespace_%s.tar.gz", dbn, nsp)

		if user.IsInGroup(p.Env.FreeTierGroup) {
			pth = fmt.Sprintf("snapshots/%s_namespace_%s_%s.tar.gz", dbn, nsp, p.Env.FreeTierGroup)
		}

		out, err := p.S3.HeadObjectWithContext(gcx.Request.Context(), &s3.HeadObjectInput{
			Bucket: aws.String(p.Env.AWSBucket),
			Key:    aws.String(pth),
		})

		if err != nil {
			log.Error(err, log.Tip("Problem with new head export, error in head object"))
			httputil.NotFound(gcx, err)
			return
		}

		s3util.WriteHeaders(gcx, out)
	}
}

// NewDownloadExportHandler creates a HTTP handler that will initialize export download.
func NewDownloadExportHandler(p *Params) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		uRaw, ok := gcx.Get("user")

		if !ok {
			log.Error("problem with new download export, error in user")
			httputil.Unauthorized(gcx, errors.New("user not found"))
			return
		}

		user, ok := uRaw.(*httputil.User)

		if !ok {
			log.Error("problem with new download export, error in user type")
			httputil.InternalServerError(gcx, errors.New("unknown user type"))
			return
		}

		nsp := gcx.Param("namespace")
		nss := p.Cfg.GetNamespaces()

		if len(nsp) > 0 && !isStringInIntSlice(nsp, nss) {
			log.Error("problem with new download export, error in namespace name")
			httputil.BadRequest(gcx, fmt.Errorf("namespace '%s' not supported", nsp))
			return
		}

		dbn := gcx.Param("project")

		if len(dbn) <= 1 || len(dbn) > 255 {
			log.Error("problem with new download export, error in project name")
			httputil.BadRequest(gcx)
			return
		}

		pth := fmt.Sprintf("snapshots/%s_namespace_%s.tar.gz", dbn, nsp)

		if user.IsInGroup(p.Env.FreeTierGroup) {
			pth = fmt.Sprintf("snapshots/%s_namespace_%s_%s.tar.gz", dbn, nsp, p.Env.FreeTierGroup)
		}

		_, err := p.S3.HeadObjectWithContext(gcx.Request.Context(), &s3.HeadObjectInput{
			Bucket: aws.String(p.Env.AWSBucket),
			Key:    aws.String(pth),
		})

		if err != nil {
			log.Error(err, log.Tip("problem with new download export, error in head object"))
			httputil.NotFound(gcx, fmt.Errorf("export for '%s' not found", dbn))
			return
		}

		req, _ := p.S3.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(p.Env.AWSBucket),
			Key:    aws.String(pth),
		})
		url, err := req.Presign(time.Second * 60)

		if err != nil {
			log.Error(err, log.Tip("problem with new download export, error in presign"))
			httputil.InternalServerError(gcx, err)
			return
		}

		gcx.Redirect(http.StatusTemporaryRedirect, url)
	}
}
