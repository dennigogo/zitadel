package instance

import (
	"context"

	"github.com/dennigogo/zitadel/internal/eventstore"

	"github.com/dennigogo/zitadel/internal/eventstore/repository"
	"github.com/dennigogo/zitadel/internal/repository/member"
)

var (
	MemberAddedEventType          = instanceEventTypePrefix + member.AddedEventType
	MemberChangedEventType        = instanceEventTypePrefix + member.ChangedEventType
	MemberRemovedEventType        = instanceEventTypePrefix + member.RemovedEventType
	MemberCascadeRemovedEventType = instanceEventTypePrefix + member.CascadeRemovedEventType
)

type MemberAddedEvent struct {
	member.MemberAddedEvent
}

func NewMemberAddedEvent(
	ctx context.Context,
	aggregate *eventstore.Aggregate,
	userID string,
	roles ...string,
) *MemberAddedEvent {

	return &MemberAddedEvent{
		MemberAddedEvent: *member.NewMemberAddedEvent(
			eventstore.NewBaseEventForPush(
				ctx,
				aggregate,
				MemberAddedEventType,
			),
			userID,
			roles...,
		),
	}
}

func MemberAddedEventMapper(event *repository.Event) (eventstore.Event, error) {
	e, err := member.MemberAddedEventMapper(event)
	if err != nil {
		return nil, err
	}

	return &MemberAddedEvent{MemberAddedEvent: *e.(*member.MemberAddedEvent)}, nil
}

type MemberChangedEvent struct {
	member.MemberChangedEvent
}

func NewMemberChangedEvent(
	ctx context.Context,
	aggregate *eventstore.Aggregate,
	userID string,
	roles ...string,
) *MemberChangedEvent {
	return &MemberChangedEvent{
		MemberChangedEvent: *member.NewMemberChangedEvent(
			eventstore.NewBaseEventForPush(
				ctx,
				aggregate,
				MemberChangedEventType,
			),
			userID,
			roles...,
		),
	}
}

func MemberChangedEventMapper(event *repository.Event) (eventstore.Event, error) {
	e, err := member.ChangedEventMapper(event)
	if err != nil {
		return nil, err
	}

	return &MemberChangedEvent{MemberChangedEvent: *e.(*member.MemberChangedEvent)}, nil
}

type MemberRemovedEvent struct {
	member.MemberRemovedEvent
}

func NewMemberRemovedEvent(
	ctx context.Context,
	aggregate *eventstore.Aggregate,
	userID string,
) *MemberRemovedEvent {
	return &MemberRemovedEvent{
		MemberRemovedEvent: *member.NewRemovedEvent(
			eventstore.NewBaseEventForPush(
				ctx,
				aggregate,
				MemberRemovedEventType,
			),
			userID,
		),
	}
}

func MemberRemovedEventMapper(event *repository.Event) (eventstore.Event, error) {
	e, err := member.RemovedEventMapper(event)
	if err != nil {
		return nil, err
	}

	return &MemberRemovedEvent{MemberRemovedEvent: *e.(*member.MemberRemovedEvent)}, nil
}

type MemberCascadeRemovedEvent struct {
	member.MemberCascadeRemovedEvent
}

func NewMemberCascadeRemovedEvent(
	ctx context.Context,
	aggregate *eventstore.Aggregate,
	userID string,
) *MemberCascadeRemovedEvent {
	return &MemberCascadeRemovedEvent{
		MemberCascadeRemovedEvent: *member.NewCascadeRemovedEvent(
			eventstore.NewBaseEventForPush(
				ctx,
				aggregate,
				MemberCascadeRemovedEventType,
			),
			userID,
		),
	}
}

func MemberCascadeRemovedEventMapper(event *repository.Event) (eventstore.Event, error) {
	e, err := member.CascadeRemovedEventMapper(event)
	if err != nil {
		return nil, err
	}

	return &MemberCascadeRemovedEvent{MemberCascadeRemovedEvent: *e.(*member.MemberCascadeRemovedEvent)}, nil
}
