package query

import (
	"context"
	"database/sql"
	errs "errors"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/dennigogo/zitadel/internal/api/authz"

	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/query/projection"
)

type UserMetadataList struct {
	SearchResponse
	Metadata []*UserMetadata
}

type UserMetadata struct {
	CreationDate  time.Time
	ChangeDate    time.Time
	ResourceOwner string
	Sequence      uint64
	Key           string
	Value         []byte
}

type UserMetadataSearchQueries struct {
	SearchRequest
	Queries []SearchQuery
}

var (
	userMetadataTable = table{
		name: projection.UserMetadataProjectionTable,
	}
	UserMetadataUserIDCol = Column{
		name:  projection.UserMetadataColumnUserID,
		table: userMetadataTable,
	}
	UserMetadataCreationDateCol = Column{
		name:  projection.UserMetadataColumnCreationDate,
		table: userMetadataTable,
	}
	UserMetadataChangeDateCol = Column{
		name:  projection.UserMetadataColumnChangeDate,
		table: userMetadataTable,
	}
	UserMetadataResourceOwnerCol = Column{
		name:  projection.UserMetadataColumnResourceOwner,
		table: userMetadataTable,
	}
	UserMetadataInstanceIDCol = Column{
		name:  projection.UserMetadataColumnInstanceID,
		table: userMetadataTable,
	}
	UserMetadataSequenceCol = Column{
		name:  projection.UserMetadataColumnSequence,
		table: userMetadataTable,
	}
	UserMetadataKeyCol = Column{
		name:  projection.UserMetadataColumnKey,
		table: userMetadataTable,
	}
	UserMetadataValueCol = Column{
		name:  projection.UserMetadataColumnValue,
		table: userMetadataTable,
	}
)

func (q *Queries) GetUserMetadataByKey(ctx context.Context, shouldTriggerBulk bool, userID, key string, queries ...SearchQuery) (*UserMetadata, error) {
	if shouldTriggerBulk {
		projection.UserMetadataProjection.Trigger(ctx)
	}

	query, scan := prepareUserMetadataQuery()
	for _, q := range queries {
		query = q.toQuery(query)
	}
	stmt, args, err := query.Where(
		sq.Eq{
			UserMetadataUserIDCol.identifier():     userID,
			UserMetadataKeyCol.identifier():        key,
			UserMetadataInstanceIDCol.identifier(): authz.GetInstance(ctx).InstanceID(),
		}).ToSql()
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-aDGG2", "Errors.Query.SQLStatment")
	}

	row := q.client.QueryRowContext(ctx, stmt, args...)
	return scan(row)
}

func (q *Queries) SearchUserMetadata(ctx context.Context, shouldTriggerBulk bool, userID string, queries *UserMetadataSearchQueries) (*UserMetadataList, error) {
	if shouldTriggerBulk {
		projection.UserMetadataProjection.Trigger(ctx)
	}

	query, scan := prepareUserMetadataListQuery()
	stmt, args, err := queries.toQuery(query).Where(
		sq.Eq{
			UserMetadataUserIDCol.identifier():     userID,
			UserMetadataInstanceIDCol.identifier(): authz.GetInstance(ctx).InstanceID(),
		}).
		ToSql()
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-Egbgd", "Errors.Query.SQLStatment")
	}

	rows, err := q.client.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-Hr2wf", "Errors.Internal")
	}
	metadata, err := scan(rows)
	if err != nil {
		return nil, err
	}
	metadata.LatestSequence, err = q.latestSequence(ctx, userMetadataTable)
	return metadata, err
}

func (q *UserMetadataSearchQueries) toQuery(query sq.SelectBuilder) sq.SelectBuilder {
	query = q.SearchRequest.toQuery(query)
	for _, q := range q.Queries {
		query = q.toQuery(query)
	}
	return query
}

func (r *UserMetadataSearchQueries) AppendMyResourceOwnerQuery(orgID string) error {
	query, err := NewUserMetadataResourceOwnerSearchQuery(orgID)
	if err != nil {
		return err
	}
	r.Queries = append(r.Queries, query)
	return nil
}

func NewUserMetadataResourceOwnerSearchQuery(value string) (SearchQuery, error) {
	return NewTextQuery(UserMetadataResourceOwnerCol, value, TextEquals)
}

func NewUserMetadataKeySearchQuery(value string, comparison TextComparison) (SearchQuery, error) {
	return NewTextQuery(UserMetadataKeyCol, value, comparison)
}

func prepareUserMetadataQuery() (sq.SelectBuilder, func(*sql.Row) (*UserMetadata, error)) {
	return sq.Select(
			UserMetadataCreationDateCol.identifier(),
			UserMetadataChangeDateCol.identifier(),
			UserMetadataResourceOwnerCol.identifier(),
			UserMetadataSequenceCol.identifier(),
			UserMetadataKeyCol.identifier(),
			UserMetadataValueCol.identifier(),
		).
			From(userMetadataTable.identifier()).
			PlaceholderFormat(sq.Dollar),
		func(row *sql.Row) (*UserMetadata, error) {
			m := new(UserMetadata)
			err := row.Scan(
				&m.CreationDate,
				&m.ChangeDate,
				&m.ResourceOwner,
				&m.Sequence,
				&m.Key,
				&m.Value,
			)

			if err != nil {
				if errs.Is(err, sql.ErrNoRows) {
					return nil, errors.ThrowNotFound(err, "QUERY-Rgh32", "Errors.User.NotFound")
				}
				return nil, errors.ThrowInternal(err, "QUERY-Hhjt2", "Errors.Internal")
			}
			return m, nil
		}
}

func prepareUserMetadataListQuery() (sq.SelectBuilder, func(*sql.Rows) (*UserMetadataList, error)) {
	return sq.Select(
			UserMetadataCreationDateCol.identifier(),
			UserMetadataChangeDateCol.identifier(),
			UserMetadataResourceOwnerCol.identifier(),
			UserMetadataSequenceCol.identifier(),
			UserMetadataKeyCol.identifier(),
			UserMetadataValueCol.identifier(),
			countColumn.identifier()).
			From(userMetadataTable.identifier()).
			PlaceholderFormat(sq.Dollar),
		func(rows *sql.Rows) (*UserMetadataList, error) {
			metadata := make([]*UserMetadata, 0)
			var count uint64
			for rows.Next() {
				m := new(UserMetadata)
				err := rows.Scan(
					&m.CreationDate,
					&m.ChangeDate,
					&m.ResourceOwner,
					&m.Sequence,
					&m.Key,
					&m.Value,
					&count,
				)
				if err != nil {
					return nil, err
				}

				metadata = append(metadata, m)
			}

			if err := rows.Close(); err != nil {
				return nil, errors.ThrowInternal(err, "QUERY-sd3gh", "Errors.Query.CloseRows")
			}

			return &UserMetadataList{
				Metadata: metadata,
				SearchResponse: SearchResponse{
					Count: count,
				},
			}, nil
		}
}
