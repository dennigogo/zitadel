package query

import (
	"context"
	"database/sql"
	errs "errors"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/dennigogo/zitadel/internal/api/authz"
	"github.com/dennigogo/zitadel/internal/database"

	"github.com/dennigogo/zitadel/internal/query/projection"

	"github.com/dennigogo/zitadel/internal/errors"
)

var (
	personalAccessTokensTable = table{
		name: projection.PersonalAccessTokenProjectionTable,
	}
	PersonalAccessTokenColumnID = Column{
		name:  projection.PersonalAccessTokenColumnID,
		table: personalAccessTokensTable,
	}
	PersonalAccessTokenColumnUserID = Column{
		name:  projection.PersonalAccessTokenColumnUserID,
		table: personalAccessTokensTable,
	}
	PersonalAccessTokenColumnExpiration = Column{
		name:  projection.PersonalAccessTokenColumnExpiration,
		table: personalAccessTokensTable,
	}
	PersonalAccessTokenColumnScopes = Column{
		name:  projection.PersonalAccessTokenColumnScopes,
		table: personalAccessTokensTable,
	}
	PersonalAccessTokenColumnCreationDate = Column{
		name:  projection.PersonalAccessTokenColumnCreationDate,
		table: personalAccessTokensTable,
	}
	PersonalAccessTokenColumnChangeDate = Column{
		name:  projection.PersonalAccessTokenColumnChangeDate,
		table: personalAccessTokensTable,
	}
	PersonalAccessTokenColumnResourceOwner = Column{
		name:  projection.PersonalAccessTokenColumnResourceOwner,
		table: personalAccessTokensTable,
	}
	PersonalAccessTokenColumnInstanceID = Column{
		name:  projection.PersonalAccessTokenColumnInstanceID,
		table: personalAccessTokensTable,
	}
	PersonalAccessTokenColumnSequence = Column{
		name:  projection.PersonalAccessTokenColumnSequence,
		table: personalAccessTokensTable,
	}
)

type PersonalAccessTokens struct {
	SearchResponse
	PersonalAccessTokens []*PersonalAccessToken
}

type PersonalAccessToken struct {
	ID            string
	CreationDate  time.Time
	ChangeDate    time.Time
	ResourceOwner string
	Sequence      uint64

	UserID     string
	Expiration time.Time
	Scopes     database.StringArray
}

type PersonalAccessTokenSearchQueries struct {
	SearchRequest
	Queries []SearchQuery
}

func (q *Queries) PersonalAccessTokenByID(ctx context.Context, shouldTriggerBulk bool, id string, queries ...SearchQuery) (*PersonalAccessToken, error) {
	if shouldTriggerBulk {
		projection.PersonalAccessTokenProjection.Trigger(ctx)
	}

	query, scan := preparePersonalAccessTokenQuery()
	for _, q := range queries {
		query = q.toQuery(query)
	}
	stmt, args, err := query.Where(sq.Eq{
		PersonalAccessTokenColumnID.identifier():         id,
		PersonalAccessTokenColumnInstanceID.identifier(): authz.GetInstance(ctx).InstanceID(),
	}).ToSql()
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-Dgfb4", "Errors.Query.SQLStatment")
	}

	row := q.client.QueryRowContext(ctx, stmt, args...)
	return scan(row)
}

func (q *Queries) SearchPersonalAccessTokens(ctx context.Context, queries *PersonalAccessTokenSearchQueries) (personalAccessTokens *PersonalAccessTokens, err error) {
	query, scan := preparePersonalAccessTokensQuery()
	stmt, args, err := queries.toQuery(query).
		Where(sq.Eq{
			PersonalAccessTokenColumnInstanceID.identifier(): authz.GetInstance(ctx).InstanceID(),
		}).ToSql()
	if err != nil {
		return nil, errors.ThrowInvalidArgument(err, "QUERY-Hjw2w", "Errors.Query.InvalidRequest")
	}

	rows, err := q.client.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-Bmz63", "Errors.Internal")
	}
	personalAccessTokens, err = scan(rows)
	if err != nil {
		return nil, err
	}
	personalAccessTokens.LatestSequence, err = q.latestSequence(ctx, personalAccessTokensTable)
	return personalAccessTokens, err
}

func NewPersonalAccessTokenResourceOwnerSearchQuery(value string) (SearchQuery, error) {
	return NewTextQuery(PersonalAccessTokenColumnResourceOwner, value, TextEquals)
}

func NewPersonalAccessTokenUserIDSearchQuery(value string) (SearchQuery, error) {
	return NewTextQuery(PersonalAccessTokenColumnUserID, value, TextEquals)
}

func (r *PersonalAccessTokenSearchQueries) AppendMyResourceOwnerQuery(orgID string) error {
	query, err := NewPersonalAccessTokenResourceOwnerSearchQuery(orgID)
	if err != nil {
		return err
	}
	r.Queries = append(r.Queries, query)
	return nil
}

func (q *PersonalAccessTokenSearchQueries) toQuery(query sq.SelectBuilder) sq.SelectBuilder {
	query = q.SearchRequest.toQuery(query)
	for _, q := range q.Queries {
		query = q.toQuery(query)
	}
	return query
}

func preparePersonalAccessTokenQuery() (sq.SelectBuilder, func(*sql.Row) (*PersonalAccessToken, error)) {
	return sq.Select(
			PersonalAccessTokenColumnID.identifier(),
			PersonalAccessTokenColumnCreationDate.identifier(),
			PersonalAccessTokenColumnChangeDate.identifier(),
			PersonalAccessTokenColumnResourceOwner.identifier(),
			PersonalAccessTokenColumnSequence.identifier(),
			PersonalAccessTokenColumnUserID.identifier(),
			PersonalAccessTokenColumnExpiration.identifier(),
			PersonalAccessTokenColumnScopes.identifier()).
			From(personalAccessTokensTable.identifier()).PlaceholderFormat(sq.Dollar),
		func(row *sql.Row) (*PersonalAccessToken, error) {
			p := new(PersonalAccessToken)
			err := row.Scan(
				&p.ID,
				&p.CreationDate,
				&p.ChangeDate,
				&p.ResourceOwner,
				&p.Sequence,
				&p.UserID,
				&p.Expiration,
				&p.Scopes,
			)
			if err != nil {
				if errs.Is(err, sql.ErrNoRows) {
					return nil, errors.ThrowNotFound(err, "QUERY-fk2fs", "Errors.PersonalAccessToken.NotFound")
				}
				return nil, errors.ThrowInternal(err, "QUERY-dj2FF", "Errors.Internal")
			}
			return p, nil
		}
}

func preparePersonalAccessTokensQuery() (sq.SelectBuilder, func(*sql.Rows) (*PersonalAccessTokens, error)) {
	return sq.Select(
			PersonalAccessTokenColumnID.identifier(),
			PersonalAccessTokenColumnCreationDate.identifier(),
			PersonalAccessTokenColumnChangeDate.identifier(),
			PersonalAccessTokenColumnResourceOwner.identifier(),
			PersonalAccessTokenColumnSequence.identifier(),
			PersonalAccessTokenColumnUserID.identifier(),
			PersonalAccessTokenColumnExpiration.identifier(),
			PersonalAccessTokenColumnScopes.identifier(),
			countColumn.identifier()).
			From(personalAccessTokensTable.identifier()).PlaceholderFormat(sq.Dollar),
		func(rows *sql.Rows) (*PersonalAccessTokens, error) {
			personalAccessTokens := make([]*PersonalAccessToken, 0)
			var count uint64
			for rows.Next() {
				token := new(PersonalAccessToken)
				err := rows.Scan(
					&token.ID,
					&token.CreationDate,
					&token.ChangeDate,
					&token.ResourceOwner,
					&token.Sequence,
					&token.UserID,
					&token.Expiration,
					&token.Scopes,
					&count,
				)
				if err != nil {
					return nil, err
				}
				personalAccessTokens = append(personalAccessTokens, token)
			}

			if err := rows.Close(); err != nil {
				return nil, errors.ThrowInternal(err, "QUERY-QMXJv", "Errors.Query.CloseRows")
			}

			return &PersonalAccessTokens{
				PersonalAccessTokens: personalAccessTokens,
				SearchResponse: SearchResponse{
					Count: count,
				},
			}, nil
		}
}
