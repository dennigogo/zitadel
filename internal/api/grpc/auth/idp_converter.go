package auth

import (
	"context"

	"github.com/dennigogo/zitadel/internal/api/authz"
	"github.com/dennigogo/zitadel/internal/api/grpc/object"
	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/query"
	auth_pb "github.com/dennigogo/zitadel/pkg/grpc/auth"
)

func ListMyLinkedIDPsRequestToQuery(ctx context.Context, req *auth_pb.ListMyLinkedIDPsRequest) (*query.IDPUserLinksSearchQuery, error) {
	offset, limit, asc := object.ListQueryToModel(req.Query)
	q, err := query.NewIDPUserLinksUserIDSearchQuery(authz.GetCtxData(ctx).UserID)
	if err != nil {
		return nil, err
	}
	return &query.IDPUserLinksSearchQuery{
		SearchRequest: query.SearchRequest{
			Offset: offset,
			Limit:  limit,
			Asc:    asc,
		},
		Queries: []query.SearchQuery{q},
	}, nil
}

func RemoveMyLinkedIDPRequestToDomain(ctx context.Context, req *auth_pb.RemoveMyLinkedIDPRequest) *domain.UserIDPLink {
	return &domain.UserIDPLink{
		ObjectRoot:     ctxToObjectRoot(ctx),
		IDPConfigID:    req.IdpId,
		ExternalUserID: req.LinkedUserId,
	}
}
