package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/dig"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"wikimedia-enterprise/api/main/config/env"
	v1 "wikimedia-enterprise/api/main/handlers/v1"
	v2 "wikimedia-enterprise/api/main/handlers/v2"
	"wikimedia-enterprise/api/main/handlers/v2/status"
	"wikimedia-enterprise/api/main/packages/container"
	"wikimedia-enterprise/api/main/packages/shutdown"
	"wikimedia-enterprise/general/httputil"
	"wikimedia-enterprise/general/log"
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
	cnt, err := container.New()

	if err != nil {
		log.Fatal(err, log.Tip("problem creating container for dependency injection"))
	}

	app := func(p Params) error {
		gin.SetMode(p.Env.ServerMode)

		go func() {
			if err := p.Recorder.Serve(); err != nil {
				log.Fatal(err, log.Tip("problem serving metrics"))
			}
		}()

		rtr := gin.New()
		rtr.Use(httputil.Recovery(log.GetZap()))
		rtr.Use(httputil.Logger(log.GetZap()))
		rtr.Use(requestid.New())
		rtr.Use(httputil.CORS())
		rtr.NoRoute(httputil.NoRoute())
		rtr.GET("/v2/status", status.NewHandler())
		rtr.Use(httputil.IPAuth(p.Env.IPAllowList))
		rtr.Use(httputil.Auth(httputil.NewAuthParams(p.Env.CognitoClientID, p.Cache, p.Provider)))
		rtr.Use(httputil.Metrics(p.Recorder))
		rtr.Use(httputil.RBAC(httputil.CasbinRBACAuthorizer(p.Enforcer)))
		rtr.Use(httputil.Limit(p.Env.RateLimitsByGroup))

		if _, err := v1.NewGroup(cnt, rtr); err != nil {
			log.Error(err, log.Tip("problem creating v1 group"))
			return err
		}

		if _, err := v2.NewGroup(cnt, rtr); err != nil {
			log.Error(err, log.Tip("problem creating v2 group"))
			return err
		}

		srv := &http.Server{
			Addr:              fmt.Sprintf(":%s", p.Env.ServerPort),
			Handler:           h2c.NewHandler(rtr, &http2.Server{}),
			ReadHeaderTimeout: 10 * time.Second,
		}

		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal(err, log.Tip("problem error while listen and serve"))
			}
		}()

		return shutdown.
			NewHelper(context.Background()).
			Wait(srv)
	}

	if err := cnt.Invoke(app); err != nil {
		log.Fatal(err, log.Tip("problem invoking dependency injection"))
	}
}
