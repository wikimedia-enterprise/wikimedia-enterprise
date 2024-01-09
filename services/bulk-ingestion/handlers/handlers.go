// Package handlers gRPC server and handlers.
package handlers

import (
	"context"
	"wikimedia-enterprise/services/bulk-ingestion/handlers/articles"
	"wikimedia-enterprise/services/bulk-ingestion/handlers/namespaces"
	"wikimedia-enterprise/services/bulk-ingestion/handlers/projects"
	pb "wikimedia-enterprise/services/bulk-ingestion/handlers/protos"

	"go.uber.org/dig"
)

// NewServer constructor initiates and returns an instance of Server.
func NewServer(cont *dig.Container) (*Server, error) {
	srv := new(Server)
	return srv, cont.Invoke(func(pParams projects.Parameters, nParams namespaces.Parameters, aParams articles.Parameters) {
		srv.PParams = pParams
		srv.NParams = nParams
		srv.AParams = aParams
	})
}

// Server is a composition of gRPC bulk server and RPC methods' dependency structs.
type Server struct {
	pb.UnimplementedBulkServer
	PParams projects.Parameters
	NParams namespaces.Parameters
	AParams articles.Parameters
}

// Projects calls handler to get all the projects. Returns the total number of projects & err as a response.
func (srv *Server) Projects(ctx context.Context, req *pb.ProjectsRequest) (*pb.ProjectsResponse, error) {
	return projects.Handler(ctx, &srv.PParams, req)
}

// Namespaces calls handler to get all the namespaces/project. Returns the total number of namespaces & err as a response.
func (srv *Server) Namespaces(ctx context.Context, req *pb.NamespacesRequest) (*pb.NamespacesResponse, error) {
	return namespaces.Handler(ctx, &srv.NParams, req)
}

// Articles calls handler to provide batch article names. Returns total number of produced messages & err as a response.
func (srv *Server) Articles(ctx context.Context, req *pb.ArticlesRequest) (*pb.ArticlesResponse, error) {
	return articles.Handler(ctx, &srv.AParams, req)
}
