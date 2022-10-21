package view

import (
	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore/v1/models"
	"github.com/dennigogo/zitadel/internal/project/repository/view"
	"github.com/dennigogo/zitadel/internal/project/repository/view/model"
	"github.com/dennigogo/zitadel/internal/view/repository"
)

const (
	orgPrgojectMappingTable = "auth.org_project_mapping"
)

func (v *View) OrgProjectMappingByIDs(orgID, projectID, instanceID string) (*model.OrgProjectMapping, error) {
	return view.OrgProjectMappingByIDs(v.Db, orgPrgojectMappingTable, orgID, projectID, instanceID)
}

func (v *View) PutOrgProjectMapping(mapping *model.OrgProjectMapping, event *models.Event) error {
	err := view.PutOrgProjectMapping(v.Db, orgPrgojectMappingTable, mapping)
	if err != nil {
		return err
	}
	return v.ProcessedOrgProjectMappingSequence(event)
}

func (v *View) DeleteOrgProjectMapping(orgID, projectID, instanceID string, event *models.Event) error {
	err := view.DeleteOrgProjectMapping(v.Db, orgPrgojectMappingTable, orgID, projectID, instanceID)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return v.ProcessedOrgProjectMappingSequence(event)
}

func (v *View) DeleteOrgProjectMappingsByProjectID(projectID, instanceID string) error {
	return view.DeleteOrgProjectMappingsByProjectID(v.Db, orgPrgojectMappingTable, projectID, instanceID)
}

func (v *View) DeleteOrgProjectMappingsByProjectGrantID(projectGrantID, instanceID string) error {
	return view.DeleteOrgProjectMappingsByProjectGrantID(v.Db, orgPrgojectMappingTable, projectGrantID, instanceID)
}

func (v *View) GetLatestOrgProjectMappingSequence(instanceID string) (*repository.CurrentSequence, error) {
	return v.latestSequence(orgPrgojectMappingTable, instanceID)
}

func (v *View) GetLatestOrgProjectMappingSequences(instanceIDs ...string) ([]*repository.CurrentSequence, error) {
	return v.latestSequences(orgPrgojectMappingTable, instanceIDs...)
}

func (v *View) ProcessedOrgProjectMappingSequence(event *models.Event) error {
	return v.saveCurrentSequence(orgPrgojectMappingTable, event)
}

func (v *View) UpdateOrgProjectMappingSpoolerRunTimestamp() error {
	return v.updateSpoolerRunSequence(orgPrgojectMappingTable)
}

func (v *View) GetLatestOrgProjectMappingFailedEvent(sequence uint64, instanceID string) (*repository.FailedEvent, error) {
	return v.latestFailedEvent(orgPrgojectMappingTable, instanceID, sequence)
}

func (v *View) ProcessedOrgProjectMappingFailedEvent(failedEvent *repository.FailedEvent) error {
	return v.saveFailedEvent(failedEvent)
}
