package admin

import (
	_ "embed"
	"errors"

	"github.com/spf13/cobra"

	"github.com/dennigogo/zitadel/cmd/initialise"
	"github.com/dennigogo/zitadel/cmd/key"
	"github.com/dennigogo/zitadel/cmd/setup"
	"github.com/dennigogo/zitadel/cmd/start"
)

func New() *cobra.Command {
	adminCMD := &cobra.Command{
		Use:        "admin",
		Short:      "The ZITADEL admin CLI lets you interact with your instance",
		Long:       `The ZITADEL admin CLI lets you interact with your instance`,
		Deprecated: "please use subcommands directly, e.g. `zitadel start`",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("no additional command provided")
		},
	}

	adminCMD.AddCommand(
		initialise.New(),
		setup.New(),
		start.New(),
		start.NewStartFromInit(),
		key.New(),
	)

	return adminCMD
}
