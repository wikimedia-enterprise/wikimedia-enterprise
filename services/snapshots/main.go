package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"wikimedia-enterprise/services/snapshots/config/env"
	"wikimedia-enterprise/services/snapshots/handlers"
	pb "wikimedia-enterprise/services/snapshots/handlers/protos"
	"wikimedia-enterprise/services/snapshots/packages/container"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const timeout = 72 * time.Hour
const check = 12 * time.Hour

func StartGRPCServer(ctx context.Context, env *env.Environment, hsr handlers.Server) error {
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

	ers := make(chan error, 1)
	go func() {
		ers <- srv.Serve(lis)
	}()

	select {
	case err := <-ers:
		// Return to cancel the context.
		return fmt.Errorf("error in Serve: %w", err)
	case <-ctx.Done():
		srv.GracefulStop()
		// lis is closed by GracefulStop() too, no need to do it ourselves.
	}

	return nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cnt, err := container.New()
	if err != nil {
		log.Panic(err)
	}

	wg := &sync.WaitGroup{}
	app := func(env *env.Environment, hsr handlers.Server) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			reg := prometheus.DefaultRegisterer
			hdl := promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{Registry: reg})
			http.Handle("/metrics", hdl)
			_ = http.ListenAndServe(fmt.Sprintf(":%v", env.PrometheusPort), nil) // #nosec G114
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := StartGRPCServer(ctx, env, hsr)
			if err != nil {
				if ctx.Err() == nil {
					// StartGRPCServer is critical, if it fails early we need to stop the service.
					cancel()
				}
				log.Printf("error in StartGRPCServer: %v", err)
			}
		}()

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigs)

		select {
		case <-sigs:
			log.Println("signal received, shutting down...")
			cancel()
		case <-ctx.Done():
		}

		// Wait for cleanup.
		wg.Wait()
		return nil
	}

	if err := cnt.Invoke(app); err != nil {
		log.Panic(err)
	}
}
