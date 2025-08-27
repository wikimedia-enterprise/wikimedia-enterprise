// Package main.
package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"wikimedia-enterprise/services/content-integrity/config/env"
	server "wikimedia-enterprise/services/content-integrity/handlers/server"
	"wikimedia-enterprise/services/content-integrity/packages/container"

	pb "wikimedia-enterprise/services/content-integrity/handlers/server/protos"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"wikimedia-enterprise/services/content-integrity/submodules/subscriber"

	"wikimedia-enterprise/services/content-integrity/handlers/listener/aggregate"

	"wikimedia-enterprise/services/content-integrity/submodules/log"
)

const timeout = 72 * time.Hour
const check = 12 * time.Hour

func main() {
	cnt, err := container.New()

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		lsn := func(env *env.Environment, agp aggregate.Parameters, sub *subscriber.Subscriber, hsr server.Server) error {
			ctx := context.Background()
			hdl := aggregate.NewAggregate(&agp)

			scf := &subscriber.Config{
				NumberOfWorkers: env.NumberOfWorkers,
				Topics:          []string{env.TopicArticleCreate, env.TopicArticleUpdate, env.TopicArticleMove},
				Events:          make(chan *subscriber.Event, env.EventChannelSize),
			}

			go func() {
				for evt := range scf.Events {
					if evt.Error != nil {
						log.Error(evt.Error)
					}
				}
			}()

			return sub.Subscribe(
				ctx,
				hdl,
				scf,
			)
		}

		if err := cnt.Invoke(lsn); err != nil {
			log.Fatal(err)
		}
	}()

	app := func(env *env.Environment, hsr server.Server) error {
		go func() {
			reg := prometheus.NewRegistry()
			reg.MustRegister(
				collectors.NewGoCollector(),
				collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
			)

			hdl := promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg})
			http.Handle("/metrics", hdl)
			_ = http.ListenAndServe(fmt.Sprintf(":%v", env.PrometheusPort), nil) // #nosec G114
		}()

		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", env.ServerPort))

		if err != nil {
			log.Error(err)
			return err
		}

		ops := []grpc.ServerOption{grpc.ConnectionTimeout(timeout),
			grpc.KeepaliveParams(keepalive.ServerParameters{
				MaxConnectionIdle:     timeout,
				MaxConnectionAge:      timeout,
				MaxConnectionAgeGrace: timeout,
				Time:                  check,
				Timeout:               check,
			})}

		srv := grpc.NewServer(ops...)
		pb.RegisterContentIntegrityServer(srv, &hsr)

		log.Info(fmt.Sprintf("Listening on :%s", env.ServerPort))
		return srv.Serve(lis)
	}

	if err := cnt.Invoke(app); err != nil {
		log.Fatal(err)
	}
}
