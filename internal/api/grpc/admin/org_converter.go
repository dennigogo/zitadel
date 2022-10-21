package admin

import (
	"github.com/dennigogo/zitadel/internal/api/grpc/object"
	org_grpc "github.com/dennigogo/zitadel/internal/api/grpc/org"
	"github.com/dennigogo/zitadel/internal/query"
	"github.com/dennigogo/zitadel/pkg/grpc/admin"
	"github.com/dennigogo/zitadel/pkg/grpc/org"
)

func listOrgRequestToModel(req *admin.ListOrgsRequest) (*query.OrgSearchQueries, error) {
	offset, limit, asc := object.ListQueryToModel(req.Query)
	queries, err := org_grpc.OrgQueriesToModel(req.Queries)
	if err != nil {
		return nil, err
	}
	return &query.OrgSearchQueries{
		SearchRequest: query.SearchRequest{
			Offset:        offset,
			Limit:         limit,
			SortingColumn: fieldNameToOrgColumn(req.SortingColumn),
			Asc:           asc,
		},
		Queries: queries,
	}, nil
}

func fieldNameToOrgColumn(fieldName org.OrgFieldName) query.Column {
	switch fieldName {
	case org.OrgFieldName_ORG_FIELD_NAME_NAME:
		return query.OrgColumnName
	default:
		return query.Column{}
	}
}
