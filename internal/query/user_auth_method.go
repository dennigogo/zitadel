package query

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/dennigogo/zitadel/internal/api/authz"
	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/query/projection"
)

var (
	userAuthMethodTable = table{
		name: projection.UserAuthMethodTable,
	}
	UserAuthMethodColumnTokenID = Column{
		name:  projection.UserAuthMethodTokenIDCol,
		table: userAuthMethodTable,
	}
	UserAuthMethodColumnCreationDate = Column{
		name:  projection.UserAuthMethodCreationDateCol,
		table: userAuthMethodTable,
	}
	UserAuthMethodColumnChangeDate = Column{
		name:  projection.UserAuthMethodChangeDateCol,
		table: userAuthMethodTable,
	}
	UserAuthMethodColumnResourceOwner = Column{
		name:  projection.UserAuthMethodResourceOwnerCol,
		table: userAuthMethodTable,
	}
	UserAuthMethodColumnInstanceID = Column{
		name:  projection.UserAuthMethodInstanceIDCol,
		table: userAuthMethodTable,
	}
	UserAuthMethodColumnUserID = Column{
		name:  projection.UserAuthMethodUserIDCol,
		table: userAuthMethodTable,
	}
	UserAuthMethodColumnSequence = Column{
		name:  projection.UserAuthMethodSequenceCol,
		table: userAuthMethodTable,
	}
	UserAuthMethodColumnName = Column{
		name:  projection.UserAuthMethodNameCol,
		table: userAuthMethodTable,
	}
	UserAuthMethodColumnState = Column{
		name:  projection.UserAuthMethodStateCol,
		table: userAuthMethodTable,
	}
	UserAuthMethodColumnMethodType = Column{
		name:  projection.UserAuthMethodTypeCol,
		table: userAuthMethodTable,
	}
)

type AuthMethods struct {
	SearchResponse
	AuthMethods []*AuthMethod
}
type AuthMethod struct {
	UserID        string
	CreationDate  time.Time
	ChangeDate    time.Time
	ResourceOwner string
	State         domain.MFAState
	Sequence      uint64

	TokenID string
	Name    string
	Type    domain.UserAuthMethodType
}

type UserAuthMethodSearchQueries struct {
	SearchRequest
	Queries []SearchQuery
}

func (q *Queries) SearchUserAuthMethods(ctx context.Context, queries *UserAuthMethodSearchQueries) (userAuthMethods *AuthMethods, err error) {
	query, scan := prepareUserAuthMethodsQuery()
	stmt, args, err := queries.toQuery(query).
		Where(sq.Eq{
			UserAuthMethodColumnInstanceID.identifier(): authz.GetInstance(ctx).InstanceID(),
		}).ToSql()
	if err != nil {
		return nil, errors.ThrowInvalidArgument(err, "QUERY-j9NJd", "Errors.Query.InvalidRequest")
	}

	rows, err := q.client.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-3n99f", "Errors.Internal")
	}
	userAuthMethods, err = scan(rows)
	if err != nil {
		return nil, err
	}
	userAuthMethods.LatestSequence, err = q.latestSequence(ctx, userAuthMethodTable)
	return userAuthMethods, err
}

func NewUserAuthMethodUserIDSearchQuery(value string) (SearchQuery, error) {
	return NewTextQuery(UserAuthMethodColumnUserID, value, TextEquals)
}

func NewUserAuthMethodTokenIDSearchQuery(value string) (SearchQuery, error) {
	return NewTextQuery(UserAuthMethodColumnTokenID, value, TextEquals)
}

func NewUserAuthMethodResourceOwnerSearchQuery(value string) (SearchQuery, error) {
	return NewTextQuery(UserAuthMethodColumnResourceOwner, value, TextEquals)
}

func NewUserAuthMethodTypeSearchQuery(value domain.UserAuthMethodType) (SearchQuery, error) {
	return NewNumberQuery(UserAuthMethodColumnMethodType, value, NumberEquals)
}

func NewUserAuthMethodStateSearchQuery(value domain.MFAState) (SearchQuery, error) {
	return NewNumberQuery(UserAuthMethodColumnState, value, NumberEquals)
}

func NewUserAuthMethodTypesSearchQuery(values ...domain.UserAuthMethodType) (SearchQuery, error) {
	list := make([]interface{}, len(values))
	for i, value := range values {
		list[i] = value
	}
	return NewListQuery(UserAuthMethodColumnMethodType, list, ListIn)
}

func (r *UserAuthMethodSearchQueries) AppendResourceOwnerQuery(orgID string) error {
	query, err := NewUserAuthMethodResourceOwnerSearchQuery(orgID)
	if err != nil {
		return err
	}
	r.Queries = append(r.Queries, query)
	return nil
}

func (r *UserAuthMethodSearchQueries) AppendUserIDQuery(userID string) error {
	query, err := NewUserAuthMethodUserIDSearchQuery(userID)
	if err != nil {
		return err
	}
	r.Queries = append(r.Queries, query)
	return nil
}

func (r *UserAuthMethodSearchQueries) AppendTokenIDQuery(tokenID string) error {
	query, err := NewUserAuthMethodTokenIDSearchQuery(tokenID)
	if err != nil {
		return err
	}
	r.Queries = append(r.Queries, query)
	return nil
}

func (r *UserAuthMethodSearchQueries) AppendStateQuery(state domain.MFAState) error {
	query, err := NewUserAuthMethodStateSearchQuery(state)
	if err != nil {
		return err
	}
	r.Queries = append(r.Queries, query)
	return nil
}

func (r *UserAuthMethodSearchQueries) AppendAuthMethodQuery(authMethod domain.UserAuthMethodType) error {
	query, err := NewUserAuthMethodTypeSearchQuery(authMethod)
	if err != nil {
		return err
	}
	r.Queries = append(r.Queries, query)
	return nil
}

func (r *UserAuthMethodSearchQueries) AppendAuthMethodsQuery(authMethod ...domain.UserAuthMethodType) error {
	query, err := NewUserAuthMethodTypesSearchQuery(authMethod...)
	if err != nil {
		return err
	}
	r.Queries = append(r.Queries, query)
	return nil
}

func (q *UserAuthMethodSearchQueries) toQuery(query sq.SelectBuilder) sq.SelectBuilder {
	query = q.SearchRequest.toQuery(query)
	for _, q := range q.Queries {
		query = q.toQuery(query)
	}
	return query
}

func prepareUserAuthMethodsQuery() (sq.SelectBuilder, func(*sql.Rows) (*AuthMethods, error)) {
	return sq.Select(
			UserAuthMethodColumnTokenID.identifier(),
			UserAuthMethodColumnCreationDate.identifier(),
			UserAuthMethodColumnChangeDate.identifier(),
			UserAuthMethodColumnResourceOwner.identifier(),
			UserAuthMethodColumnUserID.identifier(),
			UserAuthMethodColumnSequence.identifier(),
			UserAuthMethodColumnName.identifier(),
			UserAuthMethodColumnState.identifier(),
			UserAuthMethodColumnMethodType.identifier(),
			countColumn.identifier()).
			From(userAuthMethodTable.identifier()).PlaceholderFormat(sq.Dollar),
		func(rows *sql.Rows) (*AuthMethods, error) {
			userAuthMethods := make([]*AuthMethod, 0)
			var count uint64
			for rows.Next() {
				authMethod := new(AuthMethod)
				err := rows.Scan(
					&authMethod.TokenID,
					&authMethod.CreationDate,
					&authMethod.ChangeDate,
					&authMethod.ResourceOwner,
					&authMethod.UserID,
					&authMethod.Sequence,
					&authMethod.Name,
					&authMethod.State,
					&authMethod.Type,
					&count,
				)
				if err != nil {
					return nil, err
				}
				userAuthMethods = append(userAuthMethods, authMethod)
			}

			if err := rows.Close(); err != nil {
				return nil, errors.ThrowInternal(err, "QUERY-3n9fl", "Errors.Query.CloseRows")
			}

			return &AuthMethods{
				AuthMethods: userAuthMethods,
				SearchResponse: SearchResponse{
					Count: count,
				},
			}, nil
		}
}
