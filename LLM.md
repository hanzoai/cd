# hanzoai/cd — agent guide

Hanzo CD: the continuous-delivery control plane behind **cd.hanzo.ai**. It
reconciles declared state from `hanzoai/universe` into the clusters.

**Lineage, stated plainly.** This is a fork of `argoproj/argo-cd`, Apache-2.0.
Module `github.com/hanzoai/cd`. Upstream's `Application`, `ApplicationSet` and
`AppProject` are here under `apps.hanzo.ai/v1`, and the controllers keep their
upstream shape (`controller/appcontroller.go`, `applicationset/`, `cmpserver/`,
`commitserver/`). Knowing that is the fastest way to understand any file here — but
do not carry upstream's *project* rules across: Hanzo CD is not a CNCF project, has
no upstream PR gate, and per the house rule work lands on `main`, not in a PR queue.

## What CD actually reconciles here

Two planes, and mixing them up is the most common mistake:

| plane | where | who reconciles |
|---|---|---|
| **chart values** | `universe charts/app/values/<ns>/<name>.yaml` | the `fleet` ApplicationSet (git files generator `charts/app/values/*/*.yaml`) |
| **operator CRs** | `universe infra/k8s/operator/crs/*.yaml` | the `universe-crs` Application, selfHeal |

**New services go in the chart plane.** 57 services already moved there, and
`crs/kustomization.yaml` states outright that the App CRD is no longer how those
are declared. A CR dropped into `crs/` that is not listed in that kustomization is
not "inactive" — CD keeps the live object flagged `requiresPruning` and holds the
whole application OutOfSync, so the file is silently inert. That failure looks like
Synced/Healthy with the service simply absent.

**Sync is manual by convention.** 268 of 277 Applications carry no `automated`
policy; 215 sit `OutOfSync / Healthy` on purpose. A generated Application at
`OutOfSync / Missing` is a new workload awaiting a person, not a fault.

## `mirror/` — which repos live on the forge, and which way they sync

`mirror/` plus `cmd/hanzo-mirror` reconcile git.hanzo.ai against github.com. This
replaced a Python script (`hanzoai/mirrors`), and the differences are the bugs that
script had — each of which failed **silently**:

- **Names are resolved, never matched against a listing.** `Resolve` and `List` are
  separate methods so a call site must choose. `GET /repos/{org}/{name}` follows a
  rename and a transfer permanently; a listing reports a repo only under its
  current name in its current org. Six entries whose repos had moved were never
  visited once, because an absent name raises nothing.
- **The table is config** (`mirror/repos.json`), not a list literal. Adding a repo
  used to be a code edit. `hanzoai/cloud` was in *neither* leg — `is_mirror=0` so
  nothing pulled in, absent from the table so nothing pushed in — and its CI ran on
  a commit four behind for a day while merely looking red.
- **`Direction` is one typed value.** `mirror` previously meant both a vestigial git
  config flag and the forge's real push refspecs; reading the wrong one gives a
  confident wrong answer about whether a push can delete history.
- **Fast-forward is git's guarantee, not the program's** — no leading `+` in the
  refspec. A check in Go can be reasoned around by a later edit; a missing `+`
  cannot. `TestFastForwardNeverForces` asserts on the *argument*, so it fails if
  anyone adds one. A push mirror once pruned the `v1.0`–`v1.31` tag range off an IAM
  repo and the objects were then collected.
- **Divergence is reported and left alone**, exiting non-zero. Two people
  disagreeing about history is a question for a person; settling it destroys
  whichever side lost.
- **Absent credentials fail closed.** An empty token yields a 401 that reads like a
  permissions problem — the shape that hid a broken npm publish for weeks, every
  request unauthenticated because a name resolved to `""` and nothing said so.

```bash
go test ./mirror/                                            # 7 tests, no network
FORGE_TOKEN=… GITHUB_TOKEN=… go run ./cmd/hanzo-mirror \
  -config mirror/repos.json -dry-run                         # resolve + report only
```

Not yet on a schedule; `hanzoai/mirrors` still runs. Deleting that before this
replaces it leaves a window with neither.

## Build

```bash
go build ./...          # covers the whole module — do this, not just changed pkgs
go test ./mirror/
make codegen            # REQUIRED if any API struct changed, or manifests drift
```

A wrong **internal** import does not read as a typo — it reads as a missing
dependency (`cannot find module providing package … -mod=readonly`) and sends you
hunting a `require` line that was never the problem. That is what a stale
`github.com/hanzoai/deploy/...` path looks like after the module rename to
`github.com/hanzoai/cd`.

## Conventions

- Paths are `/v1/…` — never an `/api/` prefix.
- Land on `main`. No PR queue, no branches left behind.
- `LLM.md` is the canonical guide; `CLAUDE.md` and `AGENTS.md` are symlinks to it.
