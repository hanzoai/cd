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

### The table is 37 rows, and 3 repos are deliberately NOT in it

Swept both hosts by tip. It started at 40 rows: **32 in sync, 8 out of step.**
Three of the eight are now reconciled, three left the table, and two remain for
a person. So unsuspending the job does not quietly converge the estate — it will
report the two, exit non-zero, and move neither. Worth knowing BEFORE turning it
on, because a red first run reads like a broken job.

**Reconciled** — `hanzoai/o11y` (forge was one commit ahead, GitHub was its
strict ancestor, so GitHub was moved up), `hanzoai/kms` and `hanzoai/python-sdk`
(both merged; see below). **Still split** — `hanzo-apps/id` and
`hanzoai/hanzo.network`, each with no common ancestor and ~26 files apart.
(`hanzoai/hanzo.sh` is the same shape at 4 files.)

The refusal is loud by construction: `FastForward` asks
`merge-base --is-ancestor forgeTip ghTip`, treats exit 1 as git's ANSWER rather
than an error, and pushes with no leading `+` so git itself refuses anything
else. `Diverged` prints to stderr and `failed()` returns an error so cobra sets
the exit status. Nothing is at risk of being overwritten.

**`native` in this table is a CLAIM — "GitHub is canonical, move the forge onto
it".** Three repos are out of the table because that claim cannot be made true
for them, and a row that can never go green is worse than no row: it teaches
people to ignore a red job. Named here rather than silently dropped, the same
way `kustomization.yaml` in universe names its deliberate exclusions.

| not in the table | why |
|---|---|
| `hanzoai/kms` | **archived on GitHub.** An archived repo refuses pushes, so the two can never re-converge by pushing and the forge is canonical by construction. Its one GitHub commit (a README notice) was merged onto the forge by hand — that is as reconciled as it can get. |
| `hanzoai/gui` | **archived on GitHub**, same reason. Note GitHub holds 15901 commits from 2020 (the upstream fork history) against 374 on the forge from 2026-03-21 — the forge is a deliberate restart, so "GitHub has more history" is not an argument for moving the forge onto it. |
| `hanzoai/cloud` | the claim is **inverted**. GitHub's repo was CREATED 2026-08-06 and holds 8 commits; the forge's root is 2026-05-18 with 5001, and the forge is what CI and CD read. Both are being written today. Which one is canonical going forward is an owner's decision, not a table's. |

An archived upstream is worth checking for directly — it is the one condition
that makes `native` structurally impossible rather than merely wrong, and it is
two of the three. Exactly 2 of the original 40 were archived, both listed above.

Of what remained, all forge copies are `NATIVE-on-forge` in the forge's own
`mirror` table, i.e. written directly, so both sides take writes. Two shapes:

*No common ancestor at all — two different repositories sharing a name:*

| repo | GitHub | forge | which holds the history |
|---|---|---|---|
| `hanzo-apps/id` | 45 | 43 | same root DATE, different root SHA — a re-import that then took writes on both sides (26 files differ) |
| `hanzoai/hanzo.network` | 21 | 19 | same shape (28 files differ) |
| `hanzoai/hanzo.sh` | 93 | 91 | same shape (4 files differ) |

*A real common ancestor — both were merged, and neither needed a winner:*

`hanzoai/kms` (ancestor 2026-07-27) had touched **disjoint files**: six commits
on the forge to `.hanzo/workflows/release.yml`, `go.mod` and `go.sum`, one on
GitHub to `README.md`. Nothing conflicted and both statements were true, so the
merge kept both. It lives on the forge only — GitHub is archived and refused it.

`hanzoai/python-sdk` (ancestor 2026-08-07) is the instructive one. It looked
like a 14-vs-2 divergence and was really the SAME work done twice: both sides
had applied a commit named "one generated client in the wheel", 2,213 files and
800,555 deletions each, and the forge retains nothing GitHub dropped. `git
cherry` — patch-ids, not commit ids — showed the other GitHub commit was already
present under a different sha. The forge had then gone on to five client
regenerations, 3.2.2 → 3.2.7.

Merged `-s ours`, so the histories join (GitHub fast-forwards, second head gone)
while the tree stays the newer regeneration. A plain merge reported ~140
modify/delete conflicts purely because the two deletions carry different shas —
resolving those by taking trees would have RESURRECTED modules both sides had
agreed to delete. `pkg/hanzoai/cloud/models/wrote.py` is the tell: GitHub
deleted it, the forge regenerated it, and `models/__init__.py:2345` imports it,
so the "obvious" resolution breaks the package. Verified after merging — the
tree imports, 4,660 names, version 3.2.7.

`hanzoai/o11y` was simplest: the forge was strictly one commit ahead and GitHub
was its ancestor, so GitHub was fast-forwarded onto it. Both read `0a75f53cc`.
That is exactly what this job automates, in the direction it does not run.

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
