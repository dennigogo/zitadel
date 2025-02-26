package model

import (
	"encoding/json"

	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	es_models "github.com/dennigogo/zitadel/internal/eventstore/v1/models"
	iam_es_model "github.com/dennigogo/zitadel/internal/iam/repository/eventsourcing/model"
	org_model "github.com/dennigogo/zitadel/internal/org/model"
	"github.com/dennigogo/zitadel/internal/repository/org"
)

type Org struct {
	es_models.ObjectRoot `json:"-"`

	Name  string `json:"name,omitempty"`
	State int32  `json:"-"`

	Domains      []*OrgDomain               `json:"-"`
	DomainPolicy *iam_es_model.DomainPolicy `json:"-"`
}

func OrgToModel(org *Org) *org_model.Org {
	converted := &org_model.Org{
		ObjectRoot: org.ObjectRoot,
		Name:       org.Name,
		State:      org_model.OrgState(org.State),
		Domains:    OrgDomainsToModel(org.Domains),
	}
	if org.DomainPolicy != nil {
		converted.DomainPolicy = iam_es_model.DomainPolicyToModel(org.DomainPolicy)
	}
	return converted
}

func OrgFromEvents(org *Org, events ...*es_models.Event) (*Org, error) {
	if org == nil {
		org = new(Org)
	}

	return org, org.AppendEvents(events...)
}

func (o *Org) AppendEvents(events ...*es_models.Event) error {
	for _, event := range events {
		err := o.AppendEvent(event)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Org) AppendEvent(event *es_models.Event) (err error) {
	switch eventstore.EventType(event.Type) {
	case org.OrgAddedEventType:
		err = o.SetData(event)
		if err != nil {
			return err
		}
	case org.OrgChangedEventType:
		err = o.SetData(event)
		if err != nil {
			return err
		}
	case org.OrgDeactivatedEventType:
		o.State = int32(org_model.OrgStateInactive)
	case org.OrgReactivatedEventType:
		o.State = int32(org_model.OrgStateActive)
	case org.OrgDomainAddedEventType:
		err = o.appendAddDomainEvent(event)
	case org.OrgDomainVerificationAddedEventType:
		err = o.appendVerificationDomainEvent(event)
	case org.OrgDomainVerifiedEventType:
		err = o.appendVerifyDomainEvent(event)
	case org.OrgDomainPrimarySetEventType:
		err = o.appendPrimaryDomainEvent(event)
	case org.OrgDomainRemovedEventType:
		err = o.appendRemoveDomainEvent(event)
	case org.DomainPolicyAddedEventType:
		err = o.appendAddDomainPolicyEvent(event)
	case org.DomainPolicyChangedEventType:
		err = o.appendChangeDomainPolicyEvent(event)
	case org.DomainPolicyRemovedEventType:
		o.appendRemoveDomainPolicyEvent()
	}
	if err != nil {
		return err
	}
	o.ObjectRoot.AppendEvent(event)
	return nil
}

func (o *Org) SetData(event *es_models.Event) error {
	err := json.Unmarshal(event.Data, o)
	if err != nil {
		return errors.ThrowInternal(err, "EVENT-BpbQZ", "unable to unmarshal event")
	}
	return nil
}

func (o *Org) Changes(changed *Org) map[string]interface{} {
	changes := make(map[string]interface{}, 2)

	if changed.Name != "" && changed.Name != o.Name {
		changes["name"] = changed.Name
	}

	return changes
}
