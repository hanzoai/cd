package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	appcontroller "github.com/hanzoai/cd/cmd/cd-application-controller/commands"
	applicationset "github.com/hanzoai/cd/cmd/cd-applicationset-controller/commands"
	cmpserver "github.com/hanzoai/cd/cmd/cd-cmp-server/commands"
	commitserver "github.com/hanzoai/cd/cmd/cd-commit-server/commands"
	dex "github.com/hanzoai/cd/cmd/cd-dex/commands"
	gitaskpass "github.com/hanzoai/cd/cmd/cd-git-ask-pass/commands"
	k8sauth "github.com/hanzoai/cd/cmd/cd-k8s-auth/commands"
	reposerver "github.com/hanzoai/cd/cmd/cd-repo-server/commands"
	apiserver "github.com/hanzoai/cd/cmd/cd-server/commands"
	cli "github.com/hanzoai/cd/cmd/cd/commands"
	"github.com/hanzoai/cd/common"
	"github.com/hanzoai/cd/util/log"
)

const (
	binaryNameEnv = "CD_BINARY_NAME"
)

func init() {
	// Make sure klog uses the configured log level and format.
	klog.SetLogger(log.NewLogrusLogger(log.NewWithCurrentConfig()))
}

func main() {
	var command *cobra.Command

	binaryName := filepath.Base(os.Args[0])
	if val := os.Getenv(binaryNameEnv); val != "" {
		binaryName = val
	}

	isArgocdCLI := false

	switch binaryName {
	case common.CommandCLI:
		command = cli.NewCommand()
		isArgocdCLI = true
	case common.CommandServer:
		command = apiserver.NewCommand()
	case common.CommandApplicationController:
		command = appcontroller.NewCommand()
	case common.CommandRepoServer:
		command = reposerver.NewCommand()
	case common.CommandCMPServer:
		command = cmpserver.NewCommand()
		isArgocdCLI = true
	case common.CommandCommitServer:
		command = commitserver.NewCommand()
	case common.CommandDex:
		command = dex.NewCommand()
	case common.CommandGitAskPass:
		command = gitaskpass.NewCommand()
		isArgocdCLI = true
	case common.CommandApplicationSetController:
		command = applicationset.NewCommand()
	case common.CommandK8sAuth:
		command = k8sauth.NewCommand()
		isArgocdCLI = true
	default:
		// "cd-linux-amd64", "cd-darwin-amd64", "cd-windows-amd64.exe" are also valid binary names
		command = cli.NewCommand()
		isArgocdCLI = true
	}

	if isArgocdCLI {
		// silence errors and usages since we'll be printing them manually.
		// This is because if we execute a plugin, the initial
		// errors and usage are always going to get printed that we don't want.
		command.SilenceErrors = true
		command.SilenceUsage = true
	}

	err := command.Execute()
	// if an error is present, try to look for various scenarios
	// such as if the error is from the execution of a normal cd command,
	// unknown command error or any other.
	if err != nil {
		errMsg, pluginErr := cli.NewDefaultPluginHandler().HandleCommandExecutionError(err, isArgocdCLI, os.Args)
		if pluginErr != nil {
			os.Stdout.WriteString(errMsg)
			var exitErr *exec.ExitError
			if errors.As(pluginErr, &exitErr) {
				// Return the actual plugin exit code
				os.Exit(exitErr.ExitCode())
			}
			// Fallback to exit code 1 if the error isn't an exec.ExitError
			os.Exit(1)
		}
	}
}
