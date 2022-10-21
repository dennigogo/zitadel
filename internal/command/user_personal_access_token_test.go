package command

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/dennigogo/zitadel/internal/eventstore/repository"

	"github.com/dennigogo/zitadel/internal/eventstore/v1/models"

	id_mock "github.com/dennigogo/zitadel/internal/id/mock"

	"github.com/dennigogo/zitadel/internal/repository/user"

	caos_errs "github.com/dennigogo/zitadel/internal/errors"

	"github.com/dennigogo/zitadel/internal/crypto"
	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/id"
)

func TestCommands_AddPersonalAccessToken(t *testing.T) {
	type fields struct {
		eventstore   *eventstore.Eventstore
		idGenerator  id.Generator
		keyAlgorithm crypto.EncryptionAlgorithm
	}
	type args struct {
		ctx             context.Context
		userID          string
		resourceOwner   string
		expirationDate  time.Time
		scopes          []string
		allowedUserType domain.UserType
	}
	type res struct {
		want  *domain.Token
		token string
		err   func(error) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		res    res
	}{
		{
			"user does not exist, error",
			fields{
				eventstore: eventstoreExpect(t,
					expectFilter(),
				),
			},
			args{
				ctx:            context.Background(),
				userID:         "user1",
				resourceOwner:  "org1",
				scopes:         []string{"openid"},
				expirationDate: time.Time{},
			},
			res{
				err: caos_errs.IsPreconditionFailed,
			},
		},
		{
			"user type not allowed, error",
			fields{
				eventstore: eventstoreExpect(t,
					expectFilter(
						eventFromEventPusher(
							user.NewMachineAddedEvent(context.Background(),
								&user.NewAggregate("user1", "org1").Aggregate,
								"machine",
								"Machine",
								"",
								true,
							),
						),
					),
				),
			},
			args{
				ctx:             context.Background(),
				userID:          "user1",
				resourceOwner:   "org1",
				expirationDate:  time.Time{},
				scopes:          []string{"openid"},
				allowedUserType: domain.UserTypeHuman,
			},
			res{
				err: caos_errs.IsPreconditionFailed,
			},
		},
		{
			"invalid expiration date, error",
			fields{
				eventstore: eventstoreExpect(t,
					expectFilter(
						eventFromEventPusher(
							user.NewMachineAddedEvent(context.Background(),
								&user.NewAggregate("user1", "org1").Aggregate,
								"machine",
								"Machine",
								"",
								true,
							),
						),
					),
					expectFilter(),
				),
				idGenerator: id_mock.NewIDGeneratorExpectIDs(t, "token1"),
			},
			args{
				ctx:            context.Background(),
				userID:         "user1",
				resourceOwner:  "org1",
				expirationDate: time.Now().Add(-24 * time.Hour),
				scopes:         []string{"openid"},
			},
			res{
				err: caos_errs.IsErrorInvalidArgument,
			},
		},
		{
			"token added",
			fields{
				eventstore: eventstoreExpect(t,
					expectFilter(
						eventFromEventPusher(
							user.NewMachineAddedEvent(context.Background(),
								&user.NewAggregate("user1", "org1").Aggregate,
								"machine",
								"Machine",
								"",
								true,
							),
						),
					),
					expectFilter(),
					expectPush(
						[]*repository.Event{
							eventFromEventPusher(
								user.NewPersonalAccessTokenAddedEvent(context.Background(),
									&user.NewAggregate("user1", "org1").Aggregate,
									"token1",
									time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC),
									[]string{"openid"},
								),
							),
						},
					),
				),
				idGenerator:  id_mock.NewIDGeneratorExpectIDs(t, "token1"),
				keyAlgorithm: crypto.CreateMockEncryptionAlg(gomock.NewController(t)),
			},
			args{
				ctx:            context.Background(),
				userID:         "user1",
				resourceOwner:  "org1",
				expirationDate: time.Time{},
				scopes:         []string{"openid"},
			},
			res{
				want: &domain.Token{
					ObjectRoot: models.ObjectRoot{
						AggregateID:   "user1",
						ResourceOwner: "org1",
					},
					TokenID:    "token1",
					Expiration: time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC),
				},
				token: base64.RawURLEncoding.EncodeToString([]byte("token1:user1")),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Commands{
				eventstore:   tt.fields.eventstore,
				idGenerator:  tt.fields.idGenerator,
				keyAlgorithm: tt.fields.keyAlgorithm,
			}
			got, token, err := c.AddPersonalAccessToken(tt.args.ctx, tt.args.userID, tt.args.resourceOwner, tt.args.expirationDate, tt.args.scopes, tt.args.allowedUserType)
			if tt.res.err == nil {
				assert.NoError(t, err)
			}
			if tt.res.err != nil && !tt.res.err(err) {
				t.Errorf("got wrong err: %v ", err)
			}
			if tt.res.err == nil {
				assert.Equal(t, tt.res.want, got)
				assert.Equal(t, tt.res.token, token)
			}
		})
	}
}

func TestCommands_RemovePersonalAccessToken(t *testing.T) {
	type fields struct {
		eventstore *eventstore.Eventstore
	}
	type args struct {
		ctx           context.Context
		userID        string
		tokenID       string
		resourceOwner string
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
			"token does not exist, error",
			fields{
				eventstore: eventstoreExpect(t,
					expectFilter(),
				),
			},
			args{
				ctx:           context.Background(),
				userID:        "user1",
				tokenID:       "token1",
				resourceOwner: "org1",
			},
			res{
				err: caos_errs.IsNotFound,
			},
		},
		{
			"remove token, ok",
			fields{
				eventstore: eventstoreExpect(t,
					expectFilter(
						eventFromEventPusher(
							user.NewPersonalAccessTokenAddedEvent(context.Background(),
								&user.NewAggregate("user1", "org1").Aggregate,
								"token1",
								time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC),
								[]string{"openid"},
							),
						),
					),
					expectPush(
						[]*repository.Event{
							eventFromEventPusher(
								user.NewPersonalAccessTokenRemovedEvent(context.Background(),
									&user.NewAggregate("user1", "org1").Aggregate,
									"token1",
								),
							),
						},
					),
				),
			},
			args{
				ctx:           context.Background(),
				userID:        "user1",
				tokenID:       "token1",
				resourceOwner: "org1",
			},
			res{
				want: &domain.ObjectDetails{
					ResourceOwner: "org1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Commands{
				eventstore: tt.fields.eventstore,
			}
			got, err := c.RemovePersonalAccessToken(tt.args.ctx, tt.args.userID, tt.args.tokenID, tt.args.resourceOwner)
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
