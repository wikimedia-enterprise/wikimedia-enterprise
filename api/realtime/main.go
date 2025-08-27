package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"time"
	"wikimedia-enterprise/api/realtime/config/env"
	v2 "wikimedia-enterprise/api/realtime/handlers/v2"
	"wikimedia-enterprise/api/realtime/handlers/v2/status"
	"wikimedia-enterprise/api/realtime/packages/container"
	"wikimedia-enterprise/api/realtime/packages/shutdown"
	"wikimedia-enterprise/api/realtime/submodules/api-openapi-spec/ui"
	"wikimedia-enterprise/api/realtime/submodules/httputil"
	"wikimedia-enterprise/api/realtime/submodules/log"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	healthchecks "github.com/wikimedia-enterprise/health-checker/health"
	"go.uber.org/dig"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

//go:embed submodules/api-openapi-spec/ui/*
var ufs embed.FS

//go:embed submodules/api-openapi-spec/realtime.yaml
var dcs []byte

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

		checkers := []healthchecks.HealthChecker{}

		redisChecker, err := healthchecks.NewRedisChecker(
			p.Cache,
			healthchecks.RedisCheckerConfig{
				Timeout: time.Duration(p.Env.RedisHealthCheckTimeoutSeconds) * time.Second,
				Name:    "redis-check",
			},
		)

		if err != nil {
			log.Fatal(err)
		}

		checkers = append(checkers, redisChecker)

		if p.Env.UseKsqldb {
			ksqldbChecker, err := healthchecks.NewKSQLDBAsyncChecker(
				p.Env.KSQLURL,
				time.Duration(p.Env.KSQLHealthCheckTimeoutSeconds)*time.Second,
				time.Duration(p.Env.KSQLHealthCheckIntervalSeconds)*time.Second,
				p.Env.ArticlesStream,
			)

			if err != nil {
				log.Fatal(err)
			}

			checkers = append(checkers, ksqldbChecker)
		}

		hlt, err := healthchecks.SetupHealthChecks("Realtime-API", "1.0.0", true, checkers...)
		hh := hlt.Handler()

		if err != nil {
			log.Fatal(err)
		}

		rtr := gin.New()

		rtr.GET("/healthz", func(ctx *gin.Context) {
			hh.ServeHTTP(ctx.Writer, ctx.Request)
			status := ctx.Writer.Status()
			if status != http.StatusOK {
				log.Error("health check failed", log.Any("status", status))
			} else {
				log.Debug("health check", log.Any("status", status))
			}
		})

		rtr.Use(httputil.Recovery(log.GetZap()))
		rtr.Use(httputil.Logger(log.GetZap()))
		rtr.Use(httputil.CORS())
		rtr.NoRoute(httputil.NoRoute())
		rtr.GET("/v2/status", status.NewHandler())
		rtr.GET("/spec/spec.yaml", ui.File(dcs))
		rtr.StaticFS("/docs/", ui.FS(ufs))
		rtr.Use(httputil.IPAuth(p.Env.IPAllowList))
		rtr.Use(httputil.Auth(httputil.NewAuthParams(p.Env.CognitoClientID, p.Cache, p.Provider)))
		rtr.Use(httputil.Metrics(p.Recorder))
		rtr.Use(httputil.RBAC(httputil.CasbinRBACAuthorizer(p.Enforcer)))

		if _, err := v2.NewGroup(ctx, ctr, rtr); err != nil {
			return err
		}

		srv := &http.Server{
			Addr:              fmt.Sprintf(":%s", p.Env.ServerPort),
			Handler:           h2c.NewHandler(rtr, &http2.Server{}),
			ReadHeaderTimeout: time.Minute * 10,
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
