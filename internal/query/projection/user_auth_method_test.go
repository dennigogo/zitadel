package projection

import (
	"testing"

	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/handler"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
	"github.com/dennigogo/zitadel/internal/repository/user"
)

func TestUserAuthMethodProjection_reduces(t *testing.T) {
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
			name: "reduceAddedPasswordless",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(user.HumanPasswordlessTokenAddedType),
					user.AggregateType,
					[]byte(`{
						"webAuthNTokenId": "token-id"
					}`),
				), user.HumanPasswordlessAddedEventMapper),
			},
			reduce: (&userAuthMethodProjection{}).reduceInitAuthMethod,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       UserAuthMethodTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.user_auth_methods3 (token_id, creation_date, change_date, resource_owner, instance_id, user_id, sequence, state, method_type, name) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) ON CONFLICT (instance_id, user_id, method_type, token_id) DO UPDATE SET (creation_date, change_date, resource_owner, sequence, state, name) = (EXCLUDED.creation_date, EXCLUDED.change_date, EXCLUDED.resource_owner, EXCLUDED.sequence, EXCLUDED.state, EXCLUDED.name)",
							expectedArgs: []interface{}{
								"token-id",
								anyArg{},
								anyArg{},
								"ro-id",
								"instance-id",
								"agg-id",
								uint64(15),
								domain.MFAStateNotReady,
								domain.UserAuthMethodTypePasswordless,
								"",
							},
						},
					},
				},
			},
		},
		{
			name: "reduceAddedU2F",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(user.HumanU2FTokenAddedType),
					user.AggregateType,
					[]byte(`{
						"webAuthNTokenId": "token-id"
					}`),
				), user.HumanU2FAddedEventMapper),
			},
			reduce: (&userAuthMethodProjection{}).reduceInitAuthMethod,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       UserAuthMethodTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.user_auth_methods3 (token_id, creation_date, change_date, resource_owner, instance_id, user_id, sequence, state, method_type, name) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) ON CONFLICT (instance_id, user_id, method_type, token_id) DO UPDATE SET (creation_date, change_date, resource_owner, sequence, state, name) = (EXCLUDED.creation_date, EXCLUDED.change_date, EXCLUDED.resource_owner, EXCLUDED.sequence, EXCLUDED.state, EXCLUDED.name)",
							expectedArgs: []interface{}{
								"token-id",
								anyArg{},
								anyArg{},
								"ro-id",
								"instance-id",
								"agg-id",
								uint64(15),
								domain.MFAStateNotReady,
								domain.UserAuthMethodTypeU2F,
								"",
							},
						},
					},
				},
			},
		},
		{
			name: "reduceAddedOTP",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(user.HumanMFAOTPAddedType),
					user.AggregateType,
					[]byte(`{
					}`),
				), user.HumanOTPAddedEventMapper),
			},
			reduce: (&userAuthMethodProjection{}).reduceInitAuthMethod,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       UserAuthMethodTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.user_auth_methods3 (token_id, creation_date, change_date, resource_owner, instance_id, user_id, sequence, state, method_type, name) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) ON CONFLICT (instance_id, user_id, method_type, token_id) DO UPDATE SET (creation_date, change_date, resource_owner, sequence, state, name) = (EXCLUDED.creation_date, EXCLUDED.change_date, EXCLUDED.resource_owner, EXCLUDED.sequence, EXCLUDED.state, EXCLUDED.name)",
							expectedArgs: []interface{}{
								"",
								anyArg{},
								anyArg{},
								"ro-id",
								"instance-id",
								"agg-id",
								uint64(15),
								domain.MFAStateNotReady,
								domain.UserAuthMethodTypeOTP,
								"",
							},
						},
					},
				},
			},
		},
		{
			name: "reduceVerifiedPasswordless",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(user.HumanPasswordlessTokenVerifiedType),
					user.AggregateType,
					[]byte(`{
						"webAuthNTokenId": "token-id",
						"webAuthNTokenName": "name"
					}`),
				), user.HumanPasswordlessVerifiedEventMapper),
			},
			reduce: (&userAuthMethodProjection{}).reduceActivateEvent,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       UserAuthMethodTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.user_auth_methods3 SET (change_date, sequence, name, state) = ($1, $2, $3, $4) WHERE (user_id = $5) AND (method_type = $6) AND (resource_owner = $7) AND (token_id = $8)",
							expectedArgs: []interface{}{
								anyArg{},
								uint64(15),
								"name",
								domain.MFAStateReady,
								"agg-id",
								domain.UserAuthMethodTypePasswordless,
								"ro-id",
								"token-id",
							},
						},
					},
				},
			},
		},
		{
			name: "reduceVerifiedU2F",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(user.HumanU2FTokenVerifiedType),
					user.AggregateType,
					[]byte(`{
						"webAuthNTokenId": "token-id",
						"webAuthNTokenName": "name"
					}`),
				), user.HumanU2FVerifiedEventMapper),
			},
			reduce: (&userAuthMethodProjection{}).reduceActivateEvent,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       UserAuthMethodTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.user_auth_methods3 SET (change_date, sequence, name, state) = ($1, $2, $3, $4) WHERE (user_id = $5) AND (method_type = $6) AND (resource_owner = $7) AND (token_id = $8)",
							expectedArgs: []interface{}{
								anyArg{},
								uint64(15),
								"name",
								domain.MFAStateReady,
								"agg-id",
								domain.UserAuthMethodTypeU2F,
								"ro-id",
								"token-id",
							},
						},
					},
				},
			},
		},
		{
			name: "reduceVerifiedOTP",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(user.HumanMFAOTPVerifiedType),
					user.AggregateType,
					[]byte(`{
					}`),
				), user.HumanOTPVerifiedEventMapper),
			},
			reduce: (&userAuthMethodProjection{}).reduceActivateEvent,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       UserAuthMethodTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.user_auth_methods3 SET (change_date, sequence, name, state) = ($1, $2, $3, $4) WHERE (user_id = $5) AND (method_type = $6) AND (resource_owner = $7) AND (token_id = $8)",
							expectedArgs: []interface{}{
								anyArg{},
								uint64(15),
								"",
								domain.MFAStateReady,
								"agg-id",
								domain.UserAuthMethodTypeOTP,
								"ro-id",
								"",
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
