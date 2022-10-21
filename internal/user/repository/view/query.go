package view

import (
	"github.com/dennigogo/zitadel/internal/errors"
	es_models "github.com/dennigogo/zitadel/internal/eventstore/v1/models"
	"github.com/dennigogo/zitadel/internal/repository/user"
)

func UserByIDQuery(id, instanceID string, latestSequence uint64) (*es_models.SearchQuery, error) {
	if id == "" {
		return nil, errors.ThrowPreconditionFailed(nil, "EVENT-d8isw", "Errors.User.UserIDMissing")
	}
	return es_models.NewSearchQuery().
		AddQuery().
		AggregateTypeFilter(user.AggregateType).
		AggregateIDFilter(id).
		LatestSequenceFilter(latestSequence).
		InstanceIDFilter(instanceID).
		SearchQuery(), nil
}
