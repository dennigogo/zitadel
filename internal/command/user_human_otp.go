package command

import (
	"context"

	"github.com/dennigogo/zitadel/internal/crypto"

	"github.com/dennigogo/zitadel/internal/domain"
	caos_errs "github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore/v1/models"
	"github.com/dennigogo/zitadel/internal/repository/user"
	"github.com/dennigogo/zitadel/internal/telemetry/tracing"
	"github.com/zitadel/logging"
)

func (c *Commands) ImportHumanOTP(ctx context.Context, userID, userAgentID, resourceowner string, key string) error {
	encryptedSecret, err := crypto.Encrypt([]byte(key), c.multifactors.OTP.CryptoMFA)
	if err != nil {
		return err
	}
	if err = c.checkUserExists(ctx, userID, resourceowner); err != nil {
		return err
	}

	otpWriteModel, err := c.otpWriteModelByID(ctx, userID, resourceowner)
	if err != nil {
		return err
	}
	if otpWriteModel.State == domain.MFAStateReady {
		return caos_errs.ThrowAlreadyExists(nil, "COMMAND-do9se", "Errors.User.MFA.OTP.AlreadyReady")
	}
	userAgg := UserAggregateFromWriteModel(&otpWriteModel.WriteModel)

	_, err = c.eventstore.Push(ctx,
		user.NewHumanOTPAddedEvent(ctx, userAgg, encryptedSecret),
		user.NewHumanOTPVerifiedEvent(ctx, userAgg, userAgentID),
	)
	return err
}

func (c *Commands) AddHumanOTP(ctx context.Context, userID, resourceowner string) (*domain.OTP, error) {
	if userID == "" {
		return nil, caos_errs.ThrowInvalidArgument(nil, "COMMAND-5M0sd", "Errors.User.UserIDMissing")
	}
	human, err := c.getHuman(ctx, userID, resourceowner)
	if err != nil {
		logging.Log("COMMAND-DAqe1").WithError(err).WithField("traceID", tracing.TraceIDFromCtx(ctx)).Debug("unable to get human for loginname")
		return nil, caos_errs.ThrowPreconditionFailed(err, "COMMAND-MM9fs", "Errors.User.NotFound")
	}
	org, err := c.getOrg(ctx, human.ResourceOwner)
	if err != nil {
		logging.Log("COMMAND-Cm0ds").WithError(err).WithField("traceID", tracing.TraceIDFromCtx(ctx)).Debug("unable to get org for loginname")
		return nil, caos_errs.ThrowPreconditionFailed(err, "COMMAND-55M9f", "Errors.Org.NotFound")
	}
	orgPolicy, err := c.getOrgDomainPolicy(ctx, org.AggregateID)
	if err != nil {
		logging.Log("COMMAND-y5zv9").WithError(err).WithField("traceID", tracing.TraceIDFromCtx(ctx)).Debug("unable to get org policy for loginname")
		return nil, caos_errs.ThrowPreconditionFailed(err, "COMMAND-8ugTs", "Errors.Org.DomainPolicy.NotFound")
	}

	otpWriteModel, err := c.otpWriteModelByID(ctx, userID, resourceowner)
	if err != nil {
		return nil, err
	}
	if otpWriteModel.State == domain.MFAStateReady {
		return nil, caos_errs.ThrowAlreadyExists(nil, "COMMAND-do9se", "Errors.User.MFA.OTP.AlreadyReady")
	}
	userAgg := UserAggregateFromWriteModel(&otpWriteModel.WriteModel)

	accountName := domain.GenerateLoginName(human.GetUsername(), org.PrimaryDomain, orgPolicy.UserLoginMustBeDomain)
	if accountName == "" {
		accountName = human.EmailAddress
	}
	key, secret, err := domain.NewOTPKey(c.multifactors.OTP.Issuer, accountName, c.multifactors.OTP.CryptoMFA)
	if err != nil {
		return nil, err
	}

	_, err = c.eventstore.Push(ctx, user.NewHumanOTPAddedEvent(ctx, userAgg, secret))
	if err != nil {
		return nil, err
	}
	return &domain.OTP{
		ObjectRoot: models.ObjectRoot{
			AggregateID: human.AggregateID,
		},
		SecretString: key.Secret(),
		Url:          key.URL(),
	}, nil
}

func (c *Commands) HumanCheckMFAOTPSetup(ctx context.Context, userID, code, userAgentID, resourceowner string) (*domain.ObjectDetails, error) {
	if userID == "" {
		return nil, caos_errs.ThrowPreconditionFailed(nil, "COMMAND-8N9ds", "Errors.User.UserIDMissing")
	}

	existingOTP, err := c.otpWriteModelByID(ctx, userID, resourceowner)
	if err != nil {
		return nil, err
	}
	if existingOTP.State == domain.MFAStateUnspecified || existingOTP.State == domain.MFAStateRemoved {
		return nil, caos_errs.ThrowNotFound(nil, "COMMAND-3Mif9s", "Errors.User.MFA.OTP.NotExisting")
	}
	if existingOTP.State == domain.MFAStateReady {
		return nil, caos_errs.ThrowPreconditionFailed(nil, "COMMAND-qx4ls", "Errors.Users.MFA.OTP.AlreadyReady")
	}
	if err := domain.VerifyMFAOTP(code, existingOTP.Secret, c.multifactors.OTP.CryptoMFA); err != nil {
		return nil, err
	}
	userAgg := UserAggregateFromWriteModel(&existingOTP.WriteModel)

	pushedEvents, err := c.eventstore.Push(ctx, user.NewHumanOTPVerifiedEvent(ctx, userAgg, userAgentID))
	if err != nil {
		return nil, err
	}
	err = AppendAndReduce(existingOTP, pushedEvents...)
	if err != nil {
		return nil, err
	}
	return writeModelToObjectDetails(&existingOTP.WriteModel), nil
}

func (c *Commands) HumanCheckMFAOTP(ctx context.Context, userID, code, resourceowner string, authRequest *domain.AuthRequest) error {
	if userID == "" {
		return caos_errs.ThrowPreconditionFailed(nil, "COMMAND-8N9ds", "Errors.User.UserIDMissing")
	}
	existingOTP, err := c.otpWriteModelByID(ctx, userID, resourceowner)
	if err != nil {
		return err
	}
	if existingOTP.State != domain.MFAStateReady {
		return caos_errs.ThrowPreconditionFailed(nil, "COMMAND-3Mif9s", "Errors.User.MFA.OTP.NotReady")
	}
	userAgg := UserAggregateFromWriteModel(&existingOTP.WriteModel)
	err = domain.VerifyMFAOTP(code, existingOTP.Secret, c.multifactors.OTP.CryptoMFA)
	if err == nil {
		_, err = c.eventstore.Push(ctx, user.NewHumanOTPCheckSucceededEvent(ctx, userAgg, authRequestDomainToAuthRequestInfo(authRequest)))
		return err
	}
	_, pushErr := c.eventstore.Push(ctx, user.NewHumanOTPCheckFailedEvent(ctx, userAgg, authRequestDomainToAuthRequestInfo(authRequest)))
	logging.Log("COMMAND-9fj7s").OnError(pushErr).Error("error create password check failed event")
	return err
}

func (c *Commands) HumanRemoveOTP(ctx context.Context, userID, resourceOwner string) (*domain.ObjectDetails, error) {
	if userID == "" {
		return nil, caos_errs.ThrowInvalidArgument(nil, "COMMAND-5M0sd", "Errors.User.UserIDMissing")
	}

	existingOTP, err := c.otpWriteModelByID(ctx, userID, resourceOwner)
	if err != nil {
		return nil, err
	}
	if existingOTP.State == domain.MFAStateUnspecified || existingOTP.State == domain.MFAStateRemoved {
		return nil, caos_errs.ThrowNotFound(nil, "COMMAND-Hd9sd", "Errors.User.MFA.OTP.NotExisting")
	}
	userAgg := UserAggregateFromWriteModel(&existingOTP.WriteModel)
	pushedEvents, err := c.eventstore.Push(ctx, user.NewHumanOTPRemovedEvent(ctx, userAgg))
	if err != nil {
		return nil, err
	}
	err = AppendAndReduce(existingOTP, pushedEvents...)
	if err != nil {
		return nil, err
	}
	return writeModelToObjectDetails(&existingOTP.WriteModel), nil
}

func (c *Commands) otpWriteModelByID(ctx context.Context, userID, resourceOwner string) (writeModel *HumanOTPWriteModel, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer func() { span.EndWithError(err) }()

	writeModel = NewHumanOTPWriteModel(userID, resourceOwner)
	err = c.eventstore.FilterToQueryReducer(ctx, writeModel)
	if err != nil {
		return nil, err
	}
	return writeModel, nil
}
