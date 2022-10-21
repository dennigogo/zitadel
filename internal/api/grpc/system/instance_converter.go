package system

import (
	"strings"

	"golang.org/x/text/language"

	instance_grpc "github.com/dennigogo/zitadel/internal/api/grpc/instance"
	"github.com/dennigogo/zitadel/internal/api/grpc/object"
	"github.com/dennigogo/zitadel/internal/command"
	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/query"
	instance_pb "github.com/dennigogo/zitadel/pkg/grpc/instance"
	system_pb "github.com/dennigogo/zitadel/pkg/grpc/system"
)

func AddInstancePbToSetupInstance(req *system_pb.AddInstanceRequest, defaultInstance command.InstanceSetup, externalDomain string) *command.InstanceSetup {
	if req.InstanceName != "" {
		defaultInstance.InstanceName = req.InstanceName
		defaultInstance.Org.Name = req.InstanceName
	}
	if req.CustomDomain != "" {
		defaultInstance.CustomDomain = req.CustomDomain
	}
	if req.FirstOrgName != "" {
		defaultInstance.Org.Name = req.FirstOrgName
	}
	if req.OwnerEmail.Email != "" {
		defaultInstance.Org.Human.Email.Address = req.OwnerEmail.Email
		defaultInstance.Org.Human.Email.Verified = req.OwnerEmail.IsEmailVerified
	}
	if req.OwnerProfile != nil {
		if req.OwnerProfile.FirstName != "" {
			defaultInstance.Org.Human.FirstName = req.OwnerProfile.FirstName
		}
		if req.OwnerProfile.LastName != "" {
			defaultInstance.Org.Human.LastName = req.OwnerProfile.LastName
		}
		if req.OwnerProfile.PreferredLanguage != "" {
			lang, err := language.Parse(req.OwnerProfile.PreferredLanguage)
			if err == nil {
				defaultInstance.Org.Human.PreferredLanguage = lang
			}
		}
	}
	// check if default username is email style or else append @<orgname>.<custom-domain>
	// this way we have the same value as before changing `UserLoginMustBeDomain` to false
	if !defaultInstance.DomainPolicy.UserLoginMustBeDomain && !strings.Contains(defaultInstance.Org.Human.Username, "@") {
		defaultInstance.Org.Human.Username = defaultInstance.Org.Human.Username + "@" + domain.NewIAMDomainName(defaultInstance.Org.Name, externalDomain)
	}
	if req.OwnerUserName != "" {
		defaultInstance.Org.Human.Username = req.OwnerUserName
	}
	if req.OwnerPassword != nil {
		defaultInstance.Org.Human.Password = req.OwnerPassword.Password
		defaultInstance.Org.Human.PasswordChangeRequired = req.OwnerPassword.PasswordChangeRequired
	}
	if lang := language.Make(req.DefaultLanguage); lang != language.Und {
		defaultInstance.DefaultLanguage = lang
	}

	return &defaultInstance
}
func ListInstancesRequestToModel(req *system_pb.ListInstancesRequest) (*query.InstanceSearchQueries, error) {
	offset, limit, asc := object.ListQueryToModel(req.Query)
	queries, err := instance_grpc.InstanceQueriesToModel(req.Queries)
	if err != nil {
		return nil, err
	}
	return &query.InstanceSearchQueries{
		SearchRequest: query.SearchRequest{
			Offset:        offset,
			Limit:         limit,
			Asc:           asc,
			SortingColumn: fieldNameToInstanceColumn(req.SortingColumn),
		},
		Queries: queries,
	}, nil
}

func fieldNameToInstanceColumn(fieldName instance_pb.FieldName) query.Column {
	switch fieldName {
	case instance_pb.FieldName_FIELD_NAME_ID:
		return query.InstanceColumnID
	case instance_pb.FieldName_FIELD_NAME_NAME:
		return query.InstanceColumnName
	case instance_pb.FieldName_FIELD_NAME_CREATION_DATE:
		return query.InstanceColumnCreationDate
	default:
		return query.Column{}
	}
}

func ListInstanceDomainsRequestToModel(req *system_pb.ListDomainsRequest) (*query.InstanceDomainSearchQueries, error) {
	offset, limit, asc := object.ListQueryToModel(req.Query)
	queries, err := instance_grpc.DomainQueriesToModel(req.Queries)
	if err != nil {
		return nil, err
	}
	return &query.InstanceDomainSearchQueries{
		SearchRequest: query.SearchRequest{
			Offset:        offset,
			Limit:         limit,
			Asc:           asc,
			SortingColumn: fieldNameToInstanceDomainColumn(req.SortingColumn),
		},
		Queries: queries,
	}, nil
}

func fieldNameToInstanceDomainColumn(fieldName instance_pb.DomainFieldName) query.Column {
	switch fieldName {
	case instance_pb.DomainFieldName_DOMAIN_FIELD_NAME_DOMAIN:
		return query.InstanceDomainDomainCol
	case instance_pb.DomainFieldName_DOMAIN_FIELD_NAME_GENERATED:
		return query.InstanceDomainIsGeneratedCol
	case instance_pb.DomainFieldName_DOMAIN_FIELD_NAME_PRIMARY:
		return query.InstanceDomainIsPrimaryCol
	case instance_pb.DomainFieldName_DOMAIN_FIELD_NAME_CREATION_DATE:
		return query.InstanceDomainCreationDateCol
	default:
		return query.Column{}
	}
}
