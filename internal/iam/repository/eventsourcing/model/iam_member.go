package model

import (
	"encoding/json"

	es_models "github.com/dennigogo/zitadel/internal/eventstore/v1/models"
	"github.com/zitadel/logging"
)

type IAMMember struct {
	es_models.ObjectRoot
	UserID string   `json:"userId,omitempty"`
	Roles  []string `json:"roles,omitempty"`
}

func (m *IAMMember) SetData(event *es_models.Event) error {
	m.ObjectRoot.AppendEvent(event)
	if err := json.Unmarshal(event.Data, m); err != nil {
		logging.Log("EVEN-e4dkp").WithError(err).Error("could not unmarshal event data")
		return err
	}
	return nil
}
