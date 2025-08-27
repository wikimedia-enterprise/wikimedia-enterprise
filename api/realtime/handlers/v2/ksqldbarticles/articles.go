// To be deleted
// Package ksqldbarticles holds the HTTP handler that will expose internal articles topic
// with additional filtering features.
package ksqldbarticles

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	"wikimedia-enterprise/api/realtime/config/env"
	"wikimedia-enterprise/api/realtime/libraries/metrics"
	"wikimedia-enterprise/api/realtime/libraries/resolver"
	"wikimedia-enterprise/api/realtime/submodules/httputil"
	"wikimedia-enterprise/api/realtime/submodules/ksqldb"
	"wikimedia-enterprise/api/realtime/submodules/log"
	"wikimedia-enterprise/api/realtime/submodules/schema"

	sqr "github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

// Parameters dependency injection for the handler.
type Parameters struct {
	dig.In
	KSQLDB      ksqldb.PushPuller
	Resolvers   resolver.Resolvers
	Env         *env.Environment
	Unmarshaler schema.Unmarshaler
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
	partitionsPerPart := nPartitions / maxParts

	for _, part := range m.Parts {
		for i := part * partitionsPerPart; i < (part+1)*partitionsPerPart; i++ {
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
func (m *Model) GetFormattedPartitions() string {
	sts := make([]string, len(m.partitions))

	for i, num := range m.partitions {
		sts[i] = fmt.Sprintf("%d", num)
	}

	return strings.Join(sts, ",")
}

// BuildInputFilters returns the proper WHERE clauses (that must be ANDed together) to apply the filters requested by the caller.
func (mdl *Model) BuildInputFilters(res resolver.Resolver) ([]string, error) {
	inputFilters := []string{}
	for _, filter := range mdl.Filters {
		// TODO: Remove this temporary solution,
		// this handles a case when keyword is being used
		// as a field name, in our case "namespace" field is like that
		for kwd, rpl := range res.Keywords {
			filter.Field = strings.ReplaceAll(filter.Field, kwd, rpl)
		}

		if !res.HasField(filter.Field) {
			err := fmt.Errorf("there is no such field '%s' to filter", filter.Field)
			return nil, err
		}

		if res.HasSlice(filter.Field) || res.HasStruct(filter.Field) {
			err := fmt.Errorf("can't filter on a '%s' object", filter.Field)
			return nil, err
		}

		ffd := res.GetField(filter.Field)
		rfv := reflect.ValueOf(filter.Value)

		if !rfv.Type().ConvertibleTo(ffd.Value.Type()) {
			err := fmt.Errorf("the filter '%s' should not be of type '%v'", filter.Field, rfv.Kind())
			return nil, err
		}

		if rfv.Kind() == reflect.String {
			inputFilters = append(inputFilters, fmt.Sprintf("%s = '%v'", ffd.Path, rfv.Convert(ffd.Value.Type()).Interface()))
		}

		if rfv.Kind() == reflect.Int || rfv.Kind() == reflect.Float64 || rfv.Kind() == reflect.Bool {
			inputFilters = append(inputFilters, fmt.Sprintf("%s = %v", ffd.Path, rfv.Convert(ffd.Value.Type()).Interface()))
		}
	}
	return inputFilters, nil
}

// ValidateAndProcessPartsAndPartitions validates the input from the request and sets in the model which partitions we are working with.
// For example, if we have 10 parts and 50 partitions, each part corresponds to 5 partitions. If a query comes in for part 1,
// `mdl.partitions` will be set to partitions 5, 6, 7, 8 and 9.
func (mdl *Model) ValidateAndProcessPartsAndPartitions(maxParts int, numPartitions int) error {
	if len(mdl.Parts) == 0 {
		// No "parts" requested, we will query all of them from KSQLDB.
		return nil
	}

	// Validate bounds for Parts
	if len(mdl.Parts) > maxParts {
		err := fmt.Errorf("the query param parts has length %d, max allowed length %d", len(mdl.Parts), maxParts)
		return err
	}

	for _, prt := range mdl.Parts {
		if prt < 0 || prt > maxParts-1 {
			err := fmt.Errorf("the part value is %d, allowed value should be in the range [0,%d]", prt, maxParts-1)
			return err
		}
	}

	if !mdl.Since.IsZero() {
		err := errors.New("for parallel consumption, specify either offsets or time-offsets (since_per_partition) parameter")
		return err
	}

	// Resolve partitions from Parts
	mdl.SetPartitions(numPartitions, maxParts)

	if len(mdl.Offsets) != 0 {
		// Validate bounds for Offsets
		if len(mdl.Offsets) > numPartitions {
			err := fmt.Errorf("the query param offsets has length %d, max allowed length %d", len(mdl.Offsets), numPartitions)
			return err
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
		if len(mdl.SincePerPartition) > numPartitions {
			err := fmt.Errorf("the query param since_per_partition has length %d, max allowed length %d", len(mdl.SincePerPartition), numPartitions)
			return err
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
				return err
			}
		}
	}

	return nil
}

// BuildPerPartitionFilters returns the list of WHERE clauses (that must be ANDed together) relative to per-partition filters.
// This includes: parts requested by the caller, offsets/timestamps requested by the caller and (if enabled) offsets observed so far
// in the request.
func (mdl *Model) BuildPerPartitionFilters() []string {
	filters := []string{}

	if len(mdl.Parts) == 0 {
		return filters
	}

	filters = append(filters, fmt.Sprintf("ROWPARTITION IN (%s)", mdl.GetFormattedPartitions()))
	if len(mdl.Offsets) == 0 && len(mdl.SincePerPartition) == 0 {
		return filters
	}

	cse := "CASE"
	for ptn, oft := range mdl.Offsets {
		cse = fmt.Sprintf("%s WHEN ROWPARTITION = %d THEN ROWOFFSET >= %v", cse, ptn, oft)
	}

	for ptn, tme := range mdl.SincePerPartition {
		cse = fmt.Sprintf("%s WHEN ROWPARTITION = %d THEN ROWTIME >= %v", cse, ptn, int(tme.UnixMilli()))
	}

	cse = fmt.Sprintf("%s ELSE TRUE END", cse)
	filters = append(filters, cse)
	return filters
}

// BuildRequest creates a KSQLDB query request object with the appropriate parameters based on the input of the request,
// the provided configuration and, if enabled, the offsets observed so far in the request.
func (mdl *Model) BuildRequest(cols []string, stream string, res *resolver.Resolver) (*ksqldb.QueryRequest, error) {
	filters := mdl.BuildPerPartitionFilters()

	inputFilters, err := mdl.BuildInputFilters(*res)
	if err != nil {
		return nil, err
	}

	filters = append(filters, inputFilters...)

	if !mdl.Since.IsZero() {
		filters = append(filters, fmt.Sprintf("ROWTIME >= %d", int(mdl.Since.UnixNano()/int64(time.Millisecond))))
	}

	qer := sqr.Select(cols...).From(stream)
	for _, f := range filters {
		qer = qer.Where(f)
	}

	sql, _, err := qer.ToSql()
	if err != nil {
		return nil, err
	}

	qrq := &ksqldb.QueryRequest{
		SQL: fmt.Sprintf("%s EMIT CHANGES;", sql),
	}

	if !mdl.Since.IsZero() || len(mdl.Offsets) > 0 || len(mdl.SincePerPartition) > 0 {
		qrq.Properties = map[string]string{
			"ksql.streams.auto.offset.reset": "earliest",
		}
	}

	return qrq, nil
}

// RetryCounter holds the state and provides logic to retry a given number of times with expontential back-off.
type RetryCounter struct {
	MaxRetries    int
	BackOffBaseMs int
	CurrentTry    int
}

// Reset sets back the counter to zero.
func (c *RetryCounter) Reset() {
	c.CurrentTry = 0
}

// ShouldRetry returns whether the maximum number of retries have been exhausted.
func (c *RetryCounter) ShouldRetry() bool {
	return c.CurrentTry < c.MaxRetries
}

// RegisterRetryAndWait increases the retry counter and waits the appropriate amount of time based on the counter's configuration.
func (c *RetryCounter) RegisterRetryAndWait() {
	time.Sleep(time.Millisecond * time.Duration(c.BackOffBaseMs) * time.Duration(math.Pow(2, float64(c.CurrentTry))))
	c.CurrentTry++
}

// NewHandler creates new articles HTTP handler.
func NewHandler(ctx context.Context, p *Parameters) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		mdl := new(Model)
		user := httputil.NewUser(gcx)
		userField := log.Any("username", user.GetUsername())
		useragent := gcx.Request.UserAgent()

		log.Info("received request", userField, log.Any("useragent", useragent), log.Any("request", gcx.Request.URL.String()))

		if err := gcx.ShouldBind(mdl); err != nil && err != io.EOF {
			log.Error(err, userField)
			httputil.UnprocessableEntity(gcx, err)
			return
		}

		if err := gcx.ShouldBindJSON(mdl); err != nil && err != io.EOF {
			log.Error(err, userField)
			httputil.UnprocessableEntity(gcx, err)
			return
		}

		log.Info("model bound", userField, log.Any("useragent", useragent), log.Any("request", fmt.Sprintf("%+v", mdl)))
		res := p.Resolvers[schema.KeyTypeArticle]

		for idx, field := range mdl.Fields {
			// Escape ksqldb keywords from fields such as namespace
			for kwd, rpl := range res.Keywords {
				field = strings.ReplaceAll(field, kwd, rpl)
				mdl.Fields[idx] = field
			}

			name := strings.TrimSuffix(field, ".*")

			if !res.HasField(name) {
				err := fmt.Errorf("field with name '%s' not found", field)
				log.Error(err, userField)
				httputil.UnprocessableEntity(gcx, err)
				return
			}

			if strings.HasSuffix(field, ".*") {
				continue
			}

			if res.HasSlice(name) || res.HasStruct(name) {
				err := fmt.Errorf("use wildcard '%s.*' to select all fields from the object", field)
				log.Error(err, userField)
				httputil.UnprocessableEntity(gcx, err)
				return
			}
		}

		cols := []string{}

		queryFields := mdl.Fields
		eventRequested := len(queryFields) == 0
		eventIsInternalRequested := eventRequested
		if !eventRequested {
			for _, field := range queryFields {
				if strings.Index(field, "event.") == 0 {
					eventRequested = true

					if strings.Index(field, "event.*") == 0 {
						eventIsInternalRequested = true
					}
				}
			}
		}
		if !eventIsInternalRequested {
			// event.is_internal_message must always be requested, to be used in code later in this handler.
			// If nothing from event was requested by the user, it will be discarded from the response.
			queryFields = append(queryFields, "event.is_internal_message")
		}

		if col := resolver.NewSelect(res, queryFields).GetSql(); len(col) > 0 {
			cols = append(cols, col)
		}

		if len(cols) == 0 {
			cols = append(cols, "*")
		}

		cols = append(cols, []string{"ROWPARTITION", "ROWOFFSET", "ROWTIME"}...)

		err := mdl.ValidateAndProcessPartsAndPartitions(p.Env.MaxParts, p.Env.Partitions)
		if err != nil {
			log.Error(err, userField)
			httputil.UnprocessableEntity(gcx, err)
			return
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

		gcx.Header("Content-type", fmt.Sprintf("%s; charset=utf-8", ctp))

		// To get the delay in getting the first message
		fst := false
		str := time.Now()
		go func() {
			tmr := time.NewTimer(time.Minute * 5)
			defer tmr.Stop()
			select {
			case <-tmr.C:
				if !fst {
					log.Error("no messages received from KSQLDB in first 5 minutes", userField, log.Any("request", gcx.Request.URL.String()))
				}
			case <-gcx.Request.Context().Done():
			}
		}()

		reqLabels := []string{
			strconv.FormatBool(len(mdl.Fields) > 0),
			strconv.FormatBool(len(mdl.Filters) > 0),
			strconv.FormatBool(!mdl.Since.IsZero()),
			strconv.FormatBool(len(mdl.Offsets) > 0),
			strconv.FormatBool(len(mdl.SincePerPartition) > 0),
			"article", // schema
			"ksqldb",  // handler_type
			user.GetUsername(),
		}

		metrics.RequestsTotal.WithLabelValues(reqLabels...).Inc()
		metrics.OpenConnections.WithLabelValues(reqLabels...).Inc()
		defer metrics.OpenConnections.WithLabelValues(reqLabels...).Dec()

		throughputTracker := metrics.NewThroughputTracker(15, metrics.PerConnectionThroughputSeconds, reqLabels)
		throughputTracker.StartTracking(gcx.Request.Context())

		idleTracker := metrics.NewIdleTracker(metrics.ConnectionTimeIdleSeconds, reqLabels)
		idleTracker.StartTracking(gcx.Request.Context())

		var qrq *ksqldb.QueryRequest
		msgsPerSecond := float32(p.Env.ThrottlingMsgsPerSecond)
		if p.Env.PerPartThrottling && len(mdl.Parts) > 0 {
			msgsPerSecond = float32(msgsPerSecond) * float32(len(mdl.Parts)) / float32(p.Env.MaxParts)
		}
		throttle := Throttle{
			msgsPerSecond: msgsPerSecond,
			batchSize:     p.Env.ThrottlingBatchSize,
		}
		throttle.Start()
		cbk := func(hr *ksqldb.HeaderRow, row ksqldb.Row) error {
			throttle.Apply()

			if !fst {
				ltc := time.Since(str).Seconds()
				log.Info(
					"delay in getting the first message in seconds",
					log.Any("delay", ltc),
					userField,
				)
				metrics.FirstMessageLatencySeconds.WithLabelValues(reqLabels...).Observe(ltc)
				fst = true
			}

			numCols := len(row)
			rowOffset := int64(row[numCols-2].(float64))
			rowPartition := row.Int(numCols - 3)

			art := new(schema.Article)
			err := ksqldb.
				NewDecoder(row, hr.ColumnNames).
				Decode(art)

			if err != nil {
				return err
			}

			if art.Event == nil {
				log.Warn("no event returned by ksqldb", log.Any("article", art.Name), log.Any("query", qrq.SQL), userField)
			} else if art.Event.IsInternal != nil && *art.Event.IsInternal {
				// Discard internal-only events, like those produced by the take-down pipeline.
				return gcx.Request.Context().Err()
			}

			// Populate partition, offset and date_published in the outgoing messages.
			// Should do this only if user requested the "event" object, otherwise it must be discarded.
			if !eventRequested {
				art.Event = nil
			} else if art.Event != nil {
				art.Event.Partition = &rowPartition
				art.Event.Offset = &rowOffset

				tux := time.UnixMilli(int64(row[numCols-1].(float64)))
				art.Event.DatePublished = &tux
			}

			throughputTracker.RecordMessage()
			idleTracker.RecordMessage()
			write(art)
			gcx.Writer.Flush()
			metrics.MessagesSent.WithLabelValues(reqLabels...).Inc()

			return gcx.Request.Context().Err()
		}

		qrq, err = mdl.BuildRequest(cols, p.Env.ArticlesStream, res)
		if err != nil {
			log.Error(err, userField)
			httputil.UnprocessableEntity(gcx, err)
			return
		}

		log.Info(
			"sending ksqldb push query",
			log.Any("sql", qrq.SQL),
			log.Any("properties", qrq.Properties),
			userField,
			log.Any("useragent", useragent),
		)

		if err := p.KSQLDB.Push(gcx.Request.Context(), qrq, cbk); err != nil {
			log.Error(err, userField)
			httputil.InternalServerError(gcx, err)
		}
	}
}
