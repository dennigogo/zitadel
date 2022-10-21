package policy

import (
	"github.com/dennigogo/zitadel/internal/api/grpc/object"
	"github.com/dennigogo/zitadel/internal/query"
	policy_pb "github.com/dennigogo/zitadel/pkg/grpc/policy"
)

func ModelPasswordAgePolicyToPb(policy *query.PasswordAgePolicy) *policy_pb.PasswordAgePolicy {
	return &policy_pb.PasswordAgePolicy{
		IsDefault:      policy.IsDefault,
		MaxAgeDays:     policy.MaxAgeDays,
		ExpireWarnDays: policy.ExpireWarnDays,
		Details: object.ToViewDetailsPb(
			policy.Sequence,
			policy.CreationDate,
			policy.ChangeDate,
			policy.ResourceOwner,
		),
	}
}
