# cd

A Helm chart for Hanzo CD, derived from the kustomize bases in
[`manifests/`](../../manifests). Both installers describe the same objects
under the same names (`hanzocd-*`); the object names are lookup keys the
running components resolve at runtime, so a fork of this chart must not
rename them.

## Install

```console
helm install hanzocd charts/cd --namespace cd --create-namespace
```

Only one release of this chart can run per namespace: every object name is
fixed (not derived from the Helm release name), matching `manifests/install.yaml`
today. Installing a second release into the same namespace will collide.

## Values

| Key | Default | Description |
|-----|---------|--------------|
| `image.repository` / `image.tag` | `ghcr.io/hanzoai/cd` / chart appVersion | Image for every hanzocd-owned container. |
| `global.clusterScope` | `true` | Create the ClusterRole/ClusterRoleBinding pair for application-controller, applicationset-controller and server. Set `false` for a namespace-scoped install matching `manifests/namespace-install.yaml`. |
| `ha.enabled` | `false` | Swap the single redis Deployment for the redis-ha (sentinel + haproxy) dependency, add the zone-spread anti-affinity to server/repo-server, and default their replica counts to 2. Matches `manifests/ha/install.yaml`. |
| `hydrator.enabled` | `false` | Add the commit-server component. Matches `manifests/install-with-hydrator.yaml`. |
| `dex.enabled` / `notifications.enabled` | `true` | Toggle the bundled dex-server / notifications-controller. |
| `controller.replicaCount` | `1` | application-controller StatefulSet replicas (shards). |
| `server.replicaCount` / `repoServer.replicaCount` | unset (2 under `ha.enabled`, else 1) | Explicit override wins over the ha.enabled default. |
| `*.resources` | `{}` | Container resources for each component; unset matches upstream (no request/limit). |
| `server.ingress.*` | disabled | Ingress for hanzocd-server. |
| `config` / `rbac` / `params` | `{}` | Extra keys merged into `hanzocd-cm`, `hanzocd-rbac-cm` and `hanzocd-cmd-params-cm` on top of the built-in defaults -- e.g. `params: {server.insecure: "true"}` for a TLS-terminating ingress, or `config: {dex.config: "..."}` for SSO. |
| `networkPolicy.create` | `true` | Render the NetworkPolicy for each component. |

CRDs live in `crds/` (Helm's install-once directory: never upgraded or
deleted by Helm). Skip them with the native `helm install --skip-crds` when
CRDs are managed by another process, matching `manifests/namespace-install.yaml`.

## Source of truth

`manifests/base/**`, `manifests/cluster-rbac/**` and `manifests/crds/**` are
the source kustomize builds from; this chart's templates are hand-derived
from the same files and covered by `hack/verify-helm-chart.sh`, which fails
CI the moment `helm template` and `kustomize build` stop agreeing on the
rendered object names. Re-run it after changing either side.

## HA redis and the `hanzocd-redis` secret

`manifests/ha/base/redis-ha` vendors the upstream
[`dandydeveloper/redis-ha`](https://github.com/DandyDeveloper/charts) chart
at the version pinned in this chart's `Chart.yaml` dependency. Its own pods
run the stock redis/sentinel/haproxy images, not `cd`, so nothing in that
topology can run our `cd admin redis-initial-password` bootstrap. This chart
runs it once as a `pre-install,pre-upgrade` hook Job (`templates/redis.yaml`)
so `hanzocd-redis` exists before any component -- including redis-ha itself,
via `redis-ha.existingSecret` -- needs to read it.
