package action

import "github.com/dennigogo/zitadel/internal/eventstore"

func RegisterEventMappers(es *eventstore.Eventstore) {
	es.RegisterFilterEventMapper(AddedEventType, AddedEventMapper).
		RegisterFilterEventMapper(ChangedEventType, ChangedEventMapper).
		RegisterFilterEventMapper(DeactivatedEventType, DeactivatedEventMapper).
		RegisterFilterEventMapper(ReactivatedEventType, ReactivatedEventMapper).
		RegisterFilterEventMapper(RemovedEventType, RemovedEventMapper)
}
