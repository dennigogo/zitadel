package repository

import (
	"context"

	"github.com/dennigogo/zitadel/internal/user/model"
)

type RefreshTokenRepository interface {
	RefreshTokenByID(ctx context.Context, refreshToken string) (*model.RefreshTokenView, error)
	SearchMyRefreshTokens(ctx context.Context, userID string, request *model.RefreshTokenSearchRequest) (*model.RefreshTokenSearchResponse, error)
}
