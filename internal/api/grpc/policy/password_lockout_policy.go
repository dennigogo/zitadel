package policy

import (
	"github.com/dennigogo/zitadel/internal/api/grpc/object"
	"github.com/dennigogo/zitadel/internal/query"
	policy_pb "github.com/dennigogo/zitadel/pkg/grpc/policy"
)

func ModelLockoutPolicyToPb(policy *query.LockoutPolicy) *policy_pb.LockoutPolicy {
	return &policy_pb.LockoutPolicy{
		IsDefault:           policy.IsDefault,
		MaxPasswordAttempts: policy.MaxPasswordAttempts,
		Details: object.ToViewDetailsPb(
			policy.Sequence,
			policy.CreationDate,
			policy.ChangeDate,
			policy.ResourceOwner,
		),
	}
}
