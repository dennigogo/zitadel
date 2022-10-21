package command

import (
	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/repository/policy"
)

type IdentityProviderWriteModel struct {
	eventstore.WriteModel

	IDPConfigID     string
	IDPProviderType domain.IdentityProviderType
	State           domain.IdentityProviderState
}

func (wm *IdentityProviderWriteModel) Reduce() error {
	for _, event := range wm.Events {
		switch e := event.(type) {
		case *policy.IdentityProviderAddedEvent:
			wm.IDPConfigID = e.IDPConfigID
			wm.IDPProviderType = e.IDPProviderType
			wm.State = domain.IdentityProviderStateActive
		case *policy.IdentityProviderRemovedEvent:
			wm.State = domain.IdentityProviderStateRemoved
		case *policy.LoginPolicyRemovedEvent:
			wm.State = domain.IdentityProviderStateRemoved
		}
	}
	return wm.WriteModel.Reduce()
}
