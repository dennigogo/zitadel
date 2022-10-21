package setup

import (
	"context"

	"github.com/dennigogo/zitadel/internal/command"
	"github.com/dennigogo/zitadel/internal/config/systemdefaults"
	"github.com/dennigogo/zitadel/internal/eventstore"
)

type externalConfigChange struct {
	es             *eventstore.Eventstore
	ExternalDomain string `json:"externalDomain"`
	ExternalSecure bool   `json:"externalSecure"`
	ExternalPort   uint16 `json:"externalPort"`

	currentExternalDomain string
	currentExternalSecure bool
	currentExternalPort   uint16
}

func (mig *externalConfigChange) SetLastExecution(lastRun map[string]interface{}) {
	mig.currentExternalDomain, _ = lastRun["externalDomain"].(string)
	externalPort, _ := lastRun["externalPort"].(float64)
	mig.currentExternalPort = uint16(externalPort)
	mig.currentExternalSecure, _ = lastRun["externalSecure"].(bool)
}

func (mig *externalConfigChange) Check() bool {
	return mig.currentExternalSecure != mig.ExternalSecure ||
		mig.currentExternalPort != mig.ExternalPort ||
		mig.currentExternalDomain != mig.ExternalDomain
}

func (mig *externalConfigChange) Execute(ctx context.Context) error {
	cmd, err := command.StartCommands(mig.es,
		systemdefaults.SystemDefaults{},
		nil,
		nil,
		nil,
		mig.ExternalDomain,
		mig.ExternalSecure,
		mig.ExternalPort,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	if err != nil {
		return err
	}
	return cmd.ChangeSystemConfig(ctx, mig.currentExternalDomain, mig.currentExternalPort, mig.currentExternalSecure)
}

func (mig *externalConfigChange) String() string {
	return "config_change"
}
