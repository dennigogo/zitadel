package model

import (
	"encoding/json"
	"time"

	"github.com/zitadel/logging"

	caos_errs "github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/v1/models"
	"github.com/dennigogo/zitadel/internal/iam/model"
	"github.com/dennigogo/zitadel/internal/query"
	"github.com/dennigogo/zitadel/internal/repository/instance"
	"github.com/dennigogo/zitadel/internal/repository/org"
)

const (
	PasswordComplexityKeyAggregateID = "aggregate_id"
)

type PasswordComplexityPolicyView struct {
	AggregateID  string    `json:"-" gorm:"column:aggregate_id;primary_key"`
	CreationDate time.Time `json:"-" gorm:"column:creation_date"`
	ChangeDate   time.Time `json:"-" gorm:"column:change_date"`
	State        int32     `json:"-" gorm:"column:complexity_policy_state"`

	MinLength    uint64 `json:"minLength" gorm:"column:min_length"`
	HasLowercase bool   `json:"hasLowercase" gorm:"column:has_lowercase"`
	HasUppercase bool   `json:"hasUppercase" gorm:"column:has_uppercase"`
	HasSymbol    bool   `json:"hasSymbol" gorm:"column:has_symbol"`
	HasNumber    bool   `json:"hasNumber" gorm:"column:has_number"`
	Default      bool   `json:"-" gorm:"-"`

	Sequence uint64 `json:"-" gorm:"column:sequence"`
}

func PasswordComplexityViewToModel(policy *query.PasswordComplexityPolicy) *model.PasswordComplexityPolicyView {
	return &model.PasswordComplexityPolicyView{
		AggregateID:  policy.ID,
		Sequence:     policy.Sequence,
		CreationDate: policy.CreationDate,
		ChangeDate:   policy.ChangeDate,
		MinLength:    policy.MinLength,
		HasLowercase: policy.HasLowercase,
		HasUppercase: policy.HasUppercase,
		HasSymbol:    policy.HasSymbol,
		HasNumber:    policy.HasNumber,
		Default:      policy.IsDefault,
	}
}

func (i *PasswordComplexityPolicyView) AppendEvent(event *models.Event) (err error) {
	i.Sequence = event.Sequence
	i.ChangeDate = event.CreationDate
	switch eventstore.EventType(event.Type) {
	case instance.PasswordComplexityPolicyAddedEventType,
		org.PasswordComplexityPolicyAddedEventType:
		i.setRootData(event)
		i.CreationDate = event.CreationDate
		err = i.SetData(event)
	case instance.PasswordComplexityPolicyChangedEventType,
		org.PasswordComplexityPolicyChangedEventType:
		err = i.SetData(event)
	}
	return err
}

func (r *PasswordComplexityPolicyView) setRootData(event *models.Event) {
	r.AggregateID = event.AggregateID
}

func (r *PasswordComplexityPolicyView) SetData(event *models.Event) error {
	if err := json.Unmarshal(event.Data, r); err != nil {
		logging.Log("EVEN-Dmi9g").WithError(err).Error("could not unmarshal event data")
		return caos_errs.ThrowInternal(err, "MODEL-Hs8uf", "Could not unmarshal data")
	}
	return nil
}
