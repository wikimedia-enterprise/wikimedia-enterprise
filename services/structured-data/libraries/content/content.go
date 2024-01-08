// Package content dependency injection provider for content integrity grpc service.
package content

import (
	"wikimedia-enterprise/services/structured-data/config/env"
	pb "wikimedia-enterprise/services/structured-data/packages/contentintegrity"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// New creates new instance of ContentIntegrityClient for dependency injection.
func New(env *env.Environment) (pb.ContentIntegrityClient, error) {
	cnn, err := grpc.Dial(env.ContentIntegrityURL, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	return pb.NewContentIntegrityClient(cnn), nil
}
