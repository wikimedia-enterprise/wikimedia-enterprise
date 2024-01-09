package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"wikimedia-enterprise/general/schema"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	pts := flag.String("p", "", "comma-separated list of partitions, e.g., 0,1,2,3,4")
	flag.Parse()
	ctx := context.Background()
	mck, err := schema.NewMock()

	if err != nil {
		log.Panic(err)
	}

	tps := []*schema.MockTopic{
		{
			Topic:  "aws.structured-data.articles.v1",
			Config: schema.ConfigArticle,
			Type:   schema.Article{},
		},
	}

	if *pts != "" {
		pss := strings.Split(*pts, ",")
		psi := make([]int32, 0, len(pss))

		for _, ptn := range pss {
			pti, err := strconv.ParseInt(ptn, 10, 32)
			if err != nil {
				log.Println(err)
			}
			psi = append(psi, int32(pti))
		}

		for _, tpc := range tps {
			tpc.Partitions = psi
		}
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
