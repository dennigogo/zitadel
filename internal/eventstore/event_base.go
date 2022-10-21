package eventstore

import (
	"context"
	"time"

	"github.com/dennigogo/zitadel/internal/api/authz"
	"github.com/dennigogo/zitadel/internal/api/service"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
)

// BaseEvent represents the minimum metadata of an event
type BaseEvent struct {
	EventType EventType `json:"-"`

	aggregate Aggregate

	sequence                      uint64
	creationDate                  time.Time
	previousAggregateSequence     uint64
	previousAggregateTypeSequence uint64

	//User who created the event
	User string `json:"-"`
	//Service which created the event
	Service string `json:"-"`
	Data    []byte `json:"-"`
}

// EditorService implements Command
func (e *BaseEvent) EditorService() string {
	return e.Service
}

// EditorUser implements Command
func (e *BaseEvent) EditorUser() string {
	return e.User
}

// Type implements Command
func (e *BaseEvent) Type() EventType {
	return e.EventType
}

// Sequence is an upcounting unique number of the event
func (e *BaseEvent) Sequence() uint64 {
	return e.sequence
}

// CreationDate is the the time, the event is inserted into the eventstore
func (e *BaseEvent) CreationDate() time.Time {
	return e.creationDate
}

// Aggregate represents the metadata of the event's aggregate
func (e *BaseEvent) Aggregate() Aggregate {
	return e.aggregate
}

// Data returns the payload of the event. It represent the changed fields by the event
func (e *BaseEvent) DataAsBytes() []byte {
	return e.Data
}

// PreviousAggregateSequence implements EventReader
func (e *BaseEvent) PreviousAggregateSequence() uint64 {
	return e.previousAggregateSequence
}

// PreviousAggregateTypeSequence implements EventReader
func (e *BaseEvent) PreviousAggregateTypeSequence() uint64 {
	return e.previousAggregateTypeSequence
}

// BaseEventFromRepo maps a stored event to a BaseEvent
func BaseEventFromRepo(event *repository.Event) *BaseEvent {
	return &BaseEvent{
		aggregate: Aggregate{
			ID:            event.AggregateID,
			Type:          AggregateType(event.AggregateType),
			ResourceOwner: event.ResourceOwner.String,
			InstanceID:    event.InstanceID,
			Version:       Version(event.Version),
		},
		EventType:                     EventType(event.Type),
		creationDate:                  event.CreationDate,
		sequence:                      event.Sequence,
		previousAggregateSequence:     event.PreviousAggregateSequence,
		previousAggregateTypeSequence: event.PreviousAggregateTypeSequence,
		Service:                       event.EditorService,
		User:                          event.EditorUser,
		Data:                          event.Data,
	}
}

// NewBaseEventForPush is the constructor for event's which will be pushed into the eventstore
// the resource owner of the aggregate is only used if it's the first event of this aggregate type
// afterwards the resource owner of the first previous events is taken
func NewBaseEventForPush(ctx context.Context, aggregate *Aggregate, typ EventType) *BaseEvent {
	return &BaseEvent{
		aggregate: *aggregate,
		User:      authz.GetCtxData(ctx).UserID,
		Service:   service.FromContext(ctx),
		EventType: typ,
	}
}
