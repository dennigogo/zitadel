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
	PasswordAgeTable = "projections.password_age_policies"

	AgePolicyIDCol             = "id"
	AgePolicyCreationDateCol   = "creation_date"
	AgePolicyChangeDateCol     = "change_date"
	AgePolicySequenceCol       = "sequence"
	AgePolicyStateCol          = "state"
	AgePolicyIsDefaultCol      = "is_default"
	AgePolicyResourceOwnerCol  = "resource_owner"
	AgePolicyInstanceIDCol     = "instance_id"
	AgePolicyExpireWarnDaysCol = "expire_warn_days"
	AgePolicyMaxAgeDaysCol     = "max_age_days"
)

type passwordAgeProjection struct {
	crdb.StatementHandler
}

func newPasswordAgeProjection(ctx context.Context, config crdb.StatementHandlerConfig) *passwordAgeProjection {
	p := new(passwordAgeProjection)
	config.ProjectionName = PasswordAgeTable
	config.Reducers = p.reducers()
	config.InitCheck = crdb.NewTableCheck(
		crdb.NewTable([]*crdb.Column{
			crdb.NewColumn(AgePolicyIDCol, crdb.ColumnTypeText),
			crdb.NewColumn(AgePolicyCreationDateCol, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(AgePolicyChangeDateCol, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(AgePolicySequenceCol, crdb.ColumnTypeInt64),
			crdb.NewColumn(AgePolicyStateCol, crdb.ColumnTypeEnum),
			crdb.NewColumn(AgePolicyIsDefaultCol, crdb.ColumnTypeBool, crdb.Default(false)),
			crdb.NewColumn(AgePolicyResourceOwnerCol, crdb.ColumnTypeText),
			crdb.NewColumn(AgePolicyInstanceIDCol, crdb.ColumnTypeText),
			crdb.NewColumn(AgePolicyExpireWarnDaysCol, crdb.ColumnTypeInt64),
			crdb.NewColumn(AgePolicyMaxAgeDaysCol, crdb.ColumnTypeInt64),
		},
			crdb.NewPrimaryKey(AgePolicyInstanceIDCol, AgePolicyIDCol),
		),
	)
	p.StatementHandler = crdb.NewStatementHandler(ctx, config)
	return p
}

func (p *passwordAgeProjection) reducers() []handler.AggregateReducer {
	return []handler.AggregateReducer{
		{
			Aggregate: org.AggregateType,
			EventRedusers: []handler.EventReducer{
				{
					Event:  org.PasswordAgePolicyAddedEventType,
					Reduce: p.reduceAdded,
				},
				{
					Event:  org.PasswordAgePolicyChangedEventType,
					Reduce: p.reduceChanged,
				},
				{
					Event:  org.PasswordAgePolicyRemovedEventType,
					Reduce: p.reduceRemoved,
				},
			},
		},
		{
			Aggregate: instance.AggregateType,
			EventRedusers: []handler.EventReducer{
				{
					Event:  instance.PasswordAgePolicyAddedEventType,
					Reduce: p.reduceAdded,
				},
				{
					Event:  instance.PasswordAgePolicyChangedEventType,
					Reduce: p.reduceChanged,
				},
			},
		},
	}
}

func (p *passwordAgeProjection) reduceAdded(event eventstore.Event) (*handler.Statement, error) {
	var policyEvent policy.PasswordAgePolicyAddedEvent
	var isDefault bool
	switch e := event.(type) {
	case *org.PasswordAgePolicyAddedEvent:
		policyEvent = e.PasswordAgePolicyAddedEvent
		isDefault = false
	case *instance.PasswordAgePolicyAddedEvent:
		policyEvent = e.PasswordAgePolicyAddedEvent
		isDefault = true
	default:
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-CJqF0", "reduce.wrong.event.type %v", []eventstore.EventType{org.PasswordAgePolicyAddedEventType, instance.PasswordAgePolicyAddedEventType})
	}
	return crdb.NewCreateStatement(
		&policyEvent,
		[]handler.Column{
			handler.NewCol(AgePolicyCreationDateCol, policyEvent.CreationDate()),
			handler.NewCol(AgePolicyChangeDateCol, policyEvent.CreationDate()),
			handler.NewCol(AgePolicySequenceCol, policyEvent.Sequence()),
			handler.NewCol(AgePolicyIDCol, policyEvent.Aggregate().ID),
			handler.NewCol(AgePolicyStateCol, domain.PolicyStateActive),
			handler.NewCol(AgePolicyExpireWarnDaysCol, policyEvent.ExpireWarnDays),
			handler.NewCol(AgePolicyMaxAgeDaysCol, policyEvent.MaxAgeDays),
			handler.NewCol(AgePolicyIsDefaultCol, isDefault),
			handler.NewCol(AgePolicyResourceOwnerCol, policyEvent.Aggregate().ResourceOwner),
			handler.NewCol(AgePolicyInstanceIDCol, policyEvent.Aggregate().InstanceID),
		}), nil
}

func (p *passwordAgeProjection) reduceChanged(event eventstore.Event) (*handler.Statement, error) {
	var policyEvent policy.PasswordAgePolicyChangedEvent
	switch e := event.(type) {
	case *org.PasswordAgePolicyChangedEvent:
		policyEvent = e.PasswordAgePolicyChangedEvent
	case *instance.PasswordAgePolicyChangedEvent:
		policyEvent = e.PasswordAgePolicyChangedEvent
	default:
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-i7FZt", "reduce.wrong.event.type %v", []eventstore.EventType{org.PasswordAgePolicyChangedEventType, instance.PasswordAgePolicyChangedEventType})
	}
	cols := []handler.Column{
		handler.NewCol(AgePolicyChangeDateCol, policyEvent.CreationDate()),
		handler.NewCol(AgePolicySequenceCol, policyEvent.Sequence()),
	}
	if policyEvent.ExpireWarnDays != nil {
		cols = append(cols, handler.NewCol(AgePolicyExpireWarnDaysCol, *policyEvent.ExpireWarnDays))
	}
	if policyEvent.MaxAgeDays != nil {
		cols = append(cols, handler.NewCol(AgePolicyMaxAgeDaysCol, *policyEvent.MaxAgeDays))
	}
	return crdb.NewUpdateStatement(
		&policyEvent,
		cols,
		[]handler.Condition{
			handler.NewCond(AgePolicyIDCol, policyEvent.Aggregate().ID),
		}), nil
}

func (p *passwordAgeProjection) reduceRemoved(event eventstore.Event) (*handler.Statement, error) {
	policyEvent, ok := event.(*org.PasswordAgePolicyRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-EtHWB", "reduce.wrong.event.type %s", org.PasswordAgePolicyRemovedEventType)
	}
	return crdb.NewDeleteStatement(
		policyEvent,
		[]handler.Condition{
			handler.NewCond(AgePolicyIDCol, policyEvent.Aggregate().ID),
		}), nil
}
