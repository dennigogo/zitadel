package projection

import (
	"context"

	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/handler"
	"github.com/dennigogo/zitadel/internal/eventstore/handler/crdb"
	"github.com/dennigogo/zitadel/internal/repository/project"
)

const (
	ProjectProjectionTable = "projections.projects2"

	ProjectColumnID                     = "id"
	ProjectColumnCreationDate           = "creation_date"
	ProjectColumnChangeDate             = "change_date"
	ProjectColumnSequence               = "sequence"
	ProjectColumnState                  = "state"
	ProjectColumnResourceOwner          = "resource_owner"
	ProjectColumnInstanceID             = "instance_id"
	ProjectColumnName                   = "name"
	ProjectColumnProjectRoleAssertion   = "project_role_assertion"
	ProjectColumnProjectRoleCheck       = "project_role_check"
	ProjectColumnHasProjectCheck        = "has_project_check"
	ProjectColumnPrivateLabelingSetting = "private_labeling_setting"
)

type projectProjection struct {
	crdb.StatementHandler
}

func newProjectProjection(ctx context.Context, config crdb.StatementHandlerConfig) *projectProjection {
	p := new(projectProjection)
	config.ProjectionName = ProjectProjectionTable
	config.Reducers = p.reducers()
	config.InitCheck = crdb.NewTableCheck(
		crdb.NewTable([]*crdb.Column{
			crdb.NewColumn(ProjectColumnID, crdb.ColumnTypeText),
			crdb.NewColumn(ProjectColumnCreationDate, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(ProjectColumnChangeDate, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(ProjectColumnSequence, crdb.ColumnTypeInt64),
			crdb.NewColumn(ProjectColumnState, crdb.ColumnTypeEnum),
			crdb.NewColumn(ProjectColumnResourceOwner, crdb.ColumnTypeText),
			crdb.NewColumn(ProjectColumnInstanceID, crdb.ColumnTypeText),
			crdb.NewColumn(ProjectColumnName, crdb.ColumnTypeText),
			crdb.NewColumn(ProjectColumnProjectRoleAssertion, crdb.ColumnTypeBool),
			crdb.NewColumn(ProjectColumnProjectRoleCheck, crdb.ColumnTypeBool),
			crdb.NewColumn(ProjectColumnHasProjectCheck, crdb.ColumnTypeBool),
			crdb.NewColumn(ProjectColumnPrivateLabelingSetting, crdb.ColumnTypeEnum),
		},
			crdb.NewPrimaryKey(ProjectColumnInstanceID, ProjectColumnID),
			crdb.WithIndex(crdb.NewIndex("project_ro_idx", []string{ProjectColumnResourceOwner})),
		),
	)
	p.StatementHandler = crdb.NewStatementHandler(ctx, config)
	return p
}

func (p *projectProjection) reducers() []handler.AggregateReducer {
	return []handler.AggregateReducer{
		{
			Aggregate: project.AggregateType,
			EventRedusers: []handler.EventReducer{
				{
					Event:  project.ProjectAddedType,
					Reduce: p.reduceProjectAdded,
				},
				{
					Event:  project.ProjectChangedType,
					Reduce: p.reduceProjectChanged,
				},
				{
					Event:  project.ProjectDeactivatedType,
					Reduce: p.reduceProjectDeactivated,
				},
				{
					Event:  project.ProjectReactivatedType,
					Reduce: p.reduceProjectReactivated,
				},
				{
					Event:  project.ProjectRemovedType,
					Reduce: p.reduceProjectRemoved,
				},
			},
		},
	}
}

func (p *projectProjection) reduceProjectAdded(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*project.ProjectAddedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-l000S", "reduce.wrong.event.type %s", project.ProjectAddedType)
	}
	return crdb.NewCreateStatement(
		e,
		[]handler.Column{
			handler.NewCol(ProjectColumnID, e.Aggregate().ID),
			handler.NewCol(ProjectColumnCreationDate, e.CreationDate()),
			handler.NewCol(ProjectColumnChangeDate, e.CreationDate()),
			handler.NewCol(ProjectColumnResourceOwner, e.Aggregate().ResourceOwner),
			handler.NewCol(ProjectColumnInstanceID, e.Aggregate().InstanceID),
			handler.NewCol(ProjectColumnSequence, e.Sequence()),
			handler.NewCol(ProjectColumnName, e.Name),
			handler.NewCol(ProjectColumnProjectRoleAssertion, e.ProjectRoleAssertion),
			handler.NewCol(ProjectColumnProjectRoleCheck, e.ProjectRoleCheck),
			handler.NewCol(ProjectColumnHasProjectCheck, e.HasProjectCheck),
			handler.NewCol(ProjectColumnPrivateLabelingSetting, e.PrivateLabelingSetting),
			handler.NewCol(ProjectColumnState, domain.ProjectStateActive),
		},
	), nil
}

func (p *projectProjection) reduceProjectChanged(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*project.ProjectChangeEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-s00Fs", "reduce.wrong.event.type %s", project.ProjectChangedType)
	}
	if e.Name == nil && e.HasProjectCheck == nil && e.ProjectRoleAssertion == nil && e.ProjectRoleCheck == nil && e.PrivateLabelingSetting == nil {
		return crdb.NewNoOpStatement(e), nil
	}

	columns := make([]handler.Column, 0, 7)
	columns = append(columns, handler.NewCol(ProjectColumnChangeDate, e.CreationDate()),
		handler.NewCol(ProjectColumnSequence, e.Sequence()))
	if e.Name != nil {
		columns = append(columns, handler.NewCol(ProjectColumnName, *e.Name))
	}
	if e.ProjectRoleAssertion != nil {
		columns = append(columns, handler.NewCol(ProjectColumnProjectRoleAssertion, *e.ProjectRoleAssertion))
	}
	if e.ProjectRoleCheck != nil {
		columns = append(columns, handler.NewCol(ProjectColumnProjectRoleCheck, *e.ProjectRoleCheck))
	}
	if e.HasProjectCheck != nil {
		columns = append(columns, handler.NewCol(ProjectColumnHasProjectCheck, *e.HasProjectCheck))
	}
	if e.PrivateLabelingSetting != nil {
		columns = append(columns, handler.NewCol(ProjectColumnPrivateLabelingSetting, *e.PrivateLabelingSetting))
	}
	return crdb.NewUpdateStatement(
		e,
		columns,
		[]handler.Condition{
			handler.NewCond(ProjectColumnID, e.Aggregate().ID),
		},
	), nil
}

func (p *projectProjection) reduceProjectDeactivated(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*project.ProjectDeactivatedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-LLp0f", "reduce.wrong.event.type %s", project.ProjectDeactivatedType)
	}
	return crdb.NewUpdateStatement(
		e,
		[]handler.Column{
			handler.NewCol(ProjectColumnChangeDate, e.CreationDate()),
			handler.NewCol(ProjectColumnSequence, e.Sequence()),
			handler.NewCol(ProjectColumnState, domain.ProjectStateInactive),
		},
		[]handler.Condition{
			handler.NewCond(ProjectColumnID, e.Aggregate().ID),
		},
	), nil
}

func (p *projectProjection) reduceProjectReactivated(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*project.ProjectReactivatedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-9J98f", "reduce.wrong.event.type %s", project.ProjectReactivatedType)
	}
	return crdb.NewUpdateStatement(
		e,
		[]handler.Column{
			handler.NewCol(ProjectColumnChangeDate, e.CreationDate()),
			handler.NewCol(ProjectColumnSequence, e.Sequence()),
			handler.NewCol(ProjectColumnState, domain.ProjectStateActive),
		},
		[]handler.Condition{
			handler.NewCond(ProjectColumnID, e.Aggregate().ID),
		},
	), nil
}

func (p *projectProjection) reduceProjectRemoved(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*project.ProjectRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-5N9fs", "reduce.wrong.event.type %s", project.ProjectRemovedType)
	}
	return crdb.NewDeleteStatement(
		e,
		[]handler.Condition{
			handler.NewCond(ProjectColumnID, e.Aggregate().ID),
		},
	), nil
}
