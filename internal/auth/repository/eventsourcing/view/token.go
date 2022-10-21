package view

import (
	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore/v1/models"
	usr_view "github.com/dennigogo/zitadel/internal/user/repository/view"
	"github.com/dennigogo/zitadel/internal/user/repository/view/model"
	"github.com/dennigogo/zitadel/internal/view/repository"
)

const (
	tokenTable = "auth.tokens"
)

func (v *View) TokenByIDs(tokenID, userID, instanceID string) (*model.TokenView, error) {
	return usr_view.TokenByIDs(v.Db, tokenTable, tokenID, userID, instanceID)
}

func (v *View) TokensByUserID(userID, instanceID string) ([]*model.TokenView, error) {
	return usr_view.TokensByUserID(v.Db, tokenTable, userID, instanceID)
}

func (v *View) PutToken(token *model.TokenView, event *models.Event) error {
	err := usr_view.PutToken(v.Db, tokenTable, token)
	if err != nil {
		return err
	}
	return v.ProcessedTokenSequence(event)
}

func (v *View) PutTokens(token []*model.TokenView, event *models.Event) error {
	err := usr_view.PutTokens(v.Db, tokenTable, token...)
	if err != nil {
		return err
	}
	return v.ProcessedTokenSequence(event)
}

func (v *View) DeleteToken(tokenID, instanceID string, event *models.Event) error {
	err := usr_view.DeleteToken(v.Db, tokenTable, tokenID, instanceID)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return v.ProcessedTokenSequence(event)
}

func (v *View) DeleteSessionTokens(agentID, userID, instanceID string, event *models.Event) error {
	err := usr_view.DeleteSessionTokens(v.Db, tokenTable, agentID, userID, instanceID)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return v.ProcessedTokenSequence(event)
}

func (v *View) DeleteUserTokens(userID, instanceID string, event *models.Event) error {
	err := usr_view.DeleteUserTokens(v.Db, tokenTable, userID, instanceID)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return v.ProcessedTokenSequence(event)
}

func (v *View) DeleteApplicationTokens(event *models.Event, ids ...string) error {
	err := usr_view.DeleteApplicationTokens(v.Db, tokenTable, event.InstanceID, ids)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return v.ProcessedTokenSequence(event)
}

func (v *View) DeleteTokensFromRefreshToken(refreshTokenID, instanceID string, event *models.Event) error {
	err := usr_view.DeleteTokensFromRefreshToken(v.Db, tokenTable, refreshTokenID, instanceID)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return v.ProcessedTokenSequence(event)
}

func (v *View) GetLatestTokenSequence(instanceID string) (*repository.CurrentSequence, error) {
	return v.latestSequence(tokenTable, instanceID)
}

func (v *View) GetLatestTokenSequences(instanceIDs ...string) ([]*repository.CurrentSequence, error) {
	return v.latestSequences(tokenTable, instanceIDs...)
}

func (v *View) ProcessedTokenSequence(event *models.Event) error {
	return v.saveCurrentSequence(tokenTable, event)
}

func (v *View) UpdateTokenSpoolerRunTimestamp() error {
	return v.updateSpoolerRunSequence(tokenTable)
}

func (v *View) GetLatestTokenFailedEvent(sequence uint64, instanceID string) (*repository.FailedEvent, error) {
	return v.latestFailedEvent(tokenTable, instanceID, sequence)
}

func (v *View) ProcessedTokenFailedEvent(failedEvent *repository.FailedEvent) error {
	return v.saveFailedEvent(failedEvent)
}
