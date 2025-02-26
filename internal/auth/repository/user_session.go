package repository

import (
	"context"

	"github.com/dennigogo/zitadel/internal/user/model"
)

type UserSessionRepository interface {
	GetMyUserSessions(ctx context.Context) ([]*model.UserSessionView, error)
	ActiveUserSessionCount() int64
}
