package user

import (
	"github.com/dennigogo/zitadel/internal/eventstore"
)

const (
	AggregateType    = "user"
	AggregateVersion = "v2"
)

type Aggregate struct {
	eventstore.Aggregate
}

func NewAggregate(id, resourceOwner string) *Aggregate {
	return &Aggregate{
		Aggregate: eventstore.Aggregate{
			Type:          AggregateType,
			Version:       AggregateVersion,
			ID:            id,
			ResourceOwner: resourceOwner,
		},
	}
}
