package projection

import (
	"testing"

	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/handler"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
	"github.com/dennigogo/zitadel/internal/repository/instance"
	"github.com/dennigogo/zitadel/internal/repository/org"
	"github.com/dennigogo/zitadel/internal/repository/user"
)

func TestLoginNameProjection_reduces(t *testing.T) {
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
			name: "user.HumanAddedType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(user.HumanAddedType),
					user.AggregateType,
					[]byte(`{
					"userName": "human-added"
				}`),
				), user.HumanAddedEventMapper),
			},
			reduce: (&loginNameProjection{}).reduceUserCreated,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       LoginNameProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.login_names_users (id, user_name, resource_owner, instance_id) VALUES ($1, $2, $3, $4)",
							expectedArgs: []interface{}{
								"agg-id",
								"human-added",
								"ro-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
		{
			name: "user.HumanRegisteredType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(user.HumanRegisteredType),
					user.AggregateType,
					[]byte(`{
					"userName": "human-registered"
				}`),
				), user.HumanRegisteredEventMapper),
			},
			reduce: (&loginNameProjection{}).reduceUserCreated,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       LoginNameProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.login_names_users (id, user_name, resource_owner, instance_id) VALUES ($1, $2, $3, $4)",
							expectedArgs: []interface{}{
								"agg-id",
								"human-registered",
								"ro-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
		{
			name: "user.MachineAddedEventType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(user.MachineAddedEventType),
					user.AggregateType,
					[]byte(`{
					"userName": "machine-added"
				}`),
				), user.MachineAddedEventMapper),
			},
			reduce: (&loginNameProjection{}).reduceUserCreated,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       LoginNameProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.login_names_users (id, user_name, resource_owner, instance_id) VALUES ($1, $2, $3, $4)",
							expectedArgs: []interface{}{
								"agg-id",
								"machine-added",
								"ro-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
		{
			name: "user.UserRemovedType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(user.UserRemovedType),
					user.AggregateType,
					[]byte(`{}`),
				), user.UserRemovedEventMapper),
			},
			reduce: (&loginNameProjection{}).reduceUserRemoved,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       LoginNameProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM projections.login_names_users WHERE (id = $1)",
							expectedArgs: []interface{}{
								"agg-id",
							},
						},
					},
				},
			},
		},
		{
			name: "user.UserUserNameChangedType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(user.UserUserNameChangedType),
					user.AggregateType,
					[]byte(`{
					"userName": "changed"
				}`),
				), user.UsernameChangedEventMapper),
			},
			reduce: (&loginNameProjection{}).reduceUserNameChanged,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       LoginNameProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.login_names_users SET user_name = $1 WHERE (id = $2)",
							expectedArgs: []interface{}{
								"changed",
								"agg-id",
							},
						},
					},
				},
			},
		},
		{
			name: "user.UserDomainClaimedType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(user.UserDomainClaimedType),
					user.AggregateType,
					[]byte(`{
					"userName": "claimed"
				}`),
				), user.DomainClaimedEventMapper),
			},
			reduce: (&loginNameProjection{}).reduceUserDomainClaimed,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       LoginNameProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.login_names_users SET user_name = $1 WHERE (id = $2)",
							expectedArgs: []interface{}{
								"claimed",
								"agg-id",
							},
						},
					},
				},
			},
		},
		{
			name: "org.OrgDomainPolicyAddedEventType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.DomainPolicyAddedEventType),
					user.AggregateType,
					[]byte(`{
					"userLoginMustBeDomain": true
				}`),
				), org.DomainPolicyAddedEventMapper),
			},
			reduce: (&loginNameProjection{}).reduceOrgIAMPolicyAdded,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       LoginNameProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.login_names_policies (must_be_domain, is_default, resource_owner, instance_id) VALUES ($1, $2, $3, $4)",
							expectedArgs: []interface{}{
								true,
								false,
								"ro-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
		{
			name: "org.OrgDomainPolicyChangedEventType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.DomainPolicyChangedEventType),
					user.AggregateType,
					[]byte(`{
					"userLoginMustBeDomain": false
				}`),
				), org.DomainPolicyChangedEventMapper),
			},
			reduce: (&loginNameProjection{}).reduceDomainPolicyChanged,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       LoginNameProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.login_names_policies SET must_be_domain = $1 WHERE (resource_owner = $2)",
							expectedArgs: []interface{}{
								false,
								"ro-id",
							},
						},
					},
				},
			},
		},
		{
			name: "org.OrgDomainPolicyChangedEventType no change",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.DomainPolicyChangedEventType),
					user.AggregateType,
					[]byte(`{}`),
				), org.DomainPolicyChangedEventMapper),
			},
			reduce: (&loginNameProjection{}).reduceDomainPolicyChanged,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       LoginNameProjectionTable,
				executer: &testExecuter{
					executions: []execution{},
				},
			},
		},
		{
			name: "org.OrgDomainPolicyRemovedEventType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.DomainPolicyRemovedEventType),
					user.AggregateType,
					[]byte(`{}`),
				), org.DomainPolicyRemovedEventMapper),
			},
			reduce: (&loginNameProjection{}).reduceDomainPolicyRemoved,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       LoginNameProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM projections.login_names_policies WHERE (resource_owner = $1)",
							expectedArgs: []interface{}{
								"ro-id",
							},
						},
					},
				},
			},
		},
		{
			name: "org.OrgDomainVerifiedEventType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.OrgDomainVerifiedEventType),
					user.AggregateType,
					[]byte(`{
						"domain": "verified"
					}`),
				), org.DomainVerifiedEventMapper),
			},
			reduce: (&loginNameProjection{}).reduceDomainVerified,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       LoginNameProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.login_names_domains (name, resource_owner, instance_id) VALUES ($1, $2, $3)",
							expectedArgs: []interface{}{
								"verified",
								"ro-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
		{
			name: "org.OrgDomainRemovedEventType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.OrgDomainRemovedEventType),
					user.AggregateType,
					[]byte(`{
						"domain": "remove"
					}`),
				), org.DomainRemovedEventMapper),
			},
			reduce: (&loginNameProjection{}).reduceDomainRemoved,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       LoginNameProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM projections.login_names_domains WHERE (name = $1) AND (resource_owner = $2)",
							expectedArgs: []interface{}{
								"remove",
								"ro-id",
							},
						},
					},
				},
			},
		},
		{
			name: "org.OrgDomainPrimarySetEventType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.OrgDomainPrimarySetEventType),
					user.AggregateType,
					[]byte(`{
						"domain": "primary"
					}`),
				), org.DomainPrimarySetEventMapper),
			},
			reduce: (&loginNameProjection{}).reducePrimaryDomainSet,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       LoginNameProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.login_names_domains SET is_primary = $1 WHERE (resource_owner = $2) AND (is_primary = $3)",
							expectedArgs: []interface{}{
								false,
								"ro-id",
								true,
							},
						},
						{
							expectedStmt: "UPDATE projections.login_names_domains SET is_primary = $1 WHERE (name = $2) AND (resource_owner = $3)",
							expectedArgs: []interface{}{
								true,
								"primary",
								"ro-id",
							},
						},
					},
				},
			},
		},
		{
			name: "iam.OrgDomainPolicyAddedEventType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.DomainPolicyAddedEventType),
					user.AggregateType,
					[]byte(`{
					"userLoginMustBeDomain": true
				}`),
				), instance.DomainPolicyAddedEventMapper),
			},
			reduce: (&loginNameProjection{}).reduceOrgIAMPolicyAdded,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       LoginNameProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.login_names_policies (must_be_domain, is_default, resource_owner, instance_id) VALUES ($1, $2, $3, $4)",
							expectedArgs: []interface{}{
								true,
								true,
								"ro-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
		{
			name: "iam.OrgDomainPolicyChangedEventType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.DomainPolicyChangedEventType),
					user.AggregateType,
					[]byte(`{
					"userLoginMustBeDomain": false
				}`),
				), instance.DomainPolicyChangedEventMapper),
			},
			reduce: (&loginNameProjection{}).reduceDomainPolicyChanged,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       LoginNameProjectionTable,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.login_names_policies SET must_be_domain = $1 WHERE (resource_owner = $2)",
							expectedArgs: []interface{}{
								false,
								"ro-id",
							},
						},
					},
				},
			},
		},
		{
			name: "iam.OrgDomainPolicyChangedEventType no change",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.DomainPolicyChangedEventType),
					user.AggregateType,
					[]byte(`{}`),
				), instance.DomainPolicyChangedEventMapper),
			},
			reduce: (&loginNameProjection{}).reduceDomainPolicyChanged,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				projection:       LoginNameProjectionTable,
				executer: &testExecuter{
					executions: []execution{},
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
