package command

import (
	"bytes"
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"

	"github.com/dennigogo/zitadel/internal/domain"
	caos_errs "github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
	"github.com/dennigogo/zitadel/internal/repository/user"
	"github.com/dennigogo/zitadel/internal/static"
	"github.com/dennigogo/zitadel/internal/static/mock"
)

func TestCommandSide_AddHumanAvatar(t *testing.T) {
	type fields struct {
		eventstore *eventstore.Eventstore
		storage    static.Storage
	}
	type args struct {
		ctx    context.Context
		orgID  string
		userID string
		upload *AssetUpload
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
			name: "userID empty, invalid argument error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
				),
			},
			args: args{
				ctx:    context.Background(),
				orgID:  "",
				userID: "",
				upload: &AssetUpload{
					ResourceOwner: "org1",
					ObjectName:    "avatar",
					ContentType:   "image",
					ObjectType:    static.ObjectTypeUserAvatar,
					File:          bytes.NewReader([]byte("test")),
					Size:          4,
				},
			},
			res: res{
				err: caos_errs.IsErrorInvalidArgument,
			},
		},
		{
			name: "user not existing, not found error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(),
				),
			},
			args: args{
				ctx:    context.Background(),
				orgID:  "org1",
				userID: "user1",
				upload: &AssetUpload{
					ResourceOwner: "org1",
					ObjectName:    "avatar",
					ContentType:   "image",
					ObjectType:    static.ObjectTypeUserAvatar,
					File:          bytes.NewReader([]byte("test")),
					Size:          4,
				},
			},
			res: res{
				err: caos_errs.IsNotFound,
			},
		},
		{
			name: "upload failed, error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							user.NewHumanAddedEvent(context.Background(),
								&user.NewAggregate("user1", "org1").Aggregate,
								"username",
								"firstname",
								"lastname",
								"nickname",
								"displayname",
								language.Und,
								domain.GenderUnspecified,
								"email@test.ch",
								true,
							),
						),
					),
				),
				storage: mock.NewStorage(t).ExpectPutObjectError(),
			},
			args: args{
				ctx:    context.Background(),
				orgID:  "org1",
				userID: "user1",
				upload: &AssetUpload{
					ResourceOwner: "org1",
					ObjectName:    "avatar",
					ContentType:   "image",
					ObjectType:    static.ObjectTypeUserAvatar,
					File:          bytes.NewReader([]byte("test")),
					Size:          4,
				},
			},
			res: res{
				err: caos_errs.IsInternal,
			},
		},
		{
			name: "avatar added, ok",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							user.NewHumanAddedEvent(context.Background(),
								&user.NewAggregate("user1", "org1").Aggregate,
								"username",
								"firstname",
								"lastname",
								"nickname",
								"displayname",
								language.Und,
								domain.GenderUnspecified,
								"email@test.ch",
								true,
							),
						),
					),
					expectPush(
						[]*repository.Event{
							eventFromEventPusher(
								user.NewHumanAvatarAddedEvent(context.Background(),
									&user.NewAggregate("user1", "org1").Aggregate,
									"avatar?v=test",
								),
							),
						},
					),
				),
				storage: mock.NewStorage(t).ExpectPutObject(),
			},
			args: args{
				ctx:    context.Background(),
				orgID:  "org1",
				userID: "user1",
				upload: &AssetUpload{
					ResourceOwner: "org1",
					ObjectName:    "avatar",
					ContentType:   "image",
					ObjectType:    static.ObjectTypeUserAvatar,
					File:          bytes.NewReader([]byte("test")),
					Size:          4,
				},
			},
			res: res{
				want: &domain.ObjectDetails{
					ResourceOwner: "org1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Commands{
				eventstore: tt.fields.eventstore,
				static:     tt.fields.storage,
			}
			got, err := r.AddHumanAvatar(tt.args.ctx, tt.args.orgID, tt.args.userID, tt.args.upload)
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

func TestCommandSide_RemoveHumanAvatar(t *testing.T) {
	type fields struct {
		eventstore *eventstore.Eventstore
		storage    static.Storage
	}
	type args struct {
		ctx        context.Context
		orgID      string
		userID     string
		storageKey string
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
			name: "userID empty, invalid argument error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
				),
			},
			args: args{
				ctx:        context.Background(),
				storageKey: "key",
			},
			res: res{
				err: caos_errs.IsErrorInvalidArgument,
			},
		},
		{
			name: "user not existing, not found error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(),
				),
			},
			args: args{
				ctx:        context.Background(),
				orgID:      "org1",
				userID:     "user1",
				storageKey: "key",
			},
			res: res{
				err: caos_errs.IsNotFound,
			},
		},
		{
			name: "file remove error, not found error",
			fields: fields{
				storage: mock.NewMockStorage(gomock.NewController(t)).ExpectRemoveObjectError(),
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							user.NewHumanAddedEvent(context.Background(),
								&user.NewAggregate("user1", "org1").Aggregate,
								"username",
								"firstname",
								"lastname",
								"nickname",
								"displayname",
								language.Und,
								domain.GenderUnspecified,
								"email@test.ch",
								true,
							),
						),
						eventFromEventPusher(
							user.NewHumanAvatarAddedEvent(context.Background(),
								&user.NewAggregate("user1", "org1").Aggregate,
								"key",
							),
						),
					),
				),
			},
			args: args{
				ctx:        context.Background(),
				orgID:      "org1",
				userID:     "user1",
				storageKey: "key",
			},
			res: res{
				err: caos_errs.IsInternal,
			},
		},
		{
			name: "logo removed, ok",
			fields: fields{
				storage: mock.NewMockStorage(gomock.NewController(t)).ExpectRemoveObjectNoError(),
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							user.NewHumanAddedEvent(context.Background(),
								&user.NewAggregate("user1", "org1").Aggregate,
								"username",
								"firstname",
								"lastname",
								"nickname",
								"displayname",
								language.Und,
								domain.GenderUnspecified,
								"email@test.ch",
								true,
							),
						),
						eventFromEventPusher(
							user.NewHumanAvatarAddedEvent(context.Background(),
								&user.NewAggregate("user1", "org1").Aggregate,
								"key",
							),
						),
					),
					expectPush(
						[]*repository.Event{
							eventFromEventPusher(
								user.NewHumanAvatarRemovedEvent(context.Background(),
									&user.NewAggregate("user1", "org1").Aggregate,
									"key",
								),
							),
						},
					),
				),
			},
			args: args{
				ctx:        context.Background(),
				orgID:      "org1",
				userID:     "user1",
				storageKey: "key",
			},
			res: res{
				want: &domain.ObjectDetails{
					ResourceOwner: "org1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Commands{
				eventstore: tt.fields.eventstore,
				static:     tt.fields.storage,
			}
			got, err := r.RemoveHumanAvatar(tt.args.ctx, tt.args.orgID, tt.args.userID)
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
