package command

import (
	"context"

	"github.com/dennigogo/zitadel/internal/api/authz"
	"github.com/dennigogo/zitadel/internal/command/preparation"
	"github.com/dennigogo/zitadel/internal/errors"
)

func passwordComplexityPolicyWriteModel(ctx context.Context, filter preparation.FilterToQueryReducer) (*PasswordComplexityPolicyWriteModel, error) {
	wm, err := customPasswordComplexityPolicy(ctx, filter)
	if err != nil || wm != nil && wm.State.Exists() {
		return wm, err
	}
	wm, err = defaultPasswordComplexityPolicy(ctx, filter)
	if err != nil || wm != nil {
		return wm, err
	}
	return nil, errors.ThrowInternal(nil, "USER-uQ96e", "Errors.Internal")
}

func customPasswordComplexityPolicy(ctx context.Context, filter preparation.FilterToQueryReducer) (*PasswordComplexityPolicyWriteModel, error) {
	policy := NewOrgPasswordComplexityPolicyWriteModel(authz.GetCtxData(ctx).OrgID)
	events, err := filter(ctx, policy.Query())
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, nil
	}
	policy.AppendEvents(events...)
	err = policy.Reduce()
	return &policy.PasswordComplexityPolicyWriteModel, err
}

func defaultPasswordComplexityPolicy(ctx context.Context, filter preparation.FilterToQueryReducer) (*PasswordComplexityPolicyWriteModel, error) {
	policy := NewInstancePasswordComplexityPolicyWriteModel(ctx)
	events, err := filter(ctx, policy.Query())
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, nil
	}
	policy.AppendEvents(events...)
	err = policy.Reduce()
	return &policy.PasswordComplexityPolicyWriteModel, err
}
