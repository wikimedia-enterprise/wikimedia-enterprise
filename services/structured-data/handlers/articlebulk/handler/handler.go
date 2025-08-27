package handler

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"strings"
	"wikimedia-enterprise/services/structured-data/config/env"
	"wikimedia-enterprise/services/structured-data/libraries/aggregate"
	"wikimedia-enterprise/services/structured-data/packages/abstract"
	"wikimedia-enterprise/services/structured-data/packages/builder"
	"wikimedia-enterprise/services/structured-data/packages/exponential"
	"wikimedia-enterprise/services/structured-data/submodules/config"
	"wikimedia-enterprise/services/structured-data/submodules/log"
	"wikimedia-enterprise/services/structured-data/submodules/parser"
	pr "wikimedia-enterprise/services/structured-data/submodules/prometheus"
	"wikimedia-enterprise/services/structured-data/submodules/schema"
	"wikimedia-enterprise/services/structured-data/submodules/subscriber"
	"wikimedia-enterprise/services/structured-data/submodules/tracing"

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
	Cfg        config.API
}

var (
	redirectPattern = regexp.MustCompile(`Special:Redirect/revision/(\d+)`)
)

// NewArticleBulk will produce an articles schema messages for all article titles consumed from a kafka message
//
// This handler will fetch from mediawiki API the information for all corresponding articles
// and build an article schema instance with the latest data available.
func NewArticleBulk(p *Parameters) subscriber.Handler {
	return func(ctx context.Context, msg *kafka.Message) error {
		hcr := tracing.NewHeadersCarrier()
		hcr.FromKafkaHeaders(msg.Headers)
		hcx := hcr.ExtractContext(ctx)

		end, trx := p.Tracer.StartTrace(hcx, "article-bulk", map[string]string{
			"message_partition": fmt.Sprintf("%d", msg.TopicPartition.Partition),
			"message_offset":    fmt.Sprintf("%d", msg.TopicPartition.Offset),
		})

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

		if err != nil {
			return err
		}

		// Introducing gradual delay for failed events processing
		// based on the exponential backoff.
		if pld.Event != nil && pld.Event.FailCount > 0 {
			epn := exponential.GetNth(uint(pld.Event.FailCount+1), uint(p.Env.BackOffBase))
			delay := time.Duration(epn) * time.Second
			time.Sleep(delay)

			log.Info("event failed count ",
				log.Any("count", pld.Event.FailCount),
				log.Any("failed reason", pld.Event.FailReason),
				log.Any("delay", delay),
				log.Any("identifier", pld.Event.Identifier),
				log.Any("names", pld.Names),
				log.Any("partition", pld.Event.Partition),
				log.Any("offset", pld.Event.Offset))
		}

		tls := pld.Names
		dtb := pld.IsPartOf.Identifier
		ags := map[string]*aggregate.Aggregation{}

		err = p.Aggregator.GetAggregations(
			trx, dtb, tls, ags,
			aggregate.WithPages(2, len(tls)), //no version or timestamp to tell it when to go back in time , so this is the latest revision Page
		)

		if err != nil {
			log.Error("wmf api error",
				log.Any("articles queried concurrently", len(tls)),
				log.Any("project", dtb),
				log.Any("error", err),
			)

			return err
		}

		// create lookups for revision ids and titles used by different API calls
		rtm := map[string]string{} // revision id to title mapping
		trv := map[string]int{}    // title to revision id mapping for RevertRisk API call
		rvs := []string{}          // revision ids slice
		tps := []string{}          // titles slice

		for _, agg := range ags {
			if agg.GetPageLastRevID() == 0 {
				ttl := agg.GetPageTitle()
				if len(ttl) > 0 {
					tps = append(tps, ttl)
				}
				continue
			}

			rvs = append(rvs, fmt.Sprintf("%d", agg.GetPageLastRevID()))
			rtm[fmt.Sprintf("%d", agg.GetPageLastRevID())] = agg.GetPageTitle()

			// trv slice is used to get the RevertRisk score for each article, but only if in namespace 0, is a wiki and has previous revision
			if agg.GetPageNs() == 0 && strings.HasSuffix(dtb, "wiki") && agg.GetPreviousRevisionID() != 0 {
				trv[agg.GetPageTitle()] = agg.GetPageLastRevID()
			}
		}

		if len(tps) > 0 {
			err = p.Aggregator.GetAggregations(
				trx, dtb, tps, ags,
				aggregate.WithPagesHTML(len(tps)),
			)

			if err != nil {
				log.Error("wmf api error no HTML from title API",
					log.Any("articles queried concurrently", len(rvs)),
					log.Any("project", dtb),
					log.Any("error", err),
				)
			} else {
				for ttl, agg := range ags {
					match := redirectPattern.FindStringSubmatch(agg.PageHTML.Content)
					rid := strconv.Itoa(agg.GetPageLastRevID())
					if len(match) > 1 && match[1] != rid {
						log.Warn("revision mismatch in HTML using title API",
							log.Any("project", dtb),
							log.Any("title", ttl),
							log.Any("expected revision id", rid),
							log.Any("found revision id", match[1]),
							log.Any("revision from Page object", agg.PageHTML.Revision),
						)
					}
				}
			}
		}

		if len(rvs) > 0 {
			err = p.Aggregator.GetAggregations(
				trx, dtb, rvs, ags,
				aggregate.WithRevisionsHTML(len(rvs)),
			)

			if err != nil {
				log.Error("wmf api error no HTML from revision API",
					log.Any("articles queried concurrently", len(rvs)),
					log.Any("project", dtb),
					log.Any("error", err),
				)
			}

			for rid, ttl := range rtm {
				agg, ok := ags[rid]
				if !ok {
					log.Warn("revision key not found when remapping to title",
						log.Any("revision id", rid),
						log.Any("project", dtb),
					)

					continue
				}

				if ags[ttl] == nil {
					ags[ttl] = agg
				}

				ags[ttl].PageHTML = agg.GetPageHTML()
				delete(ags, rid)

				match := redirectPattern.FindStringSubmatch(ags[ttl].PageHTML.Content)
				if len(match) > 1 && match[1] != rid {
					log.Warn("revision mismatch in HTML using revision API",
						log.Any("project", dtb),
						log.Any("expected revision id", rid),
						log.Any("found revision id", match[1]),
						log.Any("revision from Page object", ags[ttl].PageHTML.Revision),
						log.Any("title", ttl),
					)
				}
			}
		}

		// get RevertRisk scores for wiki articles in namespace 0 with previous revision
		// GetAggregations call is done after calls for `WithRevisionsHTML` and `WithPagesHTML`
		// because it has a different set of articles in `trv` compared to the other calls
		if len(trv) > 0 {
			lng := p.Cfg.GetLanguage(pld.IsPartOf.Identifier)
			tts := []string{}
			for ttl := range trv {
				tts = append(tts, ttl)
			}

			err = p.Aggregator.GetAggregations(
				trx, dtb, tts, ags,
				aggregate.WithScores(trv, lng, time.Duration(p.Env.LiftwingTimeoutMs)*time.Millisecond, "revertrisk", dtb),
			)

			if err != nil {
				log.Error("wmf api error from revert risk API",
					log.Any("articles queried concurrently", len(tts)),
					log.Any("articles", tts),
					log.Any("project", dtb),
					log.Any("error", err),
				)
			}
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
				log.Warn(
					"error reading HTML Content",
					log.Any("name", agg.GetPageTitle()),
					log.Any("project", dtb),
					log.Any("url", pld.IsPartOf.URL),
				)

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
			scr := builder.NewScoresBuilder().RevertRisk(agg.GetRevertRiskScore()).Build()

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
				Scores(scr).
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
				log.Warn("error getting topic names",
					log.Any("project", dtb),
					log.Any("namespace", agg.GetPageNs()),
					log.Any("error", err),
				)
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
