package user

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dennigogo/zitadel/internal/eventstore"

	"github.com/dennigogo/zitadel/internal/crypto"
	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
)

const (
	passwordEventPrefix             = humanEventPrefix + "password."
	HumanPasswordChangedType        = passwordEventPrefix + "changed"
	HumanPasswordCodeAddedType      = passwordEventPrefix + "code.added"
	HumanPasswordCodeSentType       = passwordEventPrefix + "code.sent"
	HumanPasswordCheckSucceededType = passwordEventPrefix + "check.succeeded"
	HumanPasswordCheckFailedType    = passwordEventPrefix + "check.failed"
)

type HumanPasswordChangedEvent struct {
	eventstore.BaseEvent `json:"-"`

	Secret         *crypto.CryptoValue `json:"secret,omitempty"`
	ChangeRequired bool                `json:"changeRequired"`
	UserAgentID    string              `json:"userAgentID,omitempty"`
}

func (e *HumanPasswordChangedEvent) Data() interface{} {
	return e
}

func (e *HumanPasswordChangedEvent) UniqueConstraints() []*eventstore.EventUniqueConstraint {
	return nil
}

func NewHumanPasswordChangedEvent(
	ctx context.Context,
	aggregate *eventstore.Aggregate,
	secret *crypto.CryptoValue,
	changeRequired bool,
	userAgentID string,
) *HumanPasswordChangedEvent {
	return &HumanPasswordChangedEvent{
		BaseEvent: *eventstore.NewBaseEventForPush(
			ctx,
			aggregate,
			HumanPasswordChangedType,
		),
		Secret:         secret,
		ChangeRequired: changeRequired,
		UserAgentID:    userAgentID,
	}
}

func HumanPasswordChangedEventMapper(event *repository.Event) (eventstore.Event, error) {
	humanAdded := &HumanPasswordChangedEvent{
		BaseEvent: *eventstore.BaseEventFromRepo(event),
	}
	err := json.Unmarshal(event.Data, humanAdded)
	if err != nil {
		return nil, errors.ThrowInternal(err, "USER-4M0sd", "unable to unmarshal human password changed")
	}

	return humanAdded, nil
}

type HumanPasswordCodeAddedEvent struct {
	eventstore.BaseEvent `json:"-"`

	Code             *crypto.CryptoValue     `json:"code,omitempty"`
	Expiry           time.Duration           `json:"expiry,omitempty"`
	NotificationType domain.NotificationType `json:"notificationType,omitempty"`
}

func (e *HumanPasswordCodeAddedEvent) Data() interface{} {
	return e
}

func (e *HumanPasswordCodeAddedEvent) UniqueConstraints() []*eventstore.EventUniqueConstraint {
	return nil
}

func NewHumanPasswordCodeAddedEvent(
	ctx context.Context,
	aggregate *eventstore.Aggregate,
	code *crypto.CryptoValue,
	expiry time.Duration,
	notificationType domain.NotificationType,
) *HumanPasswordCodeAddedEvent {
	return &HumanPasswordCodeAddedEvent{
		BaseEvent: *eventstore.NewBaseEventForPush(
			ctx,
			aggregate,
			HumanPasswordCodeAddedType,
		),
		Code:             code,
		Expiry:           expiry,
		NotificationType: notificationType,
	}
}

func HumanPasswordCodeAddedEventMapper(event *repository.Event) (eventstore.Event, error) {
	humanAdded := &HumanPasswordCodeAddedEvent{
		BaseEvent: *eventstore.BaseEventFromRepo(event),
	}
	err := json.Unmarshal(event.Data, humanAdded)
	if err != nil {
		return nil, errors.ThrowInternal(err, "USER-Ms90d", "unable to unmarshal human password code added")
	}

	return humanAdded, nil
}

type HumanPasswordCodeSentEvent struct {
	eventstore.BaseEvent `json:"-"`
}

func (e *HumanPasswordCodeSentEvent) Data() interface{} {
	return nil
}

func (e *HumanPasswordCodeSentEvent) UniqueConstraints() []*eventstore.EventUniqueConstraint {
	return nil
}

func NewHumanPasswordCodeSentEvent(ctx context.Context, aggregate *eventstore.Aggregate) *HumanPasswordCodeSentEvent {
	return &HumanPasswordCodeSentEvent{
		BaseEvent: *eventstore.NewBaseEventForPush(
			ctx,
			aggregate,
			HumanPasswordCodeSentType,
		),
	}
}

func HumanPasswordCodeSentEventMapper(event *repository.Event) (eventstore.Event, error) {
	return &HumanPasswordCodeSentEvent{
		BaseEvent: *eventstore.BaseEventFromRepo(event),
	}, nil
}

type HumanPasswordCheckSucceededEvent struct {
	eventstore.BaseEvent `json:"-"`
	*AuthRequestInfo
}

func (e *HumanPasswordCheckSucceededEvent) Data() interface{} {
	return e
}

func (e *HumanPasswordCheckSucceededEvent) UniqueConstraints() []*eventstore.EventUniqueConstraint {
	return nil
}

func NewHumanPasswordCheckSucceededEvent(
	ctx context.Context,
	aggregate *eventstore.Aggregate,
	info *AuthRequestInfo,
) *HumanPasswordCheckSucceededEvent {
	return &HumanPasswordCheckSucceededEvent{
		BaseEvent: *eventstore.NewBaseEventForPush(
			ctx,
			aggregate,
			HumanPasswordCheckSucceededType,
		),
		AuthRequestInfo: info,
	}
}

func HumanPasswordCheckSucceededEventMapper(event *repository.Event) (eventstore.Event, error) {
	humanAdded := &HumanPasswordCheckSucceededEvent{
		BaseEvent: *eventstore.BaseEventFromRepo(event),
	}
	err := json.Unmarshal(event.Data, humanAdded)
	if err != nil {
		return nil, errors.ThrowInternal(err, "USER-5M9sd", "unable to unmarshal human password check succeeded")
	}

	return humanAdded, nil
}

type HumanPasswordCheckFailedEvent struct {
	eventstore.BaseEvent `json:"-"`
	*AuthRequestInfo
}

func (e *HumanPasswordCheckFailedEvent) Data() interface{} {
	return e
}

func (e *HumanPasswordCheckFailedEvent) UniqueConstraints() []*eventstore.EventUniqueConstraint {
	return nil
}

func NewHumanPasswordCheckFailedEvent(
	ctx context.Context,
	aggregate *eventstore.Aggregate,
	info *AuthRequestInfo,
) *HumanPasswordCheckFailedEvent {
	return &HumanPasswordCheckFailedEvent{
		BaseEvent: *eventstore.NewBaseEventForPush(
			ctx,
			aggregate,
			HumanPasswordCheckFailedType,
		),
		AuthRequestInfo: info,
	}
}

func HumanPasswordCheckFailedEventMapper(event *repository.Event) (eventstore.Event, error) {
	humanAdded := &HumanPasswordCheckFailedEvent{
		BaseEvent: *eventstore.BaseEventFromRepo(event),
	}
	err := json.Unmarshal(event.Data, humanAdded)
	if err != nil {
		return nil, errors.ThrowInternal(err, "USER-4m9fs", "unable to unmarshal human password check failed")
	}

	return humanAdded, nil
}
