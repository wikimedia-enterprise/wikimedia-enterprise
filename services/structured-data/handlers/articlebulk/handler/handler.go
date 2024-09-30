package handler

import (
	"context"
	"fmt"
	"time"

	"strings"
	"wikimedia-enterprise/general/log"
	"wikimedia-enterprise/general/parser"
	pr "wikimedia-enterprise/general/prometheus"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/general/subscriber"
	"wikimedia-enterprise/general/tracing"
	"wikimedia-enterprise/services/structured-data/config/env"
	"wikimedia-enterprise/services/structured-data/libraries/aggregate"
	"wikimedia-enterprise/services/structured-data/packages/abstract"
	"wikimedia-enterprise/services/structured-data/packages/builder"
	"wikimedia-enterprise/services/structured-data/packages/exponential"

	"github.com/PuerkitoBio/goquery"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/dig"
)

// Parameters dependency injection parameters for the handler.
type Parameters struct {
	dig.In
	Env        *env.Environment
	Stream     schema.UnmarshalProducer
	Aggregator aggregate.Aggregator
	Parser     parser.API
	Tracer     tracing.Tracer
	Metrics    *pr.Metrics
}

// NewArticleBulk will produce an articles schema messages for all article titles consumed from a kafka message
//
// This handler will fetch from mediawiki API the information for all corresponding articles
// and build an article schema instance with the latest data available.
func NewArticleBulk(p *Parameters) subscriber.Handler {
	return func(ctx context.Context, msg *kafka.Message) error {
		hcr := tracing.NewHeadersCarrier()
		hcr.FromKafkaHeaders(msg.Headers)
		hcx := hcr.ExtractContext(ctx)
		end, trx := p.Tracer.StartTrace(hcx, "article-bulk", map[string]string{})
		var err error
		defer func() {
			if err != nil {
				end(err, "error processing article-bulk event")
			} else {
				end(nil, "article-bulk event processed")
			}
		}()

		pld := new(schema.ArticleNames)

		if err := p.Stream.Unmarshal(trx, msg.Value, pld); err != nil {
			return err
		}

		// Introducing gradual delay for failed events processing
		// based on the exponential backoff.
		if pld.Event != nil && pld.Event.FailCount > 0 {
			epn := exponential.GetNth(uint(pld.Event.FailCount+1), uint(p.Env.BackOffBase))
			time.Sleep(time.Duration(epn) * time.Second)
		}

		tls := pld.Names
		dtb := pld.IsPartOf.Identifier
		ags := map[string]*aggregate.Aggregation{}

		err = p.Aggregator.GetAggregations(
			trx, dtb, tls, ags,
			aggregate.WithPages(1, len(tls)),
			// aggregate.WithRevisions(len(tls)),
			aggregate.WithPagesHTML(len(tls)),
		)

		if err != nil {
			log.Error("wmf api error",
				log.Any("articles queried concurrently", len(tls)),
				log.Any("project", dtb),
				log.Any("error", err),
			)

			return err
		}

		mgs := []*schema.Message{}

		for _, agg := range ags {
			if err := agg.GetPageHTMLError(); err != nil {
				log.Warn("wmf page html api error",
					log.Any("name", agg.GetPageTitle()),
					log.Any("project", dtb),
					log.Any("error", err),
				)

				continue
			}

			if agg.GetPageMissing() {
				log.Warn(
					"page is missing",
					log.Any("name", agg.GetPageTitle()),
					log.Any("project", dtb),
					log.Any("url", pld.IsPartOf.URL),
				)

				continue
			}

			doc, err := goquery.
				NewDocumentFromReader(
					strings.NewReader(agg.GetPageHTMLContent()))

			if err != nil {
				return err
			}

			abs := ""

			if abstract.IsValid(agg.GetPageNs(), agg.GetCurrentRevisionContentModel()) {
				abs, err = p.Parser.GetAbstract(doc.Selection)

				if err != nil {
					log.Warn("abstract processing failed",
						log.Any("name", agg.GetPageTitle()),
						log.Any("url", pld.IsPartOf.URL),
						log.Any("error", err),
					)
				}
			}

			vid := fmt.Sprintf("/%s/%d", pld.IsPartOf.Identifier, agg.GetPageLastRevID())
			crc := agg.GetCurrentRevisionContent()
			cts := p.Parser.GetCategories(doc.Selection)
			tps := p.Parser.GetTemplates(doc.Selection)

			edr := builder.NewEditorBuilder().
				Identifier(agg.GetCurrentRevisionUserID()).
				Name(agg.GetCurrentRevisionUser()).
				IsAnonymous(agg.GetCurrentRevisionUserID() == 0).
				Build()

			vrs := builder.NewVersionBuilder().
				Identifier(agg.GetPageLastRevID()).
				NumberOfCharacters(crc).
				Comment(agg.GetCurrentRevisionComment()).
				Tags(agg.GetCurrentRevisionTags()).
				IsMinorEdit(agg.GetCurrentRevisionMinor()).
				IsFlaggedStable(agg.GetPageFlaggedStableRevID() == agg.GetPageLastRevID()).
				Editor(edr).
				Size(crc).
				Noindex(tps.HasSubstringInNames(p.Env.NoindexTemplatePatterns) || cts.HasSubstringInNames(p.Env.NoindexCategoryPatterns)).
				Build()

			art := builder.NewArticleBuilder(pld.IsPartOf.URL).
				Event(schema.NewEvent(schema.EventTypeUpdate)).
				Name(agg.GetPageTitle()).
				Abstract(abs).
				WatchersCount(agg.GetPageWatchers()).
				Identifier(agg.GetPageID()).
				// DateCreated(agg.GetFirstRevisionTimestamp()).
				DateModified(agg.GetCurrentRevisionTimestamp()).
				// DatePreviouslyModified(agg.GetPreviousRevisionTimestamp()).
				// PreviousVersion(agg.GetPreviousRevisionID(), agg.GetPreviousRevisionContent()).
				URL(agg.GetPageCanonicalURL()).
				IsPartOf(pld.IsPartOf).
				Version(vrs).
				VersionIdentifier(vid).
				Body(crc, agg.GetPageHTMLContent()).
				License(schema.NewLicense()).
				MainEntity(agg.GetPagePropsWikiBaseItem()).
				AdditionalEntities(agg.GetPageWbEntityUsage()).
				Protection(agg.GetPageProtection()).
				Categories(p.Parser.GetCategories(doc.Selection)).
				Templates(p.Parser.GetTemplates(doc.Selection)).
				Redirects(agg.GetPageRedirects()).
				Namespace(&schema.Namespace{Identifier: agg.GetPageNs()}).
				InLanguage(&schema.Language{Identifier: agg.GetPageLanguage()}).
				Image(agg.GetPageOriginalImage(), agg.GetPageThumbnailImage()).
				Build()

			key := &schema.Key{
				Identifier: fmt.Sprintf("/%s/%s", pld.IsPartOf.Identifier, strings.ReplaceAll(art.Name, " ", "_")),
				Type:       schema.KeyTypeArticle,
			}

			tcs, err := p.Env.Topics.GetNames(dtb, agg.GetPageNs())

			if err != nil {
				return err
			}

			hds := tracing.NewHeadersCarrier().InjectContext(hcx)

			for _, tpc := range tcs {
				mgs = append(mgs, &schema.Message{
					Config:  schema.ConfigArticle,
					Topic:   tpc,
					Value:   art,
					Key:     key,
					Headers: hds,
				})
			}
		}

		return p.Stream.Produce(hcx, mgs...)
	}
}
