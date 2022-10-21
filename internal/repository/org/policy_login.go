package org

import (
	"context"
	"time"

	"github.com/dennigogo/zitadel/internal/eventstore"

	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
	"github.com/dennigogo/zitadel/internal/repository/policy"
)

var (
	LoginPolicyAddedEventType   = orgEventTypePrefix + policy.LoginPolicyAddedEventType
	LoginPolicyChangedEventType = orgEventTypePrefix + policy.LoginPolicyChangedEventType
	LoginPolicyRemovedEventType = orgEventTypePrefix + policy.LoginPolicyRemovedEventType
)

type LoginPolicyAddedEvent struct {
	policy.LoginPolicyAddedEvent
}

func NewLoginPolicyAddedEvent(
	ctx context.Context,
	aggregate *eventstore.Aggregate,
	allowUsernamePassword,
	allowRegister,
	allowExternalIDP,
	forceMFA,
	hidePasswordReset,
	ignoreUnknownUsernames,
	allowDomainDiscovery,
	disableLoginWithEmail,
	disableLoginWithPhone bool,
	passwordlessType domain.PasswordlessType,
	defaultRedirectURI string,
	passwordCheckLifetime,
	externalLoginCheckLifetime,
	mfaInitSkipLifetime,
	secondFactorCheckLifetime,
	multiFactorCheckLifetime time.Duration,
) *LoginPolicyAddedEvent {
	return &LoginPolicyAddedEvent{
		LoginPolicyAddedEvent: *policy.NewLoginPolicyAddedEvent(
			eventstore.NewBaseEventForPush(
				ctx,
				aggregate,
				LoginPolicyAddedEventType),
			allowUsernamePassword,
			allowRegister,
			allowExternalIDP,
			forceMFA,
			hidePasswordReset,
			ignoreUnknownUsernames,
			allowDomainDiscovery,
			disableLoginWithEmail,
			disableLoginWithPhone,
			passwordlessType,
			defaultRedirectURI,
			passwordCheckLifetime,
			externalLoginCheckLifetime,
			mfaInitSkipLifetime,
			secondFactorCheckLifetime,
			multiFactorCheckLifetime,
		),
	}
}

func LoginPolicyAddedEventMapper(event *repository.Event) (eventstore.Event, error) {
	e, err := policy.LoginPolicyAddedEventMapper(event)
	if err != nil {
		return nil, err
	}

	return &LoginPolicyAddedEvent{LoginPolicyAddedEvent: *e.(*policy.LoginPolicyAddedEvent)}, nil
}

type LoginPolicyChangedEvent struct {
	policy.LoginPolicyChangedEvent
}

func NewLoginPolicyChangedEvent(
	ctx context.Context,
	aggregate *eventstore.Aggregate,
	changes []policy.LoginPolicyChanges,
) (*LoginPolicyChangedEvent, error) {
	changedEvent, err := policy.NewLoginPolicyChangedEvent(
		eventstore.NewBaseEventForPush(
			ctx,
			aggregate,
			LoginPolicyChangedEventType),
		changes,
	)
	if err != nil {
		return nil, err
	}
	return &LoginPolicyChangedEvent{LoginPolicyChangedEvent: *changedEvent}, nil
}

func LoginPolicyChangedEventMapper(event *repository.Event) (eventstore.Event, error) {
	e, err := policy.LoginPolicyChangedEventMapper(event)
	if err != nil {
		return nil, err
	}

	return &LoginPolicyChangedEvent{LoginPolicyChangedEvent: *e.(*policy.LoginPolicyChangedEvent)}, nil
}

type LoginPolicyRemovedEvent struct {
	policy.LoginPolicyRemovedEvent
}

func NewLoginPolicyRemovedEvent(
	ctx context.Context,
	aggregate *eventstore.Aggregate,
) *LoginPolicyRemovedEvent {
	return &LoginPolicyRemovedEvent{
		LoginPolicyRemovedEvent: *policy.NewLoginPolicyRemovedEvent(
			eventstore.NewBaseEventForPush(
				ctx,
				aggregate,
				LoginPolicyRemovedEventType),
		),
	}
}

func LoginPolicyRemovedEventMapper(event *repository.Event) (eventstore.Event, error) {
	e, err := policy.LoginPolicyRemovedEventMapper(event)
	if err != nil {
		return nil, err
	}

	return &LoginPolicyRemovedEvent{LoginPolicyRemovedEvent: *e.(*policy.LoginPolicyRemovedEvent)}, nil
}
