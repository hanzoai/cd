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

# Dispatch is by argv[0], so the name is NOT a subcommand argument — passing it as
# one gets "unknown command" plus a plugin-not-found warning that reads like a
# broken install. From source, name the binary through the env var the switch
# already reads; from the image, the symlink is the name.
CD_BINARY_NAME=hanzocd-mirror FORGE_TOKEN=… GITHUB_TOKEN=… \
  go run ./cmd --table mirror/repos.json --dry-run
```

The fakes in `sync_test.go` all agree with the code about what git is, which is
exactly why none of them caught the missing working directory.
`TestFastForwardMovesARealRepository` builds two bare repositories, leaves the forge
a commit behind, requires it to reach GitHub's tip — then rewinds GitHub so the two
genuinely disagree and requires the forge NOT to move.

**Declared but OFF**: `universe charts/app/values/hanzo/mirror.yaml`, `suspend:
true`. ONE precondition is left, and it is a credential.

The image half is DONE — the pinned digest carries `hanzocd-mirror` and the
repository table, proven by running a probe pod on that exact digest rather than
by reading its tag's date. (The tag `v3.7.0` genuinely predates the subcommand;
the digest beside it does not. Ask the image, not the version number.)

What remains is `hanzo/mirror/{FORGE_TOKEN,GITHUB_TOKEN}`, which is an owner
write. Everything around them is built: the KMSSecret `mirror-env` is live and
authenticates (`LoadedKMSToken: True`), and it fails on exactly one sentence —
`fetch secret "FORGE_TOKEN": kms: secret not found`. So the two values simply
have not been written; nothing else is missing, and turning the job on is then a
one-line diff.

⚠️ Do not check that with `kubectl get kmssecret`. The short name resolves to
kmssecrets.hanzo.ai, a DIFFERENT and empty CRD, and answers "No resources found"
about the wrong thing — which is how this was reported as "no KMSSecret exists".
The real resource is **kmssecrets.secrets.lux.network** and holds 122 objects.

`hanzoai/mirrors` still runs all four legs.

### 7 of the 40 native repos are split-brained, and the job will go RED on them

Measured across the whole table, both sides, by tip: **32 in sync, 7 diverged,
1 resolved.** So unsuspending the job does not quietly converge the estate — it
reports seven repos that need a person, exits non-zero, and moves none of them.
That is the correct behaviour and it is worth knowing BEFORE turning it on,
because a red job on its first run reads like a broken job.

The refusal is loud by construction: `FastForward` asks
`merge-base --is-ancestor forgeTip ghTip`, treats exit 1 as git's ANSWER rather
than an error, and pushes with no leading `+` so git itself refuses anything
else. `Diverged` prints to stderr and `failed()` returns an error so cobra sets
the exit status. Nothing is at risk of being overwritten.

**`native` in this table is a CLAIM — "GitHub is canonical, move the forge onto
it" — and for these seven it is not true.** All seven forge copies are
`NATIVE-on-forge` in the forge's own `mirror` table, i.e. written directly. Both
sides are taking writes. Two shapes:

*No common ancestor at all — two different repositories sharing a name:*

| repo | GitHub | forge | which holds the history |
|---|---|---|---|
| `hanzoai/cloud` | 8 commits, root 2026-08-04 | **5001 commits**, root 2026-05-18 | **the forge**; GitHub is a fresh stub and the classification is INVERTED |
| `hanzoai/gui` | **15901 commits**, root 2020-10-16 | 374, root 2026-03-21 | GitHub (it carries the upstream fork history) |
| `hanzo-apps/id` | 45 | 43 | same root DATE, different root SHA — a re-import that then took writes on both sides (26 files differ) |
| `hanzoai/hanzo.network` | 21 | 19 | same shape (28 files differ) |
| `hanzoai/hanzo.sh` | 93 | 91 | same shape (4 files differ) |

*A real common ancestor, small divergence — mergeable by a person:*

| repo | ancestor | forge-only | github-only |
|---|---|---|---|
| `hanzoai/kms` | 2026-07-27 | 6 | 1 |
| `hanzoai/python-sdk` | 2026-08-07 | 14 | 2 |

`hanzoai/o11y` was the eighth and is DONE: the forge was strictly one commit
ahead, GitHub was its ancestor, so GitHub was fast-forwarded onto it. Both sides
now read `0a75f53cc`. That is the whole class of thing this job automates, in
the direction it does not run.

⚠️ Do not "fix" a row by force-pushing the side you happened to measure first.
`cloud` and `gui` diverge in OPPOSITE directions, so a rule applied uniformly
destroys 5001 commits on one or 15901 on the other. Ancestry decides, per repo,
every time — and where there is no ancestor, a person decides.

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
