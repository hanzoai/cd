package admin

import (
	"github.com/hanzoai/deploy/test/e2e/fixture"
)

// For admin CLI with kubernetes context
func RunCli(args ...string) (string, error) {
	return RunCliWithStdin("", args...)
}

func RunCliWithStdin(stdin string, args ...string) (string, error) {
	args = append([]string{"admin", "--namespace", fixture.TestNamespace()}, args...)
	return fixture.RunCliWithStdin(stdin, true, args...)
}
