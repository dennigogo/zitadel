package query

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/dennigogo/zitadel/internal/api/authz"
	"github.com/dennigogo/zitadel/internal/database"
	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/query/projection"
)

type Memberships struct {
	SearchResponse
	Memberships []*Membership
}

type Membership struct {
	UserID        string
	Roles         database.StringArray
	CreationDate  time.Time
	ChangeDate    time.Time
	Sequence      uint64
	ResourceOwner string

	Org          *OrgMembership
	IAM          *IAMMembership
	Project      *ProjectMembership
	ProjectGrant *ProjectGrantMembership
}

type OrgMembership struct {
	OrgID string
	Name  string
}

type IAMMembership struct {
	IAMID string
	Name  string
}

type ProjectMembership struct {
	ProjectID string
	Name      string
}

type ProjectGrantMembership struct {
	ProjectID    string
	ProjectName  string
	GrantID      string
	GrantedOrgID string
}

type MembershipSearchQuery struct {
	SearchRequest
	Queries []SearchQuery
}

func NewMembershipUserIDQuery(userID string) (SearchQuery, error) {
	return NewTextQuery(membershipUserID.setTable(membershipAlias), userID, TextEquals)
}

func NewMembershipResourceOwnerQuery(value string) (SearchQuery, error) {
	return NewTextQuery(membershipResourceOwner.setTable(membershipAlias), value, TextEquals)
}

func NewMembershipOrgIDQuery(value string) (SearchQuery, error) {
	return NewTextQuery(membershipOrgID, value, TextEquals)
}

func NewMembershipResourceOwnersSearchQuery(ids ...string) (SearchQuery, error) {
	list := make([]interface{}, len(ids))
	for i, value := range ids {
		list[i] = value
	}
	return NewListQuery(membershipResourceOwner, list, ListIn)
}

func NewMembershipGrantedOrgIDSearchQuery(id string) (SearchQuery, error) {
	return NewTextQuery(ProjectGrantColumnGrantedOrgID, id, TextEquals)
}

func NewMembershipProjectIDQuery(value string) (SearchQuery, error) {
	return NewTextQuery(membershipProjectID, value, TextEquals)
}

func NewMembershipProjectGrantIDQuery(value string) (SearchQuery, error) {
	return NewTextQuery(membershipGrantID, value, TextEquals)
}

func NewMembershipIsIAMQuery() (SearchQuery, error) {
	return NewNotNullQuery(membershipIAMID)
}

func (q *MembershipSearchQuery) toQuery(query sq.SelectBuilder) sq.SelectBuilder {
	query = q.SearchRequest.toQuery(query)
	for _, q := range q.Queries {
		query = q.toQuery(query)
	}
	return query
}

func (q *Queries) Memberships(ctx context.Context, queries *MembershipSearchQuery) (*Memberships, error) {
	query, scan := prepareMembershipsQuery()
	stmt, args, err := queries.toQuery(query).
		Where(sq.Eq{
			membershipInstanceID.identifier(): authz.GetInstance(ctx).InstanceID(),
		}).ToSql()
	if err != nil {
		return nil, errors.ThrowInvalidArgument(err, "QUERY-T84X9", "Errors.Query.InvalidRequest")
	}
	latestSequence, err := q.latestSequence(ctx, orgMemberTable, instanceMemberTable, projectMemberTable, projectGrantMemberTable)
	if err != nil {
		return nil, err
	}

	rows, err := q.client.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-eAV2x", "Errors.Internal")
	}
	memberships, err := scan(rows)
	if err != nil {
		return nil, err
	}
	memberships.LatestSequence = latestSequence
	return memberships, nil
}

var (
	//membershipAlias is a hack to satisfy checks in the queries
	membershipAlias = table{
		name: "memberships",
	}
	membershipUserID = Column{
		name:  projection.MemberUserIDCol,
		table: membershipAlias,
	}
	membershipRoles = Column{
		name:  projection.MemberRolesCol,
		table: membershipAlias,
	}
	membershipCreationDate = Column{
		name:  projection.MemberCreationDate,
		table: membershipAlias,
	}
	membershipChangeDate = Column{
		name:  projection.MemberChangeDate,
		table: membershipAlias,
	}
	membershipSequence = Column{
		name:  projection.MemberSequence,
		table: membershipAlias,
	}
	membershipResourceOwner = Column{
		name:  projection.MemberResourceOwner,
		table: membershipAlias,
	}
	membershipInstanceID = Column{
		name:  projection.MemberInstanceID,
		table: membershipAlias,
	}
	membershipOrgID = Column{
		name:  projection.OrgMemberOrgIDCol,
		table: membershipAlias,
	}
	membershipIAMID = Column{
		name:  projection.InstanceMemberIAMIDCol,
		table: membershipAlias,
	}
	membershipProjectID = Column{
		name:  projection.ProjectMemberProjectIDCol,
		table: membershipAlias,
	}
	membershipGrantID = Column{
		name:  projection.ProjectGrantMemberGrantIDCol,
		table: membershipAlias,
	}
	membershipGrantGrantedOrgID = Column{
		name:  projection.ProjectGrantColumnGrantedOrgID,
		table: membershipAlias,
	}

	membershipFrom = "(" +
		prepareOrgMember() +
		" UNION ALL " +
		prepareIAMMember() +
		" UNION ALL " +
		prepareProjectMember() +
		" UNION ALL " +
		prepareProjectGrantMember() +
		") AS " + membershipAlias.identifier()
)

func prepareMembershipsQuery() (sq.SelectBuilder, func(*sql.Rows) (*Memberships, error)) {
	return sq.Select(
			membershipUserID.identifier(),
			membershipRoles.identifier(),
			membershipCreationDate.identifier(),
			membershipChangeDate.identifier(),
			membershipSequence.identifier(),
			membershipResourceOwner.identifier(),
			membershipOrgID.identifier(),
			membershipIAMID.identifier(),
			membershipProjectID.identifier(),
			membershipGrantID.identifier(),
			ProjectGrantColumnGrantedOrgID.identifier(),
			ProjectColumnName.identifier(),
			OrgColumnName.identifier(),
			countColumn.identifier(),
		).From(membershipFrom).
			LeftJoin(join(ProjectColumnID, membershipProjectID)).
			LeftJoin(join(OrgColumnID, membershipOrgID)).
			LeftJoin(join(ProjectGrantColumnGrantID, membershipGrantID)).
			PlaceholderFormat(sq.Dollar),
		func(rows *sql.Rows) (*Memberships, error) {
			memberships := make([]*Membership, 0)
			var count uint64
			for rows.Next() {

				var (
					membership   = new(Membership)
					orgID        = sql.NullString{}
					iamID        = sql.NullString{}
					projectID    = sql.NullString{}
					grantID      = sql.NullString{}
					grantedOrgID = sql.NullString{}
					projectName  = sql.NullString{}
					orgName      = sql.NullString{}
				)

				err := rows.Scan(
					&membership.UserID,
					&membership.Roles,
					&membership.CreationDate,
					&membership.ChangeDate,
					&membership.Sequence,
					&membership.ResourceOwner,
					&orgID,
					&iamID,
					&projectID,
					&grantID,
					&grantedOrgID,
					&projectName,
					&orgName,
					&count,
				)

				if err != nil {
					return nil, err
				}

				if orgID.Valid {
					membership.Org = &OrgMembership{
						OrgID: orgID.String,
						Name:  orgName.String,
					}
				} else if iamID.Valid {
					membership.IAM = &IAMMembership{
						IAMID: iamID.String,
						Name:  iamID.String,
					}
				} else if projectID.Valid && grantID.Valid && grantedOrgID.Valid {
					membership.ProjectGrant = &ProjectGrantMembership{
						ProjectID:    projectID.String,
						ProjectName:  projectName.String,
						GrantID:      grantID.String,
						GrantedOrgID: grantedOrgID.String,
					}
				} else if projectID.Valid {
					membership.Project = &ProjectMembership{
						ProjectID: projectID.String,
						Name:      projectName.String,
					}
				}

				memberships = append(memberships, membership)
			}

			if err := rows.Close(); err != nil {
				return nil, errors.ThrowInternal(err, "QUERY-N34NV", "Errors.Query.CloseRows")
			}

			return &Memberships{
				Memberships: memberships,
				SearchResponse: SearchResponse{
					Count: count,
				},
			}, nil
		}
}

func prepareOrgMember() string {
	stmt, _ := sq.Select(
		OrgMemberUserID.identifier(),
		OrgMemberRoles.identifier(),
		OrgMemberCreationDate.identifier(),
		OrgMemberChangeDate.identifier(),
		OrgMemberSequence.identifier(),
		OrgMemberResourceOwner.identifier(),
		OrgMemberInstanceID.identifier(),
		OrgMemberOrgID.identifier(),
		"NULL::TEXT AS "+membershipIAMID.name,
		"NULL::TEXT AS "+membershipProjectID.name,
		"NULL::TEXT AS "+membershipGrantID.name,
	).From(orgMemberTable.identifier()).MustSql()
	return stmt
}

func prepareIAMMember() string {
	stmt, _ := sq.Select(
		InstanceMemberUserID.identifier(),
		InstanceMemberRoles.identifier(),
		InstanceMemberCreationDate.identifier(),
		InstanceMemberChangeDate.identifier(),
		InstanceMemberSequence.identifier(),
		InstanceMemberResourceOwner.identifier(),
		InstanceMemberInstanceID.identifier(),
		"NULL::TEXT AS "+membershipOrgID.name,
		InstanceMemberIAMID.identifier(),
		"NULL::TEXT AS "+membershipProjectID.name,
		"NULL::TEXT AS "+membershipGrantID.name,
	).From(instanceMemberTable.identifier()).MustSql()
	return stmt
}

func prepareProjectMember() string {
	stmt, _ := sq.Select(
		ProjectMemberUserID.identifier(),
		ProjectMemberRoles.identifier(),
		ProjectMemberCreationDate.identifier(),
		ProjectMemberChangeDate.identifier(),
		ProjectMemberSequence.identifier(),
		ProjectMemberResourceOwner.identifier(),
		ProjectMemberInstanceID.identifier(),
		"NULL::TEXT AS "+membershipOrgID.name,
		"NULL::TEXT AS "+membershipIAMID.name,
		ProjectMemberProjectID.identifier(),
		"NULL::TEXT AS "+membershipGrantID.name,
	).From(projectMemberTable.identifier()).MustSql()

	return stmt
}

func prepareProjectGrantMember() string {
	stmt, _ := sq.Select(
		ProjectGrantMemberUserID.identifier(),
		ProjectGrantMemberRoles.identifier(),
		ProjectGrantMemberCreationDate.identifier(),
		ProjectGrantMemberChangeDate.identifier(),
		ProjectGrantMemberSequence.identifier(),
		ProjectGrantMemberResourceOwner.identifier(),
		ProjectGrantMemberInstanceID.identifier(),
		"NULL::TEXT AS "+membershipOrgID.name,
		"NULL::TEXT AS "+membershipIAMID.name,
		ProjectGrantMemberProjectID.identifier(),
		ProjectGrantMemberGrantID.identifier(),
	).From(projectGrantMemberTable.identifier()).
		MustSql()

	return stmt
}
