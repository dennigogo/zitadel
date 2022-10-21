package projection

import (
	"testing"

	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/handler"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
	"github.com/dennigogo/zitadel/internal/repository/instance"
)

func TestInstanceProjection_reduces(t *testing.T) {
	type args struct {
		event func(t *testing.T) eventstore.Event
	}
	tests := []struct {
		name   string
		args   args
		reduce func(event eventstore.Event) (*handler.Statement, error)
		want   wantReduce
	}{{
		name: "reduceInstanceAdded",
		args: args{
			event: getEvent(testEvent(
				repository.EventType(instance.InstanceAddedEventType),
				instance.AggregateType,
				[]byte(`{"name": "Name"}`),
			), instance.InstanceAddedEventMapper),
		},
		reduce: (&instanceProjection{}).reduceInstanceAdded,
		want: wantReduce{
			projection:       InstanceProjectionTable,
			aggregateType:    eventstore.AggregateType("instance"),
			sequence:         15,
			previousSequence: 10,
			executer: &testExecuter{
				executions: []execution{
					{
						expectedStmt: "INSERT INTO projections.instances (id, creation_date, change_date, sequence, name) VALUES ($1, $2, $3, $4, $5)",
						expectedArgs: []interface{}{
							"instance-id",
							anyArg{},
							anyArg{},
							uint64(15),
							"Name",
						},
					},
				},
			},
		},
	},
		{
			name: "reduceDefaultOrgSet",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.DefaultOrgSetEventType),
					instance.AggregateType,
					[]byte(`{"orgId": "orgid"}`),
				), instance.DefaultOrgSetMapper),
			},
			reduce: (&instanceProjection{}).reduceDefaultOrgSet,
			want: wantReduce{
				projection:       InstanceProjectionTable,
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.instances SET (change_date, sequence, default_org_id) = ($1, $2, $3) WHERE (id = $4)",
							expectedArgs: []interface{}{
								anyArg{},
								uint64(15),
								"orgid",
								"instance-id",
							},
						},
					},
				},
			},
		},
		{
			name: "reduceProjectIDSet",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.ProjectSetEventType),
					instance.AggregateType,
					[]byte(`{"iamProjectId": "project-id"}`),
				), instance.ProjectSetMapper),
			},
			reduce: (&instanceProjection{}).reduceIAMProjectSet,
			want: wantReduce{
				projection:       InstanceProjectionTable,
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.instances SET (change_date, sequence, iam_project_id) = ($1, $2, $3) WHERE (id = $4)",
							expectedArgs: []interface{}{
								anyArg{},
								uint64(15),
								"project-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
		{
			name: "reduceDefaultLanguageSet",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.DefaultLanguageSetEventType),
					instance.AggregateType,
					[]byte(`{"language": "en"}`),
				), instance.DefaultLanguageSetMapper),
			},
			reduce: (&instanceProjection{}).reduceDefaultLanguageSet,
			want: wantReduce{
				projection:       InstanceProjectionTable,
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.instances SET (change_date, sequence, default_language) = ($1, $2, $3) WHERE (id = $4)",
							expectedArgs: []interface{}{
								anyArg{},
								uint64(15),
								"en",
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
