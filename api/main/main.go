package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/casbin/casbin/v2"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/dig"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"wikimedia-enterprise/api/main/config/env"
	v2 "wikimedia-enterprise/api/main/handlers/v2"
	"wikimedia-enterprise/api/main/handlers/v2/status"
	"wikimedia-enterprise/api/main/packages/container"
	"wikimedia-enterprise/api/main/packages/shutdown"
	"wikimedia-enterprise/api/main/submodules/api-openapi-spec/ui"
	"wikimedia-enterprise/api/main/submodules/httputil"
	"wikimedia-enterprise/api/main/submodules/log"

	health "github.com/wikimedia-enterprise/health-checker/health"
)

//go:embed submodules/api-openapi-spec/ui/*
var ufs embed.FS

//go:embed submodules/api-openapi-spec/main.yaml
var dcs []byte

type Params struct {
	dig.In
	Env      *env.Environment
	Cache    redis.Cmdable
	Provider httputil.AuthProvider
	Recorder httputil.MetricsRecorderAPI
	Enforcer *casbin.Enforcer
	S3Client s3iface.S3API
}

func syncLogger() {
	err := log.Sync()
	if err != nil && !errors.Is(err, syscall.EINVAL) {
		// err == EINVAL when logging to stdout/stderr, which are not buffered.
		log.Fatal(err, log.Tip("problem syncing log"))
	}
}

func setUpHealthChecker(env *env.Environment, s3Client s3iface.S3API, cache redis.Cmdable) http.Handler {
	s3Checker, err := health.NewS3Checker(health.S3CheckerConfig{
		BucketName: env.AWSBucket,
		Region:     env.AWSRegion,
		S3Client:   s3Client,
		Name:       "s3-check",
	})
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to create S3 checker: %v\n", err))
	}

	redisChecker, err := health.NewRedisChecker(cache, health.RedisCheckerConfig{
		Name: "redis-check",
	})
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to create Redis checker: %v\n", err))
	}

	h, err := health.SetupHealthChecks("Main-API", "1.0.0", true, nil, env.HealthCheckMaxRetries, s3Checker, redisChecker)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to set up health checks: %v\n", err))
	}
	return health.Handler(h)
}

func main() {
	cnt, err := container.New()

	if err != nil {
		log.Fatal(err, log.Tip("problem creating container for dependency injection"))
	}

	defer syncLogger()

	app := func(p Params) error {
		gin.SetMode(p.Env.ServerMode)

		go func() {
			if err := p.Recorder.Serve(); err != nil {
				log.Fatal(err, log.Tip("problem serving metrics"))
			}
		}()

		hnd := setUpHealthChecker(p.Env, p.S3Client, p.Cache)
		rtr := gin.New()
		rtr.GET("/healthz", func(ctx *gin.Context) {
			hnd.ServeHTTP(ctx.Writer, ctx.Request)
			status := ctx.Writer.Status()
			if status != http.StatusOK {
				log.Error("health check failed", log.Any("status", status))
			} else {
				log.Debug("health check", log.Any("status", status))
			}
		})
		if p.Env.EnablePProf {
			rtr.GET("/debug/pprof/", gin.WrapF(pprof.Index))
			rtr.GET("/debug/pprof/cmdline", gin.WrapF(pprof.Cmdline))
			rtr.GET("/debug/pprof/profile", gin.WrapF(pprof.Profile))
			rtr.GET("/debug/pprof/symbol", gin.WrapF(pprof.Symbol))
			rtr.GET("/debug/pprof/trace", gin.WrapF(pprof.Trace))
			rtr.GET("/debug/pprof/goroutine", gin.WrapH(pprof.Handler("goroutine")))
			rtr.GET("/debug/pprof/heap", gin.WrapH(pprof.Handler("heap")))
			rtr.GET("/debug/pprof/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
			rtr.GET("/debug/pprof/block", gin.WrapH(pprof.Handler("block")))
			rtr.GET("/debug/pprof/mutex", gin.WrapH(pprof.Handler("mutex")))
		}
		rtr.Use(httputil.RestoreBodyMiddleware())
		rtr.Use(httputil.Recovery(log.GetZap()))
		rtr.Use(httputil.Logger(log.GetZap()))
		rtr.Use(requestid.New())
		rtr.Use(httputil.CORS())
		rtr.NoRoute(httputil.NoRoute())
		rtr.GET("/v2/status", status.NewHandler())
		rtr.GET("/spec/spec.yaml", ui.File(dcs))
		rtr.StaticFS("/docs/", ui.FS(ufs))

		if p.Env.TestOnlySkipAuth {
			rtr.Use(httputil.FakeAuth(p.Env.TestOnlyGroup))
		}

		rtr.Use(httputil.IPAuth(*p.Env.IPAllowList))
		rtr.Use(httputil.Auth(httputil.NewAuthParams(p.Env.CognitoClientID, p.Cache, p.Provider)))
		rtr.Use(httputil.Metrics(p.Recorder))
		rtr.Use(httputil.RBAC(httputil.CasbinRBACAuthorizer(p.Enforcer)))
		rtr.Use(httputil.Limit(p.Env.RateLimitsByGroup))

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
