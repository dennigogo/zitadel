package auth

import (
	"context"

	"github.com/dennigogo/zitadel/internal/api/authz"
	object_grpc "github.com/dennigogo/zitadel/internal/api/grpc/object"
	user_grpc "github.com/dennigogo/zitadel/internal/api/grpc/user"
	auth_pb "github.com/dennigogo/zitadel/pkg/grpc/auth"
)

func (s *Server) GetMyProfile(ctx context.Context, req *auth_pb.GetMyProfileRequest) (*auth_pb.GetMyProfileResponse, error) {
	profile, err := s.query.GetHumanProfile(ctx, authz.GetCtxData(ctx).UserID)
	if err != nil {
		return nil, err
	}
	return &auth_pb.GetMyProfileResponse{
		Profile: user_grpc.ProfileToPb(profile, s.assetsAPIDomain(ctx)),
		Details: object_grpc.ToViewDetailsPb(
			profile.Sequence,
			profile.CreationDate,
			profile.ChangeDate,
			profile.ResourceOwner,
		),
	}, nil
}

func (s *Server) UpdateMyProfile(ctx context.Context, req *auth_pb.UpdateMyProfileRequest) (*auth_pb.UpdateMyProfileResponse, error) {
	profile, err := s.command.ChangeHumanProfile(ctx, UpdateProfileToDomain(ctx, req))
	if err != nil {
		return nil, err
	}
	return &auth_pb.UpdateMyProfileResponse{
		Details: object_grpc.ChangeToDetailsPb(
			profile.Sequence,
			profile.ChangeDate,
			profile.ResourceOwner,
		),
	}, nil
}
