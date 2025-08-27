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
	"wikimedia-enterprise/api/auth/submodules/httputil"
	"wikimedia-enterprise/api/auth/submodules/log"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/gin-gonic/gin"
	health "github.com/wikimedia-enterprise/health-checker/health"
	"go.uber.org/dig"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type Params struct {
	dig.In
	Env        *env.Environment
	Recorder   httputil.MetricsRecorderAPI
	CognitoAPI cognitoidentityprovideriface.CognitoIdentityProviderAPI
}

func setUpHealthChecker(env *env.Environment, cognitoAPI cognitoidentityprovideriface.CognitoIdentityProviderAPI) http.Handler {
	c := &health.CognitoChecker{
		CheckerName:      "cognito-check",
		UserPoolId:       env.CognitoUserPoolID,
		CognitoClientId:  env.CognitoClientID,
		CognitoSecret:    env.CognitoSecret,
		TestUserName:     env.HealthChecksTestUserName,
		TestUserPassword: env.HealthChecksTestPassword,
		CognitoAPI:       cognitoAPI,
	}
	permChecker := &health.CognitoPermissionChecker{
		CheckerName: "cognito-permission-check",
		UserPoolId:  env.CognitoUserPoolID,
		CognitoAPI:  cognitoAPI,
	}
	h, err := health.SetupHealthChecks("Auth", "1.0.0", true, nil, 2, c, permChecker)
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

	app := func(p Params) error {
		gin.SetMode(p.Env.ServerMode)

		go func() {
			if err := p.Recorder.Serve(); err != nil {
				log.Fatal(err)
			}
		}()

		hnd := setUpHealthChecker(p.Env, p.CognitoAPI)
		rtr := gin.New()
		rtr.GET("/healthz", func(ctx *gin.Context) {
			hnd.ServeHTTP(ctx.Writer, ctx.Request)
			log.Debug("health check", log.Any("status", ctx.Writer.Status()))
		})

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
