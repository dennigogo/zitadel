package management

import (
	"github.com/dennigogo/zitadel/internal/domain"
	mgmt_pb "github.com/dennigogo/zitadel/pkg/grpc/management"
)

func AddPasswordAgePolicyToDomain(policy *mgmt_pb.AddCustomPasswordAgePolicyRequest) *domain.PasswordAgePolicy {
	return &domain.PasswordAgePolicy{
		MaxAgeDays:     uint64(policy.MaxAgeDays),
		ExpireWarnDays: uint64(policy.ExpireWarnDays),
	}
}

func UpdatePasswordAgePolicyToDomain(policy *mgmt_pb.UpdateCustomPasswordAgePolicyRequest) *domain.PasswordAgePolicy {
	return &domain.PasswordAgePolicy{
		MaxAgeDays:     uint64(policy.MaxAgeDays),
		ExpireWarnDays: uint64(policy.ExpireWarnDays),
	}
}
