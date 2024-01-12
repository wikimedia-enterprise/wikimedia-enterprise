// Package articledata offers a grpc endpoint to get content integrity information
// for a given article.
package articledata

import (
	"context"
	"wikimedia-enterprise/general/log"
	pb "wikimedia-enterprise/services/content-integrity/handlers/server/protos"
	"wikimedia-enterprise/services/content-integrity/libraries/integrity"

	"go.uber.org/dig"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handler export handler execution struct with dependency injection.
type Handler struct {
	dig.In
	Integrity    integrity.API
	BreakingNews integrity.BreakingNews
}

// GetArticleData grpc endpoint to get content integrity information for a given article.
func (h *Handler) GetArticleData(ctx context.Context, req *pb.ArticleDataRequest) (*pb.ArticleDataResponse, error) {
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
