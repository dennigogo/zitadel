package command

import (
	"context"
	"github.com/dennigogo/zitadel/internal/domain"
	caos_errs "github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/repository/user"
	"github.com/dennigogo/zitadel/internal/telemetry/tracing"
)

func (c *Commands) AddMachine(ctx context.Context, orgID string, machine *domain.Machine) (*domain.Machine, error) {
	if !machine.IsValid() {
		return nil, caos_errs.ThrowInvalidArgument(nil, "COMMAND-bm9Ds", "Errors.User.Invalid")
	}
	domainPolicy, err := c.getOrgDomainPolicy(ctx, orgID)
	if err != nil {
		return nil, caos_errs.ThrowPreconditionFailed(err, "COMMAND-3M9fs", "Errors.Org.DomainPolicy.NotFound")
	}
	userID, err := c.idGenerator.Next()
	if err != nil {
		return nil, err
	}
	return c.addMachineWithID(ctx, orgID, userID, machine, domainPolicy)
}

func (c *Commands) AddMachineWithID(ctx context.Context, orgID string, userID string, machine *domain.Machine) (*domain.Machine, error) {
	existingMachine, err := c.machineWriteModelByID(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	if isUserStateExists(existingMachine.UserState) {
		return nil, caos_errs.ThrowPreconditionFailed(nil, "COMMAND-k2una", "Errors.User.AlreadyExisting")
	}
	domainPolicy, err := c.getOrgDomainPolicy(ctx, orgID)
	if err != nil {
		return nil, caos_errs.ThrowPreconditionFailed(err, "COMMAND-3M9fs", "Errors.Org.DomainPolicy.NotFound")
	}
	if !domainPolicy.UserLoginMustBeDomain {
		return nil, caos_errs.ThrowPreconditionFailed(nil, "COMMAND-6M0dd", "Errors.User.Invalid")
	}
	return c.addMachineWithID(ctx, orgID, userID, machine, domainPolicy)
}

func (c *Commands) addMachineWithID(ctx context.Context, orgID string, userID string, machine *domain.Machine, domainPolicy *domain.DomainPolicy) (*domain.Machine, error) {

	machine.AggregateID = userID
	addedMachine := NewMachineWriteModel(machine.AggregateID, orgID)
	userAgg := UserAggregateFromWriteModel(&addedMachine.WriteModel)
	events, err := c.eventstore.Push(ctx, user.NewMachineAddedEvent(
		ctx,
		userAgg,
		machine.Username,
		machine.Name,
		machine.Description,
		domainPolicy.UserLoginMustBeDomain,
	))
	if err != nil {
		return nil, err
	}
	err = AppendAndReduce(addedMachine, events...)
	if err != nil {
		return nil, err
	}
	return writeModelToMachine(addedMachine), nil
}

func (c *Commands) ChangeMachine(ctx context.Context, machine *domain.Machine) (*domain.Machine, error) {
	existingMachine, err := c.machineWriteModelByID(ctx, machine.AggregateID, machine.ResourceOwner)
	if err != nil {
		return nil, err
	}
	if !isUserStateExists(existingMachine.UserState) {
		return nil, caos_errs.ThrowNotFound(nil, "COMMAND-5M0od", "Errors.User.NotFound")
	}

	userAgg := UserAggregateFromWriteModel(&existingMachine.WriteModel)
	changedEvent, hasChanged, err := existingMachine.NewChangedEvent(ctx, userAgg, machine.Name, machine.Description)
	if err != nil {
		return nil, err
	}
	if !hasChanged {
		return nil, caos_errs.ThrowPreconditionFailed(nil, "COMMAND-2n8vs", "Errors.User.NotChanged")
	}

	events, err := c.eventstore.Push(ctx, changedEvent)
	if err != nil {
		return nil, err
	}
	err = AppendAndReduce(existingMachine, events...)
	if err != nil {
		return nil, err
	}
	return writeModelToMachine(existingMachine), nil
}

func (c *Commands) machineWriteModelByID(ctx context.Context, userID, resourceOwner string) (writeModel *MachineWriteModel, err error) {
	if userID == "" {
		return nil, caos_errs.ThrowInvalidArgument(nil, "COMMAND-0Plof", "Errors.User.UserIDMissing")
	}
	ctx, span := tracing.NewSpan(ctx)
	defer func() { span.EndWithError(err) }()

	writeModel = NewMachineWriteModel(userID, resourceOwner)
	err = c.eventstore.FilterToQueryReducer(ctx, writeModel)
	if err != nil {
		return nil, err
	}
	return writeModel, nil
}
