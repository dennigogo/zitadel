package projection

import (
	"context"

	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/handler"
	"github.com/dennigogo/zitadel/internal/eventstore/handler/crdb"
	"github.com/dennigogo/zitadel/internal/repository/instance"
	"github.com/dennigogo/zitadel/internal/repository/org"
	"github.com/dennigogo/zitadel/internal/repository/user"
)

const (
	IDPUserLinkTable             = "projections.idp_user_links2"
	IDPUserLinkIDPIDCol          = "idp_id"
	IDPUserLinkUserIDCol         = "user_id"
	IDPUserLinkExternalUserIDCol = "external_user_id"
	IDPUserLinkCreationDateCol   = "creation_date"
	IDPUserLinkChangeDateCol     = "change_date"
	IDPUserLinkSequenceCol       = "sequence"
	IDPUserLinkResourceOwnerCol  = "resource_owner"
	IDPUserLinkInstanceIDCol     = "instance_id"
	IDPUserLinkDisplayNameCol    = "display_name"
)

type idpUserLinkProjection struct {
	crdb.StatementHandler
}

func newIDPUserLinkProjection(ctx context.Context, config crdb.StatementHandlerConfig) *idpUserLinkProjection {
	p := new(idpUserLinkProjection)
	config.ProjectionName = IDPUserLinkTable
	config.Reducers = p.reducers()
	config.InitCheck = crdb.NewTableCheck(
		crdb.NewTable([]*crdb.Column{
			crdb.NewColumn(IDPUserLinkIDPIDCol, crdb.ColumnTypeText),
			crdb.NewColumn(IDPUserLinkUserIDCol, crdb.ColumnTypeText),
			crdb.NewColumn(IDPUserLinkExternalUserIDCol, crdb.ColumnTypeText),
			crdb.NewColumn(IDPUserLinkCreationDateCol, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(IDPUserLinkChangeDateCol, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(IDPUserLinkSequenceCol, crdb.ColumnTypeInt64),
			crdb.NewColumn(IDPUserLinkResourceOwnerCol, crdb.ColumnTypeText),
			crdb.NewColumn(IDPUserLinkInstanceIDCol, crdb.ColumnTypeText),
			crdb.NewColumn(IDPUserLinkDisplayNameCol, crdb.ColumnTypeText),
		},
			crdb.NewPrimaryKey(IDPUserLinkInstanceIDCol, IDPUserLinkIDPIDCol, IDPUserLinkExternalUserIDCol),
			crdb.WithIndex(crdb.NewIndex("idp_user_idx", []string{IDPUserLinkUserIDCol})),
		),
	)
	p.StatementHandler = crdb.NewStatementHandler(ctx, config)
	return p
}

func (p *idpUserLinkProjection) reducers() []handler.AggregateReducer {
	return []handler.AggregateReducer{
		{
			Aggregate: user.AggregateType,
			EventRedusers: []handler.EventReducer{
				{
					Event:  user.UserIDPLinkAddedType,
					Reduce: p.reduceAdded,
				},
				{
					Event:  user.UserIDPLinkCascadeRemovedType,
					Reduce: p.reduceCascadeRemoved,
				},
				{
					Event:  user.UserIDPLinkRemovedType,
					Reduce: p.reduceRemoved,
				},
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
					Event:  org.IDPConfigRemovedEventType,
					Reduce: p.reduceIDPConfigRemoved,
				},
				{
					Event:  org.OrgRemovedEventType,
					Reduce: p.reduceOrgRemoved,
				},
			},
		},
		{
			Aggregate: instance.AggregateType,
			EventRedusers: []handler.EventReducer{
				{
					Event:  instance.IDPConfigRemovedEventType,
					Reduce: p.reduceIDPConfigRemoved,
				},
			},
		},
	}
}

func (p *idpUserLinkProjection) reduceAdded(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*user.UserIDPLinkAddedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-DpmXq", "reduce.wrong.event.type %s", user.UserIDPLinkAddedType)
	}

	return crdb.NewCreateStatement(e,
		[]handler.Column{
			handler.NewCol(IDPUserLinkIDPIDCol, e.IDPConfigID),
			handler.NewCol(IDPUserLinkUserIDCol, e.Aggregate().ID),
			handler.NewCol(IDPUserLinkExternalUserIDCol, e.ExternalUserID),
			handler.NewCol(IDPUserLinkCreationDateCol, e.CreationDate()),
			handler.NewCol(IDPUserLinkChangeDateCol, e.CreationDate()),
			handler.NewCol(IDPUserLinkSequenceCol, e.Sequence()),
			handler.NewCol(IDPUserLinkResourceOwnerCol, e.Aggregate().ResourceOwner),
			handler.NewCol(IDPUserLinkInstanceIDCol, e.Aggregate().InstanceID),
			handler.NewCol(IDPUserLinkDisplayNameCol, e.DisplayName),
		},
	), nil
}

func (p *idpUserLinkProjection) reduceRemoved(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*user.UserIDPLinkRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-AZmfJ", "reduce.wrong.event.type %s", user.UserIDPLinkRemovedType)
	}

	return crdb.NewDeleteStatement(e,
		[]handler.Condition{
			handler.NewCond(IDPUserLinkIDPIDCol, e.IDPConfigID),
			handler.NewCond(IDPUserLinkUserIDCol, e.Aggregate().ID),
			handler.NewCond(IDPUserLinkExternalUserIDCol, e.ExternalUserID),
		},
	), nil
}

func (p *idpUserLinkProjection) reduceCascadeRemoved(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*user.UserIDPLinkCascadeRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-jQpv9", "reduce.wrong.event.type %s", user.UserIDPLinkCascadeRemovedType)
	}

	return crdb.NewDeleteStatement(e,
		[]handler.Condition{
			handler.NewCond(IDPUserLinkIDPIDCol, e.IDPConfigID),
			handler.NewCond(IDPUserLinkUserIDCol, e.Aggregate().ID),
			handler.NewCond(IDPUserLinkExternalUserIDCol, e.ExternalUserID),
		},
	), nil
}

func (p *idpUserLinkProjection) reduceOrgRemoved(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*org.OrgRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-AZmfJ", "reduce.wrong.event.type %s", org.OrgRemovedEventType)
	}

	return crdb.NewDeleteStatement(e,
		[]handler.Condition{
			handler.NewCond(IDPUserLinkResourceOwnerCol, e.Aggregate().ID),
		},
	), nil
}

func (p *idpUserLinkProjection) reduceUserRemoved(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*user.UserRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-uwlWE", "reduce.wrong.event.type %s", user.UserRemovedType)
	}

	return crdb.NewDeleteStatement(e,
		[]handler.Condition{
			handler.NewCond(IDPUserLinkUserIDCol, e.Aggregate().ID),
		},
	), nil
}

func (p *idpUserLinkProjection) reduceIDPConfigRemoved(event eventstore.Event) (*handler.Statement, error) {
	var idpID string

	switch e := event.(type) {
	case *org.IDPConfigRemovedEvent:
		idpID = e.ConfigID
	case *instance.IDPConfigRemovedEvent:
		idpID = e.ConfigID
	default:
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-iCKSj", "reduce.wrong.event.type %v", []eventstore.EventType{org.IDPConfigRemovedEventType, instance.IDPConfigRemovedEventType})
	}

	return crdb.NewDeleteStatement(event,
		[]handler.Condition{
			handler.NewCond(IDPUserLinkIDPIDCol, idpID),
			handler.NewCond(IDPUserLinkResourceOwnerCol, event.Aggregate().ResourceOwner),
		},
	), nil
}
