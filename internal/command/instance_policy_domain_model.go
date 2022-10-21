package command

import (
	"context"

	"github.com/dennigogo/zitadel/internal/api/authz"
	"github.com/dennigogo/zitadel/internal/eventstore"

	"github.com/dennigogo/zitadel/internal/repository/instance"
	"github.com/dennigogo/zitadel/internal/repository/policy"
)

type InstanceDomainPolicyWriteModel struct {
	PolicyDomainWriteModel
}

func NewInstanceDomainPolicyWriteModel(ctx context.Context) *InstanceDomainPolicyWriteModel {
	return &InstanceDomainPolicyWriteModel{
		PolicyDomainWriteModel{
			WriteModel: eventstore.WriteModel{
				AggregateID:   authz.GetInstance(ctx).InstanceID(),
				ResourceOwner: authz.GetInstance(ctx).InstanceID(),
			},
		},
	}
}

func (wm *InstanceDomainPolicyWriteModel) AppendEvents(events ...eventstore.Event) {
	for _, event := range events {
		switch e := event.(type) {
		case *instance.DomainPolicyAddedEvent:
			wm.PolicyDomainWriteModel.AppendEvents(&e.DomainPolicyAddedEvent)
		case *instance.DomainPolicyChangedEvent:
			wm.PolicyDomainWriteModel.AppendEvents(&e.DomainPolicyChangedEvent)
		}
	}
}

func (wm *InstanceDomainPolicyWriteModel) Reduce() error {
	return wm.PolicyDomainWriteModel.Reduce()
}

func (wm *InstanceDomainPolicyWriteModel) Query() *eventstore.SearchQueryBuilder {
	return eventstore.NewSearchQueryBuilder(eventstore.ColumnsEvent).
		ResourceOwner(wm.ResourceOwner).
		AddQuery().
		AggregateTypes(instance.AggregateType).
		AggregateIDs(wm.PolicyDomainWriteModel.AggregateID).
		EventTypes(
			instance.DomainPolicyAddedEventType,
			instance.DomainPolicyChangedEventType).
		Builder()
}

func (wm *InstanceDomainPolicyWriteModel) NewChangedEvent(
	ctx context.Context,
	aggregate *eventstore.Aggregate,
	userLoginMustBeDomain,
	validateOrgDomain,
	smtpSenderAddresssMatchesInstanceDomain bool) (*instance.DomainPolicyChangedEvent, bool) {
	changes := make([]policy.DomainPolicyChanges, 0)
	if wm.UserLoginMustBeDomain != userLoginMustBeDomain {
		changes = append(changes, policy.ChangeUserLoginMustBeDomain(userLoginMustBeDomain))
	}
	if wm.ValidateOrgDomains != validateOrgDomain {
		changes = append(changes, policy.ChangeValidateOrgDomains(validateOrgDomain))
	}
	if wm.SMTPSenderAddressMatchesInstanceDomain != smtpSenderAddresssMatchesInstanceDomain {
		changes = append(changes, policy.ChangeSMTPSenderAddressMatchesInstanceDomain(smtpSenderAddresssMatchesInstanceDomain))
	}
	if len(changes) == 0 {
		return nil, false
	}
	changedEvent, err := instance.NewDomainPolicyChangedEvent(ctx, aggregate, changes)
	if err != nil {
		return nil, false
	}
	return changedEvent, true
}
