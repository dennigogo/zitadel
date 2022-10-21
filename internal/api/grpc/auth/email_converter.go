package auth

import (
	"context"

	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/pkg/grpc/auth"
)

func UpdateMyEmailToDomain(ctx context.Context, email *auth.SetMyEmailRequest) *domain.Email {
	return &domain.Email{
		ObjectRoot:   ctxToObjectRoot(ctx),
		EmailAddress: email.Email,
	}
}
