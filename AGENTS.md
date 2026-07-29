# AI Agent Directives for Hanzo CD

This repository is **hanzoai/cd** — a hard fork of argo-cd, not a copy of it and
not a contributor to it. Nothing here goes upstream; do not open issues or PRs
against argoproj/argo-cd, and do not follow their PR template.

The Go module is still `github.com/hanzoai/deploy` (the repo's former name).
Renaming it is a coordinated change: `hanzoai/cloud` imports
`github.com/hanzoai/deploy/gitops-engine`, so the module path cannot move alone.

## 1. argoproj.io means two different things. Learn the difference before editing.

This is the single rule most likely to cause an outage, and it has already been
broken twice.

**(a) Our identity → `apps.hanzo.ai`.** The API group, and every annotation,
label, finalizer and secret-type the product defines. Upstream spells all of
these `argocd.argoproj.io`; here they are `apps.hanzo.ai`. The cluster serves
`applications`/`applicationsets`/`appprojects` **.apps.hanzo.ai** and has never
served `argoproj.io`.

**(b) Someone else's CRDs → leave `argoproj.io` alone.** Argo Rollouts, Argo
Workflows, Argo Promoter, and the `notifications-engine` we consume as an
external module. We only *observe* these resources. Their health checks and
`ignoreResourceUpdates` keys live under `resource_customizations/` (113 vendor
directories, most of which have nothing to do with Argo). Renaming them breaks
health assessment for any user running those products.

`manifests/base/config/hanzocd-cm.yaml` shows both on adjacent lines:
`ignoreResourceUpdates.apps.hanzo.ai_Application` is ours,
`ignoreResourceUpdates.argoproj.io_Rollout` is not.

**Never blanket-sed `argoproj.io`.** Match the full token. `argocd.argoproj.io`
is always ours; bare `argoproj.io` needs the Kind checked.

One deliberate exception: `common.LabelKeyLegacyApplicationName` keeps
`applications.argoproj.io/app-name`. It names upstream's v0.10 label. Renaming
it would invent a legacy Hanzo install that never existed.

## 2. Do not restate a group or APIVersion

Derive it: `application.Group`, `application.*Plural`,
`v1alpha1.SchemeGroupVersion`. Hardcoded copies are how (a) drifted — the UI
wrote one finalizer while the controller read another, and cascade delete
silently orphaned resources.

The same applies to resource names in test fixtures: use
`common.ArgoCDConfigMapName`, not `"argocd-cm"`. Production reads `hanzocd-cm`
through an informer selecting `app.kubernetes.io/part-of=hanzocd`; a fixture
that hardcodes the old name *or* the old label is invisible to the code under
test, and the test fails for a reason that has nothing to do with its subject.

## 3. This deploys the fleet

`hanzo-cd` on hanzo-k8s runs the application-controller, repo-server and server,
and reconciles every Application in the cluster. Changes land on a branch and
wait for review. Do not scale those workloads, apply manifests to the cluster,
or edit live Applications.

## 4. Local checks

* `go build ./...` — and `cd gitops-engine && go build ./...`, a separate module
  wired in by a `replace` directive.
* `go test ./...` — a number of packages fail for environmental reasons, not
  yours. `reposerver/repository` and `util/kustomize` need the `kustomize`
  binary on PATH; `util/sourceintegrity` needs `gpg-wrapper.sh`; `util/git`
  reaches the network. Establish a baseline on an untouched worktree and compare
  failure sets — never assume a red package is your fault, and never assume it
  is not.
* `make codegen` if API structs change. `hack/gen-crd-spec` derives CRD
  filenames from `application.Group`.
* `manifests/*.yaml` are generated from `manifests/base` by
  `hack/update-manifests.sh` and need `kustomize`. They have gone stale before —
  base granted `apps.hanzo.ai` while the rendered HA set still granted
  `argoproj.io`.

CI is native (`git.hanzo.ai`, `cd.hanzo.ai`). There is no GitHub Actions
pipeline here; it was removed deliberately.

## 5. PR titles

Semantic prefixes: `ci: fix: feat: test: docs: chore: refactor: revert:`.
