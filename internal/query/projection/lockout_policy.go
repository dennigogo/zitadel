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
	LockoutPolicyTable = "projections.lockout_policies"

	LockoutPolicyIDCol                  = "id"
	LockoutPolicyCreationDateCol        = "creation_date"
	LockoutPolicyChangeDateCol          = "change_date"
	LockoutPolicySequenceCol            = "sequence"
	LockoutPolicyStateCol               = "state"
	LockoutPolicyIsDefaultCol           = "is_default"
	LockoutPolicyResourceOwnerCol       = "resource_owner"
	LockoutPolicyInstanceIDCol          = "instance_id"
	LockoutPolicyMaxPasswordAttemptsCol = "max_password_attempts"
	LockoutPolicyShowLockOutFailuresCol = "show_failure"
)

type lockoutPolicyProjection struct {
	crdb.StatementHandler
}

func newLockoutPolicyProjection(ctx context.Context, config crdb.StatementHandlerConfig) *lockoutPolicyProjection {
	p := new(lockoutPolicyProjection)
	config.ProjectionName = LockoutPolicyTable
	config.Reducers = p.reducers()
	config.InitCheck = crdb.NewTableCheck(
		crdb.NewTable([]*crdb.Column{
			crdb.NewColumn(LockoutPolicyIDCol, crdb.ColumnTypeText),
			crdb.NewColumn(LockoutPolicyCreationDateCol, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(LockoutPolicyChangeDateCol, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(LockoutPolicySequenceCol, crdb.ColumnTypeInt64),
			crdb.NewColumn(LockoutPolicyStateCol, crdb.ColumnTypeEnum),
			crdb.NewColumn(LockoutPolicyIsDefaultCol, crdb.ColumnTypeBool, crdb.Default(false)),
			crdb.NewColumn(LockoutPolicyResourceOwnerCol, crdb.ColumnTypeText),
			crdb.NewColumn(LockoutPolicyInstanceIDCol, crdb.ColumnTypeText),
			crdb.NewColumn(LockoutPolicyMaxPasswordAttemptsCol, crdb.ColumnTypeInt64),
			crdb.NewColumn(LockoutPolicyShowLockOutFailuresCol, crdb.ColumnTypeBool),
		},
			crdb.NewPrimaryKey(LockoutPolicyInstanceIDCol, LockoutPolicyIDCol),
		),
	)
	p.StatementHandler = crdb.NewStatementHandler(ctx, config)
	return p
}

func (p *lockoutPolicyProjection) reducers() []handler.AggregateReducer {
	return []handler.AggregateReducer{
		{
			Aggregate: org.AggregateType,
			EventRedusers: []handler.EventReducer{
				{
					Event:  org.LockoutPolicyAddedEventType,
					Reduce: p.reduceAdded,
				},
				{
					Event:  org.LockoutPolicyChangedEventType,
					Reduce: p.reduceChanged,
				},
				{
					Event:  org.LockoutPolicyRemovedEventType,
					Reduce: p.reduceRemoved,
				},
			},
		},
		{
			Aggregate: instance.AggregateType,
			EventRedusers: []handler.EventReducer{
				{
					Event:  instance.LockoutPolicyAddedEventType,
					Reduce: p.reduceAdded,
				},
				{
					Event:  instance.LockoutPolicyChangedEventType,
					Reduce: p.reduceChanged,
				},
			},
		},
	}
}

func (p *lockoutPolicyProjection) reduceAdded(event eventstore.Event) (*handler.Statement, error) {
	var policyEvent policy.LockoutPolicyAddedEvent
	var isDefault bool
	switch e := event.(type) {
	case *org.LockoutPolicyAddedEvent:
		policyEvent = e.LockoutPolicyAddedEvent
		isDefault = false
	case *instance.LockoutPolicyAddedEvent:
		policyEvent = e.LockoutPolicyAddedEvent
		isDefault = true
	default:
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-d8mZO", "reduce.wrong.event.type, %v", []eventstore.EventType{org.LockoutPolicyAddedEventType, instance.LockoutPolicyAddedEventType})
	}
	return crdb.NewCreateStatement(
		&policyEvent,
		[]handler.Column{
			handler.NewCol(LockoutPolicyCreationDateCol, policyEvent.CreationDate()),
			handler.NewCol(LockoutPolicyChangeDateCol, policyEvent.CreationDate()),
			handler.NewCol(LockoutPolicySequenceCol, policyEvent.Sequence()),
			handler.NewCol(LockoutPolicyIDCol, policyEvent.Aggregate().ID),
			handler.NewCol(LockoutPolicyStateCol, domain.PolicyStateActive),
			handler.NewCol(LockoutPolicyMaxPasswordAttemptsCol, policyEvent.MaxPasswordAttempts),
			handler.NewCol(LockoutPolicyShowLockOutFailuresCol, policyEvent.ShowLockOutFailures),
			handler.NewCol(LockoutPolicyIsDefaultCol, isDefault),
			handler.NewCol(LockoutPolicyResourceOwnerCol, policyEvent.Aggregate().ResourceOwner),
			handler.NewCol(LockoutPolicyInstanceIDCol, policyEvent.Aggregate().InstanceID),
		}), nil
}

func (p *lockoutPolicyProjection) reduceChanged(event eventstore.Event) (*handler.Statement, error) {
	var policyEvent policy.LockoutPolicyChangedEvent
	switch e := event.(type) {
	case *org.LockoutPolicyChangedEvent:
		policyEvent = e.LockoutPolicyChangedEvent
	case *instance.LockoutPolicyChangedEvent:
		policyEvent = e.LockoutPolicyChangedEvent
	default:
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-pT3mQ", "reduce.wrong.event.type, %v", []eventstore.EventType{org.LockoutPolicyChangedEventType, instance.LockoutPolicyChangedEventType})
	}
	cols := []handler.Column{
		handler.NewCol(LockoutPolicyChangeDateCol, policyEvent.CreationDate()),
		handler.NewCol(LockoutPolicySequenceCol, policyEvent.Sequence()),
	}
	if policyEvent.MaxPasswordAttempts != nil {
		cols = append(cols, handler.NewCol(LockoutPolicyMaxPasswordAttemptsCol, *policyEvent.MaxPasswordAttempts))
	}
	if policyEvent.ShowLockOutFailures != nil {
		cols = append(cols, handler.NewCol(LockoutPolicyShowLockOutFailuresCol, *policyEvent.ShowLockOutFailures))
	}
	return crdb.NewUpdateStatement(
		&policyEvent,
		cols,
		[]handler.Condition{
			handler.NewCond(LockoutPolicyIDCol, policyEvent.Aggregate().ID),
		}), nil
}

func (p *lockoutPolicyProjection) reduceRemoved(event eventstore.Event) (*handler.Statement, error) {
	policyEvent, ok := event.(*org.LockoutPolicyRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-Bqut9", "reduce.wrong.event.type %s", org.LockoutPolicyRemovedEventType)
	}
	return crdb.NewDeleteStatement(
		policyEvent,
		[]handler.Condition{
			handler.NewCond(LockoutPolicyIDCol, policyEvent.Aggregate().ID),
		}), nil
}
