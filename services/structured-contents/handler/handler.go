// Handler for both structured-update and structured-bulk service
package handler

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"
	"wikimedia-enterprise/services/structured-contents/config/env"
	"wikimedia-enterprise/services/structured-contents/libraries/aggregate"
	"wikimedia-enterprise/services/structured-contents/packages/builder"
	"wikimedia-enterprise/services/structured-contents/submodules/config"
	"wikimedia-enterprise/services/structured-contents/submodules/log"
	pr "wikimedia-enterprise/services/structured-contents/submodules/prometheus"
	"wikimedia-enterprise/services/structured-contents/submodules/schema"
	"wikimedia-enterprise/services/structured-contents/submodules/subscriber"
	"wikimedia-enterprise/services/structured-contents/submodules/tracing"

	"github.com/PuerkitoBio/goquery"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/dig"
)

// Parameters dependency injection parameters for the handler.
type Parameters struct {
	dig.In
	Env        *env.Environment
	Aggregator aggregate.Aggregator
	Stream     schema.UnmarshalProducer
	Cfg        config.API
	Tracer     tracing.Tracer
	Metrics    *pr.Metrics
}

// New will produce an structured-contents article based on AvroStructured schema by consuming events from structured-data topics.
func New(p *Parameters) subscriber.Handler {
	return func(ctx context.Context, msg *kafka.Message) error {
		hcr := tracing.NewHeadersCarrier()
		hcr.FromKafkaHeaders(msg.Headers)
		hcx := hcr.ExtractContext(ctx)
		end, trx := p.Tracer.StartTrace(hcx, "structured-contents", map[string]string{})
		var err error

		defer func() {
			if err != nil {
				end(err, "error processing structured-content event")
			} else {
				end(nil, "structured-content event processed")
			}
		}()

		key := new(schema.Key)

		if err := p.Stream.Unmarshal(trx, msg.Key, key); err != nil {
			return err
		}

		art := new(schema.Article)

		if err := p.Stream.Unmarshal(trx, msg.Value, art); err != nil {
			return err
		}

		// Only generate structured-contents for some wiki articles in namespace 0
		if slices.Contains(p.Cfg.GetStructuredProjects(), art.IsPartOf.Identifier) && art.Namespace.Identifier == 0 &&
			art.Event != nil && len(art.Event.Type) > 0 {
			dtb := art.IsPartOf.Identifier
			// Build common fields for all event types
			bld := builder.NewStructuredBuilder().
				Name(art.Name).
				Identifier(art.Identifier).
				Version(art.Version).
				URL(art.URL).
				DateCreated(art.DateCreated).
				DateModified(art.DateModified).
				IsPartOf(art.IsPartOf).
				InLanguage(art.InLanguage)

			// Add more fields for specific event types
			switch art.Event.Type {
			case schema.EventTypeUpdate:
				doc, err := goquery.
					NewDocumentFromReader(
						strings.NewReader(art.ArticleBody.HTML))

				if err != nil {
					return err
				}

				// Get all parsed content and API data concurrently
				agg := new(aggregate.Aggregation)
				err = p.Aggregator.GetAggregation(
					trx, dtb, art.Name, doc, agg,
					aggregate.WithPageSummary(),
					aggregate.WithInfoboxes(),
					aggregate.WithSections(),
					aggregate.WithReferences(),
				)

				if err != nil {

					if "pagewasdeleted" == strings.ToLower(err.Error()) {
						log.Warn("page not found",
							log.Any("name", art.Name),
							log.Any("revision", art.Version.Identifier),
							log.Any("project", dtb),
							log.Any("event_id", art.Event.Identifier),
							log.Any("error", err))
					} else {
						log.Error("parsing error",
							log.Any("name", art.Name),
							log.Any("revision", art.Version.Identifier),
							log.Any("project", dtb),
							log.Any("event_id", art.Event.Identifier),
							log.Any("error", err))
					}

					return err
				}

				ibxs := []*schema.AvroHasParts{}

				for _, ibx := range agg.GetInfoboxes() {
					avrInf, err := builder.NewHasPartsBuilder().FromPart(ibx).Build().ToAvroStruct()

					if err != nil {
						fmt.Println(err)
						return err
					}
					ibxs = append(ibxs, avrInf)
				}

				secs := []*schema.AvroHasParts{}

				for _, sec := range agg.GetSections() {
					avrSec, err := builder.NewHasPartsBuilder().FromPart(sec).Build().ToAvroStruct()

					if err != nil {
						fmt.Println(err)
						return err
					}
					secs = append(secs, avrSec)
				}

				refs := make([]*schema.StructuredReference, 0)

				for _, ref := range agg.GetReferences() {

					fref := builder.NewStructuredReferenceBuilder().FromParserReference(ref).Build()

					if err != nil {
						fmt.Println(err)
						return err
					}
					refs = append(refs, fref)

				}
				lcs := schema.NewLicense()

				bld = bld.Abstract(art.Abstract).
					MainEntity(art.MainEntity).
					AdditionalEntities(art.AdditionalEntities).
					License([]*schema.License{lcs}). //License(art.License).
					Image(art.Image).
					Infoboxes(ibxs).
					Sections(secs).
					References(refs)

				if agg.GetPageSummary() != "" {
					bld = bld.Description(agg.GetPageSummary())
				}

			case schema.EventTypeVisibilityChange:
				bld = bld.Visibility(art.Visibility)
			}

			// Update datepublished to check the processing time performance
			evt := *art.Event
			dtn := time.Now().UTC()
			evt.SetDatePublished(&dtn)

			std := bld.Event(&evt).
				Build()

			tcs, err := p.Env.OutputCompactedTopics.
				GetNames(dtb, art.Namespace.Identifier)

			if err != nil {
				log.Error("error resolving compacted topic",
					log.Any("name", art.Name),
					log.Any("revision", art.Version.Identifier),
					log.Any("project", dtb),
					log.Any("event_id", art.Event.Identifier),
					log.Any("error", err),
				)

				return err
			}

			key.Type = schema.KeyTypeStructured
			mgs := []*schema.Message{}
			hds := tracing.NewHeadersCarrier().InjectContext(trx)

			// Only publish to noncompacted topic for structured-update service
			if len(p.Env.OutputNoncompactedTopic) > 0 {
				mgs = append(mgs, &schema.Message{
					Config:  schema.ConfigAvroStructured,
					Topic:   p.Env.OutputNoncompactedTopic,
					Value:   std,
					Key:     key,
					Headers: hds,
				})
			}

			// Events for various versions of compacted topic
			// No visibility event in compacted topic
			if art.Event.Type != schema.EventTypeVisibilityChange {
				for _, tpc := range tcs {
					mgs = append(mgs, &schema.Message{
						Config:  schema.ConfigAvroStructured,
						Topic:   tpc,
						Value:   std,
						Key:     key,
						Headers: hds,
					})
				}
			}

			return p.Stream.Produce(ctx, mgs...)
		}

		return nil
	}
}
