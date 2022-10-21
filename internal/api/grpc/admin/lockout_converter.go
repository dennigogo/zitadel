package admin

import (
	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/pkg/grpc/admin"
)

func UpdateLockoutPolicyToDomain(p *admin.UpdateLockoutPolicyRequest) *domain.LockoutPolicy {
	return &domain.LockoutPolicy{
		MaxPasswordAttempts: uint64(p.MaxPasswordAttempts),
	}
}
