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
	PasswordComplexityTable = "projections.password_complexity_policies"

	ComplexityPolicyIDCol            = "id"
	ComplexityPolicyCreationDateCol  = "creation_date"
	ComplexityPolicyChangeDateCol    = "change_date"
	ComplexityPolicySequenceCol      = "sequence"
	ComplexityPolicyStateCol         = "state"
	ComplexityPolicyIsDefaultCol     = "is_default"
	ComplexityPolicyResourceOwnerCol = "resource_owner"
	ComplexityPolicyInstanceIDCol    = "instance_id"
	ComplexityPolicyMinLengthCol     = "min_length"
	ComplexityPolicyHasLowercaseCol  = "has_lowercase"
	ComplexityPolicyHasUppercaseCol  = "has_uppercase"
	ComplexityPolicyHasSymbolCol     = "has_symbol"
	ComplexityPolicyHasNumberCol     = "has_number"
)

type passwordComplexityProjection struct {
	crdb.StatementHandler
}

func newPasswordComplexityProjection(ctx context.Context, config crdb.StatementHandlerConfig) *passwordComplexityProjection {
	p := new(passwordComplexityProjection)
	config.ProjectionName = PasswordComplexityTable
	config.Reducers = p.reducers()
	config.InitCheck = crdb.NewTableCheck(
		crdb.NewTable([]*crdb.Column{
			crdb.NewColumn(ComplexityPolicyIDCol, crdb.ColumnTypeText),
			crdb.NewColumn(ComplexityPolicyCreationDateCol, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(ComplexityPolicyChangeDateCol, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(ComplexityPolicySequenceCol, crdb.ColumnTypeInt64),
			crdb.NewColumn(ComplexityPolicyStateCol, crdb.ColumnTypeEnum),
			crdb.NewColumn(ComplexityPolicyIsDefaultCol, crdb.ColumnTypeBool, crdb.Default(false)),
			crdb.NewColumn(ComplexityPolicyResourceOwnerCol, crdb.ColumnTypeText),
			crdb.NewColumn(ComplexityPolicyInstanceIDCol, crdb.ColumnTypeText),
			crdb.NewColumn(ComplexityPolicyMinLengthCol, crdb.ColumnTypeInt64),
			crdb.NewColumn(ComplexityPolicyHasLowercaseCol, crdb.ColumnTypeBool),
			crdb.NewColumn(ComplexityPolicyHasUppercaseCol, crdb.ColumnTypeBool),
			crdb.NewColumn(ComplexityPolicyHasSymbolCol, crdb.ColumnTypeBool),
			crdb.NewColumn(ComplexityPolicyHasNumberCol, crdb.ColumnTypeBool),
		},
			crdb.NewPrimaryKey(ComplexityPolicyInstanceIDCol, ComplexityPolicyIDCol),
		),
	)
	p.StatementHandler = crdb.NewStatementHandler(ctx, config)
	return p
}

func (p *passwordComplexityProjection) reducers() []handler.AggregateReducer {
	return []handler.AggregateReducer{
		{
			Aggregate: org.AggregateType,
			EventRedusers: []handler.EventReducer{
				{
					Event:  org.PasswordComplexityPolicyAddedEventType,
					Reduce: p.reduceAdded,
				},
				{
					Event:  org.PasswordComplexityPolicyChangedEventType,
					Reduce: p.reduceChanged,
				},
				{
					Event:  org.PasswordComplexityPolicyRemovedEventType,
					Reduce: p.reduceRemoved,
				},
			},
		},
		{
			Aggregate: instance.AggregateType,
			EventRedusers: []handler.EventReducer{
				{
					Event:  instance.PasswordComplexityPolicyAddedEventType,
					Reduce: p.reduceAdded,
				},
				{
					Event:  instance.PasswordComplexityPolicyChangedEventType,
					Reduce: p.reduceChanged,
				},
			},
		},
	}
}

func (p *passwordComplexityProjection) reduceAdded(event eventstore.Event) (*handler.Statement, error) {
	var policyEvent policy.PasswordComplexityPolicyAddedEvent
	var isDefault bool
	switch e := event.(type) {
	case *org.PasswordComplexityPolicyAddedEvent:
		policyEvent = e.PasswordComplexityPolicyAddedEvent
		isDefault = false
	case *instance.PasswordComplexityPolicyAddedEvent:
		policyEvent = e.PasswordComplexityPolicyAddedEvent
		isDefault = true
	default:
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-KTHmJ", "reduce.wrong.event.type %v", []eventstore.EventType{org.PasswordComplexityPolicyAddedEventType, instance.PasswordComplexityPolicyAddedEventType})
	}
	return crdb.NewCreateStatement(
		&policyEvent,
		[]handler.Column{
			handler.NewCol(ComplexityPolicyCreationDateCol, policyEvent.CreationDate()),
			handler.NewCol(ComplexityPolicyChangeDateCol, policyEvent.CreationDate()),
			handler.NewCol(ComplexityPolicySequenceCol, policyEvent.Sequence()),
			handler.NewCol(ComplexityPolicyIDCol, policyEvent.Aggregate().ID),
			handler.NewCol(ComplexityPolicyStateCol, domain.PolicyStateActive),
			handler.NewCol(ComplexityPolicyMinLengthCol, policyEvent.MinLength),
			handler.NewCol(ComplexityPolicyHasLowercaseCol, policyEvent.HasLowercase),
			handler.NewCol(ComplexityPolicyHasUppercaseCol, policyEvent.HasUppercase),
			handler.NewCol(ComplexityPolicyHasSymbolCol, policyEvent.HasSymbol),
			handler.NewCol(ComplexityPolicyHasNumberCol, policyEvent.HasNumber),
			handler.NewCol(ComplexityPolicyResourceOwnerCol, policyEvent.Aggregate().ResourceOwner),
			handler.NewCol(ComplexityPolicyInstanceIDCol, policyEvent.Aggregate().InstanceID),
			handler.NewCol(ComplexityPolicyIsDefaultCol, isDefault),
		}), nil
}

func (p *passwordComplexityProjection) reduceChanged(event eventstore.Event) (*handler.Statement, error) {
	var policyEvent policy.PasswordComplexityPolicyChangedEvent
	switch e := event.(type) {
	case *org.PasswordComplexityPolicyChangedEvent:
		policyEvent = e.PasswordComplexityPolicyChangedEvent
	case *instance.PasswordComplexityPolicyChangedEvent:
		policyEvent = e.PasswordComplexityPolicyChangedEvent
	default:
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-cf3Xb", "reduce.wrong.event.type %v", []eventstore.EventType{org.PasswordComplexityPolicyChangedEventType, instance.PasswordComplexityPolicyChangedEventType})
	}
	cols := []handler.Column{
		handler.NewCol(ComplexityPolicyChangeDateCol, policyEvent.CreationDate()),
		handler.NewCol(ComplexityPolicySequenceCol, policyEvent.Sequence()),
	}
	if policyEvent.MinLength != nil {
		cols = append(cols, handler.NewCol(ComplexityPolicyMinLengthCol, *policyEvent.MinLength))
	}
	if policyEvent.HasLowercase != nil {
		cols = append(cols, handler.NewCol(ComplexityPolicyHasLowercaseCol, *policyEvent.HasLowercase))
	}
	if policyEvent.HasUppercase != nil {
		cols = append(cols, handler.NewCol(ComplexityPolicyHasUppercaseCol, *policyEvent.HasUppercase))
	}
	if policyEvent.HasSymbol != nil {
		cols = append(cols, handler.NewCol(ComplexityPolicyHasSymbolCol, *policyEvent.HasSymbol))
	}
	if policyEvent.HasNumber != nil {
		cols = append(cols, handler.NewCol(ComplexityPolicyHasNumberCol, *policyEvent.HasNumber))
	}
	return crdb.NewUpdateStatement(
		&policyEvent,
		cols,
		[]handler.Condition{
			handler.NewCond(ComplexityPolicyIDCol, policyEvent.Aggregate().ID),
		}), nil
}

func (p *passwordComplexityProjection) reduceRemoved(event eventstore.Event) (*handler.Statement, error) {
	policyEvent, ok := event.(*org.PasswordComplexityPolicyRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-wttCd", "reduce.wrong.event.type %s", org.PasswordComplexityPolicyRemovedEventType)
	}
	return crdb.NewDeleteStatement(
		policyEvent,
		[]handler.Condition{
			handler.NewCond(ComplexityPolicyIDCol, policyEvent.Aggregate().ID),
		}), nil
}
