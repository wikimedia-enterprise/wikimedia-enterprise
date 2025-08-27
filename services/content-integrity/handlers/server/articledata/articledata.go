// Package articledata offers a grpc endpoint to get content integrity information
// for a given article.
package articledata

import (
	"context"
	"fmt"
	pb "wikimedia-enterprise/services/content-integrity/handlers/server/protos"
	"wikimedia-enterprise/services/content-integrity/libraries/integrity"
	"wikimedia-enterprise/services/content-integrity/submodules/log"

	"go.uber.org/dig"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handler export handler execution struct with dependency injection.
type Handler struct {
	dig.In
	Integrity    integrity.API
	BreakingNews integrity.BreakingNews
}

func validateRequest(req *pb.ArticleDataRequest) error {
	if req.Project == "" {
		return status.Errorf(codes.InvalidArgument, "invalid request: %s", req)
	}

	if req.VersionIdentifier == 0 || req.Identifier == 0 {
		// TODO: check if this ever happens. If not, start rejecting these requests as invalid.
		log.Warn(fmt.Sprintf("invalid version identifier or last version identifier, proceeding: %s\n", req))
	}

	return nil
}

// GetArticleData grpc endpoint to get content integrity information for a given article.
func (h *Handler) GetArticleData(ctx context.Context, req *pb.ArticleDataRequest) (*pb.ArticleDataResponse, error) {
	err := validateRequest(req)
	if err != nil {
		return nil, err
	}

	pms := &integrity.ArticleParams{
		Project:           req.GetProject(),
		Identifier:        int(req.GetIdentifier()),
		VersionIdentifier: int(req.GetVersionIdentifier()),
		DateModified:      req.GetDateModified().AsTime(),
		Templates:         req.GetTemplates(),
		Categories:        req.GetCategories(),
	}

	art, err := h.Integrity.GetArticle(ctx, pms, &h.BreakingNews)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	res := &pb.ArticleDataResponse{
		Identifier:         req.GetIdentifier(),
		Project:            req.GetProject(),
		VersionIdentifier:  int64(art.GetVersionIdentifier()),
		IsBreakingNews:     art.GetIsBreakingNews(),
		EditsCount:         int32(art.GetEditsCount()),
		UniqueEditorsCount: int32(art.GetUniqueEditorsCount()),
	}

	if dct := art.GetDateCreated(); dct != nil {
		res.DateCreated = timestamppb.New(*dct)
	}

	if dnm := art.GetDateNamespaceMoved(); dnm != nil {
		res.DateNamespaceMoved = timestamppb.New(*dnm)
	}

	return res, nil

}
