package admin

import (
	"github.com/dennigogo/zitadel/internal/domain"
	admin_pb "github.com/dennigogo/zitadel/pkg/grpc/admin"
)

func UpdatePasswordAgePolicyToDomain(policy *admin_pb.UpdatePasswordAgePolicyRequest) *domain.PasswordAgePolicy {
	return &domain.PasswordAgePolicy{
		MaxAgeDays:     uint64(policy.MaxAgeDays),
		ExpireWarnDays: uint64(policy.ExpireWarnDays),
	}
}
