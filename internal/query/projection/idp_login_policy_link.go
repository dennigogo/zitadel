package projection

import (
	"context"

	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/handler"
	"github.com/dennigogo/zitadel/internal/eventstore/handler/crdb"
	"github.com/dennigogo/zitadel/internal/repository/instance"
	"github.com/dennigogo/zitadel/internal/repository/org"
	"github.com/dennigogo/zitadel/internal/repository/policy"
)

const (
	IDPLoginPolicyLinkTable = "projections.idp_login_policy_links3"

	IDPLoginPolicyLinkIDPIDCol         = "idp_id"
	IDPLoginPolicyLinkAggregateIDCol   = "aggregate_id"
	IDPLoginPolicyLinkCreationDateCol  = "creation_date"
	IDPLoginPolicyLinkChangeDateCol    = "change_date"
	IDPLoginPolicyLinkSequenceCol      = "sequence"
	IDPLoginPolicyLinkResourceOwnerCol = "resource_owner"
	IDPLoginPolicyLinkInstanceIDCol    = "instance_id"
	IDPLoginPolicyLinkProviderTypeCol  = "provider_type"
)

type idpLoginPolicyLinkProjection struct {
	crdb.StatementHandler
}

func newIDPLoginPolicyLinkProjection(ctx context.Context, config crdb.StatementHandlerConfig) *idpLoginPolicyLinkProjection {
	p := new(idpLoginPolicyLinkProjection)
	config.ProjectionName = IDPLoginPolicyLinkTable
	config.Reducers = p.reducers()
	config.InitCheck = crdb.NewTableCheck(
		crdb.NewTable([]*crdb.Column{
			crdb.NewColumn(IDPLoginPolicyLinkIDPIDCol, crdb.ColumnTypeText),
			crdb.NewColumn(IDPLoginPolicyLinkAggregateIDCol, crdb.ColumnTypeText),
			crdb.NewColumn(IDPLoginPolicyLinkCreationDateCol, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(IDPLoginPolicyLinkChangeDateCol, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(IDPLoginPolicyLinkSequenceCol, crdb.ColumnTypeInt64),
			crdb.NewColumn(IDPLoginPolicyLinkResourceOwnerCol, crdb.ColumnTypeText),
			crdb.NewColumn(IDPLoginPolicyLinkInstanceIDCol, crdb.ColumnTypeText),
			crdb.NewColumn(IDPLoginPolicyLinkProviderTypeCol, crdb.ColumnTypeEnum),
		},
			crdb.NewPrimaryKey(IDPLoginPolicyLinkInstanceIDCol, IDPLoginPolicyLinkAggregateIDCol, IDPLoginPolicyLinkIDPIDCol),
			crdb.WithIndex(crdb.NewIndex("link_ro_idx", []string{IDPLoginPolicyLinkResourceOwnerCol})),
		),
	)
	p.StatementHandler = crdb.NewStatementHandler(ctx, config)
	return p
}

func (p *idpLoginPolicyLinkProjection) reducers() []handler.AggregateReducer {
	return []handler.AggregateReducer{
		{
			Aggregate: org.AggregateType,
			EventRedusers: []handler.EventReducer{
				{
					Event:  org.LoginPolicyIDPProviderAddedEventType,
					Reduce: p.reduceAdded,
				},
				{
					Event:  org.LoginPolicyIDPProviderCascadeRemovedEventType,
					Reduce: p.reduceCascadeRemoved,
				},
				{
					Event:  org.LoginPolicyIDPProviderRemovedEventType,
					Reduce: p.reduceRemoved,
				},
				{
					Event:  org.LoginPolicyRemovedEventType,
					Reduce: p.reducePolicyRemoved,
				},
				{
					Event:  org.OrgRemovedEventType,
					Reduce: p.reduceOrgRemoved,
				},
				{
					Event:  org.IDPConfigRemovedEventType,
					Reduce: p.reduceIDPConfigRemoved,
				},
			},
		},
		{
			Aggregate: instance.AggregateType,
			EventRedusers: []handler.EventReducer{
				{
					Event:  instance.LoginPolicyIDPProviderAddedEventType,
					Reduce: p.reduceAdded,
				},
				{
					Event:  instance.LoginPolicyIDPProviderCascadeRemovedEventType,
					Reduce: p.reduceCascadeRemoved,
				},
				{
					Event:  instance.LoginPolicyIDPProviderRemovedEventType,
					Reduce: p.reduceRemoved,
				},
				{
					Event:  instance.IDPConfigRemovedEventType,
					Reduce: p.reduceIDPConfigRemoved,
				},
			},
		},
	}
}

func (p *idpLoginPolicyLinkProjection) reduceAdded(event eventstore.Event) (*handler.Statement, error) {
	var (
		idp          policy.IdentityProviderAddedEvent
		providerType domain.IdentityProviderType
	)

	switch e := event.(type) {
	case *org.IdentityProviderAddedEvent:
		idp = e.IdentityProviderAddedEvent
		providerType = domain.IdentityProviderTypeOrg
	case *instance.IdentityProviderAddedEvent:
		idp = e.IdentityProviderAddedEvent
		providerType = domain.IdentityProviderTypeSystem
	default:
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-Nlp55", "reduce.wrong.event.type %v", []eventstore.EventType{org.LoginPolicyIDPProviderAddedEventType, instance.LoginPolicyIDPProviderAddedEventType})
	}

	return crdb.NewCreateStatement(&idp,
		[]handler.Column{
			handler.NewCol(IDPLoginPolicyLinkIDPIDCol, idp.IDPConfigID),
			handler.NewCol(IDPLoginPolicyLinkAggregateIDCol, idp.Aggregate().ID),
			handler.NewCol(IDPLoginPolicyLinkCreationDateCol, idp.CreationDate()),
			handler.NewCol(IDPLoginPolicyLinkChangeDateCol, idp.CreationDate()),
			handler.NewCol(IDPLoginPolicyLinkSequenceCol, idp.Sequence()),
			handler.NewCol(IDPLoginPolicyLinkResourceOwnerCol, idp.Aggregate().ResourceOwner),
			handler.NewCol(IDPLoginPolicyLinkInstanceIDCol, idp.Aggregate().InstanceID),
			handler.NewCol(IDPLoginPolicyLinkProviderTypeCol, providerType),
		},
	), nil
}

func (p *idpLoginPolicyLinkProjection) reduceRemoved(event eventstore.Event) (*handler.Statement, error) {
	var idp policy.IdentityProviderRemovedEvent

	switch e := event.(type) {
	case *org.IdentityProviderRemovedEvent:
		idp = e.IdentityProviderRemovedEvent
	case *instance.IdentityProviderRemovedEvent:
		idp = e.IdentityProviderRemovedEvent
	default:
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-tUMYY", "reduce.wrong.event.type %v", []eventstore.EventType{org.LoginPolicyIDPProviderRemovedEventType, instance.LoginPolicyIDPProviderRemovedEventType})
	}

	return crdb.NewDeleteStatement(&idp,
		[]handler.Condition{
			handler.NewCond(IDPLoginPolicyLinkIDPIDCol, idp.IDPConfigID),
			handler.NewCond(IDPLoginPolicyLinkAggregateIDCol, idp.Aggregate().ID),
		},
	), nil
}

func (p *idpLoginPolicyLinkProjection) reduceCascadeRemoved(event eventstore.Event) (*handler.Statement, error) {
	var idp policy.IdentityProviderCascadeRemovedEvent

	switch e := event.(type) {
	case *org.IdentityProviderCascadeRemovedEvent:
		idp = e.IdentityProviderCascadeRemovedEvent
	case *instance.IdentityProviderCascadeRemovedEvent:
		idp = e.IdentityProviderCascadeRemovedEvent
	default:
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-iCKSj", "reduce.wrong.event.type %v", []eventstore.EventType{org.LoginPolicyIDPProviderCascadeRemovedEventType, instance.LoginPolicyIDPProviderCascadeRemovedEventType})
	}

	return crdb.NewDeleteStatement(&idp,
		[]handler.Condition{
			handler.NewCond(IDPLoginPolicyLinkIDPIDCol, idp.IDPConfigID),
			handler.NewCond(IDPLoginPolicyLinkAggregateIDCol, idp.Aggregate().ID),
		},
	), nil
}

func (p *idpLoginPolicyLinkProjection) reduceIDPConfigRemoved(event eventstore.Event) (*handler.Statement, error) {
	var idpID string

	switch e := event.(type) {
	case *org.IDPConfigRemovedEvent:
		idpID = e.ConfigID
	case *instance.IDPConfigRemovedEvent:
		idpID = e.ConfigID
	default:
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-u6tze", "reduce.wrong.event.type %v", []eventstore.EventType{org.IDPConfigRemovedEventType, instance.IDPConfigRemovedEventType})
	}

	return crdb.NewDeleteStatement(event,
		[]handler.Condition{
			handler.NewCond(IDPLoginPolicyLinkIDPIDCol, idpID),
			handler.NewCond(IDPLoginPolicyLinkResourceOwnerCol, event.Aggregate().ResourceOwner),
		},
	), nil
}

func (p *idpLoginPolicyLinkProjection) reduceOrgRemoved(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*org.OrgRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-QSoSe", "reduce.wrong.event.type %s", org.OrgRemovedEventType)
	}
	return crdb.NewDeleteStatement(e,
		[]handler.Condition{
			handler.NewCond(IDPLoginPolicyLinkResourceOwnerCol, e.Aggregate().ID),
		},
	), nil
}

func (p *idpLoginPolicyLinkProjection) reducePolicyRemoved(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*org.LoginPolicyRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-SF3dg", "reduce.wrong.event.type %s", org.LoginPolicyRemovedEventType)
	}
	return crdb.NewDeleteStatement(e,
		[]handler.Condition{
			handler.NewCond(IDPLoginPolicyLinkAggregateIDCol, e.Aggregate().ID),
		},
	), nil
}
