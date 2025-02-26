package auth

import (
	"context"

	"github.com/dennigogo/zitadel/internal/api/authz"
	"github.com/dennigogo/zitadel/internal/api/grpc/object"
	auth_pb "github.com/dennigogo/zitadel/pkg/grpc/auth"
)

func (s *Server) RemoveMyAvatar(ctx context.Context, req *auth_pb.RemoveMyAvatarRequest) (*auth_pb.RemoveMyAvatarResponse, error) {
	ctxData := authz.GetCtxData(ctx)
	objectDetails, err := s.command.RemoveHumanAvatar(ctx, ctxData.ResourceOwner, ctxData.UserID)
	if err != nil {
		return nil, err
	}
	return &auth_pb.RemoveMyAvatarResponse{
		Details: object.DomainToChangeDetailsPb(objectDetails),
	}, nil
}
