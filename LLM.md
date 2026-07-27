# Hanzo CD

A fork of Argo CD (Apache-2.0). See `NOTICE` for attribution and the list of
changes. `LICENSE` is upstream's Apache-2.0 text, verbatim.

## One lineage

`main` is the only branch that carries the fork. It tracks upstream's
development line, not a release branch.

| | |
|---|---|
| VERSION | 3.6.0 |
| module | `github.com/hanzoai/deploy` |
| engine module | `github.com/hanzoai/deploy/gitops-engine` |
| API group | `cd.hanzo.ai` |
| objects | `hanzocd-*` (`manifests/`) |
| binary | `hanzocd` (`BIN_NAME`) |
| kubernetes | v0.36.1 (`k8s.io/kubernetes v1.36.1`) |

The rebrand is applied at the source, not carried as a patch on a side branch:
module paths, the CRD group, the manifest object names and the served UI are
all renamed in tree. There is nothing to reapply when upstream is merged.

## The engine submodule

`gitops-engine/` is its own Go module, which is the seam a consumer imports.
The parent module wires it with a filesystem replace:

    // go.mod
    require github.com/hanzoai/deploy/gitops-engine v0.0.0-...
    replace github.com/hanzoai/deploy/gitops-engine => ./gitops-engine

Cutting a release tags the submodule `gitops-engine/vX.Y.Z`, which is what an
external consumer pins to instead of a filesystem replace — no code change,
same import path.

Consuming any engine package drags in the `k8s.io/kubernetes` monorepo:
`pkg/health`, `pkg/diff` and `pkg/sync/*` all funnel through `pkg/utils/kube`,
whose `scheme` package needs the legacyscheme install tree for `pkg/diff`'s
internal-version conversion and defaulting. A consumer on an older k8s stack
has to pin the staging modules plus `k8s.io/kubernetes` down to its own
version.

## Images

| image | built from | workflow |
|---|---|---|
| `ghcr.io/hanzoai/cd-ui` | `Dockerfile.cd-ui` | `.github/workflows/cd-ui.yml` |
| `ghcr.io/hanzoai/deploy-ui-embed` | `Dockerfile.embed` | `.github/workflows/deploy-ui-embed.yml` |

`cd-ui` serves the dashboard as a standalone SPA on `ghcr.io/hanzoai/spa`;
`deploy-ui-embed` is the same bundle as a bare rootfs at `/dist`, for a
consumer to `COPY --from` into its own `go:embed` tree. Both build the UI the
same way as the `argocd-ui` stage in `./Dockerfile`, which still bundles the UI
into `hanzocd-server` for the standalone deployment.

## Brand

Marks are source assets (`ui/src/assets/images/hanzo-*.svg`), referenced
directly by the components. `ui/src/assets/hanzo-cd.css` is colour only — it
recolours upstream's teal/orange/purple accents to the Hanzo monochrome
palette and leaves the functional health/sync colours alone. It ships in the
image via `ui/src/app/index.html`, and can also be mounted into
`hanzocd-server` at `/shared/app/custom` and selected with `hanzocd-cm`
`ui.cssurl: ./custom/hanzo-cd.css`.

## Build

Build with `GOWORK=off`: `~/work/hanzo/go.work` otherwise captures this repo
and the build fails with "directory prefix cmd does not contain modules listed
in go.work".
