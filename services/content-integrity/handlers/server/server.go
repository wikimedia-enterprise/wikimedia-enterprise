// Package server contains gRPC handlers for content integrity service.
package server

import (
	"context"
	"wikimedia-enterprise/services/content-integrity/handlers/server/articledata"
	pb "wikimedia-enterprise/services/content-integrity/handlers/server/protos"

	"go.uber.org/dig"
)

// Server is a composition of gRPC diffs server and RPC methods' dependency struct.
type Server struct {
	dig.In
	pb.UnimplementedContentIntegrityServer `optional:"true"`
	ArticleData                            articledata.Handler
}

// GetArticleData grpc endpoint to get content integrity informatio for a given article.
func (s *Server) GetArticleData(ctx context.Context, req *pb.ArticleDataRequest) (*pb.ArticleDataResponse, error) {
	return s.ArticleData.GetArticleData(ctx, req)
}
