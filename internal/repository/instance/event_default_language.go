package instance

import (
	"context"
	"encoding/json"

	"golang.org/x/text/language"

	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
)

const (
	DefaultLanguageSetEventType eventstore.EventType = "instance.default.language.set"
)

type DefaultLanguageSetEvent struct {
	eventstore.BaseEvent `json:"-"`

	Language language.Tag `json:"language"`
}

func (e *DefaultLanguageSetEvent) Data() interface{} {
	return e
}

func (e *DefaultLanguageSetEvent) UniqueConstraints() []*eventstore.EventUniqueConstraint {
	return nil
}

func NewDefaultLanguageSetEvent(
	ctx context.Context,
	aggregate *eventstore.Aggregate,
	language language.Tag,
) *DefaultLanguageSetEvent {
	return &DefaultLanguageSetEvent{
		BaseEvent: *eventstore.NewBaseEventForPush(
			ctx,
			aggregate,
			DefaultLanguageSetEventType,
		),
		Language: language,
	}
}

func DefaultLanguageSetMapper(event *repository.Event) (eventstore.Event, error) {
	e := &DefaultLanguageSetEvent{
		BaseEvent: *eventstore.BaseEventFromRepo(event),
	}
	err := json.Unmarshal(event.Data, e)
	if err != nil {
		return nil, errors.ThrowInternal(err, "IAM-3j9fs", "unable to unmarshal default language set")
	}

	return e, nil
}
