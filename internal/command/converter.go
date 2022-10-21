package command

import (
	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/eventstore"
)

func writeModelToObjectDetails(writeModel *eventstore.WriteModel) *domain.ObjectDetails {
	return &domain.ObjectDetails{
		Sequence:      writeModel.ProcessedSequence,
		ResourceOwner: writeModel.ResourceOwner,
		EventDate:     writeModel.ChangeDate,
	}
}

func pushedEventsToObjectDetails(events []eventstore.Event) *domain.ObjectDetails {
	return &domain.ObjectDetails{
		Sequence:      events[len(events)-1].Sequence(),
		EventDate:     events[len(events)-1].CreationDate(),
		ResourceOwner: events[len(events)-1].Aggregate().ResourceOwner,
	}
}
