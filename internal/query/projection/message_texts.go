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
	MessageTextTable = "projections.message_texts"

	MessageTextAggregateIDCol  = "aggregate_id"
	MessageTextInstanceIDCol   = "instance_id"
	MessageTextCreationDateCol = "creation_date"
	MessageTextChangeDateCol   = "change_date"
	MessageTextSequenceCol     = "sequence"
	MessageTextStateCol        = "state"
	MessageTextTypeCol         = "type"
	MessageTextLanguageCol     = "language"
	MessageTextTitleCol        = "title"
	MessageTextPreHeaderCol    = "pre_header"
	MessageTextSubjectCol      = "subject"
	MessageTextGreetingCol     = "greeting"
	MessageTextTextCol         = "text"
	MessageTextButtonTextCol   = "button_text"
	MessageTextFooterCol       = "footer_text"
)

type messageTextProjection struct {
	crdb.StatementHandler
}

func newMessageTextProjection(ctx context.Context, config crdb.StatementHandlerConfig) *messageTextProjection {
	p := new(messageTextProjection)
	config.ProjectionName = MessageTextTable
	config.Reducers = p.reducers()
	config.InitCheck = crdb.NewTableCheck(
		crdb.NewTable([]*crdb.Column{
			crdb.NewColumn(MessageTextAggregateIDCol, crdb.ColumnTypeText),
			crdb.NewColumn(MessageTextInstanceIDCol, crdb.ColumnTypeText),
			crdb.NewColumn(MessageTextCreationDateCol, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(MessageTextChangeDateCol, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(MessageTextSequenceCol, crdb.ColumnTypeInt64),
			crdb.NewColumn(MessageTextStateCol, crdb.ColumnTypeEnum),
			crdb.NewColumn(MessageTextTypeCol, crdb.ColumnTypeText),
			crdb.NewColumn(MessageTextLanguageCol, crdb.ColumnTypeText),
			crdb.NewColumn(MessageTextTitleCol, crdb.ColumnTypeText, crdb.Nullable()),
			crdb.NewColumn(MessageTextPreHeaderCol, crdb.ColumnTypeText, crdb.Nullable()),
			crdb.NewColumn(MessageTextSubjectCol, crdb.ColumnTypeText, crdb.Nullable()),
			crdb.NewColumn(MessageTextGreetingCol, crdb.ColumnTypeText, crdb.Nullable()),
			crdb.NewColumn(MessageTextTextCol, crdb.ColumnTypeText, crdb.Nullable()),
			crdb.NewColumn(MessageTextButtonTextCol, crdb.ColumnTypeText, crdb.Nullable()),
			crdb.NewColumn(MessageTextFooterCol, crdb.ColumnTypeText, crdb.Nullable()),
		},
			crdb.NewPrimaryKey(MessageTextInstanceIDCol, MessageTextAggregateIDCol, MessageTextTypeCol, MessageTextLanguageCol),
		),
	)
	p.StatementHandler = crdb.NewStatementHandler(ctx, config)
	return p
}

func (p *messageTextProjection) reducers() []handler.AggregateReducer {
	return []handler.AggregateReducer{
		{
			Aggregate: org.AggregateType,
			EventRedusers: []handler.EventReducer{
				{
					Event:  org.CustomTextSetEventType,
					Reduce: p.reduceAdded,
				},
				{
					Event:  org.CustomTextRemovedEventType,
					Reduce: p.reduceRemoved,
				},
				{
					Event:  org.CustomTextTemplateRemovedEventType,
					Reduce: p.reduceTemplateRemoved,
				},
			},
		},
		{
			Aggregate: instance.AggregateType,
			EventRedusers: []handler.EventReducer{
				{
					Event:  instance.CustomTextSetEventType,
					Reduce: p.reduceAdded,
				},
				{
					Event:  instance.CustomTextRemovedEventType,
					Reduce: p.reduceRemoved,
				},
				{
					Event:  instance.CustomTextTemplateRemovedEventType,
					Reduce: p.reduceTemplateRemoved,
				},
			},
		},
	}
}

func (p *messageTextProjection) reduceAdded(event eventstore.Event) (*handler.Statement, error) {
	var templateEvent policy.CustomTextSetEvent
	switch e := event.(type) {
	case *org.CustomTextSetEvent:
		templateEvent = e.CustomTextSetEvent
	case *instance.CustomTextSetEvent:
		templateEvent = e.CustomTextSetEvent
	default:
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-2n90r", "reduce.wrong.event.type %v", []eventstore.EventType{org.CustomTextSetEventType, instance.CustomTextSetEventType})
	}
	if !isMessageTemplate(templateEvent.Template) {
		return crdb.NewNoOpStatement(event), nil
	}

	cols := []handler.Column{
		handler.NewCol(MessageTextAggregateIDCol, templateEvent.Aggregate().ID),
		handler.NewCol(MessageTextInstanceIDCol, templateEvent.Aggregate().InstanceID),
		handler.NewCol(MessageTextCreationDateCol, templateEvent.CreationDate()),
		handler.NewCol(MessageTextChangeDateCol, templateEvent.CreationDate()),
		handler.NewCol(MessageTextSequenceCol, templateEvent.Sequence()),
		handler.NewCol(MessageTextStateCol, domain.PolicyStateActive),
		handler.NewCol(MessageTextTypeCol, templateEvent.Template),
		handler.NewCol(MessageTextLanguageCol, templateEvent.Language.String()),
	}
	if isTitle(templateEvent.Key) {
		cols = append(cols, handler.NewCol(MessageTextTitleCol, templateEvent.Text))
	}
	if isPreHeader(templateEvent.Key) {
		cols = append(cols, handler.NewCol(MessageTextPreHeaderCol, templateEvent.Text))
	}
	if isSubject(templateEvent.Key) {
		cols = append(cols, handler.NewCol(MessageTextSubjectCol, templateEvent.Text))
	}
	if isGreeting(templateEvent.Key) {
		cols = append(cols, handler.NewCol(MessageTextGreetingCol, templateEvent.Text))
	}
	if isText(templateEvent.Key) {
		cols = append(cols, handler.NewCol(MessageTextTextCol, templateEvent.Text))
	}
	if isButtonText(templateEvent.Key) {
		cols = append(cols, handler.NewCol(MessageTextButtonTextCol, templateEvent.Text))
	}
	if isFooterText(templateEvent.Key) {
		cols = append(cols, handler.NewCol(MessageTextFooterCol, templateEvent.Text))
	}
	return crdb.NewUpsertStatement(
		&templateEvent,
		[]handler.Column{
			handler.NewCol(MessageTextInstanceIDCol, nil),
			handler.NewCol(MessageTextAggregateIDCol, nil),
			handler.NewCol(MessageTextTypeCol, nil),
			handler.NewCol(MessageTextLanguageCol, nil),
		},
		cols,
	), nil
}

func (p *messageTextProjection) reduceRemoved(event eventstore.Event) (*handler.Statement, error) {
	var templateEvent policy.CustomTextRemovedEvent
	switch e := event.(type) {
	case *org.CustomTextRemovedEvent:
		templateEvent = e.CustomTextRemovedEvent
	case *instance.CustomTextRemovedEvent:
		templateEvent = e.CustomTextRemovedEvent
	default:
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-fm0ge", "reduce.wrong.event.type %v", []eventstore.EventType{org.CustomTextRemovedEventType, instance.CustomTextRemovedEventType})
	}
	if !isMessageTemplate(templateEvent.Template) {
		return crdb.NewNoOpStatement(event), nil
	}
	cols := []handler.Column{
		handler.NewCol(MessageTextChangeDateCol, templateEvent.CreationDate()),
		handler.NewCol(MessageTextSequenceCol, templateEvent.Sequence()),
	}
	if isTitle(templateEvent.Key) {
		cols = append(cols, handler.NewCol(MessageTextTitleCol, ""))
	}
	if isPreHeader(templateEvent.Key) {
		cols = append(cols, handler.NewCol(MessageTextPreHeaderCol, ""))
	}
	if isSubject(templateEvent.Key) {
		cols = append(cols, handler.NewCol(MessageTextSubjectCol, ""))
	}
	if isGreeting(templateEvent.Key) {
		cols = append(cols, handler.NewCol(MessageTextGreetingCol, ""))
	}
	if isText(templateEvent.Key) {
		cols = append(cols, handler.NewCol(MessageTextTextCol, ""))
	}
	if isButtonText(templateEvent.Key) {
		cols = append(cols, handler.NewCol(MessageTextButtonTextCol, ""))
	}
	if isFooterText(templateEvent.Key) {
		cols = append(cols, handler.NewCol(MessageTextFooterCol, ""))
	}
	return crdb.NewUpdateStatement(
		&templateEvent,
		cols,
		[]handler.Condition{
			handler.NewCond(MessageTextAggregateIDCol, templateEvent.Aggregate().ID),
			handler.NewCond(MessageTextTypeCol, templateEvent.Template),
			handler.NewCond(MessageTextLanguageCol, templateEvent.Language.String()),
		},
	), nil
}

func (p *messageTextProjection) reduceTemplateRemoved(event eventstore.Event) (*handler.Statement, error) {
	var templateEvent policy.CustomTextTemplateRemovedEvent
	switch e := event.(type) {
	case *org.CustomTextTemplateRemovedEvent:
		templateEvent = e.CustomTextTemplateRemovedEvent
	case *instance.CustomTextTemplateRemovedEvent:
		templateEvent = e.CustomTextTemplateRemovedEvent
	default:
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-2n9rs", "reduce.wrong.event.type %s", org.CustomTextTemplateRemovedEventType)
	}
	if !isMessageTemplate(templateEvent.Template) {
		return crdb.NewNoOpStatement(event), nil
	}
	return crdb.NewDeleteStatement(
		event,
		[]handler.Condition{
			handler.NewCond(MessageTextAggregateIDCol, templateEvent.Aggregate().ID),
			handler.NewCond(MessageTextTypeCol, templateEvent.Template),
			handler.NewCond(MessageTextLanguageCol, templateEvent.Language.String()),
		},
	), nil
}

func isMessageTemplate(template string) bool {
	return template == domain.InitCodeMessageType ||
		template == domain.PasswordResetMessageType ||
		template == domain.VerifyEmailMessageType ||
		template == domain.VerifyPhoneMessageType ||
		template == domain.DomainClaimedMessageType ||
		template == domain.PasswordlessRegistrationMessageType
}
func isTitle(key string) bool {
	return key == domain.MessageTitle
}
func isPreHeader(key string) bool {
	return key == domain.MessagePreHeader
}
func isSubject(key string) bool {
	return key == domain.MessageSubject
}
func isGreeting(key string) bool {
	return key == domain.MessageGreeting
}
func isText(key string) bool {
	return key == domain.MessageText
}
func isButtonText(key string) bool {
	return key == domain.MessageButtonText
}
func isFooterText(key string) bool {
	return key == domain.MessageFooterText
}
