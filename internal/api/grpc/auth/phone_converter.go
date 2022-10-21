package auth

import (
	"context"

	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/pkg/grpc/auth"
)

func UpdateMyPhoneToDomain(ctx context.Context, phone *auth.SetMyPhoneRequest) *domain.Phone {
	return &domain.Phone{
		ObjectRoot:  ctxToObjectRoot(ctx),
		PhoneNumber: phone.Phone,
	}
}
