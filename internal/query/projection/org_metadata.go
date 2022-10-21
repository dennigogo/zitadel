package projection

import (
	"context"

	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/handler"
	"github.com/dennigogo/zitadel/internal/eventstore/handler/crdb"
	"github.com/dennigogo/zitadel/internal/repository/org"
)

const (
	OrgMetadataProjectionTable = "projections.org_metadata"

	OrgMetadataColumnOrgID         = "org_id"
	OrgMetadataColumnCreationDate  = "creation_date"
	OrgMetadataColumnChangeDate    = "change_date"
	OrgMetadataColumnSequence      = "sequence"
	OrgMetadataColumnResourceOwner = "resource_owner"
	OrgMetadataColumnInstanceID    = "instance_id"
	OrgMetadataColumnKey           = "key"
	OrgMetadataColumnValue         = "value"
)

type orgMetadataProjection struct {
	crdb.StatementHandler
}

func newOrgMetadataProjection(ctx context.Context, config crdb.StatementHandlerConfig) *orgMetadataProjection {
	p := new(orgMetadataProjection)
	config.ProjectionName = OrgMetadataProjectionTable
	config.Reducers = p.reducers()
	config.InitCheck = crdb.NewTableCheck(
		crdb.NewTable([]*crdb.Column{
			crdb.NewColumn(OrgMetadataColumnOrgID, crdb.ColumnTypeText),
			crdb.NewColumn(OrgMetadataColumnCreationDate, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(OrgMetadataColumnChangeDate, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(OrgMetadataColumnSequence, crdb.ColumnTypeInt64),
			crdb.NewColumn(OrgMetadataColumnResourceOwner, crdb.ColumnTypeText),
			crdb.NewColumn(OrgMetadataColumnInstanceID, crdb.ColumnTypeText),
			crdb.NewColumn(OrgMetadataColumnKey, crdb.ColumnTypeText),
			crdb.NewColumn(OrgMetadataColumnValue, crdb.ColumnTypeBytes, crdb.Nullable()),
		},
			crdb.NewPrimaryKey(OrgMetadataColumnInstanceID, OrgMetadataColumnOrgID, OrgMetadataColumnKey),
		),
	)

	p.StatementHandler = crdb.NewStatementHandler(ctx, config)
	return p
}

func (p *orgMetadataProjection) reducers() []handler.AggregateReducer {
	return []handler.AggregateReducer{
		{
			Aggregate: org.AggregateType,
			EventRedusers: []handler.EventReducer{
				{
					Event:  org.MetadataSetType,
					Reduce: p.reduceMetadataSet,
				},
				{
					Event:  org.MetadataRemovedType,
					Reduce: p.reduceMetadataRemoved,
				},
				{
					Event:  org.MetadataRemovedAllType,
					Reduce: p.reduceMetadataRemovedAll,
				},
				{
					Event:  org.OrgRemovedEventType,
					Reduce: p.reduceMetadataRemovedAll,
				},
			},
		},
	}
}

func (p *orgMetadataProjection) reduceMetadataSet(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*org.MetadataSetEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-Ghn53", "reduce.wrong.event.type %s", org.MetadataSetType)
	}
	return crdb.NewUpsertStatement(
		e,
		[]handler.Column{
			handler.NewCol(OrgMetadataColumnInstanceID, nil),
			handler.NewCol(OrgMetadataColumnOrgID, nil),
			handler.NewCol(OrgMetadataColumnKey, e.Key),
		},
		[]handler.Column{
			handler.NewCol(OrgMetadataColumnInstanceID, e.Aggregate().InstanceID),
			handler.NewCol(OrgMetadataColumnOrgID, e.Aggregate().ID),
			handler.NewCol(OrgMetadataColumnKey, e.Key),
			handler.NewCol(OrgMetadataColumnResourceOwner, e.Aggregate().ResourceOwner),
			handler.NewCol(OrgMetadataColumnCreationDate, e.CreationDate()),
			handler.NewCol(OrgMetadataColumnChangeDate, e.CreationDate()),
			handler.NewCol(OrgMetadataColumnSequence, e.Sequence()),
			handler.NewCol(OrgMetadataColumnValue, e.Value),
		},
	), nil
}

func (p *orgMetadataProjection) reduceMetadataRemoved(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*org.MetadataRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-Bm542", "reduce.wrong.event.type %s", org.MetadataRemovedType)
	}
	return crdb.NewDeleteStatement(
		e,
		[]handler.Condition{
			handler.NewCond(OrgMetadataColumnOrgID, e.Aggregate().ID),
			handler.NewCond(OrgMetadataColumnKey, e.Key),
		},
	), nil
}

func (p *orgMetadataProjection) reduceMetadataRemovedAll(event eventstore.Event) (*handler.Statement, error) {
	switch event.(type) {
	case *org.MetadataRemovedAllEvent,
		*org.OrgRemovedEvent:
		//ok
	default:
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-Bmnf3", "reduce.wrong.event.type %v", []eventstore.EventType{org.MetadataRemovedAllType, org.OrgRemovedEventType})
	}
	return crdb.NewDeleteStatement(
		event,
		[]handler.Condition{
			handler.NewCond(OrgMetadataColumnOrgID, event.Aggregate().ID),
		},
	), nil
}
