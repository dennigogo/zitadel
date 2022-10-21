package projection

import (
	"testing"

	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/handler"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
	"github.com/dennigogo/zitadel/internal/repository/user"
)

func TestUserMetadataProjection_reduces(t *testing.T) {
	type args struct {
		event func(t *testing.T) eventstore.Event
	}
	tests := []struct {
		name   string
		args   args
		reduce func(event eventstore.Event) (*handler.Statement, error)
		want   wantReduce
	}{
		{
			name: "reduceMetadataSet",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(user.MetadataSetType),
					user.AggregateType,
					[]byte(`{
						"key": "key",
						"value": "dmFsdWU="
					}`),
				), user.MetadataSetEventMapper),
			},
			reduce: (&userMetadataProjection{}).reduceMetadataSet,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       UserMetadataProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.user_metadata3 (instance_id, user_id, key, resource_owner, creation_date, change_date, sequence, value) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (instance_id, user_id, key) DO UPDATE SET (resource_owner, creation_date, change_date, sequence, value) = (EXCLUDED.resource_owner, EXCLUDED.creation_date, EXCLUDED.change_date, EXCLUDED.sequence, EXCLUDED.value)",
							expectedArgs: []interface{}{
								"instance-id",
								"agg-id",
								"key",
								"ro-id",
								anyArg{},
								anyArg{},
								uint64(15),
								[]byte("value"),
							},
						},
					},
				},
			},
		},
		{
			name: "reduceMetadataRemoved",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(user.MetadataRemovedType),
					user.AggregateType,
					[]byte(`{
						"key": "key"
					}`),
				), user.MetadataRemovedEventMapper),
			},
			reduce: (&userMetadataProjection{}).reduceMetadataRemoved,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       UserMetadataProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM projections.user_metadata3 WHERE (user_id = $1) AND (key = $2)",
							expectedArgs: []interface{}{
								"agg-id",
								"key",
							},
						},
					},
				},
			},
		},
		{
			name: "reduceMetadataRemovedAll",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(user.MetadataRemovedAllType),
					user.AggregateType,
					nil,
				), user.MetadataRemovedAllEventMapper),
			},
			reduce: (&userMetadataProjection{}).reduceMetadataRemovedAll,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       UserMetadataProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM projections.user_metadata3 WHERE (user_id = $1)",
							expectedArgs: []interface{}{
								"agg-id",
							},
						},
					},
				},
			},
		},
		{
			name: "reduceMetadataRemovedAll (user removed)",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(user.UserRemovedType),
					user.AggregateType,
					nil,
				), user.UserRemovedEventMapper),
			},
			reduce: (&userMetadataProjection{}).reduceMetadataRemovedAll,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       UserMetadataProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM projections.user_metadata3 WHERE (user_id = $1)",
							expectedArgs: []interface{}{
								"agg-id",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := baseEvent(t)
			got, err := tt.reduce(event)
			if _, ok := err.(errors.InvalidArgument); !ok {
				t.Errorf("no wrong event mapping: %v, got: %v", err, got)
			}

			event = tt.args.event(t)
			got, err = tt.reduce(event)
			assertReduce(t, got, err, tt.want)
		})
	}
}
