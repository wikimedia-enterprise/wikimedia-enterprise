// Package articles holds the HTTP handler that will expose internal articles topic
// with additional filtering features.
package articles

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"wikimedia-enterprise/api/realtime/config/env"
	libkafka "wikimedia-enterprise/api/realtime/libraries/kafka"
	"wikimedia-enterprise/api/realtime/libraries/metrics"
	"wikimedia-enterprise/api/realtime/libraries/resolver"
	"wikimedia-enterprise/api/realtime/submodules/httputil"
	"wikimedia-enterprise/api/realtime/submodules/log"
	"wikimedia-enterprise/api/realtime/submodules/schema"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/dig"
	"golang.org/x/time/rate"
)

// Parameters dependency injection for the handler.
type Parameters struct {
	dig.In
	Resolvers   resolver.Resolvers
	Env         *env.Environment
	Unmarshaler schema.Unmarshaler
	Consumers   libkafka.ConsumerGetter
}

// NewHandler creates a new schema-agnostic HTTP handler to serve events from a kafka topic as SSE.
func NewHandler(ctx context.Context, p *Parameters, topic string, schemaType string) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		user := httputil.NewUser(gcx).GetUsername()
		useragent := gcx.Request.UserAgent()
		log.Info("received request",
			log.Any("schema", schemaType),
			log.Any("username", user),
			log.Any("useragent", useragent),
			log.Any("request", gcx.Request.URL.String()))

		mdl := NewModel(p.Env.MaxParts, p.Env.Partitions)
		res := p.Resolvers[schemaType]

		if err := mdl.Parse(gcx, res, p.Env.MaxParts, p.Env.Partitions); err != nil {
			log.Error("error parsing request",
				log.Any("schema", schemaType),
				log.Any("username", user),
				log.Any("useragent", useragent),
				log.Any("request", gcx.Request.URL.String()),
				log.Any("error", err))
			httputil.UnprocessableEntity(gcx, err)
			return
		}

		log.Debug("parsed model", log.Any("schema", schemaType),
			log.Any("username", user),
			log.Any("useragent", useragent),
			log.Any("filters", mdl.Filters),
			log.Any("fields", mdl.Fields),
			log.Any("parsed_partitions", mdl.partitions),
			log.Any("offsets", mdl.Offsets),
			log.Any("since", mdl.Since),
			log.Any("partitions_timestamps", mdl.SincePerPartition),
			log.Any("request", gcx.Request.URL.String()))

		var (
			valueFilter func(any) bool
			modifier    func(any)
		)

		if len(mdl.Filters) > 0 {
			valueFilter = res.NewValueFilter(mdl.Filters)
		}

		if len(mdl.Fields) > 0 {
			modifier = res.NewModifier(mdl.Fields)
		}

		sch := p.Resolvers.GetSchema(schemaType)
		if sch == nil {
			log.Error(resolver.ErrSchemaTypeNotDefined, log.Any("schema", schemaType),
				log.Any("username", user), log.Any("useragent", useragent),
				log.Any("request", gcx.Request.URL.String()))
			httputil.InternalServerError(gcx, resolver.ErrSchemaTypeNotDefined)
			return
		}

		schPtr := reflect.TypeOf(sch)
		articles := make(chan any, p.Env.ArticleChannelSize)

		labels := []string{
			strconv.FormatBool(len(mdl.Fields) > 0),
			strconv.FormatBool(len(mdl.Filters) > 0),
			strconv.FormatBool(!mdl.Since.IsZero()),
			strconv.FormatBool(len(mdl.Offsets) > 0),
			strconv.FormatBool(len(mdl.SincePerPartition) > 0),
			schemaType,
			"articles", // handler_type
			user,
		}

		requestsTotal := metrics.RequestsTotal.WithLabelValues(labels...)
		openConnections := metrics.OpenConnections.WithLabelValues(labels...)

		requestsTotal.Inc()
		openConnections.Inc()
		defer openConnections.Dec()

		// Initialize throughput and idle trackers
		throughputTracker := metrics.NewThroughputTracker(15, metrics.PerConnectionThroughputSeconds, labels)
		throughputTracker.StartTracking(gcx.Request.Context())

		idleTracker := metrics.NewIdleTracker(metrics.ConnectionTimeIdleSeconds, labels)
		idleTracker.StartTracking(gcx.Request.Context())
		// To get the delay in getting the first message.
		fst := false
		str := time.Now()

		ctx := gcx.Request.Context()
		callback := func(msg *kafka.Message) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			if !fst {
				ltc := time.Since(str).Seconds()
				log.Info(
					"delay in getting the first message in seconds",
					log.Any("delay", ltc),
					log.Any("schema", schemaType),
					log.Any("username", user),
					log.Any("useragent", useragent),
					log.Any("request", gcx.Request.URL.String()),
				)
				metrics.FirstMessageLatencySeconds.WithLabelValues(labels...).Observe(ltc)
				fst = true
			}

			inst := reflect.New(schPtr.Elem()).Interface()
			err := p.Unmarshaler.Unmarshal(ctx, msg.Value, inst)

			if err != nil {
				log.Error("error unmarshaling message", log.Any("key", msg.Key),
					log.Any("username", user), log.Any("useragent", useragent),
					log.Any("request", gcx.Request.URL.String()), log.Any("error", err))
				return err
			}

			if isInternalEvent(inst) {
				log.Debug("skipping internal event", log.Any("topic", topic), log.Any("partition", msg.TopicPartition.Partition),
					log.Any("offset", msg.TopicPartition.Offset),
					log.Any("username", user),
					log.Any("useragent", useragent),
					log.Any("request", gcx.Request.URL.String()))

				return ctx.Err()
			}

			if valueFilter != nil {
				var keep bool
				timer := prometheus.NewTimer(metrics.FilterLatencySeconds.WithLabelValues(labels...))
				keep = valueFilter(inst)
				timer.ObserveDuration()

				if !keep {
					log.Debug("filtered out message", log.Any("topic", topic), log.Any("partition", msg.TopicPartition.Partition),
						log.Any("offset", msg.TopicPartition.Offset),
						log.Any("username", user),
						log.Any("useragent", useragent),
						log.Any("request", gcx.Request.URL.String()))
					return ctx.Err()
				}
			}

			if len(mdl.Fields) > 0 {
				timer := prometheus.NewTimer(metrics.ModifyLatencySeconds.WithLabelValues(labels...))
				modifier(inst)
				timer.ObserveDuration()

				log.Debug("modified message", log.Any("topic", topic), log.Any("partition", msg.TopicPartition.Partition),
					log.Any("offset", msg.TopicPartition.Offset),
					log.Any("username", user),
					log.Any("useragent", useragent),
					log.Any("request", gcx.Request.URL.String()))
			}

			setPartitionOffset(inst, msg)

			log.Debug("appending message to channel", log.Any("topic", topic), log.Any("partition", msg.TopicPartition.Partition),
				log.Any("offset", msg.TopicPartition.Offset),
				log.Any("username", user),
				log.Any("useragent", useragent),
				log.Any("request", gcx.Request.URL.String()))

			// Record message for tracking
			throughputTracker.RecordMessage()
			idleTracker.RecordMessage()

			articles <- inst
			return nil
		}

		readParams := &libkafka.ReadParams{
			Since:              mdl.Since,
			SincePerPartition:  mdl.SincePerPartition,
			OffsetPerPartition: mdl.Offsets,
		}

		partitionsPerWorker := mdl.DistributePartitions(p.Env.Workers)
		readContext, cancel := context.WithCancel(ctx)
		defer cancel()

		// errCh will get exactly one “first” error
		errCh := make(chan error, 1)
		var wg sync.WaitGroup

		log.Debug("about to open consumers",
			log.Any("consumers_count", p.Env.Workers),
			log.Any("offsets_config", readParams),
			log.Any("schema", schemaType),
			log.Any("username", user),
			log.Any("useragent", useragent),
			log.Any("filters", mdl.Filters),
			log.Any("fields", mdl.Fields),
			log.Any("parsed_partitions", mdl.partitions),
			log.Any("offsets", mdl.Offsets),
			log.Any("since", mdl.Since),
			log.Any("partitions_timestamps", mdl.SincePerPartition),
			log.Any("request", gcx.Request.URL.String()))

		for worker, subPartitions := range partitionsPerWorker {
			wg.Add(1)
			go func(worker int, subPartitions []int) {
				defer wg.Done()

				cid := getConsumerId(user, worker)
				consumer, err := p.Consumers.GetConsumer(cid)

				if err != nil {
					log.Error("error getting consumer",
						log.Any("consumer_id", cid),
						log.Any("username", user),
						log.Any("useragent", useragent),
						log.Any("request", gcx.Request.URL.String()),
						log.Any("error", err),
					)

					// send first error (non-blocking) and cancel everyone
					select {
					case errCh <- fmt.Errorf("error getting consumer: %w", err):
					default:
					}
					cancel()
					return
				}

				defer func() {
					if err := consumer.Close(); err != nil {
						log.Error("error closing consumer",
							log.Any("consumer_id", cid),
							log.Any("username", user),
							log.Any("useragent", useragent),
							log.Any("request", gcx.Request.URL.String()),
							log.Any("error", err),
						)
					}
				}()

				// this blocks until readContext is done or a callback error bubbles up (only on serious schema problem)
				if err := consumer.ReadAll(readContext, readParams, topic, subPartitions, callback); err != nil {
					// if it wasn’t just our cancellation, record it
					if err != readContext.Err() {
						select {
						case errCh <- fmt.Errorf("readAll error (worker %d): %w", worker, err):
						default:
						}
						cancel()
					}
					return
				}
			}(worker, subPartitions)
		}

		// once all read workers finish, log any error and close articles
		go func() {
			wg.Wait()
			// log the first error, if any
			select {
			case err := <-errCh:
				log.Error("one or more read workers failed", log.Any("error", err))
			default:
			}
			close(articles)
		}()

		write := getWriteFunc(gcx)
		msgsPerSecond := float32(p.Env.ThrottlingMsgsPerSecond)
		if p.Env.PerPartThrottling && len(mdl.Parts) > 0 {
			msgsPerSecond = float32(msgsPerSecond) * float32(len(mdl.Parts)) / float32(p.Env.MaxParts)
		}
		limiter := rate.NewLimiter(rate.Limit(msgsPerSecond), 1)

		for {
			select {
			case <-ctx.Done():
				return
			case art, ok := <-articles:
				if !ok {
					return // all read workers finished
				}

				// blocks till next event should be written, maintains output rate
				if err := limiter.Wait(ctx); err != nil {
					// context cancelled
					return
				}

				write(art)
				gcx.Writer.Flush()
				metrics.MessagesSent.WithLabelValues(labels...).Inc()
			}
		}
	}
}

// getWriteFunc takes a gin.Context, negotiates content type, sets the
// Content-Type header, and returns a WriteFunc you can use to emit events.
func getWriteFunc(c *gin.Context) func(any) {
	ctp := c.NegotiateFormat(httputil.MIMEEVENTSTEAM, httputil.MIMENDJSON)
	c.Header("Content-Type", fmt.Sprintf("%s; charset=utf-8", ctp))

	// if SSE was requested, wrap gin.Context.JSON in the SSE framing
	if ctp == httputil.MIMEEVENTSTEAM {
		return func(v any) {
			_, _ = c.Writer.Write([]byte("event: message\n"))
			_, _ = c.Writer.Write([]byte("data: "))
			c.JSON(-1, v)
			_, _ = c.Writer.Write([]byte("\n"))
			_, _ = c.Writer.Write([]byte("\n"))
		}
	}

	// fallback to ndjson
	return func(v any) {
		c.JSON(-1, v)
		_, _ = c.Writer.WriteString("\n")
	}
}

// getConsumerId gets a consumer id in the format realtime-<userid>-<timestamp>-<worker>-<suffix>
func getConsumerId(user string, worker int) string {
	return fmt.Sprintf("realtime-%s-%s-%d-%s", user, time.Now().UTC().Format("20060102T150405"), worker, strings.Split(uuid.New().String(), "-")[0])
}

func isInternalEvent(instance any) bool {
	v := reflect.ValueOf(instance).Elem()
	evtFld := v.FieldByName("Event")

	if evtFld.IsValid() && !evtFld.IsNil() {
		ev := evtFld.Elem()
		isIntFld := ev.FieldByName("IsInternal")
		if isIntFld.IsValid() && !isIntFld.IsNil() && isIntFld.Elem().Bool() {
			return true
		}
	}

	return false
}

func setPartitionOffset(instance any, msg *kafka.Message) {
	v := reflect.ValueOf(instance).Elem()
	evtFld := v.FieldByName("Event")

	if evtFld.IsValid() && !evtFld.IsNil() {
		evt := evtFld.Elem()

		partFld := evt.FieldByName("Partition")
		if partFld.IsValid() && partFld.CanSet() {
			newPart := reflect.New(partFld.Type().Elem()) // *int
			newPart.Elem().SetInt(int64(msg.TopicPartition.Partition))
			partFld.Set(newPart)
		}

		offFld := evt.FieldByName("Offset")
		if offFld.IsValid() && offFld.CanSet() {
			newOff := reflect.New(offFld.Type().Elem()) // *int64
			newOff.Elem().SetInt(int64(msg.TopicPartition.Offset))
			offFld.Set(newOff)
		}
	}
}
