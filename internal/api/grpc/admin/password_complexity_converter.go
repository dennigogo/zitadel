package admin

import (
	"github.com/dennigogo/zitadel/internal/domain"
	admin_pb "github.com/dennigogo/zitadel/pkg/grpc/admin"
)

func UpdatePasswordComplexityPolicyToDomain(req *admin_pb.UpdatePasswordComplexityPolicyRequest) *domain.PasswordComplexityPolicy {
	return &domain.PasswordComplexityPolicy{
		MinLength:    uint64(req.MinLength),
		HasLowercase: req.HasLowercase,
		HasUppercase: req.HasUppercase,
		HasNumber:    req.HasNumber,
		HasSymbol:    req.HasSymbol,
	}
}
