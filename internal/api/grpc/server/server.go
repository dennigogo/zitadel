package server

import (
	"crypto/tls"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/dennigogo/zitadel/internal/api/authz"
	grpc_api "github.com/dennigogo/zitadel/internal/api/grpc"
	"github.com/dennigogo/zitadel/internal/api/grpc/server/middleware"
	"github.com/dennigogo/zitadel/internal/query"
	"github.com/dennigogo/zitadel/internal/telemetry/metrics"
	system_pb "github.com/dennigogo/zitadel/pkg/grpc/system"
)

type Server interface {
	Gateway
	RegisterServer(*grpc.Server)
	AppName() string
	MethodPrefix() string
	AuthMethods() authz.MethodMapping
}

func CreateServer(verifier *authz.TokenVerifier, authConfig authz.Config, queries *query.Queries, hostHeaderName string, tlsConfig *tls.Config) *grpc.Server {
	metricTypes := []metrics.MetricType{metrics.MetricTypeTotalCount, metrics.MetricTypeRequestCount, metrics.MetricTypeStatusCode}
	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				middleware.DefaultTracingServer(),
				middleware.MetricsHandler(metricTypes, grpc_api.Probes...),
				middleware.NoCacheInterceptor(),
				middleware.ErrorHandler(),
				middleware.InstanceInterceptor(queries, hostHeaderName, system_pb.SystemService_MethodPrefix),
				middleware.AuthorizationInterceptor(verifier, authConfig),
				middleware.TranslationHandler(),
				middleware.ValidationHandler(),
				middleware.ServiceHandler(),
			),
		),
	}
	if tlsConfig != nil {
		serverOptions = append(serverOptions, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}
	return grpc.NewServer(serverOptions...)
}
