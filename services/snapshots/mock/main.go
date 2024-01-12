package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"wikimedia-enterprise/general/schema"
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
			Topic:  "aws.structured-data.enwiki-articles-compacted.v1",
			Config: schema.ConfigArticle,
			Type:   schema.Article{},
		},
		{
			Topic:  "aws.structured-data.eswiki-articles-compacted.v1",
			Config: schema.ConfigArticle,
			Type:   schema.Article{},
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
}
