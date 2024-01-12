package main

import (
	"fmt"
	"log"
	"net"
	"wikimedia-enterprise/services/bulk-ingestion/config/env"
	"wikimedia-enterprise/services/bulk-ingestion/handlers"
	pb "wikimedia-enterprise/services/bulk-ingestion/handlers/protos"

	"wikimedia-enterprise/services/bulk-ingestion/packages/container"
	"wikimedia-enterprise/services/bulk-ingestion/packages/creds"

	"google.golang.org/grpc"
)

func main() {
	cnt, err := container.New()

	if err != nil {
		log.Panic(err)
	}

	err = cnt.Invoke(func(env *env.Environment) error {
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
		hsr, err := handlers.NewServer(cnt)

		if err != nil {
			return err
		}

		pb.RegisterBulkServer(srv, hsr)

		return srv.Serve(lis)
	})

	if err != nil {
		log.Panic(err)
	}
}
