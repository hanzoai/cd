#!/usr/bin/env bash
# Fails the moment charts/cd stops agreeing with the kustomize manifests it
# was derived from. Run after changing either manifests/base/** or
# charts/cd/templates/**.
set -o errexit
set -o nounset
set -o pipefail

SRCROOT="$(CDPATH='' cd -- "$(dirname "$0")/.." && pwd -P)"
CHART="${SRCROOT}/charts/cd"
WORK="$(mktemp -d)"
trap 'rm -rf "${WORK}"' EXIT

for bin in helm kustomize kubectl python3; do
	if ! command -v "$bin" >/dev/null 2>&1; then
		echo "verify-helm-chart: '$bin' is not installed, cannot verify" >&2
		exit 1
	fi
done

kind_name() {
	python3 - "$1" <<'PY'
import sys, yaml
with open(sys.argv[1]) as f:
    docs = yaml.safe_load_all(f)
    out = []
    for d in docs:
        if not d:
            continue
        kind = d.get("kind", "")
        name = (d.get("metadata") or {}).get("name", "")
        out.append(f"{kind}\t{name}")
for line in sorted(out):
    print(line)
PY
}

echo "==> helm dependency update"
helm dependency update "${CHART}" >/dev/null

echo "==> helm lint (default values)"
helm lint "${CHART}"

echo "==> helm lint (ha, hydrator, ingress, namespace-scoped)"
helm lint "${CHART}" \
	--set ha.enabled=true \
	--set hydrator.enabled=true \
	--set server.ingress.enabled=true \
	--set global.clusterScope=false

echo "==> helm template (default) vs manifests/install.yaml"
helm template hanzocd "${CHART}" --namespace cd --include-crds >"${WORK}/helm-default.yaml"
kind_name "${WORK}/helm-default.yaml" >"${WORK}/helm-default.kn"
kind_name "${SRCROOT}/manifests/install.yaml" >"${WORK}/kustomize-default.kn"
if ! diff -u "${WORK}/kustomize-default.kn" "${WORK}/helm-default.kn"; then
	echo "verify-helm-chart: default render does not match manifests/install.yaml" >&2
	exit 1
fi
echo "    identical ($(wc -l <"${WORK}/kustomize-default.kn" | tr -d ' ') objects)"

echo "==> kubectl apply --dry-run=client (default)"
kubectl apply --dry-run=client -f "${WORK}/helm-default.yaml" >/dev/null

echo "==> helm template (ha.enabled=true) vs manifests/ha/install.yaml, redis subsystem excluded"
# ha.enabled swaps the single redis Deployment for the redis-ha dependency
# (github.com/DandyDeveloper/charts), fullnameOverride'd to hanzocd-redis-ha;
# object names there differ from the kustomize-vendored cd-redis-ha-* by
# design (see charts/cd/README.md), and its Helm test-hook Pods only run on
# `helm test`, never on install/upgrade. Everything else must match exactly.
helm template hanzocd "${CHART}" --namespace cd --include-crds --set ha.enabled=true >"${WORK}/helm-ha.yaml"
kind_name "${WORK}/helm-ha.yaml" | grep -vi redis | grep -v '^Pod\b' >"${WORK}/helm-ha.kn"
kind_name "${SRCROOT}/manifests/ha/install.yaml" | grep -vi redis >"${WORK}/kustomize-ha.kn"
if ! diff -u "${WORK}/kustomize-ha.kn" "${WORK}/helm-ha.kn"; then
	echo "verify-helm-chart: ha render does not match manifests/ha/install.yaml outside the redis subsystem" >&2
	exit 1
fi
echo "    identical ($(wc -l <"${WORK}/kustomize-ha.kn" | tr -d ' ') non-redis objects)"

echo "==> kubectl apply --dry-run=client (ha)"
kubectl apply --dry-run=client -f "${WORK}/helm-ha.yaml" >/dev/null

echo "verify-helm-chart: OK"
