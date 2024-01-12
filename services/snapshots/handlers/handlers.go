package handlers

import (
	"context"
	"wikimedia-enterprise/services/snapshots/handlers/aggregate"
	"wikimedia-enterprise/services/snapshots/handlers/aggregatecopy"
	"wikimedia-enterprise/services/snapshots/handlers/copy"
	"wikimedia-enterprise/services/snapshots/handlers/export"
	pb "wikimedia-enterprise/services/snapshots/handlers/protos"

	"go.uber.org/dig"
)

// Server is a composition of gRPC diffs server and RPC methods' dependency struct.
type Server struct {
	pb.UnimplementedSnapshotsServer `optional:"true"`
	dig.In
	Exporter        export.Handler
	Copier          copy.Handler
	Aggregator      aggregate.Handler
	AggregateCopier aggregatecopy.Handler
}

// Export triggers export handler.
func (s Server) Export(ctx context.Context, req *pb.ExportRequest) (*pb.ExportResponse, error) {
	return s.Exporter.Export(ctx, req)
}

// Copy triggers copy handler.
func (s Server) Copy(ctx context.Context, req *pb.CopyRequest) (*pb.CopyResponse, error) {
	return s.Copier.Copy(ctx, req)
}

// Aggregate triggers aggregate handler.
func (s Server) Aggregate(ctx context.Context, req *pb.AggregateRequest) (*pb.AggregateResponse, error) {
	return s.Aggregator.Aggregate(ctx, req)
}

// AggregateCopy triggers aggregatecopy handler.
func (s Server) AggregateCopy(ctx context.Context, req *pb.AggregateCopyRequest) (*pb.AggregateCopyResponse, error) {
	return s.AggregateCopier.AggregateCopy(ctx, req)
}
