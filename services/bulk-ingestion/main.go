package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"wikimedia-enterprise/services/bulk-ingestion/config/env"
	"wikimedia-enterprise/services/bulk-ingestion/handlers"
	pb "wikimedia-enterprise/services/bulk-ingestion/handlers/protos"

	"wikimedia-enterprise/services/bulk-ingestion/packages/container"
	"wikimedia-enterprise/services/bulk-ingestion/packages/creds"

	"google.golang.org/grpc"
)

func startGRPCServer(ctx context.Context, env *env.Environment, hsr *handlers.Server) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", env.ServerPort))
	if err != nil {
		return err
	}

	ops := []grpc.ServerOption{}
	if env.TLSEnabled() {
		crd, err := creds.New(env)
		if err != nil {
			return err
		}

		ops = append(ops, grpc.Creds(crd))
	}

	srv := grpc.NewServer(ops...)
	pb.RegisterBulkServer(srv, hsr)

	ers := make(chan error, 1)
	go func() {
		log.Printf("Now listening on port %s\n", env.ServerPort)
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
	cnt, err := container.New()

	if err != nil {
		log.Panic(err)
	}

	wg := &sync.WaitGroup{}
	err = cnt.Invoke(func(env *env.Environment) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wg.Add(1)
		go func() {
			defer wg.Done()
			hsr, err := handlers.NewServer(cnt)
			if err != nil {
				log.Printf("error creating handler: %s", err.Error())
				cancel()
				return
			}

			err = startGRPCServer(ctx, env, hsr)
			if err != nil {
				if ctx.Err() == nil {
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
	})

	if err != nil {
		log.Panic(err)
	}
}
