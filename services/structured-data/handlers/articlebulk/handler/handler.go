package handler

import (
	"context"
	"fmt"
	"log"
	"strings"
	"wikimedia-enterprise/general/parser"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/general/subscriber"
	"wikimedia-enterprise/services/structured-data/config/env"
	"wikimedia-enterprise/services/structured-data/libraries/aggregate"
	"wikimedia-enterprise/services/structured-data/packages/abstract"
	"wikimedia-enterprise/services/structured-data/packages/builder"

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
}

// NewArticleBulk will produce an articles schema messages for all article titles consumed from a kafka message
//
// This handler will fetch from mediawiki API the information for all corresponding articles
// and build an article schema instance with the latest data available.
func NewArticleBulk(p *Parameters) subscriber.Handler {
	return func(ctx context.Context, msg *kafka.Message) error {
		pld := new(schema.ArticleNames)

		if err := p.Stream.Unmarshal(ctx, msg.Value, pld); err != nil {
			return err
		}

		tls := pld.Names
		dtb := pld.IsPartOf.Identifier
		ags := map[string]*aggregate.Aggregation{}

		err := p.Aggregator.GetAggregations(
			ctx, dtb, tls, ags,
			aggregate.WithPages(1, len(tls)),
			// aggregate.WithRevisions(len(tls)),
			aggregate.WithPagesHTML(len(tls)),
		)

		if err != nil {
			return err
		}

		mgs := []*schema.Message{}

		for _, agg := range ags {
			if err := agg.GetPageHTMLError(); err != nil {
				log.Println(err)
				continue
			}

			if agg.GetPageMissing() {
				log.Printf("page '%s/wiki/%s' is missing\n", pld.IsPartOf.URL, agg.GetPageTitle())
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
					log.Printf(
						"abstract processing for page '%s/wiki/%s' failed with error '%s'\n",
						pld.IsPartOf.URL,
						agg.GetPageTitle(),
						err,
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

			tpc, err := p.Env.Topics.GetName(dtb, agg.GetPageNs())

			if err != nil {
				return err
			}

			msg := &schema.Message{
				Config: schema.ConfigArticle,
				Topic:  tpc,
				Value:  art,
				Key:    key,
			}

			mgs = append(mgs, msg)
		}

		return p.Stream.Produce(ctx, mgs...)
	}
}
