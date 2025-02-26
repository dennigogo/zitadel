package keypair

import (
	"github.com/dennigogo/zitadel/internal/eventstore"
)

const (
	AggregateType    = "key_pair"
	AggregateVersion = "v1"
)

type Aggregate struct {
	eventstore.Aggregate
}
