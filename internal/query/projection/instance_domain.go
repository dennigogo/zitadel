package projection

import (
	"context"

	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/handler"
	"github.com/dennigogo/zitadel/internal/eventstore/handler/crdb"
	"github.com/dennigogo/zitadel/internal/repository/instance"
)

const (
	InstanceDomainTable = "projections.instance_domains"

	InstanceDomainInstanceIDCol   = "instance_id"
	InstanceDomainCreationDateCol = "creation_date"
	InstanceDomainChangeDateCol   = "change_date"
	InstanceDomainSequenceCol     = "sequence"
	InstanceDomainDomainCol       = "domain"
	InstanceDomainIsGeneratedCol  = "is_generated"
	InstanceDomainIsPrimaryCol    = "is_primary"
)

type instanceDomainProjection struct {
	crdb.StatementHandler
}

func newInstanceDomainProjection(ctx context.Context, config crdb.StatementHandlerConfig) *instanceDomainProjection {
	p := new(instanceDomainProjection)
	config.ProjectionName = InstanceDomainTable
	config.Reducers = p.reducers()
	config.InitCheck = crdb.NewTableCheck(
		crdb.NewTable([]*crdb.Column{
			crdb.NewColumn(InstanceDomainInstanceIDCol, crdb.ColumnTypeText),
			crdb.NewColumn(InstanceDomainCreationDateCol, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(InstanceDomainChangeDateCol, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(InstanceDomainSequenceCol, crdb.ColumnTypeInt64),
			crdb.NewColumn(InstanceDomainDomainCol, crdb.ColumnTypeText),
			crdb.NewColumn(InstanceDomainIsGeneratedCol, crdb.ColumnTypeBool),
			crdb.NewColumn(InstanceDomainIsPrimaryCol, crdb.ColumnTypeBool),
		},
			crdb.NewPrimaryKey(InstanceDomainInstanceIDCol, InstanceDomainDomainCol),
		),
	)
	p.StatementHandler = crdb.NewStatementHandler(ctx, config)
	return p
}

func (p *instanceDomainProjection) reducers() []handler.AggregateReducer {
	return []handler.AggregateReducer{
		{
			Aggregate: instance.AggregateType,
			EventRedusers: []handler.EventReducer{
				{
					Event:  instance.InstanceDomainAddedEventType,
					Reduce: p.reduceDomainAdded,
				},
				{
					Event:  instance.InstanceDomainPrimarySetEventType,
					Reduce: p.reduceDomainPrimarySet,
				},
				{
					Event:  instance.InstanceDomainRemovedEventType,
					Reduce: p.reduceDomainRemoved,
				},
			},
		},
	}
}

func (p *instanceDomainProjection) reduceDomainAdded(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*instance.DomainAddedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-38nNf", "reduce.wrong.event.type %s", instance.InstanceDomainAddedEventType)
	}
	return crdb.NewCreateStatement(
		e,
		[]handler.Column{
			handler.NewCol(InstanceDomainCreationDateCol, e.CreationDate()),
			handler.NewCol(InstanceDomainChangeDateCol, e.CreationDate()),
			handler.NewCol(InstanceDomainSequenceCol, e.Sequence()),
			handler.NewCol(InstanceDomainDomainCol, e.Domain),
			handler.NewCol(InstanceDomainInstanceIDCol, e.Aggregate().ID),
			handler.NewCol(InstanceDomainIsGeneratedCol, e.Generated),
			handler.NewCol(InstanceDomainIsPrimaryCol, false),
		},
	), nil
}

func (p *instanceDomainProjection) reduceDomainPrimarySet(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*instance.DomainPrimarySetEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-f8nlw", "reduce.wrong.event.type %s", instance.InstanceDomainPrimarySetEventType)
	}
	return crdb.NewMultiStatement(
		e,
		crdb.AddUpdateStatement(
			[]handler.Column{
				handler.NewCol(InstanceDomainChangeDateCol, e.CreationDate()),
				handler.NewCol(InstanceDomainSequenceCol, e.Sequence()),
				handler.NewCol(InstanceDomainIsPrimaryCol, false),
			},
			[]handler.Condition{
				handler.NewCond(InstanceDomainInstanceIDCol, e.Aggregate().InstanceID),
				handler.NewCond(InstanceDomainIsPrimaryCol, true),
			},
		),
		crdb.AddUpdateStatement(
			[]handler.Column{
				handler.NewCol(InstanceDomainChangeDateCol, e.CreationDate()),
				handler.NewCol(InstanceDomainSequenceCol, e.Sequence()),
				handler.NewCol(InstanceDomainIsPrimaryCol, true),
			},
			[]handler.Condition{
				handler.NewCond(InstanceDomainDomainCol, e.Domain),
				handler.NewCond(InstanceDomainInstanceIDCol, e.Aggregate().ID),
			},
		),
	), nil
}

func (p *instanceDomainProjection) reduceDomainRemoved(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*instance.DomainRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-388Nk", "reduce.wrong.event.type %s", instance.InstanceDomainRemovedEventType)
	}
	return crdb.NewDeleteStatement(
		e,
		[]handler.Condition{
			handler.NewCond(InstanceDomainDomainCol, e.Domain),
			handler.NewCond(InstanceDomainInstanceIDCol, e.Aggregate().ID),
		},
	), nil
}
