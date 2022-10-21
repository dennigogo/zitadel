package model

import (
	"encoding/json"

	"github.com/dennigogo/zitadel/internal/errors"
	es_models "github.com/dennigogo/zitadel/internal/eventstore/v1/models"
	iam_model "github.com/dennigogo/zitadel/internal/iam/model"
)

type DomainPolicy struct {
	es_models.ObjectRoot

	State                 int32 `json:"-"`
	UserLoginMustBeDomain bool  `json:"userLoginMustBeDomain"`
}

func DomainPolicyToModel(policy *DomainPolicy) *iam_model.DomainPolicy {
	return &iam_model.DomainPolicy{
		ObjectRoot:            policy.ObjectRoot,
		State:                 iam_model.PolicyState(policy.State),
		UserLoginMustBeDomain: policy.UserLoginMustBeDomain,
	}
}

func (p *DomainPolicy) Changes(changed *DomainPolicy) map[string]interface{} {
	changes := make(map[string]interface{}, 1)

	if p.UserLoginMustBeDomain != changed.UserLoginMustBeDomain {
		changes["userLoginMustBeDomain"] = changed.UserLoginMustBeDomain
	}
	return changes
}

func (p *DomainPolicy) SetData(event *es_models.Event) error {
	err := json.Unmarshal(event.Data, p)
	if err != nil {
		return errors.ThrowInternal(err, "EVENT-7JS9d", "unable to unmarshal data")
	}
	return nil
}
