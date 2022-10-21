package management

import (
	"context"

	"github.com/zitadel/oidc/v2/pkg/oidc"

	"github.com/dennigogo/zitadel/internal/api/authz"
	"github.com/dennigogo/zitadel/internal/api/http"
	mgmt_pb "github.com/dennigogo/zitadel/pkg/grpc/management"
)

func (s *Server) Healthz(context.Context, *mgmt_pb.HealthzRequest) (*mgmt_pb.HealthzResponse, error) {
	return &mgmt_pb.HealthzResponse{}, nil
}

func (s *Server) GetOIDCInformation(ctx context.Context, _ *mgmt_pb.GetOIDCInformationRequest) (*mgmt_pb.GetOIDCInformationResponse, error) {
	issuer := http.BuildOrigin(authz.GetInstance(ctx).RequestedHost(), s.externalSecure)
	return &mgmt_pb.GetOIDCInformationResponse{
		Issuer:            issuer,
		DiscoveryEndpoint: issuer + oidc.DiscoveryEndpoint,
	}, nil
}
