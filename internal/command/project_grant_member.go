package command

import (
	"context"
	"reflect"

	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/repository/project"
	"github.com/dennigogo/zitadel/internal/telemetry/tracing"
)

func (c *Commands) AddProjectGrantMember(ctx context.Context, member *domain.ProjectGrantMember) (*domain.ProjectGrantMember, error) {
	if !member.IsValid() {
		return nil, errors.ThrowInvalidArgument(nil, "PROJECT-8fi7G", "Errors.Project.Grant.Member.Invalid")
	}
	if len(domain.CheckForInvalidRoles(member.Roles, domain.ProjectGrantRolePrefix, c.zitadelRoles)) > 0 {
		return nil, errors.ThrowInvalidArgument(nil, "PROJECT-m9gKK", "Errors.Project.Grant.Member.Invalid")
	}
	err := c.checkUserExists(ctx, member.UserID, "")
	if err != nil {
		return nil, err
	}
	addedMember := NewProjectGrantMemberWriteModel(member.AggregateID, member.UserID, member.GrantID)
	err = c.eventstore.FilterToQueryReducer(ctx, addedMember)
	if err != nil {
		return nil, err
	}
	if addedMember.State == domain.MemberStateActive {
		return nil, errors.ThrowAlreadyExists(nil, "PROJECT-16dVN", "Errors.Project.Member.AlreadyExists")
	}
	projectAgg := ProjectAggregateFromWriteModel(&addedMember.WriteModel)
	pushedEvents, err := c.eventstore.Push(
		ctx,
		project.NewProjectGrantMemberAddedEvent(ctx, projectAgg, member.UserID, member.GrantID, member.Roles...))
	if err != nil {
		return nil, err
	}
	err = AppendAndReduce(addedMember, pushedEvents...)
	if err != nil {
		return nil, err
	}

	return memberWriteModelToProjectGrantMember(addedMember), nil
}

// ChangeProjectGrantMember updates an existing member
func (c *Commands) ChangeProjectGrantMember(ctx context.Context, member *domain.ProjectGrantMember) (*domain.ProjectGrantMember, error) {
	if !member.IsValid() {
		return nil, errors.ThrowInvalidArgument(nil, "PROJECT-109fs", "Errors.Project.Member.Invalid")
	}
	if len(domain.CheckForInvalidRoles(member.Roles, domain.ProjectGrantRolePrefix, c.zitadelRoles)) > 0 {
		return nil, errors.ThrowInvalidArgument(nil, "PROJECT-m0sDf", "Errors.Project.Member.Invalid")
	}

	existingMember, err := c.projectGrantMemberWriteModelByID(ctx, member.AggregateID, member.UserID, member.GrantID)
	if err != nil {
		return nil, err
	}

	if reflect.DeepEqual(existingMember.Roles, member.Roles) {
		return nil, errors.ThrowPreconditionFailed(nil, "PROJECT-2n8vx", "Errors.Project.Member.RolesNotChanged")
	}
	projectAgg := ProjectAggregateFromWriteModel(&existingMember.WriteModel)
	pushedEvents, err := c.eventstore.Push(
		ctx,
		project.NewProjectGrantMemberChangedEvent(ctx, projectAgg, member.UserID, member.GrantID, member.Roles...))
	if err != nil {
		return nil, err
	}
	err = AppendAndReduce(existingMember, pushedEvents...)
	if err != nil {
		return nil, err
	}

	return memberWriteModelToProjectGrantMember(existingMember), nil
}

func (c *Commands) RemoveProjectGrantMember(ctx context.Context, projectID, userID, grantID string) (*domain.ObjectDetails, error) {
	if projectID == "" || userID == "" || grantID == "" {
		return nil, errors.ThrowInvalidArgument(nil, "PROJECT-66mHd", "Errors.Project.Member.Invalid")
	}
	m, err := c.projectGrantMemberWriteModelByID(ctx, projectID, userID, grantID)
	if err != nil {
		return nil, err
	}

	projectAgg := ProjectAggregateFromWriteModel(&m.WriteModel)
	removeEvent := c.removeProjectGrantMember(ctx, projectAgg, userID, grantID, false)
	pushedEvents, err := c.eventstore.Push(ctx, removeEvent)
	if err != nil {
		return nil, err
	}
	err = AppendAndReduce(m, pushedEvents...)
	if err != nil {
		return nil, err
	}
	return writeModelToObjectDetails(&m.WriteModel), nil
}

func (c *Commands) removeProjectGrantMember(ctx context.Context, projectAgg *eventstore.Aggregate, userID, grantID string, cascade bool) eventstore.Command {
	if cascade {
		return project.NewProjectGrantMemberCascadeRemovedEvent(
			ctx,
			projectAgg,
			userID,
			grantID)
	} else {
		return project.NewProjectGrantMemberRemovedEvent(ctx, projectAgg, userID, grantID)
	}
}

func (c *Commands) projectGrantMemberWriteModelByID(ctx context.Context, projectID, userID, grantID string) (member *ProjectGrantMemberWriteModel, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer func() { span.EndWithError(err) }()

	writeModel := NewProjectGrantMemberWriteModel(projectID, userID, grantID)
	err = c.eventstore.FilterToQueryReducer(ctx, writeModel)
	if err != nil {
		return nil, err
	}

	if writeModel.State == domain.MemberStateUnspecified || writeModel.State == domain.MemberStateRemoved {
		return nil, errors.ThrowNotFound(nil, "PROJECT-37fug", "Errors.NotFound")
	}

	return writeModel, nil
}
