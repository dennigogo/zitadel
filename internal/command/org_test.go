package command

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"

	"github.com/dennigogo/zitadel/internal/api/authz"
	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
	"github.com/dennigogo/zitadel/internal/eventstore/v1/models"
	"github.com/dennigogo/zitadel/internal/id"
	id_mock "github.com/dennigogo/zitadel/internal/id/mock"
	"github.com/dennigogo/zitadel/internal/repository/member"
	"github.com/dennigogo/zitadel/internal/repository/org"
	"github.com/dennigogo/zitadel/internal/repository/user"
)

func TestAddOrg(t *testing.T) {
	type args struct {
		a    *org.Aggregate
		name string
	}

	ctx := context.Background()
	agg := org.NewAggregate("test")

	tests := []struct {
		name string
		args args
		want Want
	}{
		{
			name: "invalid domain",
			args: args{
				a:    agg,
				name: "",
			},
			want: Want{
				ValidationErr: errors.ThrowInvalidArgument(nil, "ORG-mruNY", "Errors.Invalid.Argument"),
			},
		},
		{
			name: "correct",
			args: args{
				a:    agg,
				name: "caos ag",
			},
			want: Want{
				Commands: []eventstore.Command{
					org.NewOrgAddedEvent(ctx, &agg.Aggregate, "caos ag"),
					org.NewDomainAddedEvent(ctx, &agg.Aggregate, "caos-ag.localhost"),
					org.NewDomainVerifiedEvent(ctx, &agg.Aggregate, "caos-ag.localhost"),
					org.NewDomainPrimarySetEvent(ctx, &agg.Aggregate, "caos-ag.localhost"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AssertValidation(t, context.Background(), AddOrgCommand(authz.WithRequestedDomain(context.Background(), "localhost"), tt.args.a, tt.args.name), nil, tt.want)
		})
	}
}

func TestCommandSide_AddOrg(t *testing.T) {
	type fields struct {
		eventstore   *eventstore.Eventstore
		idGenerator  id.Generator
		zitadelRoles []authz.RoleMapping
	}
	type args struct {
		ctx            context.Context
		name           string
		userID         string
		resourceOwner  string
		claimedUserIDs []string
	}
	type res struct {
		want *domain.Org
		err  func(error) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		res    res
	}{
		{
			name: "invalid org, error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
				),
			},
			args: args{
				ctx:           context.Background(),
				userID:        "user1",
				resourceOwner: "org1",
			},
			res: res{
				err: errors.IsErrorInvalidArgument,
			},
		},
		{
			name: "invalid org (spaces), error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
				),
			},
			args: args{
				ctx:           context.Background(),
				userID:        "user1",
				resourceOwner: "org1",
				name:          "  ",
			},
			res: res{
				err: errors.IsErrorInvalidArgument,
			},
		},
		{
			name: "user removed, error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilterOrgDomainNotFound(),
					expectFilter(
						eventFromEventPusher(
							user.NewHumanAddedEvent(context.Background(),
								&user.NewAggregate("user1", "org1").Aggregate,
								"username1",
								"firstname1",
								"lastname1",
								"nickname1",
								"displayname1",
								language.German,
								domain.GenderMale,
								"email1",
								true,
							),
						),
						eventFromEventPusher(
							user.NewUserRemovedEvent(
								context.Background(),
								&user.NewAggregate("user1", "org1").Aggregate,
								"username1",
								nil,
								true,
							),
						),
					),
				),
				idGenerator: id_mock.NewIDGeneratorExpectIDs(t, "org2"),
			},
			args: args{
				ctx:           context.Background(),
				name:          "Org",
				userID:        "user1",
				resourceOwner: "org1",
			},
			res: res{
				err: errors.IsPreconditionFailed,
			},
		},
		{
			name: "push failed unique constraint, error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilterOrgDomainNotFound(),
					expectFilter(
						eventFromEventPusher(
							user.NewHumanAddedEvent(context.Background(),
								&user.NewAggregate("user1", "org1").Aggregate,
								"username1",
								"firstname1",
								"lastname1",
								"nickname1",
								"displayname1",
								language.German,
								domain.GenderMale,
								"email1",
								true,
							),
						),
					),
					expectFilterOrgMemberNotFound(),
					expectPushFailed(errors.ThrowAlreadyExists(nil, "id", "internal"),
						[]*repository.Event{
							eventFromEventPusher(org.NewOrgAddedEvent(
								context.Background(),
								&org.NewAggregate("org2").Aggregate,
								"Org")),
							eventFromEventPusher(org.NewDomainAddedEvent(
								context.Background(),
								&org.NewAggregate("org2").Aggregate,
								"org.iam-domain")),
							eventFromEventPusher(org.NewDomainVerifiedEvent(
								context.Background(),
								&org.NewAggregate("org2").Aggregate,
								"org.iam-domain")),
							eventFromEventPusher(org.NewDomainPrimarySetEvent(
								context.Background(),
								&org.NewAggregate("org2").Aggregate,
								"org.iam-domain")),
							eventFromEventPusher(org.NewMemberAddedEvent(
								context.Background(),
								&org.NewAggregate("org2").Aggregate,
								"user1", domain.RoleOrgOwner)),
						},
						uniqueConstraintsFromEventConstraint(org.NewAddOrgNameUniqueConstraint("Org")),
						uniqueConstraintsFromEventConstraint(org.NewAddOrgDomainUniqueConstraint("org.iam-domain")),
						uniqueConstraintsFromEventConstraint(member.NewAddMemberUniqueConstraint("org2", "user1")),
					),
				),
				idGenerator: id_mock.NewIDGeneratorExpectIDs(t, "org2"),
				zitadelRoles: []authz.RoleMapping{
					{
						Role: "ORG_OWNER",
					},
				},
			},
			args: args{
				ctx:           authz.WithRequestedDomain(context.Background(), "iam-domain"),
				name:          "Org",
				userID:        "user1",
				resourceOwner: "org1",
			},
			res: res{
				err: errors.IsErrorAlreadyExists,
			},
		},
		{
			name: "push failed, error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilterOrgDomainNotFound(),
					expectFilter(
						eventFromEventPusher(
							user.NewHumanAddedEvent(context.Background(),
								&user.NewAggregate("user1", "org1").Aggregate,
								"username1",
								"firstname1",
								"lastname1",
								"nickname1",
								"displayname1",
								language.German,
								domain.GenderMale,
								"email1",
								true,
							),
						),
					),
					expectFilterOrgMemberNotFound(),
					expectPushFailed(errors.ThrowInternal(nil, "id", "internal"),
						[]*repository.Event{
							eventFromEventPusher(org.NewOrgAddedEvent(
								context.Background(),
								&org.NewAggregate("org2").Aggregate,
								"Org")),
							eventFromEventPusher(org.NewDomainAddedEvent(
								context.Background(),
								&org.NewAggregate("org2").Aggregate,
								"org.iam-domain")),
							eventFromEventPusher(org.NewDomainVerifiedEvent(
								context.Background(),
								&org.NewAggregate("org2").Aggregate,
								"org.iam-domain")),
							eventFromEventPusher(org.NewDomainPrimarySetEvent(
								context.Background(),
								&org.NewAggregate("org2").Aggregate,
								"org.iam-domain")),
							eventFromEventPusher(org.NewMemberAddedEvent(
								context.Background(),
								&org.NewAggregate("org2").Aggregate,
								"user1", domain.RoleOrgOwner)),
						},
						uniqueConstraintsFromEventConstraint(org.NewAddOrgNameUniqueConstraint("Org")),
						uniqueConstraintsFromEventConstraint(org.NewAddOrgDomainUniqueConstraint("org.iam-domain")),
						uniqueConstraintsFromEventConstraint(member.NewAddMemberUniqueConstraint("org2", "user1")),
					),
				),
				idGenerator: id_mock.NewIDGeneratorExpectIDs(t, "org2"),
				zitadelRoles: []authz.RoleMapping{
					{
						Role: "ORG_OWNER",
					},
				},
			},
			args: args{
				ctx:           authz.WithRequestedDomain(context.Background(), "iam-domain"),
				name:          "Org",
				userID:        "user1",
				resourceOwner: "org1",
			},
			res: res{
				err: errors.IsInternal,
			},
		},
		{
			name: "add org, no error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilterOrgDomainNotFound(),
					expectFilter(
						eventFromEventPusher(
							user.NewHumanAddedEvent(context.Background(),
								&user.NewAggregate("user1", "org1").Aggregate,
								"username1",
								"firstname1",
								"lastname1",
								"nickname1",
								"displayname1",
								language.German,
								domain.GenderMale,
								"email1",
								true,
							),
						),
					),
					expectFilterOrgMemberNotFound(),
					expectPush(
						[]*repository.Event{
							eventFromEventPusher(org.NewOrgAddedEvent(context.Background(),
								&org.NewAggregate("org2").Aggregate,
								"Org",
							)),
							eventFromEventPusher(org.NewDomainAddedEvent(context.Background(),
								&org.NewAggregate("org2").Aggregate, "org.iam-domain",
							)),
							eventFromEventPusher(org.NewDomainVerifiedEvent(context.Background(),
								&org.NewAggregate("org2").Aggregate,
								"org.iam-domain",
							)),
							eventFromEventPusher(org.NewDomainPrimarySetEvent(context.Background(),
								&org.NewAggregate("org2").Aggregate,
								"org.iam-domain",
							)),
							eventFromEventPusher(org.NewMemberAddedEvent(context.Background(),
								&org.NewAggregate("org2").Aggregate,
								"user1",
								domain.RoleOrgOwner,
							)),
						},
						uniqueConstraintsFromEventConstraint(org.NewAddOrgNameUniqueConstraint("Org")),
						uniqueConstraintsFromEventConstraint(org.NewAddOrgDomainUniqueConstraint("org.iam-domain")),
						uniqueConstraintsFromEventConstraint(member.NewAddMemberUniqueConstraint("org2", "user1")),
					),
				),
				idGenerator: id_mock.NewIDGeneratorExpectIDs(t, "org2"),
				zitadelRoles: []authz.RoleMapping{
					{
						Role: "ORG_OWNER",
					},
				},
			},
			args: args{
				ctx:           authz.WithRequestedDomain(context.Background(), "iam-domain"),
				name:          "Org",
				userID:        "user1",
				resourceOwner: "org1",
			},
			res: res{
				want: &domain.Org{
					ObjectRoot: models.ObjectRoot{
						AggregateID:   "org2",
						ResourceOwner: "org2",
					},
					Name:          "Org",
					State:         domain.OrgStateActive,
					PrimaryDomain: "org.iam-domain",
				},
			},
		},
		{
			name: "add org (remove spaces), no error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilterOrgDomainNotFound(),
					expectFilter(
						eventFromEventPusher(
							user.NewHumanAddedEvent(context.Background(),
								&user.NewAggregate("user1", "org1").Aggregate,
								"username1",
								"firstname1",
								"lastname1",
								"nickname1",
								"displayname1",
								language.German,
								domain.GenderMale,
								"email1",
								true,
							),
						),
					),
					expectFilterOrgMemberNotFound(),
					expectPush(
						[]*repository.Event{
							eventFromEventPusher(org.NewOrgAddedEvent(context.Background(),
								&org.NewAggregate("org2").Aggregate,
								"Org",
							)),
							eventFromEventPusher(org.NewDomainAddedEvent(context.Background(),
								&org.NewAggregate("org2").Aggregate, "org.iam-domain",
							)),
							eventFromEventPusher(org.NewDomainVerifiedEvent(context.Background(),
								&org.NewAggregate("org2").Aggregate,
								"org.iam-domain",
							)),
							eventFromEventPusher(org.NewDomainPrimarySetEvent(context.Background(),
								&org.NewAggregate("org2").Aggregate,
								"org.iam-domain",
							)),
							eventFromEventPusher(org.NewMemberAddedEvent(context.Background(),
								&org.NewAggregate("org2").Aggregate,
								"user1",
								domain.RoleOrgOwner,
							)),
						},
						uniqueConstraintsFromEventConstraint(org.NewAddOrgNameUniqueConstraint("Org")),
						uniqueConstraintsFromEventConstraint(org.NewAddOrgDomainUniqueConstraint("org.iam-domain")),
						uniqueConstraintsFromEventConstraint(member.NewAddMemberUniqueConstraint("org2", "user1")),
					),
				),
				idGenerator: id_mock.NewIDGeneratorExpectIDs(t, "org2"),
				zitadelRoles: []authz.RoleMapping{
					{
						Role: "ORG_OWNER",
					},
				},
			},
			args: args{
				ctx:           authz.WithRequestedDomain(context.Background(), "iam-domain"),
				name:          " Org ",
				userID:        "user1",
				resourceOwner: "org1",
			},
			res: res{
				want: &domain.Org{
					ObjectRoot: models.ObjectRoot{
						AggregateID:   "org2",
						ResourceOwner: "org2",
					},
					Name:          "Org",
					State:         domain.OrgStateActive,
					PrimaryDomain: "org.iam-domain",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Commands{
				eventstore:   tt.fields.eventstore,
				idGenerator:  tt.fields.idGenerator,
				zitadelRoles: tt.fields.zitadelRoles,
			}
			got, err := r.AddOrg(tt.args.ctx, tt.args.name, tt.args.userID, tt.args.resourceOwner, tt.args.claimedUserIDs)
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

func TestCommandSide_ChangeOrg(t *testing.T) {
	type fields struct {
		eventstore *eventstore.Eventstore
	}
	type args struct {
		ctx   context.Context
		orgID string
		name  string
	}
	type res struct {
		err func(error) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		res    res
	}{
		{
			name: "empty name, invalid argument error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
				),
			},
			args: args{
				ctx:   context.Background(),
				orgID: "org1",
			},
			res: res{
				err: errors.IsErrorInvalidArgument,
			},
		},
		{
			name: "empty name (spaces), invalid argument error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
				),
			},
			args: args{
				ctx:   context.Background(),
				orgID: "org1",
				name:  "  ",
			},
			res: res{
				err: errors.IsErrorInvalidArgument,
			},
		},
		{
			name: "org not found, error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(),
				),
			},
			args: args{
				ctx:   context.Background(),
				orgID: "org1",
				name:  "org",
			},
			res: res{
				err: errors.IsNotFound,
			},
		},
		{
			name: "no change (spaces), error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							org.NewOrgAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"org"),
						),
					),
				),
			},
			args: args{
				ctx:   authz.WithRequestedDomain(context.Background(), "zitadel.ch"),
				orgID: "org1",
				name:  " org ",
			},
			res: res{
				err: errors.IsPreconditionFailed,
			},
		},
		{
			name: "push failed, error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							org.NewOrgAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"org"),
						),
					),
					expectFilter(),
					expectPushFailed(
						errors.ThrowInternal(nil, "id", "message"),
						[]*repository.Event{
							eventFromEventPusher(org.NewOrgChangedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate, "org", "neworg")),
						},
						uniqueConstraintsFromEventConstraint(org.NewRemoveOrgNameUniqueConstraint("org")),
						uniqueConstraintsFromEventConstraint(org.NewAddOrgNameUniqueConstraint("neworg")),
					),
				),
			},
			args: args{
				ctx:   authz.WithRequestedDomain(context.Background(), "zitadel.ch"),
				orgID: "org1",
				name:  "neworg",
			},
			res: res{
				err: errors.IsInternal,
			},
		},
		{
			name: "change org name verified, not primary",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							org.NewOrgAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"org"),
						),
					),
					expectFilter(
						eventFromEventPusher(
							org.NewOrgAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"org"),
						),
						eventFromEventPusher(
							org.NewDomainAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"org.zitadel.ch"),
						),
						eventFromEventPusher(
							org.NewDomainVerifiedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"org.zitadel.ch"),
						),
					),
					expectPush(
						[]*repository.Event{
							eventFromEventPusher(org.NewOrgChangedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate, "org", "neworg")),
							eventFromEventPusher(org.NewDomainAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate, "neworg.zitadel.ch")),
							eventFromEventPusher(org.NewDomainVerifiedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate, "neworg.zitadel.ch")),
							eventFromEventPusher(org.NewDomainRemovedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate, "org.zitadel.ch", true)),
						},
						uniqueConstraintsFromEventConstraint(org.NewRemoveOrgNameUniqueConstraint("org")),
						uniqueConstraintsFromEventConstraint(org.NewAddOrgNameUniqueConstraint("neworg")),
						uniqueConstraintsFromEventConstraint(org.NewAddOrgDomainUniqueConstraint("neworg.zitadel.ch")),
						uniqueConstraintsFromEventConstraint(org.NewRemoveOrgDomainUniqueConstraint("org.zitadel.ch")),
					),
				),
			},
			args: args{
				ctx:   authz.WithRequestedDomain(context.Background(), "zitadel.ch"),
				orgID: "org1",
				name:  "neworg",
			},
			res: res{},
		},
		{
			name: "change org name verified, with primary",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							org.NewOrgAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"org"),
						),
					),
					expectFilter(
						eventFromEventPusher(
							org.NewOrgAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"org"),
						),
						eventFromEventPusher(
							org.NewDomainAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"org.zitadel.ch"),
						),
						eventFromEventPusher(
							org.NewDomainVerifiedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"org.zitadel.ch"),
						),
						eventFromEventPusher(
							org.NewDomainPrimarySetEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"org.zitadel.ch"),
						),
					),
					expectPush(
						[]*repository.Event{
							eventFromEventPusher(org.NewOrgChangedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate, "org", "neworg")),
							eventFromEventPusher(org.NewDomainAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate, "neworg.zitadel.ch")),
							eventFromEventPusher(org.NewDomainVerifiedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate, "neworg.zitadel.ch")),
							eventFromEventPusher(org.NewDomainPrimarySetEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate, "neworg.zitadel.ch")),
							eventFromEventPusher(org.NewDomainRemovedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate, "org.zitadel.ch", true)),
						},
						uniqueConstraintsFromEventConstraint(org.NewRemoveOrgNameUniqueConstraint("org")),
						uniqueConstraintsFromEventConstraint(org.NewAddOrgNameUniqueConstraint("neworg")),
						uniqueConstraintsFromEventConstraint(org.NewAddOrgDomainUniqueConstraint("neworg.zitadel.ch")),
						uniqueConstraintsFromEventConstraint(org.NewRemoveOrgDomainUniqueConstraint("org.zitadel.ch")),
					),
				),
			},
			args: args{
				ctx:   authz.WithRequestedDomain(context.Background(), "zitadel.ch"),
				orgID: "org1",
				name:  "neworg",
			},
			res: res{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Commands{
				eventstore: tt.fields.eventstore,
			}
			_, err := r.ChangeOrg(tt.args.ctx, tt.args.orgID, tt.args.name)
			if tt.res.err == nil {
				assert.NoError(t, err)
			}
			if tt.res.err != nil && !tt.res.err(err) {
				t.Errorf("got wrong err: %v ", err)
			}
		})
	}
}

func TestCommandSide_DeactivateOrg(t *testing.T) {
	type fields struct {
		eventstore  *eventstore.Eventstore
		idGenerator id.Generator
		iamDomain   string
	}
	type args struct {
		ctx   context.Context
		orgID string
	}
	type res struct {
		want *domain.Org
		err  func(error) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		res    res
	}{
		{
			name: "org not found, error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(),
				),
			},
			args: args{
				ctx:   context.Background(),
				orgID: "org1",
			},
			res: res{
				err: errors.IsNotFound,
			},
		},
		{
			name: "org already inactive, error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							org.NewOrgAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"org"),
						),
						eventFromEventPusher(
							org.NewOrgDeactivatedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate),
						),
					),
				),
			},
			args: args{
				ctx:   context.Background(),
				orgID: "org1",
			},
			res: res{
				err: errors.IsPreconditionFailed,
			},
		},
		{
			name: "push failed, error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							org.NewOrgAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"org"),
						),
					),
					expectPushFailed(
						errors.ThrowInternal(nil, "id", "message"),
						[]*repository.Event{
							eventFromEventPusher(org.NewOrgDeactivatedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate)),
						},
					),
				),
			},
			args: args{
				ctx:   context.Background(),
				orgID: "org1",
			},
			res: res{
				err: errors.IsInternal,
			},
		},
		{
			name: "deactivate org",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							org.NewOrgAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"org"),
						),
					),
					expectPush(
						[]*repository.Event{
							eventFromEventPusher(org.NewOrgDeactivatedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
							)),
						},
					),
				),
			},
			args: args{
				ctx:   context.Background(),
				orgID: "org1",
			},
			res: res{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Commands{
				eventstore:  tt.fields.eventstore,
				idGenerator: tt.fields.idGenerator,
			}
			_, err := r.DeactivateOrg(tt.args.ctx, tt.args.orgID)
			if tt.res.err == nil {
				assert.NoError(t, err)
			}
			if tt.res.err != nil && !tt.res.err(err) {
				t.Errorf("got wrong err: %v ", err)
			}
		})
	}
}
func TestCommandSide_ReactivateOrg(t *testing.T) {
	type fields struct {
		eventstore  *eventstore.Eventstore
		idGenerator id.Generator
		iamDomain   string
	}
	type args struct {
		ctx   context.Context
		orgID string
	}
	type res struct {
		want *domain.Org
		err  func(error) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		res    res
	}{
		{
			name: "org not found, error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(),
				),
			},
			args: args{
				ctx:   context.Background(),
				orgID: "org1",
			},
			res: res{
				err: errors.IsNotFound,
			},
		},
		{
			name: "org already active, error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							org.NewOrgAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"org"),
						),
					),
				),
			},
			args: args{
				ctx:   context.Background(),
				orgID: "org1",
			},
			res: res{
				err: errors.IsPreconditionFailed,
			},
		},
		{
			name: "push failed, error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							org.NewOrgAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"org"),
						),
						eventFromEventPusher(
							org.NewOrgDeactivatedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
							),
						),
					),
					expectPushFailed(
						errors.ThrowInternal(nil, "id", "message"),
						[]*repository.Event{
							eventFromEventPusher(org.NewOrgReactivatedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
							)),
						},
					),
				),
			},
			args: args{
				ctx:   context.Background(),
				orgID: "org1",
			},
			res: res{
				err: errors.IsInternal,
			},
		},
		{
			name: "reactivate org",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							org.NewOrgAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"org"),
						),
						eventFromEventPusher(
							org.NewOrgDeactivatedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate),
						),
					),
					expectPush(
						[]*repository.Event{
							eventFromEventPusher(org.NewOrgReactivatedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate)),
						},
					),
				),
			},
			args: args{
				ctx:   context.Background(),
				orgID: "org1",
			},
			res: res{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Commands{
				eventstore:  tt.fields.eventstore,
				idGenerator: tt.fields.idGenerator,
			}
			_, err := r.ReactivateOrg(tt.args.ctx, tt.args.orgID)
			if tt.res.err == nil {
				assert.NoError(t, err)
			}
			if tt.res.err != nil && !tt.res.err(err) {
				t.Errorf("got wrong err: %v ", err)
			}
		})
	}
}
