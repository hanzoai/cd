package revision_metadata

import (
	"fmt"
	"strings"

	"github.com/hanzoai/cd/util/errors"
	cdexec "github.com/hanzoai/cd/util/exec"
)

var Author string

func init() {
	userName, err := cdexec.RunCommand("git", cdexec.CmdOpts{}, "config", "--get", "user.name")
	errors.CheckError(err)
	userEmail, err := cdexec.RunCommand("git", cdexec.CmdOpts{}, "config", "--get", "user.email")
	errors.CheckError(err)
	Author = fmt.Sprintf("%s <%s>", strings.TrimSpace(userName), strings.TrimSpace(userEmail))
}
