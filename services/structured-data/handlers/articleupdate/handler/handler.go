package handler

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
	"wikimedia-enterprise/services/structured-data/config/env"
	"wikimedia-enterprise/services/structured-data/libraries/aggregate"
	"wikimedia-enterprise/services/structured-data/libraries/text"
	"wikimedia-enterprise/services/structured-data/packages/abstract"
	"wikimedia-enterprise/services/structured-data/packages/builder"
	pb "wikimedia-enterprise/services/structured-data/packages/contentintegrity"
	"wikimedia-enterprise/services/structured-data/packages/exponential"
	"wikimedia-enterprise/services/structured-data/packages/protected"
	"wikimedia-enterprise/services/structured-data/packages/tokenizer"
	"wikimedia-enterprise/services/structured-data/submodules/log"
	"wikimedia-enterprise/services/structured-data/submodules/parser"
	pr "wikimedia-enterprise/services/structured-data/submodules/prometheus"
	"wikimedia-enterprise/services/structured-data/submodules/schema"
	"wikimedia-enterprise/services/structured-data/submodules/subscriber"
	"wikimedia-enterprise/services/structured-data/submodules/tracing"

	"github.com/PuerkitoBio/goquery"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/dig"
	tpb "google.golang.org/protobuf/types/known/timestamppb"
)

// Parameters dependency injection parameters for the handler.
type Parameters struct {
	dig.In
	Env        *env.Environment
	Text       text.WordsPairGetter
	Stream     schema.UnmarshalProducer
	Aggregator aggregate.Aggregator
	Parser     parser.API
	Integrity  pb.ContentIntegrityClient
	Protected  *protected.Protected
	Tracer     tracing.Tracer
	Metrics    *pr.Metrics
}

// NewArticleUpdate will produce an article schema message, along with its version.
//
// This handler will fetch from mediawiki API the article's information and build an article schema instance
// with the latest data available.
// It will also call the ORES API to get the scores for the latest version of the article.
//
// Delete messages for obsolete versions of the article will also be produced.
// If the mediawiki API call fails, it produces to the error or dead letter topic based on the fail count.
func NewArticleUpdate(p *Parameters) subscriber.Handler {
	return func(ctx context.Context, msg *kafka.Message) error {
		hcr := tracing.NewHeadersCarrier()
		hcr.FromKafkaHeaders(msg.Headers)
		hcx := hcr.ExtractContext(ctx)

		end, trx := p.Tracer.StartTrace(hcx, "article-update", map[string]string{})
		var err error
		defer func() {
			if err != nil {
				end(err, "error processing article-update event")
			} else {
				end(nil, "article-update event processed")
			}
		}()

		key := new(schema.Key)

		if err := p.Stream.Unmarshal(trx, msg.Key, key); err != nil {
			return err
		}

		pld := new(schema.Article)

		if err := p.Stream.Unmarshal(trx, msg.Value, pld); err != nil {
			return err
		}

		// Introducing gradual delay for failed events processing
		// based on the exponential backoff.
		if pld.Event != nil && pld.Event.FailCount > 0 {
			epn := exponential.GetNth(uint(pld.Event.FailCount+1), uint(p.Env.BackOffBase))
			time.Sleep(time.Duration(epn) * time.Second)
		}

		dtb := pld.IsPartOf.Identifier
		agg := new(aggregate.Aggregation)
		ttl := strings.ReplaceAll(pld.Name, "_", " ")
		err = p.Aggregator.GetAggregation(
			trx, dtb, ttl, agg,
			aggregate.WithPages(2, 1),
			aggregate.WithRevisions(1),
		)

		if err != nil {
			log.Error("wmf api error",
				log.Any("name", pld.Name),
				log.Any("revision", pld.Version.Identifier),
				log.Any("project", dtb),
				log.Any("event_id", pld.Event.Identifier),
				log.Any("error", err),
			)

			if aggregate.IsNonFatalErr(err) {
				return nil
			}

			return err
		}

		if agg.GetPageMissing() {
			log.Error(
				"page is missing",
				log.Any("name", pld.Name),
				log.Any("url", pld.IsPartOf.URL),
				log.Any("event_id", pld.Event.Identifier),
			)

			return nil
		}

		if agg.Page.LastRevID != pld.Version.Identifier {
			err := p.Aggregator.GetAggregation(
				trx, dtb, pld.Name, agg,
				aggregate.WithUser(pld.Name, agg.GetCurrentRevisionUserID()),
			)

			if err != nil {
				return err
			}
		}

		if pld.Version.Identifier == 0 {
			return errors.New("invalid revision identifier")
		}

		err = p.Aggregator.GetAggregation(
			trx, dtb, fmt.Sprintf("%d", agg.Page.LastRevID), agg,
			aggregate.WithRevisionsHTML(1),
		)

		if err != nil {
			log.Error("wmf api parsoid error",
				log.Any("name", pld.Name),
				log.Any("revision", pld.Version.Identifier),
				log.Any("project", dtb),
				log.Any("event_id", pld.Event.Identifier),
				log.Any("error", err),
			)

			if aggregate.IsNonFatalErr(err) {
				return nil
			}

			return err
		}

		if err := agg.GetPageHTMLError(); err != nil {
			log.Error("wmf page html api error",
				log.Any("name", pld.Name),
				log.Any("revision", pld.Version.Identifier),
				log.Any("project", dtb),
				log.Any("event_id", pld.Event.Identifier),
				log.Any("error", err),
			)

			if aggregate.IsNonFatalErr(err) {
				return nil
			}

			return err
		}

		re := regexp.MustCompile(`Special:Redirect/revision/(\d+)`)

		match := re.FindStringSubmatch(agg.PageHTML.Content)
		rid := fmt.Sprintf("%d", agg.GetPageLastRevID())
		if len(match) > 1 && match[1] != rid {
			log.Warn("revision mismatch in Parsoid HTML using revision API - update",
				log.Any("project", dtb),
				log.Any("expected revision id", rid),
				log.Any("found revision id", match[1]),
				log.Any("revision from Page object", agg.PageHTML.Revision),
				log.Any("title", ttl),
			)
		}

		// Call revertrisk, reference need, and reference risk scores for all non-first revisions
		// from all languages wiki projects in namespace 0.
		// checking pld.PreviousVersion != nil for safety reasons to not get nil pointer deref
		if pld.Namespace.Identifier == 0 && strings.HasSuffix(dtb, "wiki") && pld.PreviousVersion != nil && pld.PreviousVersion.Identifier != 0 {
			if !p.Env.ReferenceNeedLanguagesFilter || (len(p.Env.ReferenceNeedLanguages) > 0 && slices.Contains(p.Env.ReferenceNeedLanguages, pld.InLanguage.Identifier)) {
				// If filtering is OFF (false) OR language is in the list (and list is non-empty) → Include ReferenceNeedScore
				log.Debug(fmt.Sprintf("Calling Reference risk and need with PageLastRevID: %d, Language: %s",
					agg.GetPageLastRevID(), pld.InLanguage.Identifier))
				err = p.Aggregator.GetAggregation(
					trx, dtb, ttl, agg,
					aggregate.WithScore(agg.GetPageLastRevID(), pld.InLanguage.Identifier, time.Duration(p.Env.LiftwingTimeoutMs)*time.Millisecond, ttl, "revertrisk", dtb),
					aggregate.WithReferenceNeedScore(agg.GetPageLastRevID(), pld.InLanguage.Identifier, time.Duration(p.Env.LiftwingTimeoutMs)*time.Millisecond, ttl, dtb),
					aggregate.WithReferenceRiskScore(agg.GetPageLastRevID(), pld.InLanguage.Identifier, time.Duration(p.Env.LiftwingTimeoutMs)*time.Millisecond, ttl, dtb),
				)
			} else {
				// If filtering is ON (true) but language is NOT in the list → Exclude ReferenceNeedScore
				log.Debug(fmt.Sprintf("Calling Reference risk with PageLastRevID: %d, Language: %s",
					agg.GetPageLastRevID(), pld.InLanguage.Identifier))
				err = p.Aggregator.GetAggregation(
					trx, dtb, ttl, agg,
					aggregate.WithScore(agg.GetPageLastRevID(), pld.InLanguage.Identifier, time.Duration(p.Env.LiftwingTimeoutMs)*time.Millisecond, ttl, "revertrisk", dtb),
					aggregate.WithReferenceRiskScore(agg.GetPageLastRevID(), pld.InLanguage.Identifier, time.Duration(p.Env.LiftwingTimeoutMs)*time.Millisecond, ttl, dtb),
				)
			}
			if err != nil {
				log.Debug("wmf score api error",
					log.Any("name", ttl),
					log.Any("revision", pld.Version.Identifier),
					log.Any("project", dtb),
					log.Any("language", pld.InLanguage.Identifier),
					log.Any("namespace", pld.Namespace.Identifier),
					log.Any("event_id", pld.Event.Identifier),
					log.Any("error", err),
				)
			}
		}

		var dff *schema.Diff
		crc := agg.GetCurrentRevisionContent()
		if p.Env.EnableDiffs && agg.GetPreviousRevision() != nil {
			tcx, cnc := context.WithTimeout(trx, 200*time.Millisecond)
			defer cnc()

			cws, pws, err := p.Text.GetWordsPair(tcx,
				tokenizer.Tokenize(crc),
				tokenizer.Tokenize(agg.GetPreviousRevisionContent()))

			if err == nil {
				dff = builder.
					NewDiffBuilder().
					Size(crc, agg.GetPreviousRevisionContent()).
					LongestNewRepeatedCharacter(crc, agg.GetPreviousRevisionContent()).
					DictNonDictWords(pws.Dictionary, pws.NonDictionary, cws.Dictionary, cws.NonDictionary).
					InformalWords(pws.Informal, cws.Informal).
					NonSafeWords(pws.NonSafe, cws.NonSafe).
					UpperCaseWords(pws.UpperCase, cws.UpperCase).
					Build()
			}

			if err != nil && !strings.Contains(err.Error(), "language with the given code not found") {
				log.Warn("text processor call failed",
					log.Any("name", pld.Name),
					log.Any("url", pld.IsPartOf.URL),
					log.Any("revision", pld.Version.Identifier),
					log.Any("language", pld.InLanguage.Identifier),
					log.Any("namespace", pld.Namespace.Identifier),
					log.Any("event_id", pld.Event.Identifier),
					log.Any("error", err),
				)
			}
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
					log.Any("name", pld.Name),
					log.Any("url", pld.IsPartOf.URL),
					log.Any("revision", pld.Version.Identifier),
					log.Any("language", pld.InLanguage.Identifier),
					log.Any("namespace", pld.Namespace.Identifier),
					log.Any("event_id", pld.Event.Identifier),
					log.Any("error", err),
				)
			}
		}

		ugs := agg.GetUserGroups()
		uec := agg.GetUserEditCount()
		uds := agg.GetUserRegistration()

		// If data doesn't exist in aggregation,
		// use values from the payload instead.
		if agg.GetUser() == nil && pld.Version.Editor != nil {
			ugs = pld.Version.Editor.Groups
			uec = pld.Version.Editor.EditCount
			uds = pld.Version.Editor.DateStarted
		}

		edr := builder.NewEditorBuilder().
			Identifier(agg.GetCurrentRevisionUserID()).
			Name(agg.GetCurrentRevisionUser()).
			IsAnonymous(agg.GetCurrentRevisionUserID() == 0).
			DateStarted(uds).
			EditCount(uec).
			Groups(ugs).
			IsBot(ugs).
			IsAdmin(ugs).
			IsPatroller(ugs).
			HasAdvancedRights(ugs).
			Build()

		if p.Protected.IsProtectedPage(pld.IsPartOf.Identifier, agg.GetPageTitle()) {
			edr = builder.NewEditorBuilder().Build()

			log.Warn("Protected Page, retracting user", log.Any("prj", pld.IsPartOf.Identifier), log.Any("ttl", pld.Name))
		}

		aci := new(pb.ArticleDataResponse)
		cts := p.Parser.GetCategories(doc.Selection)
		tps := p.Parser.GetTemplates(doc.Selection)

		// add is breaking news to version
		if p.Env.BreakingNewsEnabled {
			aci, err = p.Integrity.GetArticleData(trx, &pb.ArticleDataRequest{
				Identifier:        int64(agg.GetPageID()),
				Project:           dtb,
				VersionIdentifier: int64(agg.GetPageLastRevID()),
				DateModified:      tpb.New(*agg.GetCurrentRevisionTimestamp()),
				Templates:         tps.GetNames(),
				Categories:        cts.GetNames(),
			})

			if err != nil {
				log.Warn(
					"content integrity call failed",
					log.Any("name", pld.Name),
					log.Any("url", pld.IsPartOf.URL),
					log.Any("revision", pld.Version.Identifier),
					log.Any("language", pld.InLanguage.Identifier),
					log.Any("namespace", pld.Namespace.Identifier),
					log.Any("event_id", pld.Event.Identifier),
					log.Any("error", err),
				)
			}
		}

		scs := builder.NewScoresBuilder().
			RevertRisk(agg.GetRevertRiskScore()).
			ReferenceNeed(agg.GetReferenceNeedScore()).
			ReferenceRisk(agg.GetReferenceRiskScore()).
			Build()

		var tgs *schema.MaintenanceTags
		if strings.HasSuffix(pld.IsPartOf.Identifier, "wiki") && pld.Namespace.Identifier == 0 {
			tgs = builder.NewMaintenanceTagsBuilder().
				CitationNeeded(crc).
				PovIdentified(crc).
				ClarificationNeeded(crc).
				Update(crc).
				Build()
		}

		vrs := builder.NewVersionBuilder().
			Identifier(agg.GetPageLastRevID()).
			NumberOfCharacters(crc).
			IsBreakingNews(aci.GetIsBreakingNews()).
			Comment(agg.GetCurrentRevisionComment()).
			Tags(agg.GetCurrentRevisionTags()).
			IsMinorEdit(agg.GetCurrentRevisionMinor()).
			IsFlaggedStable(agg.GetPageFlaggedStableRevID() == agg.GetPageLastRevID()).
			MaintenanceTags(tgs).
			Editor(edr).
			Diff(dff).
			Size(crc).
			Scores(scs).
			Noindex(tps.HasSubstringInNames(p.Env.NoindexTemplatePatterns) || cts.HasSubstringInNames(p.Env.NoindexCategoryPatterns)).
			Build()

		dtn := time.Now().UTC()
		evt := *pld.Event
		evt.SetDatePublished(&dtn)
		evt.SetType(schema.EventTypeUpdate)

		if dur := dtn.Sub(*pld.Event.DateCreated); pld.Event.FailCount == 0 && dur.Milliseconds() > p.Env.LatencyThresholdMS {
			log.Warn("latency threshold exceeded",
				log.Any("name", pld.Name),
				log.Any("url", pld.IsPartOf.URL),
				log.Any("revision", pld.Version.Identifier),
				log.Any("language", pld.InLanguage.Identifier),
				log.Any("namespace", pld.Namespace.Identifier),
				log.Any("duration", dur.Milliseconds()),
			)
		}

		art := builder.NewArticleBuilder(pld.IsPartOf.URL).
			Event(&evt).
			Name(agg.GetPageTitle()).
			Abstract(abs).
			WatchersCount(agg.GetPageWatchers()).
			Identifier(agg.GetPageID()).
			DateCreated(agg.GetFirstRevisionTimestamp()).
			DateModified(agg.GetCurrentRevisionTimestamp()).
			DatePreviouslyModified(agg.GetPreviousRevisionTimestamp()).
			PreviousVersion(agg.GetPreviousRevisionID(), agg.GetPreviousRevisionContent()).
			URL(agg.GetPageCanonicalURL()).
			InLanguage(pld.InLanguage).
			Namespace(pld.Namespace).
			IsPartOf(pld.IsPartOf).
			Version(vrs).
			VersionIdentifier(fmt.Sprintf("/%s/%d", pld.IsPartOf.Identifier, agg.GetPageLastRevID())).
			Body(crc, agg.GetPageHTMLContent()).
			License(schema.NewLicense()).
			MainEntity(agg.GetPagePropsWikiBaseItem()).
			AdditionalEntities(agg.GetPageWbEntityUsage()).
			Protection(agg.GetPageProtection()).
			Categories(cts).
			Templates(tps).
			Redirects(agg.GetPageRedirects()).
			Image(agg.GetPageOriginalImage(), agg.GetPageThumbnailImage()).
			Build()

		tcs, err := p.Env.Topics.
			GetNames(pld.IsPartOf.Identifier, agg.GetPageNs())

		if err != nil {
			return err
		}

		hds := tracing.NewHeadersCarrier().InjectContext(trx)

		mgs := []*schema.Message{
			{
				Config:  schema.ConfigArticle,
				Topic:   p.Env.TopicArticles,
				Value:   art,
				Key:     key,
				Headers: hds,
			},
		}

		for _, tpc := range tcs {
			mgs = append(mgs, &schema.Message{
				Config:  schema.ConfigArticle,
				Topic:   tpc,
				Value:   art,
				Key:     key,
				Headers: hds,
			})
		}

		return p.Stream.Produce(trx, mgs...)
	}

}
