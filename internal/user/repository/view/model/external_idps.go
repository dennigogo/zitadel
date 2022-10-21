package model

import (
	"encoding/json"
	"time"

	"github.com/zitadel/logging"

	caos_errs "github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/v1/models"
	user_repo "github.com/dennigogo/zitadel/internal/repository/user"
)

const (
	ExternalIDPKeyExternalUserID = "external_user_id"
	ExternalIDPKeyUserID         = "user_id"
	ExternalIDPKeyIDPConfigID    = "idp_config_id"
	ExternalIDPKeyResourceOwner  = "resource_owner"
	ExternalIDPKeyInstanceID     = "instance_id"
)

type ExternalIDPView struct {
	ExternalUserID  string    `json:"userID" gorm:"column:external_user_id;primary_key"`
	IDPConfigID     string    `json:"idpConfigID" gorm:"column:idp_config_id;primary_key"`
	UserID          string    `json:"-" gorm:"column:user_id"`
	IDPName         string    `json:"-" gorm:"column:idp_name"`
	UserDisplayName string    `json:"displayName" gorm:"column:user_display_name"`
	CreationDate    time.Time `json:"-" gorm:"column:creation_date"`
	ChangeDate      time.Time `json:"-" gorm:"column:change_date"`
	ResourceOwner   string    `json:"-" gorm:"column:resource_owner"`
	Sequence        uint64    `json:"-" gorm:"column:sequence"`
	InstanceID      string    `json:"instanceID" gorm:"column:instance_id;primary_key"`
}

func (i *ExternalIDPView) AppendEvent(event *models.Event) (err error) {
	i.Sequence = event.Sequence
	i.ChangeDate = event.CreationDate
	switch eventstore.EventType(event.Type) {
	case user_repo.UserIDPLinkAddedType:
		i.setRootData(event)
		i.CreationDate = event.CreationDate
		err = i.SetData(event)
	}
	return err
}

func (r *ExternalIDPView) setRootData(event *models.Event) {
	r.UserID = event.AggregateID
	r.ResourceOwner = event.ResourceOwner
	r.InstanceID = event.InstanceID
}

func (r *ExternalIDPView) SetData(event *models.Event) error {
	if err := json.Unmarshal(event.Data, r); err != nil {
		logging.Log("EVEN-48sfs").WithError(err).Error("could not unmarshal event data")
		return caos_errs.ThrowInternal(err, "MODEL-Hs8uf", "Could not unmarshal data")
	}
	return nil
}
