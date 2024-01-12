package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"wikimedia-enterprise/api/realtime/config/env"
	v1 "wikimedia-enterprise/api/realtime/handlers/v1"
	v2 "wikimedia-enterprise/api/realtime/handlers/v2"
	"wikimedia-enterprise/api/realtime/handlers/v2/status"
	"wikimedia-enterprise/api/realtime/packages/container"
	"wikimedia-enterprise/api/realtime/packages/shutdown"
	"wikimedia-enterprise/general/httputil"
	"wikimedia-enterprise/general/log"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/dig"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type Params struct {
	dig.In
	Env      *env.Environment
	Cache    redis.Cmdable
	Provider httputil.AuthProvider
	Recorder httputil.MetricsRecorderAPI
	Enforcer *casbin.Enforcer
}

func main() {
	sht := shutdown.NewHelper(context.Background())
	ctx := sht.Ctx()
	ctr, err := container.New()

	if err != nil {
		log.Fatal(err)
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
		rtr.NoRoute(httputil.NoRoute())
		rtr.GET("/v2/status", status.NewHandler())
		rtr.Use(httputil.IPAuth(p.Env.IPAllowList))
		rtr.Use(httputil.Auth(httputil.NewAuthParams(p.Env.CognitoClientID, p.Cache, p.Provider)))
		rtr.Use(httputil.Metrics(p.Recorder))
		rtr.Use(httputil.RBAC(httputil.CasbinRBACAuthorizer(p.Enforcer)))

		if _, err := v2.NewGroup(ctx, ctr, rtr); err != nil {
			return err
		}

		if _, err := v1.NewGroup(ctx, ctr, rtr); err != nil {
			return err
		}

		srv := &http.Server{
			Addr:              fmt.Sprintf(":%s", p.Env.ServerPort),
			Handler:           h2c.NewHandler(rtr, &http2.Server{}),
			ReadHeaderTimeout: time.Second * 60,
		}

		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}()

		return sht.Wait(srv)
	}

	if err := ctr.Invoke(app); err != nil {
		log.Fatal(err)
	}
}
