package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra/doc"

	controller "github.com/hanzoai/deploy/cmd/cd-application-controller/commands"
	cdappsetcontroller "github.com/hanzoai/deploy/cmd/cd-applicationset-controller/commands"
	cddex "github.com/hanzoai/deploy/cmd/cd-dex/commands"
	reposerver "github.com/hanzoai/deploy/cmd/cd-repo-server/commands"
	cdserver "github.com/hanzoai/deploy/cmd/cd-server/commands"
	cdcli "github.com/hanzoai/deploy/cmd/cd/commands"
)

func main() {
	// set HOME env var so that default values involve user's home directory do not depend on the running user.
	os.Setenv("HOME", "/home/user")
	os.Setenv("XDG_CONFIG_HOME", "/home/user/.config")

	identity := func(s string) string { return s }
	headerPrepender := func(filename string) string {
		// The default header looks like `Argocd app get`. The leading capital letter is off-putting.
		// This header overrides the default. It's better visually and for search results.
		filename = filepath.Base(filename)
		filename = filename[:len(filename)-3] // Drop the '.md'
		return fmt.Sprintf("# `%s` Command Reference\n\n", strings.ReplaceAll(filename, "_", " "))
	}

	err := doc.GenMarkdownTreeCustom(cdcli.NewCommand(), "./docs/user-guide/commands", headerPrepender, identity)
	if err != nil {
		log.Fatal(err)
	}

	err = doc.GenMarkdownTreeCustom(cdserver.NewCommand(), "./docs/operator-manual/server-commands", headerPrepender, identity)
	if err != nil {
		log.Fatal(err)
	}

	err = doc.GenMarkdownTreeCustom(controller.NewCommand(), "./docs/operator-manual/server-commands", headerPrepender, identity)
	if err != nil {
		log.Fatal(err)
	}

	err = doc.GenMarkdownTreeCustom(reposerver.NewCommand(), "./docs/operator-manual/server-commands", headerPrepender, identity)
	if err != nil {
		log.Fatal(err)
	}

	err = doc.GenMarkdownTreeCustom(cddex.NewCommand(), "./docs/operator-manual/server-commands", headerPrepender, identity)
	if err != nil {
		log.Fatal(err)
	}

	err = doc.GenMarkdownTreeCustom(cdappsetcontroller.NewCommand(), "./docs/operator-manual/server-commands", headerPrepender, identity)
	if err != nil {
		log.Fatal(err)
	}
}
