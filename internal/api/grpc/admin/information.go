package admin

import (
	"context"

	"github.com/dennigogo/zitadel/pkg/grpc/admin"
)

func (s *Server) Healthz(context.Context, *admin.HealthzRequest) (*admin.HealthzResponse, error) {
	return &admin.HealthzResponse{}, nil
}
