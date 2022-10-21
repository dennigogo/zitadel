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

func TestCommandSide_AddDefaultPrivacyPolicy(t *testing.T) {
	type fields struct {
		eventstore *eventstore.Eventstore
	}
	type args struct {
		ctx         context.Context
		tosLink     string
		privacyLink string
		helpLink    string
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
			name: "privacy policy already existing, already exists error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							instance.NewPrivacyPolicyAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								"TOSLink",
								"PrivacyLink",
								"HelpLink",
							),
						),
					),
				),
			},
			args: args{
				ctx:         context.Background(),
				tosLink:     "TOSLink",
				privacyLink: "PrivacyLink",
				helpLink:    "HelpLink",
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
								instance.NewPrivacyPolicyAddedEvent(context.Background(),
									&instance.NewAggregate("INSTANCE").Aggregate,
									"TOSLink",
									"PrivacyLink",
									"HelpLink",
								),
							),
						},
					),
				),
			},
			args: args{
				ctx:         authz.WithInstanceID(context.Background(), "INSTANCE"),
				tosLink:     "TOSLink",
				privacyLink: "PrivacyLink",
				helpLink:    "HelpLink",
			},
			res: res{
				want: &domain.ObjectDetails{
					ResourceOwner: "INSTANCE",
				},
			},
		},
		{
			name: "add empty policy,ok",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(),
					expectPush(
						[]*repository.Event{
							eventFromEventPusherWithInstanceID(
								"INSTANCE",
								instance.NewPrivacyPolicyAddedEvent(context.Background(),
									&instance.NewAggregate("INSTANCE").Aggregate,
									"",
									"",
									"",
								),
							),
						},
					),
				),
			},
			args: args{
				ctx:         authz.WithInstanceID(context.Background(), "INSTANCE"),
				tosLink:     "",
				privacyLink: "",
				helpLink:    "",
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
			got, err := r.AddDefaultPrivacyPolicy(tt.args.ctx, tt.args.tosLink, tt.args.privacyLink, tt.args.helpLink)
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

func TestCommandSide_ChangeDefaultPrivacyPolicy(t *testing.T) {
	type fields struct {
		eventstore *eventstore.Eventstore
	}
	type args struct {
		ctx    context.Context
		policy *domain.PrivacyPolicy
	}
	type res struct {
		want *domain.PrivacyPolicy
		err  func(error) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		res    res
	}{
		{
			name: "privacy policy not existing, not found error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(),
				),
			},
			args: args{
				ctx: context.Background(),
				policy: &domain.PrivacyPolicy{
					TOSLink:     "TOSLink",
					PrivacyLink: "PrivacyLink",
					HelpLink:    "HelpLink",
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
							instance.NewPrivacyPolicyAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								"TOSLink",
								"PrivacyLink",
								"HelpLink",
							),
						),
					),
				),
			},
			args: args{
				ctx: context.Background(),
				policy: &domain.PrivacyPolicy{
					TOSLink:     "TOSLink",
					PrivacyLink: "PrivacyLink",
					HelpLink:    "HelpLink",
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
							instance.NewPrivacyPolicyAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								"TOSLink",
								"PrivacyLink",
								"HelpLink",
							),
						),
					),
					expectPush(
						[]*repository.Event{
							eventFromEventPusher(
								newDefaultPrivacyPolicyChangedEvent(context.Background(),
									"TOSLinkChanged",
									"PrivacyLinkChanged",
									"HelpLinkChanged",
								),
							),
						},
					),
				),
			},
			args: args{
				ctx: context.Background(),
				policy: &domain.PrivacyPolicy{
					TOSLink:     "TOSLinkChanged",
					PrivacyLink: "PrivacyLinkChanged",
					HelpLink:    "HelpLinkChanged",
				},
			},
			res: res{
				want: &domain.PrivacyPolicy{
					ObjectRoot: models.ObjectRoot{
						AggregateID:   "INSTANCE",
						ResourceOwner: "INSTANCE",
					},
					TOSLink:     "TOSLinkChanged",
					PrivacyLink: "PrivacyLinkChanged",
					HelpLink:    "HelpLinkChanged",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Commands{
				eventstore: tt.fields.eventstore,
			}
			got, err := r.ChangeDefaultPrivacyPolicy(tt.args.ctx, tt.args.policy)
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

func newDefaultPrivacyPolicyChangedEvent(ctx context.Context, tosLink, privacyLink, helpLink string) *instance.PrivacyPolicyChangedEvent {
	event, _ := instance.NewPrivacyPolicyChangedEvent(ctx,
		&instance.NewAggregate("INSTANCE").Aggregate,
		[]policy.PrivacyPolicyChanges{
			policy.ChangeTOSLink(tosLink),
			policy.ChangePrivacyLink(privacyLink),
			policy.ChangeHelpLink(helpLink),
		},
	)
	return event
}
