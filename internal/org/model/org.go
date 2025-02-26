package model

import (
	"strings"

	es_models "github.com/dennigogo/zitadel/internal/eventstore/v1/models"
	iam_model "github.com/dennigogo/zitadel/internal/iam/model"
)

type Org struct {
	es_models.ObjectRoot

	State   OrgState
	Name    string
	Domains []*OrgDomain

	DomainPolicy *iam_model.DomainPolicy
}

type OrgState int32

const (
	OrgStateActive OrgState = iota
	OrgStateInactive
)

func (o *Org) IsActive() bool {
	return o.State == OrgStateActive
}

func (o *Org) GetDomain(domain *OrgDomain) (int, *OrgDomain) {
	for i, d := range o.Domains {
		if d.Domain == domain.Domain {
			return i, d
		}
	}
	return -1, nil
}

func (o *Org) GetPrimaryDomain() *OrgDomain {
	for _, d := range o.Domains {
		if d.Primary {
			return d
		}
	}
	return nil
}

func (o *Org) nameForDomain(iamDomain string) string {
	return strings.ToLower(strings.ReplaceAll(o.Name, " ", "-") + "." + iamDomain)
}

func (o *Org) AddIAMDomain(iamDomain string) {
	o.Domains = append(o.Domains, &OrgDomain{Domain: o.nameForDomain(iamDomain), Verified: true, Primary: true})
}
