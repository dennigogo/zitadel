package management

import (
	"github.com/dennigogo/zitadel/internal/domain"
	mgmt "github.com/dennigogo/zitadel/pkg/grpc/management"
)

func AddLockoutPolicyToDomain(p *mgmt.AddCustomLockoutPolicyRequest) *domain.LockoutPolicy {
	return &domain.LockoutPolicy{
		MaxPasswordAttempts: uint64(p.MaxPasswordAttempts),
	}
}

func UpdateLockoutPolicyToDomain(p *mgmt.UpdateCustomLockoutPolicyRequest) *domain.LockoutPolicy {
	return &domain.LockoutPolicy{
		MaxPasswordAttempts: uint64(p.MaxPasswordAttempts),
	}
}
