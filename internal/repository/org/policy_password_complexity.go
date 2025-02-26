package org

import (
	"context"

	"github.com/dennigogo/zitadel/internal/eventstore"

	"github.com/dennigogo/zitadel/internal/eventstore/repository"
	"github.com/dennigogo/zitadel/internal/repository/policy"
)

var (
	PasswordComplexityPolicyAddedEventType   = orgEventTypePrefix + policy.PasswordComplexityPolicyAddedEventType
	PasswordComplexityPolicyChangedEventType = orgEventTypePrefix + policy.PasswordComplexityPolicyChangedEventType
	PasswordComplexityPolicyRemovedEventType = orgEventTypePrefix + policy.PasswordComplexityPolicyRemovedEventType
)

type PasswordComplexityPolicyAddedEvent struct {
	policy.PasswordComplexityPolicyAddedEvent
}

func NewPasswordComplexityPolicyAddedEvent(
	ctx context.Context,
	aggregate *eventstore.Aggregate,
	minLength uint64,
	hasLowercase,
	hasUppercase,
	hasNumber,
	hasSymbol bool,
) *PasswordComplexityPolicyAddedEvent {
	return &PasswordComplexityPolicyAddedEvent{
		PasswordComplexityPolicyAddedEvent: *policy.NewPasswordComplexityPolicyAddedEvent(
			eventstore.NewBaseEventForPush(
				ctx,
				aggregate,
				PasswordComplexityPolicyAddedEventType),
			minLength,
			hasLowercase,
			hasUppercase,
			hasNumber,
			hasSymbol),
	}
}

func PasswordComplexityPolicyAddedEventMapper(event *repository.Event) (eventstore.Event, error) {
	e, err := policy.PasswordComplexityPolicyAddedEventMapper(event)
	if err != nil {
		return nil, err
	}

	return &PasswordComplexityPolicyAddedEvent{PasswordComplexityPolicyAddedEvent: *e.(*policy.PasswordComplexityPolicyAddedEvent)}, nil
}

type PasswordComplexityPolicyChangedEvent struct {
	policy.PasswordComplexityPolicyChangedEvent
}

func NewPasswordComplexityPolicyChangedEvent(
	ctx context.Context,
	aggregate *eventstore.Aggregate,
	changes []policy.PasswordComplexityPolicyChanges,
) (*PasswordComplexityPolicyChangedEvent, error) {
	changedEvent, err := policy.NewPasswordComplexityPolicyChangedEvent(
		eventstore.NewBaseEventForPush(
			ctx,
			aggregate,
			PasswordComplexityPolicyChangedEventType),
		changes,
	)
	if err != nil {
		return nil, err
	}
	return &PasswordComplexityPolicyChangedEvent{PasswordComplexityPolicyChangedEvent: *changedEvent}, nil
}

func PasswordComplexityPolicyChangedEventMapper(event *repository.Event) (eventstore.Event, error) {
	e, err := policy.PasswordComplexityPolicyChangedEventMapper(event)
	if err != nil {
		return nil, err
	}

	return &PasswordComplexityPolicyChangedEvent{PasswordComplexityPolicyChangedEvent: *e.(*policy.PasswordComplexityPolicyChangedEvent)}, nil
}

type PasswordComplexityPolicyRemovedEvent struct {
	policy.PasswordComplexityPolicyRemovedEvent
}

func NewPasswordComplexityPolicyRemovedEvent(
	ctx context.Context,
	aggregate *eventstore.Aggregate,
) *PasswordComplexityPolicyRemovedEvent {
	return &PasswordComplexityPolicyRemovedEvent{
		PasswordComplexityPolicyRemovedEvent: *policy.NewPasswordComplexityPolicyRemovedEvent(
			eventstore.NewBaseEventForPush(
				ctx,
				aggregate,
				PasswordComplexityPolicyRemovedEventType),
		),
	}
}

func PasswordComplexityPolicyRemovedEventMapper(event *repository.Event) (eventstore.Event, error) {
	e, err := policy.PasswordComplexityPolicyRemovedEventMapper(event)
	if err != nil {
		return nil, err
	}

	return &PasswordComplexityPolicyRemovedEvent{PasswordComplexityPolicyRemovedEvent: *e.(*policy.PasswordComplexityPolicyRemovedEvent)}, nil
}
