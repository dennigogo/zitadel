package policy

import (
	"encoding/json"

	"github.com/dennigogo/zitadel/internal/eventstore"

	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
)

const (
	DomainPolicyAddedEventType   = "policy.domain.added"
	DomainPolicyChangedEventType = "policy.domain.changed"
	DomainPolicyRemovedEventType = "policy.domain.removed"
)

type DomainPolicyAddedEvent struct {
	eventstore.BaseEvent `json:"-"`

	UserLoginMustBeDomain                  bool `json:"userLoginMustBeDomain,omitempty"`
	ValidateOrgDomains                     bool `json:"validateOrgDomains,omitempty"`
	SMTPSenderAddressMatchesInstanceDomain bool `json:"smtpSenderAddressMatchesInstanceDomain,omitempty"`
}

func (e *DomainPolicyAddedEvent) Data() interface{} {
	return e
}

func (e *DomainPolicyAddedEvent) UniqueConstraints() []*eventstore.EventUniqueConstraint {
	return nil
}

func NewDomainPolicyAddedEvent(
	base *eventstore.BaseEvent,
	userLoginMustBeDomain,
	validateOrgDomains,
	smtpSenderAddressMatchesInstanceDomain bool,
) *DomainPolicyAddedEvent {

	return &DomainPolicyAddedEvent{
		BaseEvent:                              *base,
		UserLoginMustBeDomain:                  userLoginMustBeDomain,
		ValidateOrgDomains:                     validateOrgDomains,
		SMTPSenderAddressMatchesInstanceDomain: smtpSenderAddressMatchesInstanceDomain,
	}
}

func DomainPolicyAddedEventMapper(event *repository.Event) (eventstore.Event, error) {
	e := &DomainPolicyAddedEvent{
		BaseEvent: *eventstore.BaseEventFromRepo(event),
	}

	err := json.Unmarshal(event.Data, e)
	if err != nil {
		return nil, errors.ThrowInternal(err, "POLIC-TvSmA", "unable to unmarshal policy")
	}

	return e, nil
}

type DomainPolicyChangedEvent struct {
	eventstore.BaseEvent `json:"-"`

	UserLoginMustBeDomain                  *bool `json:"userLoginMustBeDomain,omitempty"`
	ValidateOrgDomains                     *bool `json:"validateOrgDomains,omitempty"`
	SMTPSenderAddressMatchesInstanceDomain *bool `json:"smtpSenderAddressMatchesInstanceDomain,omitempty"`
}

func (e *DomainPolicyChangedEvent) Data() interface{} {
	return e
}

func (e *DomainPolicyChangedEvent) UniqueConstraints() []*eventstore.EventUniqueConstraint {
	return nil
}

func NewDomainPolicyChangedEvent(
	base *eventstore.BaseEvent,
	changes []DomainPolicyChanges,
) (*DomainPolicyChangedEvent, error) {
	if len(changes) == 0 {
		return nil, errors.ThrowPreconditionFailed(nil, "POLICY-DAf3h", "Errors.NoChangesFound")
	}
	changeEvent := &DomainPolicyChangedEvent{
		BaseEvent: *base,
	}
	for _, change := range changes {
		change(changeEvent)
	}
	return changeEvent, nil
}

type DomainPolicyChanges func(*DomainPolicyChangedEvent)

func ChangeUserLoginMustBeDomain(userLoginMustBeDomain bool) func(*DomainPolicyChangedEvent) {
	return func(e *DomainPolicyChangedEvent) {
		e.UserLoginMustBeDomain = &userLoginMustBeDomain
	}
}

func ChangeValidateOrgDomains(validateOrgDomain bool) func(*DomainPolicyChangedEvent) {
	return func(e *DomainPolicyChangedEvent) {
		e.ValidateOrgDomains = &validateOrgDomain
	}
}

func ChangeSMTPSenderAddressMatchesInstanceDomain(smtpSenderAddressMatchesInstanceDomain bool) func(*DomainPolicyChangedEvent) {
	return func(e *DomainPolicyChangedEvent) {
		e.SMTPSenderAddressMatchesInstanceDomain = &smtpSenderAddressMatchesInstanceDomain
	}
}

func DomainPolicyChangedEventMapper(event *repository.Event) (eventstore.Event, error) {
	e := &DomainPolicyChangedEvent{
		BaseEvent: *eventstore.BaseEventFromRepo(event),
	}

	err := json.Unmarshal(event.Data, e)
	if err != nil {
		return nil, errors.ThrowInternal(err, "POLIC-0Pl9d", "unable to unmarshal policy")
	}

	return e, nil
}

type DomainPolicyRemovedEvent struct {
	eventstore.BaseEvent `json:"-"`
}

func (e *DomainPolicyRemovedEvent) Data() interface{} {
	return nil
}

func (e *DomainPolicyRemovedEvent) UniqueConstraints() []*eventstore.EventUniqueConstraint {
	return nil
}

func NewDomainPolicyRemovedEvent(base *eventstore.BaseEvent) *DomainPolicyRemovedEvent {
	return &DomainPolicyRemovedEvent{
		BaseEvent: *base,
	}
}

func DomainPolicyRemovedEventMapper(event *repository.Event) (eventstore.Event, error) {
	return &DomainPolicyRemovedEvent{
		BaseEvent: *eventstore.BaseEventFromRepo(event),
	}, nil
}
