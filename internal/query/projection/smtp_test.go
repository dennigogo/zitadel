package projection

import (
	"testing"

	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/handler"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
	"github.com/dennigogo/zitadel/internal/repository/instance"
)

func TestSMTPConfigProjection_reduces(t *testing.T) {
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
			name: "reduceSMTPConfigChanged",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.SMTPConfigChangedEventType),
					instance.AggregateType,
					[]byte(`{
						"tls": true,
						"senderAddress": "sender",
						"senderName": "name",
						"host": "host",
						"user": "user"
					}`,
					),
				), instance.SMTPConfigChangedEventMapper),
			},
			reduce: (&smtpConfigProjection{}).reduceSMTPConfigChanged,
			want: wantReduce{
				projection:       SMTPConfigProjectionTable,
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.smtp_configs SET (change_date, sequence, tls, sender_address, sender_name, host, username) = ($1, $2, $3, $4, $5, $6, $7) WHERE (aggregate_id = $8) AND (instance_id = $9)",
							expectedArgs: []interface{}{
								anyArg{},
								uint64(15),
								true,
								"sender",
								"name",
								"host",
								"user",
								"agg-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
		{
			name: "reduceSMTPConfigAdded",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.SMTPConfigAddedEventType),
					instance.AggregateType,
					[]byte(`{
						"tls": true,
						"senderAddress": "sender",
						"senderName": "name",
						"host": "host",
						"user": "user",
						"password": {
							"cryptoType": 0,
							"algorithm": "RSA-265",
							"keyId": "key-id"
						}
					}`),
				), instance.SMTPConfigAddedEventMapper),
			},
			reduce: (&smtpConfigProjection{}).reduceSMTPConfigAdded,
			want: wantReduce{
				projection:       SMTPConfigProjectionTable,
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.smtp_configs (aggregate_id, creation_date, change_date, resource_owner, instance_id, sequence, tls, sender_address, sender_name, host, username, password) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)",
							expectedArgs: []interface{}{
								"agg-id",
								anyArg{},
								anyArg{},
								"ro-id",
								"instance-id",
								uint64(15),
								true,
								"sender",
								"name",
								"host",
								"user",
								anyArg{},
							},
						},
					},
				},
			},
		},
		{
			name: "reduceSMTPConfigPasswordChanged",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.SMTPConfigPasswordChangedEventType),
					instance.AggregateType,
					[]byte(`{
						"password": {
							"cryptoType": 0,
							"algorithm": "RSA-265",
							"keyId": "key-id"
						}
					}`),
				), instance.SMTPConfigPasswordChangedEventMapper),
			},
			reduce: (&smtpConfigProjection{}).reduceSMTPConfigPasswordChanged,
			want: wantReduce{
				projection:       SMTPConfigProjectionTable,
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.smtp_configs SET (change_date, sequence, password) = ($1, $2, $3) WHERE (aggregate_id = $4) AND (instance_id = $5)",
							expectedArgs: []interface{}{
								anyArg{},
								uint64(15),
								anyArg{},
								"agg-id",
								"instance-id",
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
