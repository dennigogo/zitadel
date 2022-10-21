package projection

import (
	"testing"

	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/handler"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
	"github.com/dennigogo/zitadel/internal/repository/project"
)

func TestProjectProjection_reduces(t *testing.T) {
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
			name: "reduceProjectRemoved",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(project.ProjectRemovedType),
					project.AggregateType,
					nil,
				), project.ProjectRemovedEventMapper),
			},
			reduce: (&projectProjection{}).reduceProjectRemoved,
			want: wantReduce{
				projection:       ProjectProjectionTable,
				aggregateType:    eventstore.AggregateType("project"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM projections.projects2 WHERE (id = $1)",
							expectedArgs: []interface{}{
								"agg-id",
							},
						},
					},
				},
			},
		},
		{
			name: "reduceProjectReactivated",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(project.ProjectReactivatedType),
					project.AggregateType,
					nil,
				), project.ProjectReactivatedEventMapper),
			},
			reduce: (&projectProjection{}).reduceProjectReactivated,
			want: wantReduce{
				projection:       ProjectProjectionTable,
				aggregateType:    eventstore.AggregateType("project"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.projects2 SET (change_date, sequence, state) = ($1, $2, $3) WHERE (id = $4)",
							expectedArgs: []interface{}{
								anyArg{},
								uint64(15),
								domain.ProjectStateActive,
								"agg-id",
							},
						},
					},
				},
			},
		},
		{
			name: "reduceProjectDeactivated",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(project.ProjectDeactivatedType),
					project.AggregateType,
					nil,
				), project.ProjectDeactivatedEventMapper),
			},
			reduce: (&projectProjection{}).reduceProjectDeactivated,
			want: wantReduce{
				projection:       ProjectProjectionTable,
				aggregateType:    eventstore.AggregateType("project"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.projects2 SET (change_date, sequence, state) = ($1, $2, $3) WHERE (id = $4)",
							expectedArgs: []interface{}{
								anyArg{},
								uint64(15),
								domain.ProjectStateInactive,
								"agg-id",
							},
						},
					},
				},
			},
		},
		{
			name: "reduceProjectChanged",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(project.ProjectChangedType),
					project.AggregateType,
					[]byte(`{"name": "new name", "projectRoleAssertion": true, "projectRoleCheck": true, "hasProjectCheck": true, "privateLabelingSetting": 1}`),
				), project.ProjectChangeEventMapper),
			},
			reduce: (&projectProjection{}).reduceProjectChanged,
			want: wantReduce{
				projection:       ProjectProjectionTable,
				aggregateType:    eventstore.AggregateType("project"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.projects2 SET (change_date, sequence, name, project_role_assertion, project_role_check, has_project_check, private_labeling_setting) = ($1, $2, $3, $4, $5, $6, $7) WHERE (id = $8)",
							expectedArgs: []interface{}{
								anyArg{},
								uint64(15),
								"new name",
								true,
								true,
								true,
								domain.PrivateLabelingSettingEnforceProjectResourceOwnerPolicy,
								"agg-id",
							},
						},
					},
				},
			},
		},
		{
			name: "reduceProjectChanged no changes",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(project.ProjectChangedType),
					project.AggregateType,
					[]byte(`{}`),
				), project.ProjectChangeEventMapper),
			},
			reduce: (&projectProjection{}).reduceProjectChanged,
			want: wantReduce{
				projection:       ProjectProjectionTable,
				aggregateType:    eventstore.AggregateType("project"),
				sequence:         15,
				previousSequence: 10,
				executer:         &testExecuter{},
			},
		},
		{
			name: "reduceProjectAdded",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(project.ProjectAddedType),
					project.AggregateType,
					[]byte(`{"name": "name", "projectRoleAssertion": true, "projectRoleCheck": true, "hasProjectCheck": true, "privateLabelingSetting": 1}`),
				), project.ProjectAddedEventMapper),
			},
			reduce: (&projectProjection{}).reduceProjectAdded,
			want: wantReduce{
				projection:       ProjectProjectionTable,
				aggregateType:    eventstore.AggregateType("project"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.projects2 (id, creation_date, change_date, resource_owner, instance_id, sequence, name, project_role_assertion, project_role_check, has_project_check, private_labeling_setting, state) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)",
							expectedArgs: []interface{}{
								"agg-id",
								anyArg{},
								anyArg{},
								"ro-id",
								"instance-id",
								uint64(15),
								"name",
								true,
								true,
								true,
								domain.PrivateLabelingSettingEnforceProjectResourceOwnerPolicy,
								domain.ProjectStateActive,
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
