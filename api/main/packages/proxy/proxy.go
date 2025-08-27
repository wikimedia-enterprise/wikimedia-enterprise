// Package proxy provides handlers to get data form s3 storage.
package proxy

import (
	"context"
	"crypto/md5" // #nosec G501
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
	"wikimedia-enterprise/api/main/config/env"
	"wikimedia-enterprise/api/main/packages/s3util"
	"wikimedia-enterprise/api/main/submodules/config"
	"wikimedia-enterprise/api/main/submodules/httputil"
	"wikimedia-enterprise/api/main/submodules/log"

	"github.com/Masterminds/squirrel"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/tidwall/gjson"
	"go.uber.org/dig"
)

// ErrLimitToLarge error for limit too large.
var ErrLimitToLarge = errors.New("limit cannot be greater than 10")

// Params struct represents parameters needed for proxy.
type Params struct {
	dig.In
	Cfg config.API
	Env *env.Environment
	S3  s3iface.S3API
}

// Filter filtering by field payload value.
type Filter struct {
	Field string      `json:"field"`
	Value interface{} `json:"value"`
}

// Filter returns true if value needs to be filtered out.
func (f *Filter) Filter(val any) bool {
	if f == nil {
		return false
	}

	return f.Value != val
}

// Model model for incoming requests.
type Model struct {
	Fields  []string  `form:"fields" json:"fields" binding:"max=255,dive,min=1,max=255"`
	Filters []*Filter `form:"filters" json:"filters" binding:"max=255"`
	Limit   int       `form:"limit" json:"limit" binding:"max=100"`
	index   map[string]interface{}
}

func (m *Model) buildFieldsIndex(fph string, pth []string, fmp map[string]interface{}) {
	if len(pth) == 1 {
		fmp[pth[0]] = fph
		return
	}

	switch fmp[pth[0]].(type) {
	case map[string]interface{}:
		m.buildFieldsIndex(fph, pth[1:], fmp[pth[0]].(map[string]interface{}))
	default:
		tmp := map[string]interface{}{}
		m.buildFieldsIndex(fph, pth[1:], tmp)
		fmp[pth[0]] = tmp
	}
}

// BuildFieldsIndex builds fields index for the request.
func (m *Model) BuildFieldsIndex() {
	m.index = map[string]interface{}{}

	for _, fld := range m.Fields {
		fnm := strings.TrimSuffix(fld, ".*") // remove `*` for wildcard queries
		m.buildFieldsIndex(fnm, strings.Split(fnm, "."), m.index)
	}
}

// GetFieldsIndex returns fields index for the request.
func (m *Model) GetFieldsIndex() map[string]interface{} {
	if m == nil {
		return nil
	}

	return m.index
}

// GetFilterByField returns filter by field name.
func (m *Model) GetFilterByField(fln string) *Filter {
	for _, flt := range m.Filters {
		if flt.Field == fln {
			return flt
		}
	}

	return nil
}

// PathGetter is an interface that wraps GetPath method.
type PathGetter interface {
	GetPath(*gin.Context) (string, error)
}

// Modifier is an interface that wraps Modify method.
// Be careful when using the interface, if you return true from the Modify method
// the object will be filtered out.
type Modifier interface {
	Modify(context.Context, *Model, *gjson.Result) (bool, error)
}

func escapeField(fld string) string {
	parts := strings.Split(fld, ".")
	for i := range parts {
		parts[i] = fmt.Sprintf(`"%s"`, parts[i])
	}

	return "s." + strings.Join(parts, ".")
}

func formatFields(mdl Model) []string {
	var cls []string

	for _, fld := range mdl.Fields {
		if len(fld) == 0 || fld == "*" {
			return []string{"*"}
		}

		ffd := escapeField(fld)
		cls = append(cls, fmt.Sprintf(`%s as "%s"`, ffd, strings.ReplaceAll(fld, ".", "__")))
	}

	if len(cls) == 0 {
		cls = append(cls, "*")
	}

	return cls
}

// getBucketName resolve bucket for /v2/files/<filename> and /v2/files/<filename>/download
// commons files are served from commons/files/ s3 prefix
// commons file metadata are served from commons/pages/ s3 prefix
func getBucketName(p *Params, path string) string {
	bkt := p.Env.AWSBucket
	if strings.HasPrefix(path, "commons/pages/") || strings.HasPrefix(path, "commons/files/") {
		bkt = p.Env.AWSBucketCommons
	}

	return bkt
}

// NewGetEntity returns a single entity from s3 storage by path.
func NewGetEntity(p *Params, pg PathGetter) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		mdl := new(Model)

		if err := httputil.BindModel(gcx, mdl); err != nil && err != io.EOF {
			log.Error(err, log.Tip("problem in proxy with binding model"))
			httputil.UnprocessableEntity(gcx, err)
			return
		}

		cls := formatFields(*mdl)
		sql, _, err := squirrel.Select(cls...).From("S3Object as s").ToSql()

		if err != nil {
			log.Error(err, log.Tip("problem in proxy with squirrel"))
			httputil.InternalServerError(gcx, err)
			return
		}

		path, err := pg.GetPath(gcx)

		if err != nil {
			log.Error(err, log.Tip("problem in proxy with path"))
			httputil.InternalServerError(gcx, err)
			return
		}

		//Choose between Commons bucket or Primary bucket
		bkt := getBucketName(p, path)
		sip := &s3.SelectObjectContentInput{
			Bucket:         aws.String(bkt),
			Key:            aws.String(path),
			ExpressionType: aws.String(s3.ExpressionTypeSql),
			Expression:     aws.String(sql),
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
			aer, ok := err.(awserr.Error)

			if ok && aer.Code() == s3.ErrCodeNoSuchKey {
				log.Error(aer, log.Tip("problem in proxy with s3 get entity, no such key"))
				httputil.NotFound(gcx, err)
			} else if ok && aer.Code() == "ParseInvalidPathComponent" {
				log.Error(aer, log.Tip("problem in proxy with s3 entity, invalid path"))
				httputil.BadRequest(gcx, err)
			} else {
				log.Error(err, log.Tip("problem in proxy with s3 entity, unknown error"))
				httputil.InternalServerError(gcx, err)
			}

			return
		}

		defer res.EventStream.Close()

		if res.EventStream.Err() != nil {
			log.Error(res.EventStream.Err(), log.Tip("problem in proxy with s3, event stream error"))
			httputil.InternalServerError(gcx, err)
			return
		}

		err = httputil.Render(gcx, http.StatusOK, s3util.EventContent(res.EventStream))

		if err != nil {
			log.Error(err, log.Tip("problem in proxy with s3, render error"))
			httputil.InternalServerError(gcx, err)
			return
		}
	}
}

// NewGetEntities returns multiple entities from s3 storage by path.
func NewGetEntities(p *Params, pgr PathGetter) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		mdl := new(Model)

		if err := httputil.BindModel(gcx, mdl); err != nil && err != io.EOF {
			log.Error(err, log.Tip("problem in proxy with binding model"))
			httputil.UnprocessableEntity(gcx, err)
			return
		}

		cls := formatFields(*mdl)
		que := squirrel.Select(cls...).From("S3Object as s")

		for _, flt := range mdl.Filters {
			if len(flt.Field) > 0 && flt.Value != nil {
				switch val := flt.Value.(type) {
				case int:
					que = que.Where(fmt.Sprintf("%s = %d", escapeField(flt.Field), val))
				case float64:
					if strings.Contains(flt.Field, "namespace.identifier") {
						// Namespace field is stored as a float in S3 Select, we cast to int in SQL for exact match
						que = que.Where(fmt.Sprintf("CAST(%s AS INT) = %d", escapeField(flt.Field), int(val)))
					} else {
						que = que.Where(fmt.Sprintf("%s > %f", escapeField(flt.Field), val))
					}
				case string:
					que = que.Where(fmt.Sprintf("%s = '%s'", escapeField(flt.Field), val))
				}
			}
		}

		sql, _, err := que.ToSql()

		if err != nil {
			log.Error(err, log.Tip("problem in proxy with squirrel"))
			httputil.InternalServerError(gcx, err)
			return
		}

		pth, err := pgr.GetPath(gcx)

		if err != nil {
			log.Error(err, log.Tip("problem in proxy with path"))
			httputil.InternalServerError(gcx, err)
			return
		}

		sip := &s3.SelectObjectContentInput{
			Bucket:         aws.String(p.Env.AWSBucket),
			Key:            aws.String(pth),
			ExpressionType: aws.String(s3.ExpressionTypeSql),
			Expression:     aws.String(sql),
			InputSerialization: &s3.InputSerialization{
				JSON: &s3.JSONInput{
					Type: aws.String(s3.JSONTypeLines),
				},
			},
			OutputSerialization: &s3.OutputSerialization{
				JSON: &s3.JSONOutput{
					RecordDelimiter: aws.String(httputil.RecordDelimiter(gcx)),
				},
			},
		}

		res, err := p.S3.SelectObjectContentWithContext(gcx.Request.Context(), sip)

		if err != nil {
			aer, ok := err.(awserr.Error)

			if ok && aer.Code() == s3.ErrCodeNoSuchKey {
				log.Error(aer, log.Tip("problem in proxy with s3 entities, no such key"))
				httputil.NotFound(gcx, err)
			} else if ok && aer.Code() == "ParseInvalidPathComponent" {
				log.Error(aer, log.Tip("problem in proxy with s3 entities, invalid path"))
				httputil.BadRequest(gcx, err)
			} else {
				log.Error(err, log.Tip("problem in proxy with s3 entities, unknown error"))
				httputil.InternalServerError(gcx, err)
			}

			return
		}

		defer res.EventStream.Close()

		if err := res.EventStream.Err(); err != nil {
			log.Error(err, log.Tip("problem in proxy with s3 entities, event stream error"))
			httputil.InternalServerError(gcx, err)
			return
		}

		err = httputil.Render(gcx, http.StatusOK, httputil.Format(gcx, s3util.EventContent(res.EventStream)))

		if err != nil {
			log.Error(err, log.Tip("Problem in proxy with s3 entities, render error"))
			httputil.InternalServerError(gcx, err)
			return
		}
	}
}

// NewGetDownload initiates file download.
func NewGetDownload(p *Params, pgr PathGetter) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		pth, err := pgr.GetPath(gcx)

		if err != nil {
			log.Error(err, log.Tip("problem in proxy in get download path"))
			httputil.InternalServerError(gcx, err)
			return
		}

		//Choose between Commons bucket or Primary bucket
		bkt := getBucketName(p, pth)
		req, _ := p.S3.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(bkt),
			Key:    aws.String(pth),
		})
		var params = req.HTTPRequest.URL.Query()
		if prm, ok := gcx.Get("user"); ok {
			if usr, ok := prm.(*httputil.User); ok {
				if sub := usr.Sub; sub != "" {
					params.Set("sub", sub)
				}
			}
		}
		req.HTTPRequest.URL.RawQuery = params.Encode()
		url, err := req.Presign(time.Second * 60)

		if err != nil {
			log.Error(err, log.Tip("problem in proxy in get download presign"))
			httputil.InternalServerError(gcx, err)
			return
		}

		gcx.Redirect(http.StatusTemporaryRedirect, url)
	}
}

// NewHeadDownload returns info about the file.
func NewHeadDownload(p *Params, pgr PathGetter) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		pth, err := pgr.GetPath(gcx)

		if err != nil {
			log.Error(err, log.Tip("problem in proxy in head download path"))
			httputil.InternalServerError(gcx, err)
			return
		}

		//Choose between Commons bucket or Primary bucket
		bkt := getBucketName(p, pth)
		out, err := p.S3.HeadObjectWithContext(gcx.Request.Context(), &s3.HeadObjectInput{
			Bucket: aws.String(bkt),
			Key:    aws.String(pth),
		})

		if err != nil {
			log.Error(err, log.Tip("problem in proxy in head download"))
			httputil.NotFound(gcx, err)
			return
		}

		s3util.WriteHeaders(gcx, out)
	}
}

// NewGetLargeEntities this is a s3 proxy function for large entities like Articles, the ones that s3 API can't handle.
func NewGetLargeEntities(p *Params, ent string, mfs ...Modifier) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		mdl := new(Model)

		if err := httputil.BindModel(gcx, mdl); err != nil && err != io.EOF {
			log.Error(err, log.Tip("problem in proxy with large entities, bind model error"))
			httputil.UnprocessableEntity(gcx, err)
			return
		}

		nme := gcx.Param("name")

		if len(nme) < 2 {
			log.Error("problem in proxy with large entities, no name")
			httputil.NotFound(gcx)
			return
		}

		// gcx.Param("name") has a leading slash.
		nme = nme[1:]

		articlePath := strings.ReplaceAll(nme, " ", "_")
		if p.Env.UseHashedPrefixes {
			hash := md5.Sum([]byte(articlePath)) // #nosec G401
			hashString := hex.EncodeToString(hash[:])

			// For example, 5/5c/Earth
			articlePath = fmt.Sprintf("%s/%s/%s", hashString[:1], hashString[:2], articlePath)
		}

		if mdl.Limit > 10 {
			httputil.UnprocessableEntity(gcx, ErrLimitToLarge)
			return
		}

		if mdl.Limit <= 0 {
			mdl.Limit = 3
		}

		// this is required for the concurrency
		// we need to be sure the index exists before we use it
		// in the go routines
		mdl.BuildFieldsIndex()

		pjs := p.Cfg.GetProjects()
		prs := make(chan string, len(pjs))
		swg := new(sync.WaitGroup)
		swg.Add(mdl.Limit + 1)

		// this go routine goes through all the projects and filters them
		// also applies language filters if needed
		go func() {
			defer swg.Done()

			prf := mdl.GetFilterByField("is_part_of.identifier")
			lnf := mdl.GetFilterByField("in_language.identifier")

			for _, prj := range pjs {
				if prf.Filter(prj) {
					continue
				}

				lng := p.Cfg.GetLanguage(prj)

				if lnf.Filter(lng) {
					continue
				}

				prs <- prj
			}

			close(prs)
		}()

		ets := make(chan string, mdl.Limit)

		// create a list of workers to get the entities
		// apply filters to them as well
		for i := 0; i < mdl.Limit; i++ {
			go func() {
				defer swg.Done()

			QUEUE:
				for prj := range prs {
					out, err := p.S3.GetObjectWithContext(gcx.Request.Context(), &s3.GetObjectInput{
						Bucket: aws.String(p.Env.AWSBucket),
						Key:    aws.String(fmt.Sprintf("%s/%s/%s.json", ent, prj, articlePath)),
					})

					if err != nil {
						aer, ok := err.(awserr.Error)

						if ok && aer.Code() != s3.ErrCodeNoSuchKey {
							log.Error(aer, log.Tip("problem in proxy with large entities, s3 no such key"))
						}

						if !ok {
							log.Error(err, log.Tip("problem in proxy with large entities, not ok"))
						}

						continue QUEUE
					}

					dta, err := io.ReadAll(out.Body)

					if err := out.Body.Close(); err != nil {
						log.Error(err, log.Tip("problem in proxy with large entities, body close"))
						continue QUEUE
					}

					if err != nil {
						log.Error(err, log.Tip("problem in proxy with large entities, read all"))
						continue QUEUE
					}

					obj := gjson.ParseBytes(dta)

					for _, mfr := range mfs {
						ftr, err := mfr.Modify(gcx.Request.Context(), mdl, &obj)

						if err != nil {
							log.Error(err, log.Tip("problem in proxy with large entities, modifier failed"))
						}

						if ftr {
							log.Info("breaking the loop because of modifier")
							continue QUEUE
						}
					}

					ets <- obj.Raw
					break QUEUE
				}
			}()
		}

		swg.Wait()
		close(ets)

		ctp := httputil.NegotiateFormat(gcx)
		gcx.Status(http.StatusOK)
		gcx.Header("Content-type", httputil.NegotiateFormat(gcx))
		nms := 0

		for ent := range ets {
			if nms == 0 {
				if ctp == binding.MIMEJSON {
					if _, err := gcx.Writer.WriteString("["); err != nil {
						log.Error(err, log.Tip("problem in proxy with large entities, binding mime write error"))
						httputil.InternalServerError(gcx, err)
						return
					}
				}
			}

			if nms != 0 {
				if _, err := gcx.Writer.WriteString(httputil.RecordDelimiter(gcx)); err != nil {
					log.Error(err, log.Tip("problem in proxy with large entities, record delimiter error"))
					httputil.InternalServerError(gcx, err)
					return
				}
			}

			if _, err := gcx.Writer.WriteString(ent); err != nil {
				log.Error(err, log.Tip("problem in proxy with large entities, write error"))
				httputil.InternalServerError(gcx, err)
				return
			}

			nms++
		}

		if nms > 0 {
			if ctp == binding.MIMEJSON {
				if _, err := gcx.Writer.WriteString("]"); err != nil {
					log.Error(err, log.Tip("problem in proxy with large entities, write error"))
					httputil.InternalServerError(gcx, err)
				}
			}
		} else {
			log.Error("problem in proxy with large entities, no data")
			httputil.NotFound(gcx)
		}
	}
}
