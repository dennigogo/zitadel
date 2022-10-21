package admin

import (
	"github.com/dennigogo/zitadel/internal/api/grpc/object"
	policy_grpc "github.com/dennigogo/zitadel/internal/api/grpc/policy"
	"github.com/dennigogo/zitadel/internal/command"
	"github.com/dennigogo/zitadel/internal/query"
	admin_pb "github.com/dennigogo/zitadel/pkg/grpc/admin"
)

func updateLoginPolicyToCommand(p *admin_pb.UpdateLoginPolicyRequest) *command.ChangeLoginPolicy {
	return &command.ChangeLoginPolicy{
		AllowUsernamePassword:      p.AllowUsernamePassword,
		AllowRegister:              p.AllowRegister,
		AllowExternalIDP:           p.AllowExternalIdp,
		ForceMFA:                   p.ForceMfa,
		PasswordlessType:           policy_grpc.PasswordlessTypeToDomain(p.PasswordlessType),
		HidePasswordReset:          p.HidePasswordReset,
		IgnoreUnknownUsernames:     p.IgnoreUnknownUsernames,
		AllowDomainDiscovery:       p.AllowDomainDiscovery,
		DisableLoginWithEmail:      p.DisableLoginWithEmail,
		DisableLoginWithPhone:      p.DisableLoginWithPhone,
		DefaultRedirectURI:         p.DefaultRedirectUri,
		PasswordCheckLifetime:      p.PasswordCheckLifetime.AsDuration(),
		ExternalLoginCheckLifetime: p.ExternalLoginCheckLifetime.AsDuration(),
		MFAInitSkipLifetime:        p.MfaInitSkipLifetime.AsDuration(),
		SecondFactorCheckLifetime:  p.SecondFactorCheckLifetime.AsDuration(),
		MultiFactorCheckLifetime:   p.MultiFactorCheckLifetime.AsDuration(),
	}
}

func ListLoginPolicyIDPsRequestToQuery(req *admin_pb.ListLoginPolicyIDPsRequest) *query.IDPLoginPolicyLinksSearchQuery {
	offset, limit, asc := object.ListQueryToModel(req.Query)
	return &query.IDPLoginPolicyLinksSearchQuery{
		SearchRequest: query.SearchRequest{
			Offset: offset,
			Limit:  limit,
			Asc:    asc,
		},
	}
}
