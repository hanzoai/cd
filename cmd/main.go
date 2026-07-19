package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	appcontroller "github.com/hanzoai/deploy/cmd/argocd-application-controller/commands"
	applicationset "github.com/hanzoai/deploy/cmd/argocd-applicationset-controller/commands"
	cmpserver "github.com/hanzoai/deploy/cmd/argocd-cmp-server/commands"
	commitserver "github.com/hanzoai/deploy/cmd/argocd-commit-server/commands"
	dex "github.com/hanzoai/deploy/cmd/argocd-dex/commands"
	gitaskpass "github.com/hanzoai/deploy/cmd/argocd-git-ask-pass/commands"
	k8sauth "github.com/hanzoai/deploy/cmd/argocd-k8s-auth/commands"
	notification "github.com/hanzoai/deploy/cmd/argocd-notification/commands"
	reposerver "github.com/hanzoai/deploy/cmd/argocd-repo-server/commands"
	apiserver "github.com/hanzoai/deploy/cmd/argocd-server/commands"
	cli "github.com/hanzoai/deploy/cmd/argocd/commands"
	"github.com/hanzoai/deploy/common"
	"github.com/hanzoai/deploy/util/log"
)

const (
	binaryNameEnv = "ARGOCD_BINARY_NAME"
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
	case common.CommandNotifications:
		command = notification.NewCommand()
	case common.CommandGitAskPass:
		command = gitaskpass.NewCommand()
		isArgocdCLI = true
	case common.CommandApplicationSetController:
		command = applicationset.NewCommand()
	case common.CommandK8sAuth:
		command = k8sauth.NewCommand()
		isArgocdCLI = true
	default:
		// "argocd-linux-amd64", "argocd-darwin-amd64", "argocd-windows-amd64.exe" are also valid binary names
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
	// such as if the error is from the execution of a normal argocd command,
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
