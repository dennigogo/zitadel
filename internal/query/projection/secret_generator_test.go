package projection

import (
	"testing"
	"time"

	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/handler"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
	"github.com/dennigogo/zitadel/internal/repository/instance"
)

func TestSecretGeneratorProjection_reduces(t *testing.T) {
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
			name: "reduceSecretGeneratorRemoved",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.SecretGeneratorRemovedEventType),
					instance.AggregateType,
					[]byte(`{"generatorType": 1}`),
				), instance.SecretGeneratorRemovedEventMapper),
			},
			reduce: (&secretGeneratorProjection{}).reduceSecretGeneratorRemoved,
			want: wantReduce{
				projection:       SecretGeneratorProjectionTable,
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM projections.secret_generators2 WHERE (aggregate_id = $1) AND (generator_type = $2)",
							expectedArgs: []interface{}{
								"agg-id",
								domain.SecretGeneratorTypeInitCode,
							},
						},
					},
				},
			},
		},
		{
			name: "reduceSecretGeneratorChanged",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.SecretGeneratorChangedEventType),
					instance.AggregateType,
					[]byte(`{"generatorType": 1, "length": 4, "expiry": 10000000, "includeLowerLetters": true, "includeUpperLetters": true, "includeDigits": true, "includeSymbols": true}`),
				), instance.SecretGeneratorChangedEventMapper),
			},
			reduce: (&secretGeneratorProjection{}).reduceSecretGeneratorChanged,
			want: wantReduce{
				projection:       SecretGeneratorProjectionTable,
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.secret_generators2 SET (change_date, sequence, length, expiry, include_lower_letters, include_upper_letters, include_digits, include_symbols) = ($1, $2, $3, $4, $5, $6, $7, $8) WHERE (aggregate_id = $9) AND (generator_type = $10)",
							expectedArgs: []interface{}{
								anyArg{},
								uint64(15),
								uint(4),
								time.Millisecond * 10,
								true,
								true,
								true,
								true,
								"agg-id",
								domain.SecretGeneratorTypeInitCode,
							},
						},
					},
				},
			},
		},
		{
			name: "reduceSecretGeneratorAdded",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.SecretGeneratorAddedEventType),
					instance.AggregateType,
					[]byte(`{"generatorType": 1, "length": 4, "expiry": 10000000, "includeLowerLetters": true, "includeUpperLetters": true, "includeDigits": true, "includeSymbols": true}`),
				), instance.SecretGeneratorAddedEventMapper),
			},
			reduce: (&secretGeneratorProjection{}).reduceSecretGeneratorAdded,
			want: wantReduce{
				projection:       SecretGeneratorProjectionTable,
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.secret_generators2 (aggregate_id, generator_type, creation_date, change_date, resource_owner, instance_id, sequence, length, expiry, include_lower_letters, include_upper_letters, include_digits, include_symbols) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)",
							expectedArgs: []interface{}{
								"agg-id",
								domain.SecretGeneratorTypeInitCode,
								anyArg{},
								anyArg{},
								"ro-id",
								"instance-id",
								uint64(15),
								uint(4),
								time.Millisecond * 10,
								true,
								true,
								true,
								true,
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
