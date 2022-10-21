package handler

import (
	"context"

	"github.com/zitadel/logging"

	"github.com/dennigogo/zitadel/internal/config/systemdefaults"
	caos_errs "github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	v1 "github.com/dennigogo/zitadel/internal/eventstore/v1"
	es_models "github.com/dennigogo/zitadel/internal/eventstore/v1/models"
	"github.com/dennigogo/zitadel/internal/eventstore/v1/query"
	"github.com/dennigogo/zitadel/internal/eventstore/v1/spooler"
	iam_model "github.com/dennigogo/zitadel/internal/iam/model"
	iam_view_model "github.com/dennigogo/zitadel/internal/iam/repository/view/model"
	query2 "github.com/dennigogo/zitadel/internal/query"
	"github.com/dennigogo/zitadel/internal/repository/instance"
	"github.com/dennigogo/zitadel/internal/repository/org"
	"github.com/dennigogo/zitadel/internal/repository/user"
	usr_view_model "github.com/dennigogo/zitadel/internal/user/repository/view/model"
)

const (
	externalIDPTable = "auth.user_external_idps"
)

type ExternalIDP struct {
	handler
	systemDefaults systemdefaults.SystemDefaults
	subscription   *v1.Subscription
	queries        *query2.Queries
}

func newExternalIDP(
	handler handler,
	defaults systemdefaults.SystemDefaults,
	queries *query2.Queries,
) *ExternalIDP {
	h := &ExternalIDP{
		handler:        handler,
		systemDefaults: defaults,
		queries:        queries,
	}

	h.subscribe()

	return h
}

func (i *ExternalIDP) subscribe() {
	i.subscription = i.es.Subscribe(i.AggregateTypes()...)
	go func() {
		for event := range i.subscription.Events {
			query.ReduceEvent(i, event)
		}
	}()
}

func (i *ExternalIDP) ViewModel() string {
	return externalIDPTable
}

func (i *ExternalIDP) Subscription() *v1.Subscription {
	return i.subscription
}

func (_ *ExternalIDP) AggregateTypes() []es_models.AggregateType {
	return []es_models.AggregateType{user.AggregateType, instance.AggregateType, org.AggregateType}
}

func (i *ExternalIDP) CurrentSequence(instanceID string) (uint64, error) {
	sequence, err := i.view.GetLatestExternalIDPSequence(instanceID)
	if err != nil {
		return 0, err
	}
	return sequence.CurrentSequence, nil
}

func (i *ExternalIDP) EventQuery(instanceIDs ...string) (*es_models.SearchQuery, error) {
	sequences, err := i.view.GetLatestExternalIDPSequences(instanceIDs...)
	if err != nil {
		return nil, err
	}
	return newSearchQuery(sequences, i.AggregateTypes(), instanceIDs), nil
}

func (i *ExternalIDP) Reduce(event *es_models.Event) (err error) {
	switch event.AggregateType {
	case user.AggregateType:
		err = i.processUser(event)
	case instance.AggregateType, org.AggregateType:
		err = i.processIdpConfig(event)
	}
	return err
}

func (i *ExternalIDP) processUser(event *es_models.Event) (err error) {
	externalIDP := new(usr_view_model.ExternalIDPView)
	switch eventstore.EventType(event.Type) {
	case user.UserIDPLinkAddedType:
		err = externalIDP.AppendEvent(event)
		if err != nil {
			return err
		}
		err = i.fillData(externalIDP)
	case user.UserIDPLinkRemovedType, user.UserIDPLinkCascadeRemovedType:
		err = externalIDP.SetData(event)
		if err != nil {
			return err
		}
		return i.view.DeleteExternalIDP(externalIDP.ExternalUserID, externalIDP.IDPConfigID, event.InstanceID, event)
	case user.UserRemovedType:
		return i.view.DeleteExternalIDPsByUserID(event.AggregateID, event.InstanceID, event)
	default:
		return i.view.ProcessedExternalIDPSequence(event)
	}
	if err != nil {
		return err
	}
	return i.view.PutExternalIDP(externalIDP, event)
}

func (i *ExternalIDP) processIdpConfig(event *es_models.Event) (err error) {
	switch eventstore.EventType(event.Type) {
	case instance.IDPConfigChangedEventType, org.IDPConfigChangedEventType:
		configView := new(iam_view_model.IDPConfigView)
		var config *query2.IDP
		if eventstore.EventType(event.Type) == instance.IDPConfigChangedEventType {
			err = configView.AppendEvent(iam_model.IDPProviderTypeSystem, event)
		} else {
			err = configView.AppendEvent(iam_model.IDPProviderTypeOrg, event)
		}
		if err != nil {
			return err
		}
		exterinalIDPs, err := i.view.ExternalIDPsByIDPConfigID(configView.IDPConfigID, event.InstanceID)
		if err != nil {
			return err
		}
		if event.AggregateType == instance.AggregateType {
			config, err = i.getDefaultIDPConfig(event.InstanceID, configView.IDPConfigID)
		} else {
			config, err = i.getOrgIDPConfig(event.InstanceID, event.AggregateID, configView.IDPConfigID)
		}
		if err != nil {
			return err
		}
		for _, provider := range exterinalIDPs {
			i.fillConfigData(provider, config)
		}
		return i.view.PutExternalIDPs(event, exterinalIDPs...)
	default:
		return i.view.ProcessedExternalIDPSequence(event)
	}
}

func (i *ExternalIDP) fillData(externalIDP *usr_view_model.ExternalIDPView) error {
	config, err := i.getOrgIDPConfig(externalIDP.InstanceID, externalIDP.ResourceOwner, externalIDP.IDPConfigID)
	if caos_errs.IsNotFound(err) {
		config, err = i.getDefaultIDPConfig(externalIDP.InstanceID, externalIDP.IDPConfigID)
	}
	if err != nil {
		return err
	}
	i.fillConfigData(externalIDP, config)
	return nil
}

func (i *ExternalIDP) fillConfigData(externalIDP *usr_view_model.ExternalIDPView, config *query2.IDP) {
	externalIDP.IDPName = config.Name
}

func (i *ExternalIDP) OnError(event *es_models.Event, err error) error {
	logging.WithFields("id", event.AggregateID).WithError(err).Warn("something went wrong in idp provider handler")
	return spooler.HandleError(event, err, i.view.GetLatestExternalIDPFailedEvent, i.view.ProcessedExternalIDPFailedEvent, i.view.ProcessedExternalIDPSequence, i.errorCountUntilSkip)
}

func (i *ExternalIDP) OnSuccess() error {
	return spooler.HandleSuccess(i.view.UpdateExternalIDPSpoolerRunTimestamp)
}

func (i *ExternalIDP) getOrgIDPConfig(instanceID, aggregateID, idpConfigID string) (*query2.IDP, error) {
	return i.queries.IDPByIDAndResourceOwner(withInstanceID(context.Background(), instanceID), false, idpConfigID, aggregateID)
}

func (i *ExternalIDP) getDefaultIDPConfig(instanceID, idpConfigID string) (*query2.IDP, error) {
	return i.queries.IDPByIDAndResourceOwner(withInstanceID(context.Background(), instanceID), false, idpConfigID, instanceID)
}
