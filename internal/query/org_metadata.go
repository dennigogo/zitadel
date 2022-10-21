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

type OrgMetadataList struct {
	SearchResponse
	Metadata []*OrgMetadata
}

type OrgMetadata struct {
	CreationDate  time.Time
	ChangeDate    time.Time
	ResourceOwner string
	Sequence      uint64
	Key           string
	Value         []byte
}

type OrgMetadataSearchQueries struct {
	SearchRequest
	Queries []SearchQuery
}

var (
	orgMetadataTable = table{
		name: projection.OrgMetadataProjectionTable,
	}
	OrgMetadataOrgIDCol = Column{
		name:  projection.OrgMetadataColumnOrgID,
		table: orgMetadataTable,
	}
	OrgMetadataCreationDateCol = Column{
		name:  projection.OrgMetadataColumnCreationDate,
		table: orgMetadataTable,
	}
	OrgMetadataChangeDateCol = Column{
		name:  projection.OrgMetadataColumnChangeDate,
		table: orgMetadataTable,
	}
	OrgMetadataResourceOwnerCol = Column{
		name:  projection.OrgMetadataColumnResourceOwner,
		table: orgMetadataTable,
	}
	OrgMetadataInstanceIDCol = Column{
		name:  projection.OrgMetadataColumnInstanceID,
		table: orgMetadataTable,
	}
	OrgMetadataSequenceCol = Column{
		name:  projection.OrgMetadataColumnSequence,
		table: orgMetadataTable,
	}
	OrgMetadataKeyCol = Column{
		name:  projection.OrgMetadataColumnKey,
		table: orgMetadataTable,
	}
	OrgMetadataValueCol = Column{
		name:  projection.OrgMetadataColumnValue,
		table: orgMetadataTable,
	}
)

func (q *Queries) GetOrgMetadataByKey(ctx context.Context, shouldTriggerBulk bool, orgID string, key string, queries ...SearchQuery) (*OrgMetadata, error) {
	if shouldTriggerBulk {
		projection.OrgMetadataProjection.Trigger(ctx)
	}

	query, scan := prepareOrgMetadataQuery()
	for _, q := range queries {
		query = q.toQuery(query)
	}
	stmt, args, err := query.Where(
		sq.Eq{
			OrgMetadataOrgIDCol.identifier():      orgID,
			OrgMetadataKeyCol.identifier():        key,
			OrgMetadataInstanceIDCol.identifier(): authz.GetInstance(ctx).InstanceID(),
		}).ToSql()
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-aDaG2", "Errors.Query.SQLStatment")
	}

	row := q.client.QueryRowContext(ctx, stmt, args...)
	return scan(row)
}

func (q *Queries) SearchOrgMetadata(ctx context.Context, shouldTriggerBulk bool, orgID string, queries *OrgMetadataSearchQueries) (*OrgMetadataList, error) {
	if shouldTriggerBulk {
		projection.OrgMetadataProjection.Trigger(ctx)
	}

	query, scan := prepareOrgMetadataListQuery()
	stmt, args, err := queries.toQuery(query).Where(
		sq.Eq{
			OrgMetadataOrgIDCol.identifier():      orgID,
			OrgMetadataInstanceIDCol.identifier(): authz.GetInstance(ctx).InstanceID(),
		}).
		ToSql()
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-Egbld", "Errors.Query.SQLStatment")
	}

	rows, err := q.client.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-Ho2wf", "Errors.Internal")
	}
	metadata, err := scan(rows)
	if err != nil {
		return nil, err
	}
	metadata.LatestSequence, err = q.latestSequence(ctx, orgMetadataTable)
	return metadata, err
}

func (q *OrgMetadataSearchQueries) toQuery(query sq.SelectBuilder) sq.SelectBuilder {
	query = q.SearchRequest.toQuery(query)
	for _, q := range q.Queries {
		query = q.toQuery(query)
	}
	return query
}

func (r *OrgMetadataSearchQueries) AppendMyResourceOwnerQuery(orgID string) error {
	query, err := NewOrgMetadataResourceOwnerSearchQuery(orgID)
	if err != nil {
		return err
	}
	r.Queries = append(r.Queries, query)
	return nil
}

func NewOrgMetadataResourceOwnerSearchQuery(value string) (SearchQuery, error) {
	return NewTextQuery(OrgMetadataResourceOwnerCol, value, TextEquals)
}

func NewOrgMetadataKeySearchQuery(value string, comparison TextComparison) (SearchQuery, error) {
	return NewTextQuery(OrgMetadataKeyCol, value, comparison)
}

func prepareOrgMetadataQuery() (sq.SelectBuilder, func(*sql.Row) (*OrgMetadata, error)) {
	return sq.Select(
			OrgMetadataCreationDateCol.identifier(),
			OrgMetadataChangeDateCol.identifier(),
			OrgMetadataResourceOwnerCol.identifier(),
			OrgMetadataSequenceCol.identifier(),
			OrgMetadataKeyCol.identifier(),
			OrgMetadataValueCol.identifier(),
		).
			From(orgMetadataTable.identifier()).
			PlaceholderFormat(sq.Dollar),
		func(row *sql.Row) (*OrgMetadata, error) {
			m := new(OrgMetadata)
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
					return nil, errors.ThrowNotFound(err, "QUERY-Rph32", "Errors.Metadata.NotFound")
				}
				return nil, errors.ThrowInternal(err, "QUERY-Hajt2", "Errors.Internal")
			}
			return m, nil
		}
}

func prepareOrgMetadataListQuery() (sq.SelectBuilder, func(*sql.Rows) (*OrgMetadataList, error)) {
	return sq.Select(
			OrgMetadataCreationDateCol.identifier(),
			OrgMetadataChangeDateCol.identifier(),
			OrgMetadataResourceOwnerCol.identifier(),
			OrgMetadataSequenceCol.identifier(),
			OrgMetadataKeyCol.identifier(),
			OrgMetadataValueCol.identifier(),
			countColumn.identifier()).
			From(orgMetadataTable.identifier()).
			PlaceholderFormat(sq.Dollar),
		func(rows *sql.Rows) (*OrgMetadataList, error) {
			metadata := make([]*OrgMetadata, 0)
			var count uint64
			for rows.Next() {
				m := new(OrgMetadata)
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
				return nil, errors.ThrowInternal(err, "QUERY-dd3gh", "Errors.Query.CloseRows")
			}

			return &OrgMetadataList{
				Metadata: metadata,
				SearchResponse: SearchResponse{
					Count: count,
				},
			}, nil
		}
}
