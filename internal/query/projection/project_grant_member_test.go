package projection

import (
	"testing"

	"github.com/dennigogo/zitadel/internal/database"
	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/handler"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
	"github.com/dennigogo/zitadel/internal/repository/org"
	"github.com/dennigogo/zitadel/internal/repository/project"
	"github.com/dennigogo/zitadel/internal/repository/user"
)

func TestProjectGrantMemberProjection_reduces(t *testing.T) {
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
			name: "project.GrantMemberAddedType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(project.GrantMemberAddedType),
					project.AggregateType,
					[]byte(`{
					"userId": "user-id",
					"roles": ["role"],
					"grantId": "grant-id"
				}`),
				), project.GrantMemberAddedEventMapper),
			},
			reduce: (&projectGrantMemberProjection{}).reduceAdded,
			want: wantReduce{
				aggregateType:    project.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       ProjectGrantMemberProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.project_grant_members2 (user_id, roles, creation_date, change_date, sequence, resource_owner, instance_id, project_id, grant_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
							expectedArgs: []interface{}{
								"user-id",
								database.StringArray{"role"},
								anyArg{},
								anyArg{},
								uint64(15),
								"ro-id",
								"instance-id",
								"agg-id",
								"grant-id",
							},
						},
					},
				},
			},
		},
		{
			name: "project.GrantMemberChangedType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(project.GrantMemberChangedType),
					project.AggregateType,
					[]byte(`{
					"userId": "user-id",
					"roles": ["role", "changed"],
					"grantId": "grant-id"
				}`),
				), project.GrantMemberChangedEventMapper),
			},
			reduce: (&projectGrantMemberProjection{}).reduceChanged,
			want: wantReduce{
				aggregateType:    project.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       ProjectGrantMemberProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.project_grant_members2 SET (roles, change_date, sequence) = ($1, $2, $3) WHERE (user_id = $4) AND (project_id = $5) AND (grant_id = $6)",
							expectedArgs: []interface{}{
								database.StringArray{"role", "changed"},
								anyArg{},
								uint64(15),
								"user-id",
								"agg-id",
								"grant-id",
							},
						},
					},
				},
			},
		},
		{
			name: "project.GrantMemberCascadeRemovedType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(project.GrantMemberCascadeRemovedType),
					project.AggregateType,
					[]byte(`{
					"userId": "user-id",
					"grantId": "grant-id"
				}`),
				), project.GrantMemberCascadeRemovedEventMapper),
			},
			reduce: (&projectGrantMemberProjection{}).reduceCascadeRemoved,
			want: wantReduce{
				aggregateType:    project.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       ProjectGrantMemberProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM projections.project_grant_members2 WHERE (user_id = $1) AND (project_id = $2) AND (grant_id = $3)",
							expectedArgs: []interface{}{
								"user-id",
								"agg-id",
								"grant-id",
							},
						},
					},
				},
			},
		},
		{
			name: "project.GrantMemberRemovedType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(project.GrantMemberRemovedType),
					project.AggregateType,
					[]byte(`{
					"userId": "user-id",
					"grantId": "grant-id"
				}`),
				), project.GrantMemberRemovedEventMapper),
			},
			reduce: (&projectGrantMemberProjection{}).reduceRemoved,
			want: wantReduce{
				aggregateType:    project.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       ProjectGrantMemberProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM projections.project_grant_members2 WHERE (user_id = $1) AND (project_id = $2) AND (grant_id = $3)",
							expectedArgs: []interface{}{
								"user-id",
								"agg-id",
								"grant-id",
							},
						},
					},
				},
			},
		},
		{
			name: "user.UserRemovedEventType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(user.UserRemovedType),
					user.AggregateType,
					[]byte(`{}`),
				), user.UserRemovedEventMapper),
			},
			reduce: (&projectGrantMemberProjection{}).reduceUserRemoved,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       ProjectGrantMemberProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM projections.project_grant_members2 WHERE (user_id = $1)",
							expectedArgs: []interface{}{
								"agg-id",
							},
						},
					},
				},
			},
		},
		{
			name: "org.OrgRemovedEventType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.OrgRemovedEventType),
					org.AggregateType,
					[]byte(`{}`),
				), org.OrgRemovedEventMapper),
			},
			reduce: (&projectGrantMemberProjection{}).reduceOrgRemoved,
			want: wantReduce{
				aggregateType:    org.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       ProjectGrantMemberProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM projections.project_grant_members2 WHERE (resource_owner = $1)",
							expectedArgs: []interface{}{
								"agg-id",
							},
						},
					},
				},
			},
		},
		{
			name: "project.ProjectRemovedEventType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(project.ProjectRemovedType),
					project.AggregateType,
					[]byte(`{}`),
				), project.ProjectRemovedEventMapper),
			},
			reduce: (&projectGrantMemberProjection{}).reduceProjectRemoved,
			want: wantReduce{
				aggregateType:    project.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       ProjectGrantMemberProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM projections.project_grant_members2 WHERE (project_id = $1)",
							expectedArgs: []interface{}{
								"agg-id",
							},
						},
					},
				},
			},
		},
		{
			name: "project.GrantRemovedEventType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(project.GrantRemovedType),
					project.AggregateType,
					[]byte(`{"grantId": "grant-id"}`),
				), project.GrantRemovedEventMapper),
			},
			reduce: (&projectGrantMemberProjection{}).reduceProjectGrantRemoved,
			want: wantReduce{
				aggregateType:    project.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       ProjectGrantMemberProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM projections.project_grant_members2 WHERE (grant_id = $1) AND (project_id = $2)",
							expectedArgs: []interface{}{
								"grant-id",
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
