package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
	"wikimedia-enterprise/services/snapshots/config/env"
	"wikimedia-enterprise/services/snapshots/handlers"
	pb "wikimedia-enterprise/services/snapshots/handlers/protos"
	"wikimedia-enterprise/services/snapshots/packages/container"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const timeout = 72 * time.Hour
const check = 12 * time.Hour

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cnt, err := container.New()

	if err != nil {
		log.Panic(err)
	}

	app := func(env *env.Environment, hsr handlers.Server) error {
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
		pb.RegisterSnapshotsServer(srv, hsr)

		return srv.Serve(lis)
	}

	if err := cnt.Invoke(app); err != nil {
		log.Panic(err)
	}
}
