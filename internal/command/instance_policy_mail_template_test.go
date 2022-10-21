package command

import (
	"context"
	"testing"

	"github.com/dennigogo/zitadel/internal/api/authz"
	"github.com/dennigogo/zitadel/internal/domain"
	caos_errs "github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
	"github.com/dennigogo/zitadel/internal/eventstore/v1/models"
	"github.com/dennigogo/zitadel/internal/repository/instance"
	"github.com/dennigogo/zitadel/internal/repository/policy"
	"github.com/stretchr/testify/assert"
)

func TestCommandSide_AddDefaultMailTemplatePolicy(t *testing.T) {
	type fields struct {
		eventstore *eventstore.Eventstore
	}
	type args struct {
		ctx    context.Context
		policy *domain.MailTemplate
	}
	type res struct {
		want *domain.MailTemplate
		err  func(error) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		res    res
	}{
		{
			name: "mailtemplate invalid, invalid argument error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
				),
			},
			args: args{
				ctx:    context.Background(),
				policy: &domain.MailTemplate{},
			},
			res: res{
				err: caos_errs.IsErrorInvalidArgument,
			},
		},
		{
			name: "mailtemplate already existing, already exists error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							instance.NewMailTemplateAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								[]byte("template"),
							),
						),
					),
				),
			},
			args: args{
				ctx: context.Background(),
				policy: &domain.MailTemplate{
					Template: []byte("template"),
				},
			},
			res: res{
				err: caos_errs.IsErrorAlreadyExists,
			},
		},
		{
			name: "add mail template,ok",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(),
					expectPush(
						[]*repository.Event{
							eventFromEventPusherWithInstanceID(
								"INSTANCE",
								instance.NewMailTemplateAddedEvent(context.Background(),
									&instance.NewAggregate("INSTANCE").Aggregate,
									[]byte("template"),
								),
							),
						},
					),
				),
			},
			args: args{
				ctx: authz.WithInstanceID(context.Background(), "INSTANCE"),
				policy: &domain.MailTemplate{
					Template: []byte("template"),
				},
			},
			res: res{
				want: &domain.MailTemplate{
					ObjectRoot: models.ObjectRoot{
						InstanceID:    "INSTANCE",
						AggregateID:   "INSTANCE",
						ResourceOwner: "INSTANCE",
					},
					Template: []byte("template"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Commands{
				eventstore: tt.fields.eventstore,
			}
			got, err := r.AddDefaultMailTemplate(tt.args.ctx, tt.args.policy)
			if tt.res.err == nil {
				assert.NoError(t, err)
			}
			if tt.res.err != nil && !tt.res.err(err) {
				t.Errorf("got wrong err: %v ", err)
			}
			if tt.res.err == nil {
				assert.Equal(t, tt.res.want, got)
			}
		})
	}
}

func TestCommandSide_ChangeDefaultMailTemplatePolicy(t *testing.T) {
	type fields struct {
		eventstore *eventstore.Eventstore
	}
	type args struct {
		ctx    context.Context
		policy *domain.MailTemplate
	}
	type res struct {
		want *domain.MailTemplate
		err  func(error) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		res    res
	}{
		{
			name: "mailtemplate invalid, invalid argument error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
				),
			},
			args: args{
				ctx:    context.Background(),
				policy: &domain.MailTemplate{},
			},
			res: res{
				err: caos_errs.IsErrorInvalidArgument,
			},
		},
		{
			name: "mailtempalte not existing, not found error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(),
				),
			},
			args: args{
				ctx: context.Background(),
				policy: &domain.MailTemplate{
					Template: []byte("template-change"),
				},
			},
			res: res{
				err: caos_errs.IsNotFound,
			},
		},
		{
			name: "no changes, precondition error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							instance.NewMailTemplateAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								[]byte("template"),
							),
						),
					),
				),
			},
			args: args{
				ctx: context.Background(),
				policy: &domain.MailTemplate{
					Template: []byte("template"),
				},
			},
			res: res{
				err: caos_errs.IsPreconditionFailed,
			},
		},
		{
			name: "change, ok",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							instance.NewMailTemplateAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								[]byte("template"),
							),
						),
					),
					expectPush(
						[]*repository.Event{
							eventFromEventPusher(
								newDefaultMailTemplatePolicyChangedEvent(context.Background(), []byte("template-change")),
							),
						},
					),
				),
			},
			args: args{
				ctx: context.Background(),
				policy: &domain.MailTemplate{
					Template: []byte("template-change"),
				},
			},
			res: res{
				want: &domain.MailTemplate{
					ObjectRoot: models.ObjectRoot{
						AggregateID:   "INSTANCE",
						ResourceOwner: "INSTANCE",
					},
					Template: []byte("template-change"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Commands{
				eventstore: tt.fields.eventstore,
			}
			got, err := r.ChangeDefaultMailTemplate(tt.args.ctx, tt.args.policy)
			if tt.res.err == nil {
				assert.NoError(t, err)
			}
			if tt.res.err != nil && !tt.res.err(err) {
				t.Errorf("got wrong err: %v ", err)
			}
			if tt.res.err == nil {
				assert.Equal(t, tt.res.want, got)
			}
		})
	}
}

func newDefaultMailTemplatePolicyChangedEvent(ctx context.Context, template []byte) *instance.MailTemplateChangedEvent {
	event, _ := instance.NewMailTemplateChangedEvent(ctx,
		&instance.NewAggregate("INSTANCE").Aggregate,
		[]policy.MailTemplateChanges{
			policy.ChangeTemplate(template),
		},
	)
	return event
}
