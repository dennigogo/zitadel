package projection

import (
	"context"

	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/handler"
	"github.com/dennigogo/zitadel/internal/eventstore/handler/crdb"
	"github.com/dennigogo/zitadel/internal/repository/instance"
)

const (
	SMSConfigProjectionTable = "projections.sms_configs"
	SMSTwilioTable           = SMSConfigProjectionTable + "_" + smsTwilioTableSuffix

	SMSColumnID            = "id"
	SMSColumnAggregateID   = "aggregate_id"
	SMSColumnCreationDate  = "creation_date"
	SMSColumnChangeDate    = "change_date"
	SMSColumnSequence      = "sequence"
	SMSColumnState         = "state"
	SMSColumnResourceOwner = "resource_owner"
	SMSColumnInstanceID    = "instance_id"

	smsTwilioTableSuffix              = "twilio"
	SMSTwilioConfigColumnSMSID        = "sms_id"
	SMSTwilioColumnInstanceID         = "instance_id"
	SMSTwilioConfigColumnSID          = "sid"
	SMSTwilioConfigColumnSenderNumber = "sender_number"
	SMSTwilioConfigColumnToken        = "token"
)

type smsConfigProjection struct {
	crdb.StatementHandler
}

func newSMSConfigProjection(ctx context.Context, config crdb.StatementHandlerConfig) *smsConfigProjection {
	p := new(smsConfigProjection)
	config.ProjectionName = SMSConfigProjectionTable
	config.Reducers = p.reducers()
	config.InitCheck = crdb.NewMultiTableCheck(
		crdb.NewTable([]*crdb.Column{
			crdb.NewColumn(SMSColumnID, crdb.ColumnTypeText),
			crdb.NewColumn(SMSColumnAggregateID, crdb.ColumnTypeText),
			crdb.NewColumn(SMSColumnCreationDate, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(SMSColumnChangeDate, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(SMSColumnSequence, crdb.ColumnTypeInt64),
			crdb.NewColumn(SMSColumnState, crdb.ColumnTypeEnum),
			crdb.NewColumn(SMSColumnResourceOwner, crdb.ColumnTypeText),
			crdb.NewColumn(SMSColumnInstanceID, crdb.ColumnTypeText),
		},
			crdb.NewPrimaryKey(SMSColumnID, SMSColumnInstanceID),
		),
		crdb.NewSuffixedTable([]*crdb.Column{
			crdb.NewColumn(SMSTwilioConfigColumnSMSID, crdb.ColumnTypeText),
			crdb.NewColumn(SMSTwilioColumnInstanceID, crdb.ColumnTypeText),
			crdb.NewColumn(SMSTwilioConfigColumnSID, crdb.ColumnTypeText),
			crdb.NewColumn(SMSTwilioConfigColumnSenderNumber, crdb.ColumnTypeText),
			crdb.NewColumn(SMSTwilioConfigColumnToken, crdb.ColumnTypeJSONB),
		},
			crdb.NewPrimaryKey(SMSTwilioConfigColumnSMSID, SMSTwilioColumnInstanceID),
			smsTwilioTableSuffix,
			crdb.WithForeignKey(crdb.NewForeignKeyOfPublicKeys("fk_twilio_ref_sms")),
		),
	)
	p.StatementHandler = crdb.NewStatementHandler(ctx, config)
	return p
}

func (p *smsConfigProjection) reducers() []handler.AggregateReducer {
	return []handler.AggregateReducer{
		{
			Aggregate: instance.AggregateType,
			EventRedusers: []handler.EventReducer{
				{
					Event:  instance.SMSConfigTwilioAddedEventType,
					Reduce: p.reduceSMSConfigTwilioAdded,
				},
				{
					Event:  instance.SMSConfigTwilioChangedEventType,
					Reduce: p.reduceSMSConfigTwilioChanged,
				},
				{
					Event:  instance.SMSConfigTwilioTokenChangedEventType,
					Reduce: p.reduceSMSConfigTwilioTokenChanged,
				},
				{
					Event:  instance.SMSConfigActivatedEventType,
					Reduce: p.reduceSMSConfigActivated,
				},
				{
					Event:  instance.SMSConfigDeactivatedEventType,
					Reduce: p.reduceSMSConfigDeactivated,
				},
				{
					Event:  instance.SMSConfigRemovedEventType,
					Reduce: p.reduceSMSConfigRemoved,
				},
			},
		},
	}
}

func (p *smsConfigProjection) reduceSMSConfigTwilioAdded(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*instance.SMSConfigTwilioAddedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-s8efs", "reduce.wrong.event.type %s", instance.SMSConfigTwilioAddedEventType)
	}

	return crdb.NewMultiStatement(
		e,
		crdb.AddCreateStatement(
			[]handler.Column{
				handler.NewCol(SMSColumnID, e.ID),
				handler.NewCol(SMSColumnAggregateID, e.Aggregate().ID),
				handler.NewCol(SMSColumnCreationDate, e.CreationDate()),
				handler.NewCol(SMSColumnChangeDate, e.CreationDate()),
				handler.NewCol(SMSColumnResourceOwner, e.Aggregate().ResourceOwner),
				handler.NewCol(SMSColumnInstanceID, e.Aggregate().InstanceID),
				handler.NewCol(SMSColumnState, domain.SMSConfigStateInactive),
				handler.NewCol(SMSColumnSequence, e.Sequence()),
			},
		),
		crdb.AddCreateStatement(
			[]handler.Column{
				handler.NewCol(SMSTwilioConfigColumnSMSID, e.ID),
				handler.NewCol(SMSTwilioColumnInstanceID, e.Aggregate().InstanceID),
				handler.NewCol(SMSTwilioConfigColumnSID, e.SID),
				handler.NewCol(SMSTwilioConfigColumnToken, e.Token),
				handler.NewCol(SMSTwilioConfigColumnSenderNumber, e.SenderNumber),
			},
			crdb.WithTableSuffix(smsTwilioTableSuffix),
		),
	), nil
}

func (p *smsConfigProjection) reduceSMSConfigTwilioChanged(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*instance.SMSConfigTwilioChangedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-fi99F", "reduce.wrong.event.type %s", instance.SMSConfigTwilioChangedEventType)
	}
	columns := make([]handler.Column, 0)
	if e.SID != nil {
		columns = append(columns, handler.NewCol(SMSTwilioConfigColumnSID, *e.SID))
	}
	if e.SenderNumber != nil {
		columns = append(columns, handler.NewCol(SMSTwilioConfigColumnSenderNumber, *e.SenderNumber))
	}

	return crdb.NewMultiStatement(
		e,
		crdb.AddUpdateStatement(
			columns,
			[]handler.Condition{
				handler.NewCond(SMSTwilioConfigColumnSMSID, e.ID),
				handler.NewCond(SMSTwilioColumnInstanceID, e.Aggregate().InstanceID),
			},
			crdb.WithTableSuffix(smsTwilioTableSuffix),
		),
		crdb.AddUpdateStatement(
			[]handler.Column{
				handler.NewCol(SMSColumnChangeDate, e.CreationDate()),
				handler.NewCol(SMSColumnSequence, e.Sequence()),
			},
			[]handler.Condition{
				handler.NewCond(SMSColumnID, e.ID),
				handler.NewCond(SMSColumnInstanceID, e.Aggregate().InstanceID),
			},
		),
	), nil
}

func (p *smsConfigProjection) reduceSMSConfigTwilioTokenChanged(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*instance.SMSConfigTwilioTokenChangedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-fi99F", "reduce.wrong.event.type %s", instance.SMSConfigTwilioTokenChangedEventType)
	}
	columns := make([]handler.Column, 0)
	if e.Token != nil {
		columns = append(columns, handler.NewCol(SMSTwilioConfigColumnToken, e.Token))
	}

	return crdb.NewMultiStatement(
		e,
		crdb.AddUpdateStatement(
			columns,
			[]handler.Condition{
				handler.NewCond(SMSTwilioConfigColumnSMSID, e.ID),
				handler.NewCond(SMSTwilioColumnInstanceID, e.Aggregate().InstanceID),
			},
			crdb.WithTableSuffix(smsTwilioTableSuffix),
		),
		crdb.AddUpdateStatement(
			[]handler.Column{
				handler.NewCol(SMSColumnChangeDate, e.CreationDate()),
				handler.NewCol(SMSColumnSequence, e.Sequence()),
			},
			[]handler.Condition{
				handler.NewCond(SMSColumnID, e.ID),
				handler.NewCond(SMSColumnInstanceID, e.Aggregate().InstanceID),
			},
		),
	), nil
}

func (p *smsConfigProjection) reduceSMSConfigActivated(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*instance.SMSConfigActivatedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-fj9Ef", "reduce.wrong.event.type %s", instance.SMSConfigActivatedEventType)
	}
	return crdb.NewUpdateStatement(
		e,
		[]handler.Column{
			handler.NewCol(SMSColumnState, domain.SMSConfigStateActive),
			handler.NewCol(SMSColumnChangeDate, e.CreationDate()),
			handler.NewCol(SMSColumnSequence, e.Sequence()),
		},
		[]handler.Condition{
			handler.NewCond(SMSColumnID, e.ID),
			handler.NewCond(SMSColumnInstanceID, e.Aggregate().InstanceID),
		},
	), nil
}

func (p *smsConfigProjection) reduceSMSConfigDeactivated(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*instance.SMSConfigDeactivatedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-dj9Js", "reduce.wrong.event.type %s", instance.SMSConfigDeactivatedEventType)
	}
	return crdb.NewUpdateStatement(
		e,
		[]handler.Column{
			handler.NewCol(SMSColumnState, domain.SMSConfigStateInactive),
			handler.NewCol(SMSColumnChangeDate, e.CreationDate()),
			handler.NewCol(SMSColumnSequence, e.Sequence()),
		},
		[]handler.Condition{
			handler.NewCond(SMSColumnID, e.ID),
			handler.NewCond(SMSColumnInstanceID, e.Aggregate().InstanceID),
		},
	), nil
}

func (p *smsConfigProjection) reduceSMSConfigRemoved(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*instance.SMSConfigRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-s9JJf", "reduce.wrong.event.type %s", instance.SMSConfigRemovedEventType)
	}
	return crdb.NewDeleteStatement(
		e,
		[]handler.Condition{
			handler.NewCond(SMSColumnID, e.ID),
			handler.NewCond(SMSColumnInstanceID, e.Aggregate().InstanceID),
		},
	), nil
}
