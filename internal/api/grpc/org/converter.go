package org

import (
	"github.com/dennigogo/zitadel/internal/api/grpc/object"
	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/query"
	org_pb "github.com/dennigogo/zitadel/pkg/grpc/org"
)

func OrgQueriesToModel(queries []*org_pb.OrgQuery) (_ []query.SearchQuery, err error) {
	q := make([]query.SearchQuery, len(queries))
	for i, query := range queries {
		q[i], err = OrgQueryToModel(query)
		if err != nil {
			return nil, err
		}
	}
	return q, nil
}

func OrgQueryToModel(apiQuery *org_pb.OrgQuery) (query.SearchQuery, error) {
	switch q := apiQuery.Query.(type) {
	case *org_pb.OrgQuery_DomainQuery:
		return query.NewOrgDomainSearchQuery(object.TextMethodToQuery(q.DomainQuery.Method), q.DomainQuery.Domain)
	case *org_pb.OrgQuery_NameQuery:
		return query.NewOrgNameSearchQuery(object.TextMethodToQuery(q.NameQuery.Method), q.NameQuery.Name)
	default:
		return nil, errors.ThrowInvalidArgument(nil, "ORG-vR9nC", "List.Query.Invalid")
	}
}

func OrgQueriesToQuery(queries []*org_pb.OrgQuery) (_ []query.SearchQuery, err error) {
	q := make([]query.SearchQuery, len(queries))
	for i, query := range queries {
		q[i], err = OrgQueryToQuery(query)
		if err != nil {
			return nil, err
		}
	}
	return q, nil
}

func OrgQueryToQuery(search *org_pb.OrgQuery) (query.SearchQuery, error) {
	switch q := search.Query.(type) {
	case *org_pb.OrgQuery_DomainQuery:
		return query.NewOrgDomainSearchQuery(object.TextMethodToQuery(q.DomainQuery.Method), q.DomainQuery.Domain)
	case *org_pb.OrgQuery_NameQuery:
		return query.NewOrgNameSearchQuery(object.TextMethodToQuery(q.NameQuery.Method), q.NameQuery.Name)
	default:
		return nil, errors.ThrowInvalidArgument(nil, "ADMIN-ADvsd", "List.Query.Invalid")
	}
}

func OrgViewsToPb(orgs []*query.Org) []*org_pb.Org {
	o := make([]*org_pb.Org, len(orgs))
	for i, org := range orgs {
		o[i] = OrgViewToPb(org)
	}
	return o
}

func OrgViewToPb(org *query.Org) *org_pb.Org {
	return &org_pb.Org{
		Id:            org.ID,
		State:         OrgStateToPb(org.State),
		Name:          org.Name,
		PrimaryDomain: org.Domain,
		Details: object.ToViewDetailsPb(
			org.Sequence,
			org.CreationDate,
			org.ChangeDate,
			org.ResourceOwner,
		),
	}
}

func OrgsToPb(orgs []*query.Org) []*org_pb.Org {
	o := make([]*org_pb.Org, len(orgs))
	for i, org := range orgs {
		o[i] = OrgToPb(org)
	}
	return o
}

func OrgToPb(org *query.Org) *org_pb.Org {
	return &org_pb.Org{
		Id:            org.ID,
		Name:          org.Name,
		PrimaryDomain: org.Domain,
		Details:       object.ToViewDetailsPb(org.Sequence, org.CreationDate, org.ChangeDate, org.ResourceOwner),
		State:         OrgStateToPb(org.State),
	}
}

func OrgStateToPb(state domain.OrgState) org_pb.OrgState {
	switch state {
	case domain.OrgStateActive:
		return org_pb.OrgState_ORG_STATE_ACTIVE
	case domain.OrgStateInactive:
		return org_pb.OrgState_ORG_STATE_INACTIVE
	default:
		return org_pb.OrgState_ORG_STATE_UNSPECIFIED
	}
}

func DomainQueriesToModel(queries []*org_pb.DomainSearchQuery) (_ []query.SearchQuery, err error) {
	q := make([]query.SearchQuery, len(queries))
	for i, query := range queries {
		q[i], err = DomainQueryToModel(query)
		if err != nil {
			return nil, err
		}
	}
	return q, nil
}

func DomainQueryToModel(searchQuery *org_pb.DomainSearchQuery) (query.SearchQuery, error) {
	switch q := searchQuery.Query.(type) {
	case *org_pb.DomainSearchQuery_DomainNameQuery:
		return query.NewOrgDomainDomainSearchQuery(object.TextMethodToQuery(q.DomainNameQuery.Method), q.DomainNameQuery.Name)
	default:
		return nil, errors.ThrowInvalidArgument(nil, "ORG-Ags42", "List.Query.Invalid")
	}
}

func DomainsToPb(domains []*query.Domain) []*org_pb.Domain {
	d := make([]*org_pb.Domain, len(domains))
	for i, domain := range domains {
		d[i] = DomainToPb(domain)
	}
	return d
}

func DomainToPb(d *query.Domain) *org_pb.Domain {
	return &org_pb.Domain{
		OrgId:          d.OrgID,
		DomainName:     d.Domain,
		IsVerified:     d.IsVerified,
		IsPrimary:      d.IsPrimary,
		ValidationType: DomainValidationTypeFromModel(d.ValidationType),
		Details: object.ToViewDetailsPb(
			d.Sequence,
			d.CreationDate,
			d.ChangeDate,
			d.OrgID,
		),
	}
}

func DomainValidationTypeToDomain(validationType org_pb.DomainValidationType) domain.OrgDomainValidationType {
	switch validationType {
	case org_pb.DomainValidationType_DOMAIN_VALIDATION_TYPE_HTTP:
		return domain.OrgDomainValidationTypeHTTP
	case org_pb.DomainValidationType_DOMAIN_VALIDATION_TYPE_DNS:
		return domain.OrgDomainValidationTypeDNS
	default:
		return domain.OrgDomainValidationTypeUnspecified
	}
}

func DomainValidationTypeFromModel(validationType domain.OrgDomainValidationType) org_pb.DomainValidationType {
	switch validationType {
	case domain.OrgDomainValidationTypeDNS:
		return org_pb.DomainValidationType_DOMAIN_VALIDATION_TYPE_DNS
	case domain.OrgDomainValidationTypeHTTP:
		return org_pb.DomainValidationType_DOMAIN_VALIDATION_TYPE_HTTP
	default:
		return org_pb.DomainValidationType_DOMAIN_VALIDATION_TYPE_UNSPECIFIED
	}
}
