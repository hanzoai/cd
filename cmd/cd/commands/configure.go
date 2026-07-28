package commands

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	cdclient "github.com/hanzoai/deploy/pkg/apiclient"
	"github.com/hanzoai/deploy/util/errors"
	"github.com/hanzoai/deploy/util/localconfig"
)

// NewConfigureCommand returns a new instance of an `cd configure` command
func NewConfigureCommand(clientOpts *cdclient.ClientOptions) *cobra.Command {
	var promptsEnabled bool

	command := &cobra.Command{
		Use:   "configure",
		Short: "Manage local configuration",
		Example: `# Enable optional interactive prompts
cd configure --prompts-enabled
cd configure --prompts-enabled=true

# Disable optional interactive prompts
cd configure --prompts-enabled=false`,
		Run: func(_ *cobra.Command, _ []string) {
			localCfg, err := localconfig.ReadLocalConfig(clientOpts.ConfigPath)
			errors.CheckError(err)
			if localCfg == nil {
				fmt.Println("No local configuration found")
				os.Exit(1)
			}

			localCfg.PromptsEnabled = promptsEnabled

			err = localconfig.WriteLocalConfig(*localCfg, clientOpts.ConfigPath)
			errors.CheckError(err)

			fmt.Println("Successfully updated the following configuration settings:")
			fmt.Printf("prompts-enabled: %v\n", strconv.FormatBool(localCfg.PromptsEnabled))
		},
	}

	command.Flags().BoolVar(&promptsEnabled, "prompts-enabled", localconfig.GetPromptsEnabled(false), "Enable (or disable) optional interactive prompts")

	return command
}
