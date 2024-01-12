// Package pages holds HTTP handler for mimicking the logic of v1 APIs.
package pages

import (
	"context"
	"fmt"
	"strconv"
	"time"
	"wikimedia-enterprise/api/realtime/config/env"
	"wikimedia-enterprise/general/httputil"
	"wikimedia-enterprise/general/ksqldb"
	"wikimedia-enterprise/general/log"
	"wikimedia-enterprise/general/schema"

	sqr "github.com/Masterminds/squirrel"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

type msgID struct {
	Partition int        `json:"partition"`
	Offset    int        `json:"offset"`
	Topic     string     `json:"topic"`
	Dt        *time.Time `json:"dt"`
	Timestamp int        `json:"timestamp"`
}

// Parameters holds a dependency injection needed for the handler.
type Parameters struct {
	dig.In
	KSQLDB ksqldb.PushPuller
	Env    *env.Environment
}

// NewHandler creates new HTTP handler for exposing events by type.
func NewHandler(ctx context.Context, p *Parameters, etp string) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		var snc int

		if qsc := gcx.Query("since"); len(qsc) > 0 {
			var err error
			var dte time.Time
			var tmp int

			if dte, err = time.Parse(time.RFC3339, qsc); err == nil {
				snc = int(dte.UnixNano() / int64(time.Millisecond))
			} else if tmp, err = strconv.Atoi(qsc); err == nil {
				snc = tmp
			} else {
				log.Error(err)
				httputil.BadRequest(gcx, err)
				return
			}
		}

		qer := sqr.
			Select("*").
			Where(fmt.Sprintf("event->type = '%s'", etp)).
			From(p.Env.ArticlesStream)

		if snc > 0 {
			qer = qer.
				Where(fmt.Sprintf("ROWTIME >= %d", snc))
		}

		sql, _, err := qer.ToSql()

		if err != nil {
			log.Error(err)
			httputil.InternalServerError(gcx, err)
			return
		}

		qrq := &ksqldb.QueryRequest{
			SQL: fmt.Sprintf("%s EMIT CHANGES;", sql),
		}

		if snc > 0 {
			qrq.Properties = map[string]string{
				"ksql.streams.auto.offset.reset": "earliest",
			}
		}

		gcx.Header("Content-type", fmt.Sprintf("%s; charset=utf-8", httputil.MIMEEVENTSTEAM))

		cbk := func(hr *ksqldb.HeaderRow, row ksqldb.Row) error {
			art := new(schema.Article)
			err := ksqldb.
				NewDecoder(row, hr.ColumnNames).
				Decode(art)

			if err != nil {
				return err
			}

			mid := new(msgID)

			if art.Event != nil && art.Event.DateCreated != nil {
				mid.Dt = art.Event.DateCreated
				mid.Timestamp = int(art.Event.DateCreated.UTC().UnixNano() / int64(time.Millisecond))
				mid.Topic = "aws.structured-data.articles.v1"
				mid.Partition = 0
				mid.Offset = 0
			}

			_, _ = gcx.Writer.Write([]byte("id: "))
			gcx.JSON(-1, []interface{}{mid})
			_, _ = gcx.Writer.Write([]byte("\nevent: message\n"))
			_, _ = gcx.Writer.Write([]byte("data: "))
			gcx.JSON(-1, art)
			_, _ = gcx.Writer.Write([]byte("\n"))
			_, _ = gcx.Writer.Write([]byte("\n"))
			gcx.Writer.Flush()

			return ctx.Err()
		}

		if err := p.KSQLDB.Push(gcx.Request.Context(), qrq, cbk); err != nil {
			log.Error(err)
			httputil.InternalServerError(gcx, err)
		}
	}
}
