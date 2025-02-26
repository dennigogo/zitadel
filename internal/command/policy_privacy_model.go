package command

import (
	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/repository/policy"
)

type PrivacyPolicyWriteModel struct {
	eventstore.WriteModel

	TOSLink     string
	PrivacyLink string
	HelpLink    string
	State       domain.PolicyState
}

func (wm *PrivacyPolicyWriteModel) Reduce() error {
	for _, event := range wm.Events {
		switch e := event.(type) {
		case *policy.PrivacyPolicyAddedEvent:
			wm.TOSLink = e.TOSLink
			wm.PrivacyLink = e.PrivacyLink
			wm.HelpLink = e.HelpLink
			wm.State = domain.PolicyStateActive
		case *policy.PrivacyPolicyChangedEvent:
			if e.PrivacyLink != nil {
				wm.PrivacyLink = *e.PrivacyLink
			}
			if e.TOSLink != nil {
				wm.TOSLink = *e.TOSLink
			}
			if e.HelpLink != nil {
				wm.HelpLink = *e.HelpLink
			}
		case *policy.PrivacyPolicyRemovedEvent:
			wm.State = domain.PolicyStateRemoved
		}
	}
	return wm.WriteModel.Reduce()
}
