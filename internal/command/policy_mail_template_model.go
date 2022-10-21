package command

import (
	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/repository/policy"
)

type MailTemplateWriteModel struct {
	eventstore.WriteModel

	Template []byte

	State domain.PolicyState
}

func (wm *MailTemplateWriteModel) Reduce() error {
	for _, event := range wm.Events {
		switch e := event.(type) {
		case *policy.MailTemplateAddedEvent:
			wm.Template = e.Template
			wm.State = domain.PolicyStateActive
		case *policy.MailTemplateChangedEvent:
			if e.Template != nil {
				wm.Template = *e.Template
			}
		case *policy.MailTemplateRemovedEvent:
			wm.State = domain.PolicyStateRemoved
		}
	}
	return wm.WriteModel.Reduce()
}
