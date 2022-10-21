package auth

import (
	"context"
	"time"

	"google.golang.org/grpc"

	"github.com/dennigogo/zitadel/internal/api/assets"
	"github.com/dennigogo/zitadel/internal/api/authz"
	"github.com/dennigogo/zitadel/internal/api/grpc/server"
	"github.com/dennigogo/zitadel/internal/auth/repository"
	"github.com/dennigogo/zitadel/internal/auth/repository/eventsourcing"
	"github.com/dennigogo/zitadel/internal/command"
	"github.com/dennigogo/zitadel/internal/config/systemdefaults"
	"github.com/dennigogo/zitadel/internal/crypto"
	"github.com/dennigogo/zitadel/internal/query"
	"github.com/dennigogo/zitadel/pkg/grpc/auth"
)

var _ auth.AuthServiceServer = (*Server)(nil)

const (
	authName = "Auth-API"
)

type Server struct {
	auth.UnimplementedAuthServiceServer
	command           *command.Commands
	query             *query.Queries
	repo              repository.Repository
	defaults          systemdefaults.SystemDefaults
	assetsAPIDomain   func(context.Context) string
	userCodeAlg       crypto.EncryptionAlgorithm
	externalSecure    bool
	auditLogRetention time.Duration
}

type Config struct {
	Repository eventsourcing.Config
}

func CreateServer(command *command.Commands,
	query *query.Queries,
	authRepo repository.Repository,
	defaults systemdefaults.SystemDefaults,
	userCodeAlg crypto.EncryptionAlgorithm,
	externalSecure bool,
	auditLogRetention time.Duration,
) *Server {
	return &Server{
		command:           command,
		query:             query,
		repo:              authRepo,
		defaults:          defaults,
		assetsAPIDomain:   assets.AssetAPI(externalSecure),
		userCodeAlg:       userCodeAlg,
		externalSecure:    externalSecure,
		auditLogRetention: auditLogRetention,
	}
}

func (s *Server) RegisterServer(grpcServer *grpc.Server) {
	auth.RegisterAuthServiceServer(grpcServer, s)
}

func (s *Server) AppName() string {
	return authName
}

func (s *Server) MethodPrefix() string {
	return auth.AuthService_MethodPrefix
}

func (s *Server) AuthMethods() authz.MethodMapping {
	return auth.AuthService_AuthMethods
}

func (s *Server) RegisterGateway() server.GatewayFunc {
	return auth.RegisterAuthServiceHandlerFromEndpoint
}

func (s *Server) GatewayPathPrefix() string {
	return "/auth/v1"
}
