package eventstore

import (
	"context"
	"time"

	"github.com/zitadel/logging"

	"github.com/dennigogo/zitadel/internal/api/authz"
	"github.com/dennigogo/zitadel/internal/auth/repository/eventsourcing/view"
	"github.com/dennigogo/zitadel/internal/crypto"
	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/errors"
	v1 "github.com/dennigogo/zitadel/internal/eventstore/v1"
	"github.com/dennigogo/zitadel/internal/eventstore/v1/models"
	"github.com/dennigogo/zitadel/internal/telemetry/tracing"
	usr_model "github.com/dennigogo/zitadel/internal/user/model"
	usr_view "github.com/dennigogo/zitadel/internal/user/repository/view"
	"github.com/dennigogo/zitadel/internal/user/repository/view/model"
)

type RefreshTokenRepo struct {
	Eventstore   v1.Eventstore
	View         *view.View
	SearchLimit  uint64
	KeyAlgorithm crypto.EncryptionAlgorithm
}

func (r *RefreshTokenRepo) RefreshTokenByID(ctx context.Context, refreshToken string) (*usr_model.RefreshTokenView, error) {
	userID, tokenID, token, err := domain.FromRefreshToken(refreshToken, r.KeyAlgorithm)
	if err != nil {
		return nil, err
	}
	tokenView, viewErr := r.View.RefreshTokenByID(tokenID, authz.GetInstance(ctx).InstanceID())
	if viewErr != nil && !errors.IsNotFound(viewErr) {
		return nil, viewErr
	}
	if errors.IsNotFound(viewErr) {
		tokenView = new(model.RefreshTokenView)
		tokenView.ID = tokenID
		tokenView.UserID = userID
	}

	events, esErr := r.getUserEvents(ctx, userID, tokenView.InstanceID, tokenView.Sequence)
	if errors.IsNotFound(viewErr) && len(events) == 0 {
		return nil, errors.ThrowNotFound(nil, "EVENT-BHB52", "Errors.User.RefreshToken.Invalid")
	}

	if esErr != nil {
		logging.Log("EVENT-AE462").WithError(viewErr).WithField("traceID", tracing.TraceIDFromCtx(ctx)).Debug("error retrieving new events")
		return model.RefreshTokenViewToModel(tokenView), nil
	}
	viewToken := *tokenView
	for _, event := range events {
		err := tokenView.AppendEventIfMyRefreshToken(event)
		if err != nil {
			return model.RefreshTokenViewToModel(&viewToken), nil
		}
	}
	if !tokenView.Expiration.After(time.Now()) || tokenView.Token != token {
		return nil, errors.ThrowNotFound(nil, "EVENT-5Bm9s", "Errors.User.RefreshToken.Invalid")
	}
	return model.RefreshTokenViewToModel(tokenView), nil
}

func (r *RefreshTokenRepo) SearchMyRefreshTokens(ctx context.Context, userID string, request *usr_model.RefreshTokenSearchRequest) (*usr_model.RefreshTokenSearchResponse, error) {
	err := request.EnsureLimit(r.SearchLimit)
	if err != nil {
		return nil, err
	}
	sequence, err := r.View.GetLatestRefreshTokenSequence(authz.GetInstance(ctx).InstanceID())
	logging.Log("EVENT-GBdn4").OnError(err).WithField("traceID", tracing.TraceIDFromCtx(ctx)).Warn("could not read latest refresh token sequence")
	request.Queries = append(request.Queries, &usr_model.RefreshTokenSearchQuery{Key: usr_model.RefreshTokenSearchKeyUserID, Method: domain.SearchMethodEquals, Value: userID})
	tokens, count, err := r.View.SearchRefreshTokens(request)
	if err != nil {
		return nil, err
	}
	return &usr_model.RefreshTokenSearchResponse{
		Offset:      request.Offset,
		Limit:       request.Limit,
		TotalResult: count,
		Sequence:    sequence.CurrentSequence,
		Timestamp:   sequence.LastSuccessfulSpoolerRun,
		Result:      model.RefreshTokenViewsToModel(tokens),
	}, nil
}

func (r *RefreshTokenRepo) getUserEvents(ctx context.Context, userID, instanceID string, sequence uint64) ([]*models.Event, error) {
	query, err := usr_view.UserByIDQuery(userID, instanceID, sequence)
	if err != nil {
		return nil, err
	}
	return r.Eventstore.FilterEvents(ctx, query)
}
