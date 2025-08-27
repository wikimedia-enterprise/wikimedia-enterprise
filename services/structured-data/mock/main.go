package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"wikimedia-enterprise/services/structured-data/submodules/schema"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	ctx := context.Background()
	mck, err := schema.NewMock()

	if err != nil {
		log.Panic(err)
	}

	tps := []*schema.MockTopic{
		{
			Topic:  "aws.event-bridge.article-update.v1",
			Config: schema.ConfigArticle,
			Type:   schema.Article{},
		},
		{
			Topic:  "aws.event-bridge.article-create.v1",
			Config: schema.ConfigArticle,
			Type:   schema.Article{},
		},
		{
			Topic:  "aws.event-bridge.article-move.v1",
			Config: schema.ConfigArticle,
			Type:   schema.Article{},
		},
		{
			Topic:  "aws.event-bridge.article-delete.v1",
			Config: schema.ConfigArticle,
			Type:   schema.Article{},
		},
		{
			Topic:  "aws.bulk-ingestion.article-names.v1",
			Config: schema.ConfigArticleNames,
			Type:   schema.ArticleNames{},
		},
	}

	_, fnm, _, _ := runtime.Caller(0)

	for _, tpc := range tps {
		fle, err := os.Open(fmt.Sprintf("%s/%s.ndjson", filepath.Dir(fnm), tpc.Topic))

		if err != nil {
			log.Panic(err)
		}

		defer func() {
			if err := fle.Close(); err != nil {
				log.Println(err)
			}
		}()

		tpc.Reader = fle
	}

	if err := mck.Run(ctx, tps...); err != nil {
		log.Panic(err)
	}

	log.Println("Finished successfully")
}
