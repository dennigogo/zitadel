package management

import (
	"github.com/dennigogo/zitadel/internal/domain"
	mgmt_pb "github.com/dennigogo/zitadel/pkg/grpc/management"
)

func AddPrivacyPolicyToDomain(req *mgmt_pb.AddCustomPrivacyPolicyRequest) *domain.PrivacyPolicy {
	return &domain.PrivacyPolicy{
		TOSLink:     req.TosLink,
		PrivacyLink: req.PrivacyLink,
		HelpLink:    req.HelpLink,
	}
}

func UpdatePrivacyPolicyToDomain(req *mgmt_pb.UpdateCustomPrivacyPolicyRequest) *domain.PrivacyPolicy {
	return &domain.PrivacyPolicy{
		TOSLink:     req.TosLink,
		PrivacyLink: req.PrivacyLink,
		HelpLink:    req.HelpLink,
	}
}
