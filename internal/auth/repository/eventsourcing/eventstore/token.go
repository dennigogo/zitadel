package eventstore

import (
	"context"
	"time"

	"github.com/zitadel/logging"

	"github.com/dennigogo/zitadel/internal/api/authz"
	"github.com/dennigogo/zitadel/internal/auth/repository/eventsourcing/view"
	"github.com/dennigogo/zitadel/internal/errors"
	v1 "github.com/dennigogo/zitadel/internal/eventstore/v1"
	"github.com/dennigogo/zitadel/internal/eventstore/v1/models"
	"github.com/dennigogo/zitadel/internal/telemetry/tracing"
	usr_model "github.com/dennigogo/zitadel/internal/user/model"
	usr_view "github.com/dennigogo/zitadel/internal/user/repository/view"
	"github.com/dennigogo/zitadel/internal/user/repository/view/model"
)

type TokenRepo struct {
	Eventstore v1.Eventstore
	View       *view.View
}

func (repo *TokenRepo) IsTokenValid(ctx context.Context, userID, tokenID string) (bool, error) {
	token, err := repo.TokenByIDs(ctx, userID, tokenID)
	if err == nil {
		return token.Expiration.After(time.Now().UTC()), nil
	}
	if errors.IsNotFound(err) {
		return false, nil
	}
	return false, err
}

func (repo *TokenRepo) TokenByIDs(ctx context.Context, userID, tokenID string) (*usr_model.TokenView, error) {
	token, viewErr := repo.View.TokenByIDs(tokenID, userID, authz.GetInstance(ctx).InstanceID())
	if viewErr != nil && !errors.IsNotFound(viewErr) {
		return nil, viewErr
	}
	if errors.IsNotFound(viewErr) {
		token = new(model.TokenView)
		token.ID = tokenID
		token.UserID = userID
	}

	events, esErr := repo.getUserEvents(ctx, userID, token.InstanceID, token.Sequence)
	if errors.IsNotFound(viewErr) && len(events) == 0 {
		return nil, errors.ThrowNotFound(nil, "EVENT-4T90g", "Errors.Token.NotFound")
	}

	if esErr != nil {
		logging.Log("EVENT-5Nm9s").WithError(viewErr).WithField("traceID", tracing.TraceIDFromCtx(ctx)).Debug("error retrieving new events")
		return model.TokenViewToModel(token), nil
	}
	viewToken := *token
	for _, event := range events {
		err := token.AppendEventIfMyToken(event)
		if err != nil {
			return model.TokenViewToModel(&viewToken), nil
		}
	}
	if !token.Expiration.After(time.Now().UTC()) || token.Deactivated {
		return nil, errors.ThrowNotFound(nil, "EVENT-5Bm9s", "Errors.Token.NotFound")
	}
	return model.TokenViewToModel(token), nil
}

func (r *TokenRepo) getUserEvents(ctx context.Context, userID, instanceID string, sequence uint64) ([]*models.Event, error) {
	query, err := usr_view.UserByIDQuery(userID, instanceID, sequence)
	if err != nil {
		return nil, err
	}
	return r.Eventstore.FilterEvents(ctx, query)
}
