package project

import (
	"context"
	"encoding/json"

	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
)

const (
	UniqueEntityIDType    = "entity_ids"
	SAMLConfigAddedType   = applicationEventTypePrefix + "config.saml.added"
	SAMLConfigChangedType = applicationEventTypePrefix + "config.saml.changed"
)

type SAMLConfigAddedEvent struct {
	eventstore.BaseEvent `json:"-"`

	AppID       string `json:"appId"`
	EntityID    string `json:"entityId"`
	Metadata    []byte `json:"metadata,omitempty"`
	MetadataURL string `json:"metadata_url,omitempty"`
}

func (e *SAMLConfigAddedEvent) Data() interface{} {
	return e
}

func NewAddSAMLConfigEntityIDUniqueConstraint(entityID string) *eventstore.EventUniqueConstraint {
	return eventstore.NewAddEventUniqueConstraint(
		UniqueEntityIDType,
		entityID,
		"Errors.Project.App.SAMLEntityIDAlreadyExists")
}

func NewRemoveSAMLConfigEntityIDUniqueConstraint(entityID string) *eventstore.EventUniqueConstraint {
	return eventstore.NewRemoveEventUniqueConstraint(
		UniqueEntityIDType,
		entityID)
}

func (e *SAMLConfigAddedEvent) UniqueConstraints() []*eventstore.EventUniqueConstraint {
	return []*eventstore.EventUniqueConstraint{NewAddSAMLConfigEntityIDUniqueConstraint(e.EntityID)}
}

func NewSAMLConfigAddedEvent(
	ctx context.Context,
	aggregate *eventstore.Aggregate,
	appID string,
	entityID string,
	metadata []byte,
	metadataURL string,
) *SAMLConfigAddedEvent {
	return &SAMLConfigAddedEvent{
		BaseEvent: *eventstore.NewBaseEventForPush(
			ctx,
			aggregate,
			SAMLConfigAddedType,
		),
		AppID:       appID,
		EntityID:    entityID,
		Metadata:    metadata,
		MetadataURL: metadataURL,
	}
}

func SAMLConfigAddedEventMapper(event *repository.Event) (eventstore.Event, error) {
	e := &SAMLConfigAddedEvent{
		BaseEvent: *eventstore.BaseEventFromRepo(event),
	}

	err := json.Unmarshal(event.Data, e)
	if err != nil {
		return nil, errors.ThrowInternal(err, "SAML-BDd15", "unable to unmarshal saml config")
	}

	return e, nil
}

type SAMLConfigChangedEvent struct {
	eventstore.BaseEvent `json:"-"`

	AppID       string  `json:"appId"`
	EntityID    string  `json:"entityId"`
	Metadata    []byte  `json:"metadata,omitempty"`
	MetadataURL *string `json:"metadata_url,omitempty"`
	oldEntityID string
}

func (e *SAMLConfigChangedEvent) Data() interface{} {
	return e
}

func (e *SAMLConfigChangedEvent) UniqueConstraints() []*eventstore.EventUniqueConstraint {
	if e.EntityID != "" {
		return []*eventstore.EventUniqueConstraint{
			NewRemoveSAMLConfigEntityIDUniqueConstraint(e.oldEntityID),
			NewAddSAMLConfigEntityIDUniqueConstraint(e.EntityID),
		}
	}
	return nil
}

func NewSAMLConfigChangedEvent(
	ctx context.Context,
	aggregate *eventstore.Aggregate,
	appID string,
	oldEntityID string,
	changes []SAMLConfigChanges,
) (*SAMLConfigChangedEvent, error) {
	if len(changes) == 0 {
		return nil, errors.ThrowPreconditionFailed(nil, "SAML-i8idç", "Errors.NoChangesFound")
	}

	changeEvent := &SAMLConfigChangedEvent{
		BaseEvent: *eventstore.NewBaseEventForPush(
			ctx,
			aggregate,
			SAMLConfigChangedType,
		),
		AppID:       appID,
		oldEntityID: oldEntityID,
	}
	for _, change := range changes {
		change(changeEvent)
	}
	return changeEvent, nil
}

type SAMLConfigChanges func(event *SAMLConfigChangedEvent)

func ChangeMetadata(metadata []byte) func(event *SAMLConfigChangedEvent) {
	return func(e *SAMLConfigChangedEvent) {
		e.Metadata = metadata
	}
}

func ChangeMetadataURL(metadataURL string) func(event *SAMLConfigChangedEvent) {
	return func(e *SAMLConfigChangedEvent) {
		e.MetadataURL = &metadataURL
	}
}

func ChangeEntityID(entityID string) func(event *SAMLConfigChangedEvent) {
	return func(e *SAMLConfigChangedEvent) {
		e.EntityID = entityID
	}
}

func SAMLConfigChangedEventMapper(event *repository.Event) (eventstore.Event, error) {
	e := &SAMLConfigChangedEvent{
		BaseEvent: *eventstore.BaseEventFromRepo(event),
	}

	err := json.Unmarshal(event.Data, e)
	if err != nil {
		return nil, errors.ThrowInternal(err, "SAML-BFd15", "unable to unmarshal saml config")
	}

	return e, nil
}
