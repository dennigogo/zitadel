package projection

import (
	"context"

	"github.com/dennigogo/zitadel/internal/database"
	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/handler"
	"github.com/dennigogo/zitadel/internal/eventstore/handler/crdb"
	"github.com/dennigogo/zitadel/internal/repository/project"
	"github.com/dennigogo/zitadel/internal/repository/user"
	"github.com/dennigogo/zitadel/internal/repository/usergrant"
)

const (
	UserGrantProjectionTable = "projections.user_grants2"

	UserGrantID            = "id"
	UserGrantCreationDate  = "creation_date"
	UserGrantChangeDate    = "change_date"
	UserGrantSequence      = "sequence"
	UserGrantState         = "state"
	UserGrantResourceOwner = "resource_owner"
	UserGrantInstanceID    = "instance_id"
	UserGrantUserID        = "user_id"
	UserGrantProjectID     = "project_id"
	UserGrantGrantID       = "grant_id"
	UserGrantRoles         = "roles"
)

type userGrantProjection struct {
	crdb.StatementHandler
}

func newUserGrantProjection(ctx context.Context, config crdb.StatementHandlerConfig) *userGrantProjection {
	p := new(userGrantProjection)
	config.ProjectionName = UserGrantProjectionTable
	config.Reducers = p.reducers()
	config.InitCheck = crdb.NewTableCheck(
		crdb.NewTable([]*crdb.Column{
			crdb.NewColumn(UserGrantID, crdb.ColumnTypeText),
			crdb.NewColumn(UserGrantCreationDate, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(UserGrantChangeDate, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(UserGrantSequence, crdb.ColumnTypeInt64),
			crdb.NewColumn(UserGrantState, crdb.ColumnTypeEnum),
			crdb.NewColumn(UserGrantResourceOwner, crdb.ColumnTypeText),
			crdb.NewColumn(UserGrantInstanceID, crdb.ColumnTypeText),
			crdb.NewColumn(UserGrantUserID, crdb.ColumnTypeText),
			crdb.NewColumn(UserGrantProjectID, crdb.ColumnTypeText),
			crdb.NewColumn(UserGrantGrantID, crdb.ColumnTypeText),
			crdb.NewColumn(UserGrantRoles, crdb.ColumnTypeTextArray, crdb.Nullable()),
		},
			crdb.NewPrimaryKey(UserGrantInstanceID, UserGrantID),
			crdb.WithIndex(crdb.NewIndex("user_grant_user_idx", []string{UserGrantUserID})),
			crdb.WithIndex(crdb.NewIndex("user_grant_ro_idx", []string{UserGrantResourceOwner})),
		),
	)

	p.StatementHandler = crdb.NewStatementHandler(ctx, config)
	return p
}

func (p *userGrantProjection) reducers() []handler.AggregateReducer {
	return []handler.AggregateReducer{
		{
			Aggregate: usergrant.AggregateType,
			EventRedusers: []handler.EventReducer{
				{
					Event:  usergrant.UserGrantAddedType,
					Reduce: p.reduceAdded,
				},
				{
					Event:  usergrant.UserGrantChangedType,
					Reduce: p.reduceChanged,
				},
				{
					Event:  usergrant.UserGrantCascadeChangedType,
					Reduce: p.reduceChanged,
				},
				{
					Event:  usergrant.UserGrantRemovedType,
					Reduce: p.reduceRemoved,
				},
				{
					Event:  usergrant.UserGrantCascadeRemovedType,
					Reduce: p.reduceRemoved,
				},
				{
					Event:  usergrant.UserGrantDeactivatedType,
					Reduce: p.reduceDeactivated,
				},
				{
					Event:  usergrant.UserGrantReactivatedType,
					Reduce: p.reduceReactivated,
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
			Aggregate: project.AggregateType,
			EventRedusers: []handler.EventReducer{
				{
					Event:  project.ProjectRemovedType,
					Reduce: p.reduceProjectRemoved,
				},
				{
					Event:  project.GrantRemovedType,
					Reduce: p.reduceProjectGrantRemoved,
				},
				{
					Event:  project.RoleRemovedType,
					Reduce: p.reduceRoleRemoved,
				},
				{
					Event:  project.GrantChangedType,
					Reduce: p.reduceProjectGrantChanged,
				},
				{
					Event:  project.GrantCascadeChangedType,
					Reduce: p.reduceProjectGrantChanged,
				},
			},
		},
	}
}

func (p *userGrantProjection) reduceAdded(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*usergrant.UserGrantAddedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-MQHVB", "reduce.wrong.event.type %s", usergrant.UserGrantAddedType)
	}
	return crdb.NewCreateStatement(
		e,
		[]handler.Column{
			handler.NewCol(UserGrantID, e.Aggregate().ID),
			handler.NewCol(UserGrantResourceOwner, e.Aggregate().ResourceOwner),
			handler.NewCol(UserGrantInstanceID, e.Aggregate().InstanceID),
			handler.NewCol(UserGrantCreationDate, e.CreationDate()),
			handler.NewCol(UserGrantChangeDate, e.CreationDate()),
			handler.NewCol(UserGrantSequence, e.Sequence()),
			handler.NewCol(UserGrantUserID, e.UserID),
			handler.NewCol(UserGrantProjectID, e.ProjectID),
			handler.NewCol(UserGrantGrantID, e.ProjectGrantID),
			handler.NewCol(UserGrantRoles, database.StringArray(e.RoleKeys)),
			handler.NewCol(UserGrantState, domain.UserGrantStateActive),
		},
	), nil
}

func (p *userGrantProjection) reduceChanged(event eventstore.Event) (*handler.Statement, error) {
	var roles database.StringArray

	switch e := event.(type) {
	case *usergrant.UserGrantChangedEvent:
		roles = e.RoleKeys
	case *usergrant.UserGrantCascadeChangedEvent:
		roles = e.RoleKeys
	default:
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-hOr1E", "reduce.wrong.event.type %v", []eventstore.EventType{usergrant.UserGrantChangedType, usergrant.UserGrantCascadeChangedType})
	}

	return crdb.NewUpdateStatement(
		event,
		[]handler.Column{
			handler.NewCol(UserGrantChangeDate, event.CreationDate()),
			handler.NewCol(UserGrantRoles, roles),
			handler.NewCol(UserGrantSequence, event.Sequence()),
		},
		[]handler.Condition{
			handler.NewCond(UserGrantID, event.Aggregate().ID),
		},
	), nil
}

func (p *userGrantProjection) reduceRemoved(event eventstore.Event) (*handler.Statement, error) {
	switch event.(type) {
	case *usergrant.UserGrantRemovedEvent, *usergrant.UserGrantCascadeRemovedEvent:
		// ok
	default:
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-7OBEC", "reduce.wrong.event.type %v", []eventstore.EventType{usergrant.UserGrantRemovedType, usergrant.UserGrantCascadeRemovedType})
	}

	return crdb.NewDeleteStatement(
		event,
		[]handler.Condition{
			handler.NewCond(UserGrantID, event.Aggregate().ID),
		},
	), nil
}

func (p *userGrantProjection) reduceDeactivated(event eventstore.Event) (*handler.Statement, error) {
	if _, ok := event.(*usergrant.UserGrantDeactivatedEvent); !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-oP7Gm", "reduce.wrong.event.type %s", usergrant.UserGrantDeactivatedType)
	}

	return crdb.NewUpdateStatement(
		event,
		[]handler.Column{
			handler.NewCol(UserGrantChangeDate, event.CreationDate()),
			handler.NewCol(UserGrantState, domain.UserGrantStateInactive),
			handler.NewCol(UserGrantSequence, event.Sequence()),
		},
		[]handler.Condition{
			handler.NewCond(UserGrantID, event.Aggregate().ID),
		},
	), nil
}

func (p *userGrantProjection) reduceReactivated(event eventstore.Event) (*handler.Statement, error) {
	if _, ok := event.(*usergrant.UserGrantDeactivatedEvent); !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-DGsKh", "reduce.wrong.event.type %s", usergrant.UserGrantReactivatedType)
	}

	return crdb.NewUpdateStatement(
		event,
		[]handler.Column{
			handler.NewCol(UserGrantChangeDate, event.CreationDate()),
			handler.NewCol(UserGrantState, domain.UserGrantStateActive),
			handler.NewCol(UserGrantSequence, event.Sequence()),
		},
		[]handler.Condition{
			handler.NewCond(UserGrantID, event.Aggregate().ID),
		},
	), nil
}

func (p *userGrantProjection) reduceUserRemoved(event eventstore.Event) (*handler.Statement, error) {
	if _, ok := event.(*user.UserRemovedEvent); !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-Bner2a", "reduce.wrong.event.type %s", user.UserRemovedType)
	}

	return crdb.NewDeleteStatement(
		event,
		[]handler.Condition{
			handler.NewCond(UserGrantUserID, event.Aggregate().ID),
		},
	), nil
}

func (p *userGrantProjection) reduceProjectRemoved(event eventstore.Event) (*handler.Statement, error) {
	if _, ok := event.(*project.ProjectRemovedEvent); !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-Bne2a", "reduce.wrong.event.type %s", project.ProjectRemovedType)
	}

	return crdb.NewDeleteStatement(
		event,
		[]handler.Condition{
			handler.NewCond(UserGrantProjectID, event.Aggregate().ID),
		},
	), nil
}

func (p *userGrantProjection) reduceProjectGrantRemoved(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*project.GrantRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-dGr2a", "reduce.wrong.event.type %s", project.GrantRemovedType)
	}

	return crdb.NewDeleteStatement(
		event,
		[]handler.Condition{
			handler.NewCond(UserGrantGrantID, e.GrantID),
		},
	), nil
}

func (p *userGrantProjection) reduceRoleRemoved(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*project.RoleRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-dswg2", "reduce.wrong.event.type %s", project.RoleRemovedType)
	}

	return crdb.NewUpdateStatement(
		event,
		[]handler.Column{
			crdb.NewArrayRemoveCol(UserGrantRoles, e.Key),
		},
		[]handler.Condition{
			handler.NewCond(UserGrantProjectID, e.Aggregate().ID),
		},
	), nil
}

func (p *userGrantProjection) reduceProjectGrantChanged(event eventstore.Event) (*handler.Statement, error) {
	var grantID string
	var keys []string
	switch e := event.(type) {
	case *project.GrantChangedEvent:
		grantID = e.GrantID
		keys = e.RoleKeys
	case *project.GrantCascadeChangedEvent:
		grantID = e.GrantID
		keys = e.RoleKeys
	default:
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-Fh3gw", "reduce.wrong.event.type %v", []eventstore.EventType{project.GrantChangedType, project.GrantCascadeChangedType})
	}

	return crdb.NewUpdateStatement(
		event,
		[]handler.Column{
			crdb.NewArrayIntersectCol(UserGrantRoles, database.StringArray(keys)),
		},
		[]handler.Condition{
			handler.NewCond(UserGrantGrantID, grantID),
		},
	), nil
}
