package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"wikimedia-enterprise/services/content-integrity/submodules/log"
	"wikimedia-enterprise/services/content-integrity/submodules/schema"
)

func main() {
	ctx := context.Background()
	mck, err := schema.NewMock()

	if err != nil {
		log.Error(err)
	}

	tps := []*schema.MockTopic{
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
			Topic:  "aws.event-bridge.article-update.v1",
			Config: schema.ConfigArticle,
			Type:   schema.Article{},
		},
	}

	_, fnm, _, _ := runtime.Caller(0)

	for _, tpc := range tps {
		fle, err := os.Open(fmt.Sprintf("%s/%s.ndjson", filepath.Dir(fnm), tpc.Topic))

		if err != nil {
			log.Error(err)
		}

		defer func() {
			if err := fle.Close(); err != nil {
				log.Error(err)
			}
		}()

		tpc.Reader = fle
	}

	if err := mck.Run(ctx, tps...); err != nil {
		log.Error(err)
		os.Exit(1)
	}

	log.Info("Finished successfully")
}
