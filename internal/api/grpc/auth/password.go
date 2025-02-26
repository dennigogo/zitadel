package auth

import (
	"context"

	"github.com/dennigogo/zitadel/internal/api/grpc/object"

	"github.com/dennigogo/zitadel/internal/api/authz"
	auth_pb "github.com/dennigogo/zitadel/pkg/grpc/auth"
)

func (s *Server) UpdateMyPassword(ctx context.Context, req *auth_pb.UpdateMyPasswordRequest) (*auth_pb.UpdateMyPasswordResponse, error) {
	ctxData := authz.GetCtxData(ctx)
	objectDetails, err := s.command.ChangePassword(ctx, ctxData.ResourceOwner, ctxData.UserID, req.OldPassword, req.NewPassword, "")
	if err != nil {
		return nil, err
	}
	return &auth_pb.UpdateMyPasswordResponse{
		Details: object.DomainToChangeDetailsPb(objectDetails),
	}, nil
}
