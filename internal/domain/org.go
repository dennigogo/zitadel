package domain

import (
	"strings"

	"github.com/dennigogo/zitadel/internal/eventstore/v1/models"
)

type Org struct {
	models.ObjectRoot

	State OrgState
	Name  string

	PrimaryDomain string
	Domains       []*OrgDomain
}

func (o *Org) IsValid() bool {
	if o == nil {
		return false
	}
	o.Name = strings.TrimSpace(o.Name)
	return o.Name != ""
}

func (o *Org) AddIAMDomain(iamDomain string) {
	o.Domains = append(o.Domains, &OrgDomain{Domain: NewIAMDomainName(o.Name, iamDomain), Verified: true, Primary: true})
}

type OrgState int32

const (
	OrgStateUnspecified OrgState = iota
	OrgStateActive
	OrgStateInactive
	OrgStateRemoved
)
