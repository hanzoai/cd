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

`mirror/` plus `cmd/cd-mirror` push our own repos FORWARD onto git.hanzo.ai. It is
a subcommand of the one CD binary, dispatched by `argv[0]` like every other
`hanzocd-*`, so it travels in the image the control plane already builds and pins —
no second artifact to version, and no second chance for the two to disagree about
which repos exist. The table ships beside it at `common.DefaultMirrorTable`, so the
JSON validated in CI is the JSON a scheduled run reads.

**It replaces ONE of four legs in `hanzoai/mirrors`. Deleting that repo now loses
work:**

| leg | every | what it does | here? |
|---|---|---|---|
| `sync.py` | 15m | push native repos forward onto the forge | **yes** |
| `reconcile.py` | 6h | create/refresh the pull-mirror for every qualifying repo across 15 orgs — how a new repo becomes visible here at all | no |
| `releases.py` | 6h | copy releases and their assets from GitHub and GitLab | no |
| `sync-from-github` | 10m | keeps that repo itself current | dies with it |

The differences below are the bugs `sync.py` had — each of which failed
**silently**:

- **Names are resolved, never matched against a listing.** `Resolve` and `List` are
  separate methods so a call site must choose. `GET /repos/{org}/{name}` follows a
  rename and a transfer permanently; a listing reports a repo only under its
  current name in its current org. Six entries whose repos had moved were never
  visited once, because an absent name raises nothing.
- **The table is config** (`mirror/repos.json`), not a list literal. Adding a repo
  used to be a code edit. `hanzoai/cloud` was in *neither* leg — `is_mirror=0` so
  nothing pulled in, absent from the table so nothing pushed in — and its CI ran on
  a commit four behind for a day while merely looking red. **40 entries**, and the
  port originally kept 24 of them: `hanzoai/ci` was among the sixteen dropped, the
  reusable pipeline every repo in three orgs imports, where a stale pipeline runs
  green and no caller can tell. Two tests now guard the count and the allowlist.
- **`Owned` is an allowlist and deliberately broader than the table.** A declared
  repo can be transferred, and `owned` guards the name GitHub RESOLVES to — so the
  org a repo may land in tomorrow has to be listed today. The live run proves the
  point: eight of the forty resolve only because names are followed (three moved to
  `hanzo-apps`, three to `hanzo-inc`, `esign`→`sign`, `kms-v1`→`kms`), and the
  three in `hanzo-inc` were REFUSED by name until that org was listed rather than
  pushed into an org nobody had vouched for.
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

- **Every git command runs in a repository.** `RealGit` takes the directory; a
  scheduled job starts in none, and git answers `fatal: not a git repository` to the
  fetch — on the branch that reports GITHUB as unreadable. Perfect credentials and a
  perfect table would still have produced a fleet-wide "cannot read github", and the
  dry run would have stayed green throughout, because a dry run never touches git.
  `Workspace` keeps one object store per repo under a root that MAY survive between
  runs; it is an optimisation, so an empty root is equally correct.
- **Both sides are fetched, forge first.** Its tip has to be local before a real
  divergence can be NAMED rather than reported as "ancestry undetermined" — and once
  its history is present those objects are offered in negotiation, so the internet
  side of a run is a delta even on an empty store. That is what makes a ten-minute
  schedule affordable across two dozen repos, and why the job needs no volume.

```bash
go test ./mirror/ ./cmd/cd-mirror/...   # incl. a real two-repository end-to-end
FORGE_TOKEN=… GITHUB_TOKEN=… go run ./cmd hanzocd-mirror --dry-run  # resolve only
```

The fakes in `sync_test.go` all agree with the code about what git is, which is
exactly why none of them caught the missing working directory.
`TestFastForwardMovesARealRepository` builds two bare repositories, leaves the forge
a commit behind, requires it to reach GitHub's tip — then rewinds GitHub so the two
genuinely disagree and requires the forge NOT to move.

**Declared but OFF**: `universe charts/app/values/hanzo/mirror.yaml`, `suspend:
true`. Two preconditions, both stated in that file — an image carrying
`hanzocd-mirror` (v3.7.0 predates it), and `hanzo/mirror/{FORGE_TOKEN,GITHUB_TOKEN}`
in KMS, which is an owner write. `hanzoai/mirrors` still runs all four legs.

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
