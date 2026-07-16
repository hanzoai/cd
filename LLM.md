# hanzoai/deploy — Hanzo's ArgoCD v3 fork (the canonical CD library)

This is the hanzo-owned fork of Argo CD v3. Its `gitops-engine/` submodule
(`github.com/argoproj/argo-cd/gitops-engine/v3`) is the **canonical GitOps CD
library embedded inside the hanzo cloud money binary** — it drives the
`/v1/deploy` control plane (the ArgoCD-grade fleet view over the operator App CRs)
in `github.com/hanzoai/cloud` (`clients/deploy`). There is no separate `argocd`
namespace or process; the engine runs in-process.

## How cloud consumes it (module wiring — mechanism, decided)

Cloud imports the engine under its **upstream module path**
`github.com/argoproj/argo-cd/gitops-engine/v3/pkg/...` and points that path at this
fork with a **filesystem replace** onto the sibling checkout:

```
// cloud/go.mod
require github.com/argoproj/argo-cd/gitops-engine/v3 v3.0.0
replace github.com/argoproj/argo-cd/gitops-engine/v3 => ../deploy/gitops-engine
```

Why this and not a module rename:

- The `gitops-engine/` subdir is **its own Go module** (own `go.mod`), so it is
  independently importable — that is the seam cloud uses.
- We deliberately do **NOT** rename the module to `github.com/hanzoai/deploy/...`.
  The parent Argo CD module (`github.com/argoproj/argo-cd/v3`) imports the engine
  under the argoproj path in hundreds of files and wires it with its own
  `replace ... => ./gitops-engine`; renaming the submodule would break the parent's
  `go build ./...` and force a repo-wide import rewrite. Keeping the argoproj path +
  a consumer-side replace is the least-coupling wiring that keeps BOTH this fork and
  cloud building. This fork is hanzo-owned by **repository**, patched here as needed.
- **End state:** tag the `gitops-engine/` submodule (`gitops-engine/vX.Y.Z`) on this
  repo and swap cloud's filesystem replace for a pinned pseudo-version — no code
  change, same import path.

## Dependency note (why cloud pins the engine's k8s stack down)

The engine's read path (`pkg/health`, `pkg/diff`, `pkg/sync/*`) all funnel through
`pkg/utils/kube`, whose `pkg/utils/kube/scheme` imports the **`k8s.io/kubernetes`
monorepo** (the legacyscheme install tree — genuinely needed for `pkg/diff`'s
internal-version conversion + defaulting). Consuming any engine package therefore
drags `k8s.io/kubernetes` + the k8s staging-module replace block.

This fork targets k8s **v0.36.1**. Cloud runs its k8s client stack at **v0.35.3**
(its `sigs.k8s.io/controller-runtime v0.23.3` requires it). Cloud pins the engine's
whole k8s stack **down to v0.35.3** (staging replaces + `k8s.io/kubernetes => v1.35.3`)
so the money binary's k8s versions do not move. The engine compiles cleanly against
v0.35.3 for the health path.

## Do not

- Do not rename the `gitops-engine` module path (breaks the parent Argo CD module).
- Do not push these hanzo notes or the integration branch to `argoproj/argo-cd`.
  This lives on hanzo branches only. Upstream `AGENTS.md` (argo's contribution
  rules) is retained verbatim and applies only to upstream PRs, which we do not file.
