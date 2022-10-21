package command

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dennigogo/zitadel/internal/api/authz"
	"github.com/dennigogo/zitadel/internal/domain"
	caos_errs "github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
	"github.com/dennigogo/zitadel/internal/eventstore/v1/models"
	"github.com/dennigogo/zitadel/internal/repository/instance"
	"github.com/dennigogo/zitadel/internal/repository/policy"
)

func TestCommandSide_AddDefaultDomainPolicy(t *testing.T) {
	type fields struct {
		eventstore *eventstore.Eventstore
	}
	type args struct {
		ctx                                    context.Context
		userLoginMustBeDomain                  bool
		validateOrgDomains                     bool
		smtpSenderAddressMatchesInstanceDomain bool
	}
	type res struct {
		want *domain.ObjectDetails
		err  func(error) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		res    res
	}{
		{
			name: "domain policy already existing, already exists error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							instance.NewDomainPolicyAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								true,
								true,
								true,
							),
						),
					),
				),
			},
			args: args{
				ctx:                                    context.Background(),
				userLoginMustBeDomain:                  true,
				validateOrgDomains:                     true,
				smtpSenderAddressMatchesInstanceDomain: true,
			},
			res: res{
				err: caos_errs.IsErrorAlreadyExists,
			},
		},
		{
			name: "add policy,ok",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(),
					expectPush(
						[]*repository.Event{
							eventFromEventPusherWithInstanceID(
								"INSTANCE",
								instance.NewDomainPolicyAddedEvent(context.Background(),
									&instance.NewAggregate("INSTANCE").Aggregate,
									true,
									true,
									true,
								),
							),
						},
					),
				),
			},
			args: args{
				ctx:                                    authz.WithInstanceID(context.Background(), "INSTANCE"),
				userLoginMustBeDomain:                  true,
				validateOrgDomains:                     true,
				smtpSenderAddressMatchesInstanceDomain: true,
			},
			res: res{
				want: &domain.ObjectDetails{
					ResourceOwner: "INSTANCE",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Commands{
				eventstore: tt.fields.eventstore,
			}
			got, err := r.AddDefaultDomainPolicy(tt.args.ctx, tt.args.userLoginMustBeDomain, tt.args.validateOrgDomains, tt.args.smtpSenderAddressMatchesInstanceDomain)
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

func TestCommandSide_ChangeDefaultDomainPolicy(t *testing.T) {
	type fields struct {
		eventstore *eventstore.Eventstore
	}
	type args struct {
		ctx    context.Context
		policy *domain.DomainPolicy
	}
	type res struct {
		want *domain.DomainPolicy
		err  func(error) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		res    res
	}{
		{
			name: "domain policy not existing, not found error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(),
				),
			},
			args: args{
				ctx: context.Background(),
				policy: &domain.DomainPolicy{
					UserLoginMustBeDomain:                  true,
					ValidateOrgDomains:                     true,
					SMTPSenderAddressMatchesInstanceDomain: true,
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
							instance.NewDomainPolicyAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								true,
								true,
								true,
							),
						),
					),
				),
			},
			args: args{
				ctx: context.Background(),
				policy: &domain.DomainPolicy{
					UserLoginMustBeDomain:                  true,
					ValidateOrgDomains:                     true,
					SMTPSenderAddressMatchesInstanceDomain: true,
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
							instance.NewDomainPolicyAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								true,
								true,
								true,
							),
						),
					),
					expectPush(
						[]*repository.Event{
							eventFromEventPusher(
								newDefaultDomainPolicyChangedEvent(context.Background(), false, false, false),
							),
						},
					),
				),
			},
			args: args{
				ctx: context.Background(),
				policy: &domain.DomainPolicy{
					UserLoginMustBeDomain:                  false,
					ValidateOrgDomains:                     false,
					SMTPSenderAddressMatchesInstanceDomain: false,
				},
			},
			res: res{
				want: &domain.DomainPolicy{
					ObjectRoot: models.ObjectRoot{
						AggregateID:   "INSTANCE",
						ResourceOwner: "INSTANCE",
					},
					UserLoginMustBeDomain:                  false,
					ValidateOrgDomains:                     false,
					SMTPSenderAddressMatchesInstanceDomain: false,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Commands{
				eventstore: tt.fields.eventstore,
			}
			got, err := r.ChangeDefaultDomainPolicy(tt.args.ctx, tt.args.policy)
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

func newDefaultDomainPolicyChangedEvent(ctx context.Context, userLoginMustBeDomain, validateOrgDomains, smtpSenderAddressMatchesInstanceDomain bool) *instance.DomainPolicyChangedEvent {
	event, _ := instance.NewDomainPolicyChangedEvent(ctx,
		&instance.NewAggregate("INSTANCE").Aggregate,
		[]policy.DomainPolicyChanges{
			policy.ChangeUserLoginMustBeDomain(userLoginMustBeDomain),
			policy.ChangeValidateOrgDomains(validateOrgDomains),
			policy.ChangeSMTPSenderAddressMatchesInstanceDomain(smtpSenderAddressMatchesInstanceDomain),
		},
	)
	return event
}
