package handler

import (
	"github.com/zitadel/logging"

	"github.com/dennigogo/zitadel/internal/eventstore"
	v1 "github.com/dennigogo/zitadel/internal/eventstore/v1"
	"github.com/dennigogo/zitadel/internal/eventstore/v1/models"
	"github.com/dennigogo/zitadel/internal/eventstore/v1/query"
	"github.com/dennigogo/zitadel/internal/eventstore/v1/spooler"
	iam_model "github.com/dennigogo/zitadel/internal/iam/model"
	iam_view_model "github.com/dennigogo/zitadel/internal/iam/repository/view/model"
	"github.com/dennigogo/zitadel/internal/repository/instance"
	"github.com/dennigogo/zitadel/internal/repository/org"
)

const (
	idpConfigTable = "auth.idp_configs"
)

type IDPConfig struct {
	handler
	subscription *v1.Subscription
}

func newIDPConfig(h handler) *IDPConfig {
	idpConfig := &IDPConfig{
		handler: h,
	}

	idpConfig.subscribe()

	return idpConfig
}

func (i *IDPConfig) subscribe() {
	i.subscription = i.es.Subscribe(i.AggregateTypes()...)
	go func() {
		for event := range i.subscription.Events {
			query.ReduceEvent(i, event)
		}
	}()
}

func (i *IDPConfig) ViewModel() string {
	return idpConfigTable
}

func (i *IDPConfig) Subscription() *v1.Subscription {
	return i.subscription
}

func (_ *IDPConfig) AggregateTypes() []models.AggregateType {
	return []models.AggregateType{org.AggregateType, instance.AggregateType}
}

func (i *IDPConfig) CurrentSequence(instanceID string) (uint64, error) {
	sequence, err := i.view.GetLatestIDPConfigSequence(instanceID)
	if err != nil {
		return 0, err
	}
	return sequence.CurrentSequence, nil
}

func (i *IDPConfig) EventQuery(instanceIDs ...string) (*models.SearchQuery, error) {
	sequences, err := i.view.GetLatestIDPConfigSequences(instanceIDs...)
	if err != nil {
		return nil, err
	}
	return newSearchQuery(sequences, i.AggregateTypes(), instanceIDs), nil
}

func (i *IDPConfig) Reduce(event *models.Event) (err error) {
	switch event.AggregateType {
	case org.AggregateType:
		err = i.processIdpConfig(iam_model.IDPProviderTypeOrg, event)
	case instance.AggregateType:
		err = i.processIdpConfig(iam_model.IDPProviderTypeSystem, event)
	}
	return err
}

func (i *IDPConfig) processIdpConfig(providerType iam_model.IDPProviderType, event *models.Event) (err error) {
	idp := new(iam_view_model.IDPConfigView)
	switch eventstore.EventType(event.Type) {
	case org.IDPConfigAddedEventType,
		instance.IDPConfigAddedEventType:
		err = idp.AppendEvent(providerType, event)
	case org.IDPConfigChangedEventType, instance.IDPConfigChangedEventType,
		org.IDPOIDCConfigAddedEventType, instance.IDPOIDCConfigAddedEventType,
		org.IDPOIDCConfigChangedEventType, instance.IDPOIDCConfigChangedEventType,
		org.IDPJWTConfigAddedEventType, instance.IDPJWTConfigAddedEventType,
		org.IDPJWTConfigChangedEventType, instance.IDPJWTConfigChangedEventType:
		err = idp.SetData(event)
		if err != nil {
			return err
		}
		idp, err = i.view.IDPConfigByID(idp.IDPConfigID, event.InstanceID)
		if err != nil {
			return err
		}
		err = idp.AppendEvent(providerType, event)
	case org.IDPConfigDeactivatedEventType, instance.IDPConfigDeactivatedEventType,
		org.IDPConfigReactivatedEventType, instance.IDPConfigReactivatedEventType:
		err = idp.SetData(event)
		if err != nil {
			return err
		}
		idp, err = i.view.IDPConfigByID(idp.IDPConfigID, event.InstanceID)
		if err != nil {
			return err
		}
		err = idp.AppendEvent(providerType, event)
	case org.IDPConfigRemovedEventType, instance.IDPConfigRemovedEventType:
		err = idp.SetData(event)
		if err != nil {
			return err
		}
		return i.view.DeleteIDPConfig(idp.IDPConfigID, event)
	default:
		return i.view.ProcessedIDPConfigSequence(event)
	}
	if err != nil {
		return err
	}
	return i.view.PutIDPConfig(idp, event)
}

func (i *IDPConfig) OnError(event *models.Event, err error) error {
	logging.LogWithFields("SPOOL-Ejf8s", "id", event.AggregateID).WithError(err).Warn("something went wrong in idp config handler")
	return spooler.HandleError(event, err, i.view.GetLatestIDPConfigFailedEvent, i.view.ProcessedIDPConfigFailedEvent, i.view.ProcessedIDPConfigSequence, i.errorCountUntilSkip)
}

func (i *IDPConfig) OnSuccess() error {
	return spooler.HandleSuccess(i.view.UpdateIDPConfigSpoolerRunTimestamp)
}
