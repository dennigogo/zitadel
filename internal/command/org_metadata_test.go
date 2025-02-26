package command

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dennigogo/zitadel/internal/domain"
	caos_errs "github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
	"github.com/dennigogo/zitadel/internal/eventstore/v1/models"
	"github.com/dennigogo/zitadel/internal/repository/org"
)

func TestCommandSide_SetOrgMetadata(t *testing.T) {
	type fields struct {
		eventstore *eventstore.Eventstore
	}
	type (
		args struct {
			ctx      context.Context
			orgID    string
			metadata *domain.Metadata
		}
	)
	type res struct {
		want *domain.Metadata
		err  func(error) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		res    res
	}{
		{
			name: "org not existing, pre condition error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(),
				),
			},
			args: args{
				ctx:   context.Background(),
				orgID: "org1",
				metadata: &domain.Metadata{
					Key:   "key",
					Value: []byte("value"),
				},
			},
			res: res{
				err: caos_errs.IsPreconditionFailed,
			},
		},
		{
			name: "invalid metadata, pre condition error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							org.NewOrgAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"ZITADEL",
							),
						),
					),
				),
			},
			args: args{
				ctx:   context.Background(),
				orgID: "org1",
				metadata: &domain.Metadata{
					Key: "key",
				},
			},
			res: res{
				err: caos_errs.IsErrorInvalidArgument,
			},
		},
		{
			name: "add metadata, ok",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							org.NewOrgAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"ZITADEL",
							),
						),
					),
					expectPush(
						[]*repository.Event{
							eventFromEventPusher(
								org.NewMetadataSetEvent(context.Background(),
									&org.NewAggregate("org1").Aggregate,
									"key",
									[]byte("value"),
								),
							),
						},
					),
				),
			},
			args: args{
				ctx:   context.Background(),
				orgID: "org1",
				metadata: &domain.Metadata{
					Key:   "key",
					Value: []byte("value"),
				},
			},
			res: res{
				want: &domain.Metadata{
					ObjectRoot: models.ObjectRoot{
						AggregateID:   "org1",
						ResourceOwner: "org1",
					},
					Key:   "key",
					Value: []byte("value"),
					State: domain.MetadataStateActive,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Commands{
				eventstore: tt.fields.eventstore,
			}
			got, err := r.SetOrgMetadata(tt.args.ctx, tt.args.orgID, tt.args.metadata)
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

func TestCommandSide_BulkSetOrgMetadata(t *testing.T) {
	type fields struct {
		eventstore *eventstore.Eventstore
	}
	type (
		args struct {
			ctx          context.Context
			orgID        string
			metadataList []*domain.Metadata
		}
	)
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
			name: "empty meta data list, pre condition error",
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
				err: caos_errs.IsPreconditionFailed,
			},
		},
		{
			name: "org not existing, pre condition error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(),
				),
			},
			args: args{
				ctx:   context.Background(),
				orgID: "org1",
				metadataList: []*domain.Metadata{
					{Key: "key", Value: []byte("value")},
					{Key: "key1", Value: []byte("value1")},
				},
			},
			res: res{
				err: caos_errs.IsPreconditionFailed,
			},
		},
		{
			name: "invalid metadata, pre condition error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							org.NewOrgAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"ZITADEL",
							),
						),
					),
				),
			},
			args: args{
				ctx:   context.Background(),
				orgID: "org1",
				metadataList: []*domain.Metadata{
					{Key: "key"},
					{Key: "key1"},
				},
			},
			res: res{
				err: caos_errs.IsErrorInvalidArgument,
			},
		},
		{
			name: "add metadata, ok",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							org.NewOrgAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"ZITADEL",
							),
						),
					),
					expectPush(
						[]*repository.Event{
							eventFromEventPusher(
								org.NewMetadataSetEvent(context.Background(),
									&org.NewAggregate("org1").Aggregate,
									"key",
									[]byte("value"),
								),
							),
							eventFromEventPusher(
								org.NewMetadataSetEvent(context.Background(),
									&org.NewAggregate("org1").Aggregate,
									"key1",
									[]byte("value1"),
								),
							),
						},
					),
				),
			},
			args: args{
				ctx:   context.Background(),
				orgID: "org1",
				metadataList: []*domain.Metadata{
					{Key: "key", Value: []byte("value")},
					{Key: "key1", Value: []byte("value1")},
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
			}
			got, err := r.BulkSetOrgMetadata(tt.args.ctx, tt.args.orgID, tt.args.metadataList...)
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

func TestCommandSide_OrgRemoveMetadata(t *testing.T) {
	type fields struct {
		eventstore *eventstore.Eventstore
	}
	type (
		args struct {
			ctx         context.Context
			orgID       string
			metadataKey string
		}
	)
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
			name: "org not existing, pre condition error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(),
				),
			},
			args: args{
				ctx:         context.Background(),
				orgID:       "org1",
				metadataKey: "key",
			},
			res: res{
				err: caos_errs.IsPreconditionFailed,
			},
		},
		{
			name: "invalid metadata, pre condition error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
				),
			},
			args: args{
				ctx:         context.Background(),
				orgID:       "org1",
				metadataKey: "",
			},
			res: res{
				err: caos_errs.IsErrorInvalidArgument,
			},
		},
		{
			name: "meta data not existing, not found error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							org.NewOrgAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"ZITADEL",
							),
						),
					),
					expectFilter(),
				),
			},
			args: args{
				ctx:         context.Background(),
				orgID:       "org1",
				metadataKey: "key",
			},
			res: res{
				err: caos_errs.IsNotFound,
			},
		},
		{
			name: "remove metadata, ok",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							org.NewOrgAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"ZITADEL",
							),
						),
					),
					expectFilter(
						eventFromEventPusher(
							org.NewMetadataSetEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"key",
								[]byte("value"),
							),
						),
					),
					expectPush(
						[]*repository.Event{
							eventFromEventPusher(
								org.NewMetadataRemovedEvent(context.Background(),
									&org.NewAggregate("org1").Aggregate,
									"key",
								),
							),
						},
					),
				),
			},
			args: args{
				ctx:         context.Background(),
				orgID:       "org1",
				metadataKey: "key",
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
			}
			got, err := r.RemoveOrgMetadata(tt.args.ctx, tt.args.orgID, tt.args.metadataKey)
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

func TestCommandSide_BulkRemoveOrgMetadata(t *testing.T) {
	type fields struct {
		eventstore *eventstore.Eventstore
	}
	type (
		args struct {
			ctx          context.Context
			orgID        string
			metadataList []string
		}
	)
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
			name: "empty meta data list, pre condition error",
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
				err: caos_errs.IsPreconditionFailed,
			},
		},
		{
			name: "org not existing, pre condition error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(),
				),
			},
			args: args{
				ctx:          context.Background(),
				orgID:        "org1",
				metadataList: []string{"key", "key1"},
			},
			res: res{
				err: caos_errs.IsPreconditionFailed,
			},
		},
		{
			name: "remove metadata keys not existing, precondition error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							org.NewOrgAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"ZITADEL",
							),
						),
					),
					expectFilter(
						eventFromEventPusher(
							org.NewMetadataSetEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"key",
								[]byte("value"),
							),
						),
					),
				),
			},
			args: args{
				ctx:          context.Background(),
				orgID:        "org1",
				metadataList: []string{"key", "key1"},
			},
			res: res{
				err: caos_errs.IsNotFound,
			},
		},
		{
			name: "invalid metadata, pre condition error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							org.NewOrgAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"ZITADEL",
							),
						),
					),
					expectFilter(
						eventFromEventPusher(
							org.NewMetadataSetEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"key",
								[]byte("value"),
							),
						),
						eventFromEventPusher(
							org.NewMetadataSetEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"key1",
								[]byte("value1"),
							),
						),
					),
				),
			},
			args: args{
				ctx:          context.Background(),
				orgID:        "org1",
				metadataList: []string{""},
			},
			res: res{
				err: caos_errs.IsErrorInvalidArgument,
			},
		},
		{
			name: "remove metadata, ok",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							org.NewOrgAddedEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"ZITADEL",
							),
						),
					),
					expectFilter(
						eventFromEventPusher(
							org.NewMetadataSetEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"key",
								[]byte("value"),
							),
						),
						eventFromEventPusher(
							org.NewMetadataSetEvent(context.Background(),
								&org.NewAggregate("org1").Aggregate,
								"key1",
								[]byte("value1"),
							),
						),
					),
					expectPush(
						[]*repository.Event{
							eventFromEventPusher(
								org.NewMetadataRemovedEvent(context.Background(),
									&org.NewAggregate("org1").Aggregate,
									"key",
								),
							),
							eventFromEventPusher(
								org.NewMetadataRemovedEvent(context.Background(),
									&org.NewAggregate("org1").Aggregate,
									"key1",
								),
							),
						},
					),
				),
			},
			args: args{
				ctx:          context.Background(),
				orgID:        "org1",
				metadataList: []string{"key", "key1"},
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
			}
			got, err := r.BulkRemoveOrgMetadata(tt.args.ctx, tt.args.orgID, tt.args.metadataList...)
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
