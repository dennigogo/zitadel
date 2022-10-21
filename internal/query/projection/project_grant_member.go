package projection

import (
	"context"

	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/handler"
	"github.com/dennigogo/zitadel/internal/eventstore/handler/crdb"
	"github.com/dennigogo/zitadel/internal/repository/member"
	"github.com/dennigogo/zitadel/internal/repository/org"
	"github.com/dennigogo/zitadel/internal/repository/project"
	"github.com/dennigogo/zitadel/internal/repository/user"
)

const (
	ProjectGrantMemberProjectionTable = "projections.project_grant_members2"
	ProjectGrantMemberProjectIDCol    = "project_id"
	ProjectGrantMemberGrantIDCol      = "grant_id"
)

type projectGrantMemberProjection struct {
	crdb.StatementHandler
}

func newProjectGrantMemberProjection(ctx context.Context, config crdb.StatementHandlerConfig) *projectGrantMemberProjection {
	p := new(projectGrantMemberProjection)
	config.ProjectionName = ProjectGrantMemberProjectionTable
	config.Reducers = p.reducers()
	config.InitCheck = crdb.NewTableCheck(
		crdb.NewTable(
			append(memberColumns,
				crdb.NewColumn(ProjectGrantMemberProjectIDCol, crdb.ColumnTypeText),
				crdb.NewColumn(ProjectGrantMemberGrantIDCol, crdb.ColumnTypeText),
			),
			crdb.NewPrimaryKey(MemberInstanceID, ProjectGrantMemberProjectIDCol, ProjectGrantMemberGrantIDCol, MemberUserIDCol),
			crdb.WithIndex(crdb.NewIndex("proj_grant_memb_user_idx", []string{MemberUserIDCol})),
		),
	)

	p.StatementHandler = crdb.NewStatementHandler(ctx, config)
	return p
}

func (p *projectGrantMemberProjection) reducers() []handler.AggregateReducer {
	return []handler.AggregateReducer{
		{
			Aggregate: project.AggregateType,
			EventRedusers: []handler.EventReducer{
				{
					Event:  project.GrantMemberAddedType,
					Reduce: p.reduceAdded,
				},
				{
					Event:  project.GrantMemberChangedType,
					Reduce: p.reduceChanged,
				},
				{
					Event:  project.GrantMemberCascadeRemovedType,
					Reduce: p.reduceCascadeRemoved,
				},
				{
					Event:  project.GrantMemberRemovedType,
					Reduce: p.reduceRemoved,
				},
				{
					Event:  project.ProjectRemovedType,
					Reduce: p.reduceProjectRemoved,
				},
				{
					Event:  project.GrantRemovedType,
					Reduce: p.reduceProjectGrantRemoved,
				},
			},
		},
		{
			Aggregate: user.AggregateType,
			EventRedusers: []handler.EventReducer{
				{
					Event:  user.UserRemovedType,
					Reduce: p.reduceUserRemoved,
				},
			},
		},
		{
			Aggregate: org.AggregateType,
			EventRedusers: []handler.EventReducer{
				{
					Event:  org.OrgRemovedEventType,
					Reduce: p.reduceOrgRemoved,
				},
			},
		},
	}
}

func (p *projectGrantMemberProjection) reduceAdded(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*project.GrantMemberAddedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-0EBQf", "reduce.wrong.event.type %s", project.GrantMemberAddedType)
	}
	return reduceMemberAdded(
		*member.NewMemberAddedEvent(&e.BaseEvent, e.UserID, e.Roles...),
		withMemberCol(ProjectGrantMemberProjectIDCol, e.Aggregate().ID),
		withMemberCol(ProjectGrantMemberGrantIDCol, e.GrantID),
	)
}

func (p *projectGrantMemberProjection) reduceChanged(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*project.GrantMemberChangedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-YX5Tk", "reduce.wrong.event.type %s", project.GrantMemberChangedType)
	}
	return reduceMemberChanged(
		*member.NewMemberChangedEvent(&e.BaseEvent, e.UserID, e.Roles...),
		withMemberCond(ProjectGrantMemberProjectIDCol, e.Aggregate().ID),
		withMemberCond(ProjectGrantMemberGrantIDCol, e.GrantID),
	)
}

func (p *projectGrantMemberProjection) reduceCascadeRemoved(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*project.GrantMemberCascadeRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-adnHG", "reduce.wrong.event.type %s", project.GrantMemberCascadeRemovedType)
	}
	return reduceMemberCascadeRemoved(
		*member.NewCascadeRemovedEvent(&e.BaseEvent, e.UserID),
		withMemberCond(ProjectGrantMemberProjectIDCol, e.Aggregate().ID),
		withMemberCond(ProjectGrantMemberGrantIDCol, e.GrantID),
	)
}

func (p *projectGrantMemberProjection) reduceRemoved(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*project.GrantMemberRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-MGNnA", "reduce.wrong.event.type %s", project.GrantMemberRemovedType)
	}
	return reduceMemberRemoved(e,
		withMemberCond(MemberUserIDCol, e.UserID),
		withMemberCond(ProjectGrantMemberProjectIDCol, e.Aggregate().ID),
		withMemberCond(ProjectGrantMemberGrantIDCol, e.GrantID),
	)
}

func (p *projectGrantMemberProjection) reduceUserRemoved(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*user.UserRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-rufJr", "reduce.wrong.event.type %s", user.UserRemovedType)
	}
	return reduceMemberRemoved(e, withMemberCond(MemberUserIDCol, e.Aggregate().ID))
}

func (p *projectGrantMemberProjection) reduceOrgRemoved(event eventstore.Event) (*handler.Statement, error) {
	//TODO: as soon as org deletion is implemented:
	// Case: The user has resource owner A and project has resource owner B
	// if org B deleted it works
	// if org A is deleted, the membership wouldn't be deleted
	e, ok := event.(*org.OrgRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-Zzp6o", "reduce.wrong.event.type %s", org.OrgRemovedEventType)
	}
	return reduceMemberRemoved(e, withMemberCond(MemberResourceOwner, e.Aggregate().ID))
}

func (p *projectGrantMemberProjection) reduceProjectRemoved(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*project.ProjectRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-JLODy", "reduce.wrong.event.type %s", project.ProjectRemovedType)
	}
	return reduceMemberRemoved(e, withMemberCond(ProjectGrantMemberProjectIDCol, e.Aggregate().ID))
}

func (p *projectGrantMemberProjection) reduceProjectGrantRemoved(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*project.GrantRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-D1J9R", "reduce.wrong.event.type %s", project.GrantRemovedType)
	}
	return reduceMemberRemoved(e,
		withMemberCond(ProjectGrantMemberGrantIDCol, e.GrantID),
		withMemberCond(ProjectGrantMemberProjectIDCol, e.Aggregate().ID),
	)
}
