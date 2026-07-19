package security

import (
	"fmt"

	"github.com/hanzoai/deploy/util/glob"
)

func IsNamespaceEnabled(namespace string, serverNamespace string, enabledNamespaces []string) bool {
	return namespace == serverNamespace || glob.MatchStringInList(enabledNamespaces, namespace, glob.REGEXP)
}

func NamespaceNotPermittedError(namespace string) error {
	return fmt.Errorf("namespace '%s' is not permitted", namespace)
}
