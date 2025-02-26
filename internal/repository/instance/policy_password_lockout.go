package instance

import (
	"context"

	"github.com/dennigogo/zitadel/internal/eventstore"

	"github.com/dennigogo/zitadel/internal/eventstore/repository"
	"github.com/dennigogo/zitadel/internal/repository/policy"
)

var (
	LockoutPolicyAddedEventType   = instanceEventTypePrefix + policy.LockoutPolicyAddedEventType
	LockoutPolicyChangedEventType = instanceEventTypePrefix + policy.LockoutPolicyChangedEventType
)

type LockoutPolicyAddedEvent struct {
	policy.LockoutPolicyAddedEvent
}

func NewLockoutPolicyAddedEvent(
	ctx context.Context,
	aggregate *eventstore.Aggregate,
	maxAttempts uint64,
	showLockoutFailure bool,
) *LockoutPolicyAddedEvent {
	return &LockoutPolicyAddedEvent{
		LockoutPolicyAddedEvent: *policy.NewLockoutPolicyAddedEvent(
			eventstore.NewBaseEventForPush(
				ctx,
				aggregate,
				LockoutPolicyAddedEventType),
			maxAttempts,
			showLockoutFailure),
	}
}

func LockoutPolicyAddedEventMapper(event *repository.Event) (eventstore.Event, error) {
	e, err := policy.LockoutPolicyAddedEventMapper(event)
	if err != nil {
		return nil, err
	}

	return &LockoutPolicyAddedEvent{LockoutPolicyAddedEvent: *e.(*policy.LockoutPolicyAddedEvent)}, nil
}

type LockoutPolicyChangedEvent struct {
	policy.LockoutPolicyChangedEvent
}

func NewLockoutPolicyChangedEvent(
	ctx context.Context,
	aggregate *eventstore.Aggregate,
	changes []policy.LockoutPolicyChanges,
) (*LockoutPolicyChangedEvent, error) {
	changedEvent, err := policy.NewLockoutPolicyChangedEvent(
		eventstore.NewBaseEventForPush(
			ctx,
			aggregate,
			LockoutPolicyChangedEventType),
		changes,
	)
	if err != nil {
		return nil, err
	}
	return &LockoutPolicyChangedEvent{LockoutPolicyChangedEvent: *changedEvent}, nil
}

func LockoutPolicyChangedEventMapper(event *repository.Event) (eventstore.Event, error) {
	e, err := policy.LockoutPolicyChangedEventMapper(event)
	if err != nil {
		return nil, err
	}

	return &LockoutPolicyChangedEvent{LockoutPolicyChangedEvent: *e.(*policy.LockoutPolicyChangedEvent)}, nil
}
