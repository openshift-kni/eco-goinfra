package security

import (
	"fmt"

	"github.com/openshift-kni/eco-goinfra/pkg/schemes/argocd/argocdv2/util/glob"
)

func IsNamespaceEnabled(namespace string, serverNamespace string, enabledNamespaces []string) bool {
	return namespace == serverNamespace || glob.MatchStringInList(enabledNamespaces, namespace, false)
}

func NamespaceNotPermittedError(namespace string) error {
	return fmt.Errorf("namespace '%s' is not permitted", namespace)
}
