package command

import (
	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/repository/policy"
)

type PolicyDomainWriteModel struct {
	eventstore.WriteModel

	UserLoginMustBeDomain                  bool
	ValidateOrgDomains                     bool
	SMTPSenderAddressMatchesInstanceDomain bool
	State                                  domain.PolicyState
}

func (wm *PolicyDomainWriteModel) Reduce() error {
	for _, event := range wm.Events {
		switch e := event.(type) {
		case *policy.DomainPolicyAddedEvent:
			wm.UserLoginMustBeDomain = e.UserLoginMustBeDomain
			wm.ValidateOrgDomains = e.ValidateOrgDomains
			wm.SMTPSenderAddressMatchesInstanceDomain = e.SMTPSenderAddressMatchesInstanceDomain
			wm.State = domain.PolicyStateActive
		case *policy.DomainPolicyChangedEvent:
			if e.UserLoginMustBeDomain != nil {
				wm.UserLoginMustBeDomain = *e.UserLoginMustBeDomain
			}
			if e.ValidateOrgDomains != nil {
				wm.ValidateOrgDomains = *e.ValidateOrgDomains
			}
			if e.SMTPSenderAddressMatchesInstanceDomain != nil {
				wm.SMTPSenderAddressMatchesInstanceDomain = *e.SMTPSenderAddressMatchesInstanceDomain
			}
		case *policy.DomainPolicyRemovedEvent:
			wm.State = domain.PolicyStateRemoved
		}
	}
	return wm.WriteModel.Reduce()
}
