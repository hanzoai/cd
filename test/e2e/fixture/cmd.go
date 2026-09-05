package fixture

import (
	"context"
	"os"
	"os/exec"
	"strings"

	cdexec "github.com/hanzoai/cd/util/exec"
)

func Run(workDir, name string, args ...string) (string, error) {
	return RunWithStdin("", workDir, name, args...)
}

func RunWithStdin(stdin, workDir, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(context.Background(), name, args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	cmd.Env = os.Environ()
	cmd.Dir = workDir

	return cdexec.RunCommandExt(cmd, cdexec.CmdOpts{})
}

func RunWithStdinWithRedactor(stdin, workDir, name string, redactor func(string) string, args ...string) (string, error) {
	cmd := exec.CommandContext(context.Background(), name, args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	cmd.Env = os.Environ()
	cmd.Dir = workDir

	return cdexec.RunCommandExt(cmd, cdexec.CmdOpts{Redactor: redactor})
}
