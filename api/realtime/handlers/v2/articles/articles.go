// Package articles holds the HTTP handler that will expose internal articles topic
// with additional filtering features.
package articles

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	"wikimedia-enterprise/api/realtime/config/env"
	"wikimedia-enterprise/api/realtime/libraries/resolver"
	"wikimedia-enterprise/general/httputil"
	"wikimedia-enterprise/general/ksqldb"
	"wikimedia-enterprise/general/log"
	"wikimedia-enterprise/general/schema"

	sqr "github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

// Parameters dependency injection for the handler.
type Parameters struct {
	dig.In
	KSQLDB   ksqldb.PushPuller
	Resolver *resolver.Resolver
	Env      *env.Environment
}

// Filter model for the filter field and value.
type Filter struct {
	Field string      `json:"field"`
	Value interface{} `json:"value"`
}

// Model model for incoming requests. Time format is RFC 3339 (2019-10-12T07:20:50.52Z) for incoming requests for Since and SincePerPartition.
type Model struct {
	Since             time.Time         `form:"since" json:"since"`
	Fields            []string          `form:"fields" json:"fields" binding:"max=255,dive,min=1,max=255"`
	Filters           []Filter          `form:"filters" json:"filters" binding:"max=255,dive,max=2"`
	Parts             []int             `form:"parts" json:"parts"`
	Offsets           map[int]int64     `form:"offsets" json:"offsets"`
	SincePerPartition map[int]time.Time `form:"since_per_partition" json:"since_per_partition"`
	partitions        []int
}

// SetPartitions sets the partitions for the model.
func (m *Model) SetPartitions(nPartitions int, maxParts int) {
	sze := nPartitions / maxParts

	for _, prt := range m.Parts {
		for i := prt * sze; i < (prt+1)*sze; i++ {
			m.partitions = append(m.partitions, i)
		}
	}
}

// GetPartitions gets the partitions for the model.
func (m *Model) GetPartitions() []int {
	return m.partitions
}

// PartitionsContain checks if the partitions for the model contain the given partition.
func (m *Model) PartitionsContain(ptn int) bool {
	return slices.Contains(m.partitions, ptn)
}

// GetPartitions gets the partitions as a comma separated string for the model.
func (m *Model) GerFormattedPartitions() string {
	sts := make([]string, len(m.partitions))

	for i, num := range m.partitions {
		sts[i] = fmt.Sprintf("%d", num)
	}

	return strings.Join(sts, ",")
}

// NewHandler creates new articles HTTP handler.
func NewHandler(ctx context.Context, p *Parameters) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		mdl := new(Model)

		if err := gcx.ShouldBind(mdl); err != nil && err != io.EOF {
			log.Error(err)
			httputil.UnprocessableEntity(gcx, err)
			return
		}

		if err := gcx.ShouldBindJSON(mdl); err != nil && err != io.EOF {
			log.Error(err)
			httputil.UnprocessableEntity(gcx, err)
			return
		}

		for idx, field := range mdl.Fields {
			// Escape ksqldb keywords from fields such as namespace
			for kwd, rpl := range resolver.Keywords {
				field = strings.ReplaceAll(field, kwd, rpl)
				mdl.Fields[idx] = field
			}

			name := strings.TrimSuffix(field, ".*")

			if !p.Resolver.HasField(name) {
				err := fmt.Errorf("field with name '%s' not found", field)
				log.Error(err)
				httputil.UnprocessableEntity(gcx, err)
				return
			}

			if strings.HasSuffix(field, ".*") {
				continue
			}

			if p.Resolver.HasSlice(name) || p.Resolver.HasStruct(name) {
				err := fmt.Errorf("use wildcard '%s.*' to select all fields from the object", field)
				log.Error(err)
				httputil.UnprocessableEntity(gcx, err)
				return
			}
		}

		cols := []string{}

		if col := resolver.NewSelect(p.Resolver, mdl.Fields).GetSql(); len(col) > 0 {
			cols = append(cols, col)
		}

		if len(cols) == 0 {
			cols = append(cols, "*")
		}

		cols = append(cols, []string{"ROWPARTITION", "ROWOFFSET", "ROWTIME"}...)

		qer := sqr.
			Select(cols...).
			From(p.Env.ArticlesStream)

		if len(mdl.Parts) != 0 {
			// Validate bounds for Parts
			if len(mdl.Parts) > p.Env.MaxParts {
				err := fmt.Errorf("the query param parts has length %d, max allowed length %d", len(mdl.Parts), p.Env.MaxParts)
				log.Error(err)
				httputil.UnprocessableEntity(gcx, err)
				return
			}

			for _, prt := range mdl.Parts {
				if prt < 0 || prt > p.Env.MaxParts-1 {
					err := fmt.Errorf("the part value is %d, allowed value should be in the range [0,%d]", prt, p.Env.MaxParts-1)
					log.Error(err)
					httputil.UnprocessableEntity(gcx, err)
					return
				}
			}

			if !mdl.Since.IsZero() {
				err := errors.New("for parallel consumption, specify either offsets or time-offsets (since_per_partition) parameter")
				log.Error(err)
				httputil.UnprocessableEntity(gcx, err)
				return
			}

			// Resolve partitions from Parts
			mdl.SetPartitions(p.Env.Partitions, p.Env.MaxParts)

			if len(mdl.Offsets) != 0 {
				// Validate bounds for Offsets
				if len(mdl.Offsets) > p.Env.Partitions {
					err := fmt.Errorf("the query param offsets has length %d, max allowed length %d", len(mdl.Offsets), p.Env.Partitions)
					log.Error(err)
					httputil.UnprocessableEntity(gcx, err)
					return
				}

				// Pick relevant offsets related to the partitions resolved
				for ptn := range mdl.Offsets {
					if !mdl.PartitionsContain(ptn) {
						delete(mdl.Offsets, ptn)
					}
				}
			}

			if len(mdl.SincePerPartition) != 0 {
				// Validate bounds for SincePerPartition
				if len(mdl.SincePerPartition) > p.Env.Partitions {
					err := fmt.Errorf("the query param since_per_partition has length %d, max allowed length %d", len(mdl.SincePerPartition), p.Env.Partitions)
					log.Error(err)
					httputil.UnprocessableEntity(gcx, err)
					return
				}

				// Pick relevant since_per_partition related to the partitions resolved
				for ptn := range mdl.SincePerPartition {
					if !mdl.PartitionsContain(ptn) {
						delete(mdl.SincePerPartition, ptn)
					}
				}
			}

			// Query with offset and time-offset for same partion not allowed
			if len(mdl.Offsets) != 0 && len(mdl.SincePerPartition) != 0 {
				for ptn := range mdl.Offsets {
					if _, ok := mdl.SincePerPartition[ptn]; ok {
						err := errors.New("for a specific partion, specify either offset or time-offset (since_per_partition), not both")
						log.Error(err)
						httputil.UnprocessableEntity(gcx, err)
						return
					}
				}

			}

			qer = sqr.
				Select(cols...).
				From(p.Env.ArticlesStream).
				Where(fmt.Sprintf("ROWPARTITION IN (%s)", mdl.GerFormattedPartitions()))

			if len(mdl.Offsets) != 0 || len(mdl.SincePerPartition) != 0 {
				cse := "CASE"

				for ptn, oft := range mdl.Offsets {
					cse = fmt.Sprintf("%s WHEN ROWPARTITION = %d THEN ROWOFFSET >= %v", cse, ptn, oft)
				}

				for ptn, tme := range mdl.SincePerPartition {
					cse = fmt.Sprintf("%s WHEN ROWPARTITION = %d THEN ROWTIME >= %v", cse, ptn, int(tme.UnixMilli()))
				}

				cse = fmt.Sprintf("%s ELSE TRUE END", cse)
				qer = sqr.
					Select(cols...).
					From(p.Env.ArticlesStream).
					Where(fmt.Sprintf("ROWPARTITION IN (%s) AND %s", mdl.GerFormattedPartitions(), cse))
			}
		}

		for _, filter := range mdl.Filters {
			// TODO: Remove this temporary solution,
			// this handles a case when keyword is being used
			// as a field name, in our case "namespace" field is like that
			for kwd, rpl := range resolver.Keywords {
				filter.Field = strings.ReplaceAll(filter.Field, kwd, rpl)
			}

			if !p.Resolver.HasField(filter.Field) {
				err := fmt.Errorf("there is no such field '%s' to filter", filter.Field)
				log.Error(err)
				httputil.UnprocessableEntity(gcx)
				return
			}

			if p.Resolver.HasSlice(filter.Field) || p.Resolver.HasStruct(filter.Field) {
				err := fmt.Errorf("can't filter on a '%s' object", filter.Field)
				log.Error(err)
				httputil.UnprocessableEntity(gcx, err)
				return
			}

			ffd := p.Resolver.GetField(filter.Field)
			rfv := reflect.ValueOf(filter.Value)

			if !rfv.Type().ConvertibleTo(ffd.Value.Type()) {
				err := fmt.Errorf("the filter '%s' should not be of type '%v'", filter.Field, rfv.Kind())
				log.Error(err)
				httputil.UnprocessableEntity(gcx, err)
				return
			}

			if rfv.Kind() == reflect.String {
				qer = qer.Where(fmt.Sprintf("%s = '%v'", ffd.Path, rfv.Convert(ffd.Value.Type()).Interface()))
			}

			if rfv.Kind() == reflect.Int || rfv.Kind() == reflect.Float64 || rfv.Kind() == reflect.Bool {
				qer = qer.Where(fmt.Sprintf("%s = %v", ffd.Path, rfv.Convert(ffd.Value.Type()).Interface()))
			}
		}

		ctp := gcx.NegotiateFormat(httputil.MIMEEVENTSTEAM, httputil.MIMENDJSON)

		write := func(v interface{}) {
			gcx.JSON(-1, v)
			_, _ = gcx.Writer.WriteString("\n")
		}

		if ctp == httputil.MIMEEVENTSTEAM {
			write = func(v interface{}) {
				_, _ = gcx.Writer.Write([]byte("event: message\n"))
				_, _ = gcx.Writer.Write([]byte("data: "))
				gcx.JSON(-1, v)
				_, _ = gcx.Writer.Write([]byte("\n"))
				_, _ = gcx.Writer.Write([]byte("\n"))
			}
		}

		if !mdl.Since.IsZero() {
			qer = qer.
				Where(fmt.Sprintf("ROWTIME >= %d", int(mdl.Since.UnixNano()/int64(time.Millisecond))))
		}

		sql, _, err := qer.ToSql()

		if err != nil {
			log.Error(err)
			httputil.InternalServerError(gcx, err)
			return
		}

		gcx.Header("Content-type", fmt.Sprintf("%s; charset=utf-8", ctp))

		qrq := &ksqldb.QueryRequest{
			SQL: fmt.Sprintf("%s EMIT CHANGES;", sql),
		}

		if !mdl.Since.IsZero() || len(mdl.Offsets) > 0 || len(mdl.SincePerPartition) > 0 {
			qrq.Properties = map[string]string{
				"ksql.streams.auto.offset.reset": "earliest",
			}
		}

		// To get the delay in getting the first message
		fst := false
		str := time.Now()

		cbk := func(hr *ksqldb.HeaderRow, row ksqldb.Row) error {
			if !fst {
				ltc := time.Since(str).Seconds()
				log.Info(
					"delay in getting the first message in seconds",
					log.Any("delay", ltc),
				)
				fst = true
			}

			art := new(schema.Article)
			err := ksqldb.
				NewDecoder(row, hr.ColumnNames).
				Decode(art)

			if err != nil {
				return err
			}

			// Populate partition, offset and date_published to the outgoing messages.
			// Should do this only if user requested the "event" object.
			// This can easily be checked by checking if the "event" object is nil or not.
			if art.Event != nil {
				lnt := len(row)
				ptn := row.Int(lnt - 3)
				art.Event.Partition = &ptn

				oft := int64(row[lnt-2].(float64))
				art.Event.Offset = &oft

				tux := time.UnixMilli(int64(row[lnt-1].(float64)))
				art.Event.DatePublished = &tux
			}

			write(art)
			gcx.Writer.Flush()

			return ctx.Err()
		}

		log.Info(
			"sending ksqldb push query",
			log.Any("sql", qrq.SQL),
			log.Any("properties", qrq.Properties),
		)

		if err := p.KSQLDB.Push(gcx.Request.Context(), qrq, cbk); err != nil {
			log.Error(err)
			httputil.InternalServerError(gcx, err)
		}
	}
}
