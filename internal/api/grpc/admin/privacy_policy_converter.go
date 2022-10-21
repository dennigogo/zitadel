package admin

import (
	"github.com/dennigogo/zitadel/internal/domain"
	admin_pb "github.com/dennigogo/zitadel/pkg/grpc/admin"
)

func UpdatePrivacyPolicyToDomain(req *admin_pb.UpdatePrivacyPolicyRequest) *domain.PrivacyPolicy {
	return &domain.PrivacyPolicy{
		TOSLink:     req.TosLink,
		PrivacyLink: req.PrivacyLink,
		HelpLink:    req.HelpLink,
	}
}
