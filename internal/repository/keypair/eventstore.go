package keypair

import (
	"github.com/dennigogo/zitadel/internal/eventstore"
)

func RegisterEventMappers(es *eventstore.Eventstore) {
	es.RegisterFilterEventMapper(AddedEventType, AddedEventMapper)
	es.RegisterFilterEventMapper(AddedCertificateEventType, AddedCertificateEventMapper)
}
