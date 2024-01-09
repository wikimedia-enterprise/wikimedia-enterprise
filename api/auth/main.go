package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"wikimedia-enterprise/api/auth/config/env"
	v1 "wikimedia-enterprise/api/auth/handlers/v1"
	"wikimedia-enterprise/api/auth/packages/container"
	"wikimedia-enterprise/api/auth/packages/shutdown"
	"wikimedia-enterprise/general/httputil"
	"wikimedia-enterprise/general/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type Params struct {
	dig.In
	Env      *env.Environment
	Recorder httputil.MetricsRecorderAPI
}

func main() {
	cnt, err := container.New()

	if err != nil {
		log.Fatal(err, log.Tip("problem creating container for dependency injection"))
	}

	app := func(p Params) error {
		gin.SetMode(p.Env.ServerMode)

		go func() {
			if err := p.Recorder.Serve(); err != nil {
				log.Fatal(err)
			}
		}()

		rtr := gin.New()
		rtr.Use(httputil.Recovery(log.GetZap()))
		rtr.Use(httputil.Logger(log.GetZap()))
		rtr.Use(httputil.CORS())
		rtr.Use(httputil.Metrics(p.Recorder))
		rtr.NoRoute(httputil.NoRoute())

		if _, err := v1.NewGroup(cnt, rtr); err != nil {
			log.Error(err, log.Tip("problem creating v1 group"))
			return err
		}

		srv := &http.Server{
			Addr:              fmt.Sprintf(":%s", p.Env.ServerPort),
			Handler:           h2c.NewHandler(rtr, &http2.Server{}),
			ReadHeaderTimeout: time.Second * 60,
		}

		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Error(err, log.Tip("error while listen and serve"))
			}
		}()

		return shutdown.
			NewHelper(context.Background()).
			Wait(srv)
	}

	if err := cnt.Invoke(app); err != nil {
		log.Fatal(err, log.Tip("problem calling invoke for dependency injection"))
	}
}
