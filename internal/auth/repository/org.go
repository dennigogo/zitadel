package repository

import (
	"context"

	"github.com/dennigogo/zitadel/internal/domain"
	iam_model "github.com/dennigogo/zitadel/internal/iam/model"
)

type OrgRepository interface {
	GetIDPConfigByID(ctx context.Context, idpConfigID string) (*iam_model.IDPConfigView, error)
	GetMyPasswordComplexityPolicy(ctx context.Context) (*iam_model.PasswordComplexityPolicyView, error)
	GetLoginText(ctx context.Context, orgID string) ([]*domain.CustomText, error)
}
