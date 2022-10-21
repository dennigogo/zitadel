package policy

import (
	"github.com/dennigogo/zitadel/internal/api/grpc/object"
	"github.com/dennigogo/zitadel/internal/query"
	policy_pb "github.com/dennigogo/zitadel/pkg/grpc/policy"
)

func ModelPrivacyPolicyToPb(policy *query.PrivacyPolicy) *policy_pb.PrivacyPolicy {
	return &policy_pb.PrivacyPolicy{
		IsDefault:   policy.IsDefault,
		TosLink:     policy.TOSLink,
		PrivacyLink: policy.PrivacyLink,
		HelpLink:    policy.HelpLink,
		Details: object.ToViewDetailsPb(
			policy.Sequence,
			policy.CreationDate,
			policy.ChangeDate,
			policy.ResourceOwner,
		),
	}
}
