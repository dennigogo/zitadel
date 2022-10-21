package projection

import (
	"testing"

	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/handler"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
	"github.com/dennigogo/zitadel/internal/repository/instance"
	"github.com/dennigogo/zitadel/internal/repository/org"
)

func TestPasswordComplexityProjection_reduces(t *testing.T) {
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
			name: "org.reduceAdded",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.PasswordComplexityPolicyAddedEventType),
					org.AggregateType,
					[]byte(`{
	"minLength": 10,
	"hasLowercase": true,
	"hasUppercase": true,
	"HasNumber": true,
	"HasSymbol": true
}`),
				), org.PasswordComplexityPolicyAddedEventMapper),
			},
			reduce: (&passwordComplexityProjection{}).reduceAdded,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("org"),
				sequence:         15,
				previousSequence: 10,
				projection:       PasswordComplexityTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.password_complexity_policies (creation_date, change_date, sequence, id, state, min_length, has_lowercase, has_uppercase, has_symbol, has_number, resource_owner, instance_id, is_default) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)",
							expectedArgs: []interface{}{
								anyArg{},
								anyArg{},
								uint64(15),
								"agg-id",
								domain.PolicyStateActive,
								uint64(10),
								true,
								true,
								true,
								true,
								"ro-id",
								"instance-id",
								false,
							},
						},
					},
				},
			},
		},
		{
			name:   "org.reduceChanged",
			reduce: (&passwordComplexityProjection{}).reduceChanged,
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.PasswordComplexityPolicyChangedEventType),
					org.AggregateType,
					[]byte(`{
			"minLength": 11,
			"hasLowercase": true,
			"hasUppercase": true,
			"HasNumber": true,
			"HasSymbol": true
		}`),
				), org.PasswordComplexityPolicyChangedEventMapper),
			},
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("org"),
				sequence:         15,
				previousSequence: 10,
				projection:       PasswordComplexityTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.password_complexity_policies SET (change_date, sequence, min_length, has_lowercase, has_uppercase, has_symbol, has_number) = ($1, $2, $3, $4, $5, $6, $7) WHERE (id = $8)",
							expectedArgs: []interface{}{
								anyArg{},
								uint64(15),
								uint64(11),
								true,
								true,
								true,
								true,
								"agg-id",
							},
						},
					},
				},
			},
		},
		{
			name:   "org.reduceRemoved",
			reduce: (&passwordComplexityProjection{}).reduceRemoved,
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.PasswordComplexityPolicyRemovedEventType),
					org.AggregateType,
					nil,
				), org.PasswordComplexityPolicyRemovedEventMapper),
			},
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("org"),
				sequence:         15,
				previousSequence: 10,
				projection:       PasswordComplexityTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM projections.password_complexity_policies WHERE (id = $1)",
							expectedArgs: []interface{}{
								"agg-id",
							},
						},
					},
				},
			},
		},
		{
			name:   "instance.reduceAdded",
			reduce: (&passwordComplexityProjection{}).reduceAdded,
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.PasswordComplexityPolicyAddedEventType),
					instance.AggregateType,
					[]byte(`{
			"minLength": 10,
			"hasLowercase": true,
			"hasUppercase": true,
			"HasNumber": true,
			"HasSymbol": true
					}`),
				), instance.PasswordComplexityPolicyAddedEventMapper),
			},
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				projection:       PasswordComplexityTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.password_complexity_policies (creation_date, change_date, sequence, id, state, min_length, has_lowercase, has_uppercase, has_symbol, has_number, resource_owner, instance_id, is_default) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)",
							expectedArgs: []interface{}{
								anyArg{},
								anyArg{},
								uint64(15),
								"agg-id",
								domain.PolicyStateActive,
								uint64(10),
								true,
								true,
								true,
								true,
								"ro-id",
								"instance-id",
								true,
							},
						},
					},
				},
			},
		},
		{
			name:   "instance.reduceChanged",
			reduce: (&passwordComplexityProjection{}).reduceChanged,
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.PasswordComplexityPolicyChangedEventType),
					instance.AggregateType,
					[]byte(`{
			"minLength": 10,
			"hasLowercase": true,
			"hasUppercase": true,
			"HasNumber": true,
			"HasSymbol": true
					}`),
				), instance.PasswordComplexityPolicyChangedEventMapper),
			},
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				projection:       PasswordComplexityTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.password_complexity_policies SET (change_date, sequence, min_length, has_lowercase, has_uppercase, has_symbol, has_number) = ($1, $2, $3, $4, $5, $6, $7) WHERE (id = $8)",
							expectedArgs: []interface{}{
								anyArg{},
								uint64(15),
								uint64(10),
								true,
								true,
								true,
								true,
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
