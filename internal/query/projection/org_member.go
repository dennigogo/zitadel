package projection

import (
	"context"

	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/handler"
	"github.com/dennigogo/zitadel/internal/eventstore/handler/crdb"
	"github.com/dennigogo/zitadel/internal/repository/org"
	"github.com/dennigogo/zitadel/internal/repository/user"
)

const (
	OrgMemberProjectionTable = "projections.org_members2"
	OrgMemberOrgIDCol        = "org_id"
)

type orgMemberProjection struct {
	crdb.StatementHandler
}

func newOrgMemberProjection(ctx context.Context, config crdb.StatementHandlerConfig) *orgMemberProjection {
	p := new(orgMemberProjection)
	config.ProjectionName = OrgMemberProjectionTable
	config.Reducers = p.reducers()
	config.InitCheck = crdb.NewTableCheck(
		crdb.NewTable(
			append(memberColumns, crdb.NewColumn(OrgMemberOrgIDCol, crdb.ColumnTypeText)),
			crdb.NewPrimaryKey(MemberInstanceID, OrgMemberOrgIDCol, MemberUserIDCol),
			crdb.WithIndex(crdb.NewIndex("org_memb_user_idx", []string{MemberUserIDCol})),
		),
	)
	p.StatementHandler = crdb.NewStatementHandler(ctx, config)
	return p
}

func (p *orgMemberProjection) reducers() []handler.AggregateReducer {
	return []handler.AggregateReducer{
		{
			Aggregate: org.AggregateType,
			EventRedusers: []handler.EventReducer{
				{
					Event:  org.MemberAddedEventType,
					Reduce: p.reduceAdded,
				},
				{
					Event:  org.MemberChangedEventType,
					Reduce: p.reduceChanged,
				},
				{
					Event:  org.MemberCascadeRemovedEventType,
					Reduce: p.reduceCascadeRemoved,
				},
				{
					Event:  org.MemberRemovedEventType,
					Reduce: p.reduceRemoved,
				},
				{
					Event:  org.OrgRemovedEventType,
					Reduce: p.reduceOrgRemoved,
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
	}
}

func (p *orgMemberProjection) reduceAdded(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*org.MemberAddedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-uYq4r", "reduce.wrong.event.type %s", org.MemberAddedEventType)
	}
	return reduceMemberAdded(e.MemberAddedEvent, withMemberCol(OrgMemberOrgIDCol, e.Aggregate().ID))
}

func (p *orgMemberProjection) reduceChanged(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*org.MemberChangedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-Bg8oM", "reduce.wrong.event.type %s", org.MemberChangedEventType)
	}
	return reduceMemberChanged(e.MemberChangedEvent, withMemberCond(OrgMemberOrgIDCol, e.Aggregate().ID))
}

func (p *orgMemberProjection) reduceCascadeRemoved(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*org.MemberCascadeRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-4twP2", "reduce.wrong.event.type %s", org.MemberCascadeRemovedEventType)
	}
	return reduceMemberCascadeRemoved(e.MemberCascadeRemovedEvent, withMemberCond(OrgMemberOrgIDCol, e.Aggregate().ID))
}

func (p *orgMemberProjection) reduceRemoved(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*org.MemberRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-avatH", "reduce.wrong.event.type %s", org.MemberRemovedEventType)
	}
	return reduceMemberRemoved(e,
		withMemberCond(MemberUserIDCol, e.UserID),
		withMemberCond(OrgMemberOrgIDCol, e.Aggregate().ID),
	)
}

func (p *orgMemberProjection) reduceUserRemoved(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*user.UserRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-eBMqH", "reduce.wrong.event.type %s", user.UserRemovedType)
	}
	return reduceMemberRemoved(e, withMemberCond(MemberUserIDCol, e.Aggregate().ID))
}

func (p *orgMemberProjection) reduceOrgRemoved(event eventstore.Event) (*handler.Statement, error) {
	//TODO: as soon as org deletion is implemented:
	// Case: The user has resource owner A and an org has resource owner B
	// if org B deleted it works
	// if org A is deleted, the membership wouldn't be deleted
	e, ok := event.(*org.OrgRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-jnGAV", "reduce.wrong.event.type %s", org.OrgRemovedEventType)
	}
	return reduceMemberRemoved(e, withMemberCond(OrgMemberOrgIDCol, e.Aggregate().ID))
}
