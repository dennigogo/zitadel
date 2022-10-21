package command

import (
	"context"

	"github.com/dennigogo/zitadel/internal/api/authz"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/repository/settings"

	"github.com/dennigogo/zitadel/internal/repository/instance"
)

type InstanceDebugNotificationLogWriteModel struct {
	DebugNotificationWriteModel
}

func NewInstanceDebugNotificationLogWriteModel(ctx context.Context) *InstanceDebugNotificationLogWriteModel {
	return &InstanceDebugNotificationLogWriteModel{
		DebugNotificationWriteModel{
			WriteModel: eventstore.WriteModel{
				AggregateID:   authz.GetInstance(ctx).InstanceID(),
				ResourceOwner: authz.GetInstance(ctx).InstanceID(),
			},
		},
	}
}

func (wm *InstanceDebugNotificationLogWriteModel) AppendEvents(events ...eventstore.Event) {
	for _, event := range events {
		switch e := event.(type) {
		case *instance.DebugNotificationProviderLogAddedEvent:
			wm.DebugNotificationWriteModel.AppendEvents(&e.DebugNotificationProviderAddedEvent)
		case *instance.DebugNotificationProviderLogChangedEvent:
			wm.DebugNotificationWriteModel.AppendEvents(&e.DebugNotificationProviderChangedEvent)
		case *instance.DebugNotificationProviderLogRemovedEvent:
			wm.DebugNotificationWriteModel.AppendEvents(&e.DebugNotificationProviderRemovedEvent)
		}
	}
}

func (wm *InstanceDebugNotificationLogWriteModel) IsValid() bool {
	return wm.AggregateID != ""
}

func (wm *InstanceDebugNotificationLogWriteModel) Reduce() error {
	return wm.DebugNotificationWriteModel.Reduce()
}

func (wm *InstanceDebugNotificationLogWriteModel) Query() *eventstore.SearchQueryBuilder {
	return eventstore.NewSearchQueryBuilder(eventstore.ColumnsEvent).
		ResourceOwner(wm.ResourceOwner).
		AddQuery().
		AggregateTypes(instance.AggregateType).
		AggregateIDs(wm.DebugNotificationWriteModel.AggregateID).
		EventTypes(
			instance.DebugNotificationProviderLogAddedEventType,
			instance.DebugNotificationProviderLogChangedEventType,
			instance.DebugNotificationProviderLogEnabledEventType,
			instance.DebugNotificationProviderLogDisabledEventType,
			instance.DebugNotificationProviderLogRemovedEventType).
		Builder()
}

func (wm *InstanceDebugNotificationLogWriteModel) NewChangedEvent(
	ctx context.Context,
	aggregate *eventstore.Aggregate,
	compact bool) (*instance.DebugNotificationProviderLogChangedEvent, bool) {

	changes := make([]settings.DebugNotificationProviderChanges, 0)
	if wm.Compact != compact {
		changes = append(changes, settings.ChangeCompact(compact))
	}
	if len(changes) == 0 {
		return nil, false
	}
	changedEvent, err := instance.NewDebugNotificationProviderLogChangedEvent(ctx, aggregate, changes)
	if err != nil {
		return nil, false
	}
	return changedEvent, true
}
