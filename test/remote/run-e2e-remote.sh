#!/bin/sh

# Should point to the Hanzo CD API endpoint on the cluster
if test "${CD_SERVER}" = ""; then
	echo "Please set CD_SERVER to the remote Hanzo CD API endpoint to test." >&2
	exit 1
fi

# CD_E2E_REMOTE must be set to 'true' in order for remote tests to work
export CD_E2E_REMOTE=true

# The timeout for running the test suite (duration)
export CD_E2E_TEST_TIMEOUT=2h

# The default timeout for certain operations (such as sync)
export CD_E2E_DEFAULT_TIMEOUT=30

# Set CD_E2E_NAMESPACE to the namespace the Hanzo CD we're testing against is
# running in. Defaults to "cd-e2e"
export CD_E2E_NAMESPACE="${CD_E2E_NAMESPACE:-cd-e2e}"

# Name prefix the operator sets on resources created for Hanzo CD instance. This
# is usually also the name of the instance itself.
export CD_E2E_NAME_PREFIX="${CD_E2E_NAME_PREFIX:-}"

# This is to skip some (deprecated) tests
export CD_E2E_K3S=true

# Configuration for skipping certain classes of tests

# GnuPG features not yet available with GitOps Operator
export CD_E2E_SKIP_GPG="${CD_E2E_SKIP_GPG:-false}"
# Some tests do not work OOTB with OpenShift
export CD_E2E_SKIP_OPENSHIFT="${CD_E2E_SKIP_OPENSHIFT:-false}"
# Skip Helm tests
export CD_E2E_SKIP_HELM="${CD_E2E_SKIP_HELM:-false}"
# Skip Ksonnet tests
export CD_E2E_SKIP_KSONNET="${CD_E2E_SKIP_KSONNET:-false}"

## ====================================================
# no changes below this line required
## ====================================================

# Unauthenticated URLs for pushing from CI
#
# Use `kubectl port-forward service/cd-e2e-server 9081:9081` to set up the
# listener required for this.
export CD_E2E_GIT_SERVICE="http://127.0.0.1:9081/cd-e2e/testdata.git"
export CD_E2E_HELM_SERVICE="http://127.0.0.1:9081/helm-repo"
export CD_E2E_GIT_SERVICE_SUBMODULE="http://127.0.0.1:9081/cd-e2e/submodule.git"
export CD_E2E_GIT_SERVICE_SUBMODULE_PARENT="http://127.0.0.1:9081/cd-e2e/submoduleParent.git"

# URLs used during testing - usually no need to change those
export CD_E2E_REPO_SSH="ssh://root@cd-e2e-server:2222/tmp/cd-e2e/testdata.git"
export CD_E2E_REPO_SSH_SUBMODULE="ssh://root@cd-e2e-server:2222/tmp/cd-e2e/submodule.git"
export CD_E2E_REPO_SSH_SUBMODULE_PARENT="ssh://root@cd-e2e-server:2222/tmp/cd-e2e/submoduleParent.git"
export CD_E2E_REPO_HTTPS="https://cd-e2e-server:9443/cd-e2e/testdata.git"
export CD_E2E_REPO_HTTPS_CLIENT_CERT="https://cd-e2e-server:9444/cd-e2e/testdata.git"
export CD_E2E_REPO_HTTPS_SUBMODULE="https://cd-e2e-server:9443/cd-e2e/submodule.git"
export CD_E2E_REPO_HTTPS_SUBMODULE_PARENT="https://cd-e2e-server:9443/cd-e2e/submoduleParent.git"
export CD_E2E_REPO_HELM="https://cd-e2e-server:9444/helm-repo"
export CD_E2E_REPO_DEFAULT="http://cd-e2e-server:9081/cd-e2e/testdata.git"

"$@"
