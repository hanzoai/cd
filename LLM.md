# Hanzo CD

A fork of Argo CD (Apache-2.0). See NOTICE for attribution and the list of
changes. Product name is Hanzo CD; the Go module is
`github.com/hanzoai/deploy/v3`; images publish to `ghcr.io/hanzoai/deploy`.

## Lineages

One branch carries the rebrand.

| branch | VERSION | module | what it is |
|---|---|---|---|
| `hanzo/v3.4.5` | 3.4.5 | `github.com/hanzoai/deploy/v3` | the release lineage production runs |
| `master` | 3.6.0 | `github.com/argoproj/argo-cd/v3` | an upstream mirror, unbranded |

`v3.4.5` is **not an ancestor of master** (`git merge-base --is-ancestor
v3.4.5 master` exits 1). Master is 823 commits ahead of v3.4.5, and v3.4.5
carries 98 commits master lacks — release-branch backports. So an image built
from master is a different codebase, not a rebuild.

The first cutover of `hanzo-cd` off `quay.io/argoproj/argocd:v3.4.5` must come
from `hanzo/v3.4.5`. It changes the registry and nothing else. Moving
production onto the master lineage is a separate change with its own gate: it
is a 3.4.5 → 3.6.0 upgrade, and rebranding it at the same time makes one change
out of two.

### The rebrand is a rule, not a diff

Rebranding master onto a second branch and maintaining it there is the same
intent held in two places. It is also unnecessary, because the rebrand is a
pure substitution over the files where the module path is code rather than
prose, and one command reproduces it byte-for-byte:

    grep -rlZ 'github.com/argoproj/argo-cd/v3' \
      --include='*.go' --include='*.proto' --include='*.yaml' \
      --include='*.sh' --include='go.mod' --include='Makefile' \
      --include='Procfile' . \
      | xargs -0 sed -i 's|github.com/argoproj/argo-cd/v3|github.com/hanzoai/deploy/v3|g'

Run against the v3.4.5 base that is exactly the rebrand commit: 627 files,
2950 insertions, 2950 deletions — insertions equal deletions because nothing is
added, only swapped. Prose in `docs/` keeps upstream's name and URLs, which is
what attribution requires.

So the rebrand is regenerated at the moment a lineage is adopted, not carried
on a branch that rots against upstream. When 3.6 is adopted, run the rule then.

Build with `GOWORK=off`: `~/work/hanzo/go.work` otherwise captures this repo
and the build fails with "directory prefix cmd does not contain modules listed
in go.work".

## Enabling Actions arms upstream's quay.io workflows

Workflows are dormant on this fork — GitHub disables them on forks, and only
`pages-build-deployment` is registered. That dormancy is currently the only
thing preventing upstream's workflows from running:

- `image.yaml` fires on a push to `master`
- `release.yaml` fires on a `v*` tag

Both run on `ubuntu-24.04` (GitHub-hosted, which the house rule forbids) and
both publish to `quay.io/argoproj`. Turning Actions on so `hanzo-image.yml`
can run arms them at the same instant. Enabling must therefore be paired with
disabling each upstream workflow:

    PUT /repos/hanzoai/deploy/actions/workflows/{id}/disable

Doing it that way, rather than deleting or editing their YAML, keeps upstream
merges conflict-free here.

## Identity: the OIDC contract iam2 must satisfy

Dex is a husk and is not what authenticates anyone today: `argocd-dex-server`
runs at `replicas=0`, and `dex.config`, `oidc.config` and `url` in `argocd-cm`
are all empty. What actually authenticates is `admin.enabled=true` — a single
shared local password, not individually attributable, no MFA. It is bounded
only by there being no Ingress into the `hanzo-cd` namespace. That is the real
identity hole; dex is a corpse in the room, not the problem.

Hanzo CD consumes a plain OIDC provider. Switching issuers is a ConfigMap
change, not a code change, so this contract is issuer-agnostic by construction:
it can point at `hanzo.id` today and iam2 the day iam2 can serve it, with no
rebuild.

`util/settings.OIDCConfig` is the whole surface. To be Hanzo CD's IdP, an
issuer must provide:

1. **Discovery** at `{issuer}/.well-known/openid-configuration`, advertising
   `authorization_endpoint`, `token_endpoint`, `jwks_uri`, `issuer`, and
   `response_types_supported` including `code`.
2. **JWKS** at the advertised `jwks_uri`, with `kid`-matched keys. Signature
   verification is against this; keys must survive rotation.
3. **Authorization Code flow**, plus **PKCE** — the `argocd` CLI is a public
   client and sets `enablePKCEAuthentication`.
4. **Two client registrations**: a confidential `clientID`/`clientSecret` for
   the web UI, and a separate public `cliClientID` for the CLI. The CLI client
   must permit the loopback redirect the CLI listens on.
5. **Scopes**: `openid`, `profile`, `email`, `groups`. `groups` is the one that
   matters — it carries authorization.
6. **A `groups` claim in the ID token**. If it is only available from userinfo,
   the issuer must serve `userinfo` and CD sets `enableUserInfoGroups` with
   `userInfoPath`.

### Claim to privilege — one predicate, no second definition

Platform sudo in Hanzo is `owner == "admin"` (membership of the reserved
`admin` org). That predicate has exactly one definition and CD does not get to
invent a second one.

Hanzo CD's authorization is `argocd-rbac-cm`'s `policy.csv`, which maps claim
values to roles. The issuer's job is to emit the org as a group so the mapping
is a lookup, not a judgement:

    groups: ["admin"]        # member of the reserved admin org  => platform sudo
    groups: ["<org-slug>"]   # any other org                     => not sudo

    # argocd-rbac-cm
    policy.csv: |
      g, admin, role:admin
    policy.default: role:readonly

`role:admin` is granted to the `admin` **group**, which is org membership —
never to a per-org `isAdmin`. Trusting a per-org admin flag for platform gating
is the privilege-escalation bug the house rule names; keeping the mapping to
group-membership-only is what forecloses it.

Once an issuer is wired, `admin.enabled` must go to `false` in the same change.
Leaving the shared local password enabled beside a real IdP means the IdP is
decoration.

**Flagged OFF until iam2 can serve a login.** iam2 is not deployed
(`iam2.hanzo.ai` resolves only to the shared `hanzo/ingress-lb` and returns
404), and its own HEAD records that `users.VerifyPassword` calls bcrypt
unconditionally while every live row is argon2id, so cutover would fail every
existing user's login. `hanzo.id` serves discovery today and satisfies this
contract now.

## Redis: a consequence of the process split, not a caching choice

All three binaries — `argocd-server`, `argocd-repo-server`,
`argocd-application-controller` — construct their own `redis.Client`. Redis is
not a cache inside a process. It is the shared store **between four separate
processes**, and the pub/sub channel they use to invalidate each other.

Classify each use by who writes and who reads, not by what the operation is
called. `SetItem`, `GetItem` and `SetNX` read as cache primitives, and reading
them that way is how you conclude redis is droppable. The primitive is a claim.
The writer-to-reader boundary is the value.

### The resource tree is computed in one process and read in another

`argocd-server` has no cluster informers. Only the controller watches the
cluster, so the server cannot compute a resource tree at all. On a miss,
`getCachedAppState` pokes the controller via `Refresh` and re-reads the same
store. That works only because both point at one redis.

    argocd-application-controller-0  hanzo-val-37mglj    writes the tree
    argocd-server-…                  worker-pool-3ck6l8  reads the tree

Separate pods, separate nodes. Without the shared store the UI returns
`error getting cached app resource tree` and never recovers.

| flow | writer | reader |
|---|---|---|
| `SetAppResourcesTree`, `SetAppManagedResources` | controller | server |
| `NotifyUpdated` → `OnUpdated`, backing `WatchResourceTree` | controller | server |
| manifest cache invalidation (`util/webhook`) | server | repo-server |
| OIDC refresh token, encrypted under `server.secretkey` | server | server, across restarts |

`util/cache.CacheClient` is a clean seam with two implementations, `redisCache`
and `InMemoryCache`, so the swap looks free. It is not, and the trap is
specific:

    func (i *InMemoryCache) OnUpdated(...) error   { return nil }
    func (i *InMemoryCache) NotifyUpdated(...) error { return nil }

`twoLevelClient.OnUpdated` delegates only to `externalCache`, so an in-memory
L1 structurally cannot carry pub/sub, and `InMemoryCache` satisfies the
interface by making cross-process invalidation a silent no-op. Dropping it in
under the current four-process split does not degrade performance — it serves
stale state with no error anywhere. That is a wrong-state bug wearing a cache's
clothes.

### There is no in-memory path

An empty `--redis` does not mean "no redis". `AddCacheFlagsToCmd` has no
in-memory branch and no flag, and an empty address defaults to
`common.DefaultRedisAddr`. Dial errors do not map to `ErrCacheMiss` — only a
real miss does — so `getCachedAppState` never triggers a refresh and the dial
error reaches the UI.

A nil client is worse. `IsTokenRevoked(nil)` is safe, but
`userStateStorage.Init` panics inside its own goroutines, which no caller can
recover: the process exits. Around 40 tests construct
`NewUserStateStorage(nil)` and stay green only because they never call `Init`.
Their passing is not evidence of safety.

### The revoked token list, and why it does not get a store

`RevokeToken` has exactly one writer: `server/logout`. There is no admin revoke
API. The entry is written with `ttl = time.Until(exp)`, so the list's whole job
is the gap between a logout and the token's own expiry.

Down and flushed are not the same failure:

- redis **down** — `loadRevokedTokens` returns the iterator error before the
  assignment and `loadRevokedTokensSafe` retries, so the in-memory set is
  **preserved**.
- redis **flushed** — `Scan` succeeds with zero keys and the set is replaced
  wholesale with an empty one. `recentRevokedTokens` buys exactly one cycle.

Redis here is a Deployment with no PVC, so a restart is a flush. The list
already fails open in stock Argo CD, independent of anything in this fork. The
failure is also swallowed — `log.Warnf`, then the redirect — so a logout can
report success while leaving the session valid.

The gap it covers is bounded by the token's expiry, which is
`users.session.duration`. The control is therefore the session lifetime, not a
new store. A store added to shorten a window that a config value already bounds
complects session lifetime with revocation state; setting the lifetime keeps
them one thing each, and is one line.

The durable case is already covered without redis: an API key is checked
against `account.TokenIndex(id) == -1` in `argocd-cm`. The redis list only ever
served short-lived session tokens — the credential for which waiting out the
expiry is the right answer.

**There is no `revocations` store.** The value should not exist. Upstream's list
stays as upstream best-effort and nothing relies on it.

### The git-ref lock is real, and still not `ha`'s job

`SetNX` here is a genuine claim of ownership. `TryLockGitRefCache` and
`GetOrLockGitReferences` exist so that "only one process is able to claim
ownership". Reading `DisableOverwrite` as a cache-write option and concluding
there is no lock reads the primitive instead of the caller, and is wrong.

The conclusion survives the correction. On timeout the loser returns its own
lock id and proceeds as owner anyway, so the lock is advisory dedup: losing it
costs a duplicate `git ls-remote`, not correctness. Nothing here needs
`hanzoai/ha`.

### Ruling: redis stays

Redis is not droppable by deleting a Deployment. A replacement has to be a
shared store carrying app state **and** pub/sub across processes, and that is a
design task with its own gate — not a deletion.

### The trade, stated for a decision

Redis exists because the components are split. HIP-0106 — *calls when
co-resident, ZAP RPC when split* — means co-residency would delete redis and
the internal gRPC in one move, using a seam that already exists.

**What co-residency buys:** redis goes away entirely, and so does the internal
server↔repo-server↔controller gRPC. Two subsystems deleted by one structural
change rather than two ports. `InMemoryCache` becomes correct rather than
dangerous, because there is no longer a second process to invalidate.

**What it costs:** `repo-server` renders repo-supplied Helm and Kustomize. That
is untrusted code execution, and it is isolated in its own process on purpose.
Co-residency merges that trust boundary into the API server — the process
holding cluster credentials and every user's session. Deleting redis is not a
reason to do that.

**What makes it tractable:** every component is `replicas=1` with no HPA, so
nothing today needs cross-replica sharing. The only sharing that matters is
cross-component.

**The third option**, and the honest recommendation: keep `repo-server` split
and give the split a non-redis cache and invalidation path. That keeps the
trust boundary, drops the redis dependency, and leaves ZAP a real job rather
than a transport swap done for its own sake. It costs an implementation of
`OnUpdated`/`NotifyUpdated` that actually delivers, which `InMemoryCache` does
not.

This is a CTO decision. It should not be made as a side effect of deleting a
Deployment.

## Persistence inventory

What is stored, where, who reads it, and the ruling.

| what | where today | read by | ruling |
|---|---|---|---|
| `admin.password`, `admin.passwordMtime`, `server.secretkey` | `argocd-secret` | server | **KMS**, gated below |
| repository credentials (`hanzo-cd-repo-universe`: url/username/password) | k8s Secret, `argocd.argoproj.io/secret-type: repository` | repo-server, server | **KMS** |
| `argocd-initial-admin-secret` | k8s Secret | bootstrap only | **delete**, after a KMS-sourced admin credential exists |
| `argocd-cm`, `argocd-rbac-cm` | ConfigMaps | all | **stays** |
| app state, manifest, cluster state, OIDC refresh token | redis | server, repo-server, controller | **stays** — the store is the bus |
| revoked token list | redis | server | **no store** — bounded by `users.session.duration` |
| `Application`, `AppProject` | etcd (CRDs) | controller, server, CLI, kubectl | **stays in etcd** |

### Secrets move to KMS, and it costs no fork code

A k8s Secret is base64, not encryption. `hanzo-cd-repo-universe` is a live git
credential sitting in etcd in the clear, and `argocd-initial-admin-secret` — the
bootstrap admin password — is still present long after bootstrap.

The move needs **zero changes to this fork**. The house mechanism already
exists and is already running in this cluster: `KMSSecret`
(`secrets.lux.network/v1alpha1`) syncs KMS → CR → k8s Secret, and is in live
use elsewhere. Argo keeps reading k8s Secrets through its existing informer;
KMS becomes the source of truth. Prefer deletion: the right change here is a
`KMSSecret` CR, not a secret backend in Go.

`argocd-secret` carries a gate the repository credential does not.
`argocd-server` **writes back to it**: `InitializeSettings` generates a 32-byte
`server.secretkey` when absent, and `saveSignatureAndCertificate` upserts it. A
`KMSSecret` whose template omits `server.secretkey` therefore write-write flaps
against the server every resync. Both effects are fail-closed, and both are
outages:

- every session signature changes each cycle, so every session dies each cycle;
- `server.secretkey` is also the key the OIDC refresh tokens in redis are
  encrypted under, so each rotation makes every cached token undecryptable.

Put `server.secretkey` in KMS and in the template, or leave `argocd-secret`
alone. Settle this before `argocd-secret` goes under KMS, not after.

### Settings stay in ConfigMaps

`argocd-cm` and `argocd-rbac-cm` are declarative configuration, held in git at
`universe/infra/k8s/hanzo-cd` and applied with `kubectl apply -k`, read through
a k8s informer, and gated by k8s RBAC. Moving them into an embedded database
would convert declarative config into out-of-band mutable state, take them out
of git, and drop RBAC — losing the property that makes them good. They hold no
secrets. They stay.

### Application and AppProject stay in etcd

The CRD is the right value in the right place.

- **It is the public contract.** `kubectl get applications` is how these are
  operated, including by us; ApplicationSet generates Applications *as* CRs;
  k8s RBAC gates them.
- **The controller is built on watches.** Argo reconciles via client-go
  informers. An embedded SQL store has no watch primitive; serving Applications
  from it means replacing informers with polling — that is not a store swap,
  it is a rewrite of the reconciler.
- **It is the house pattern.** `hanzod` reconciles `App` CRs (`hanzo.ai/v1`)
  from etcd for exactly these reasons. Moving Argo the other way would make the
  two operators disagree about where desired state lives.
- **The live chain depends on it.** `universe-crs` *is* an Application. It syncs
  the App CRs hanzod reconciles. Moving Applications out of etcd breaks the
  path all of production ships through.

etcd is already a consistent, watchable, RBAC-gated store, and a controller's
desired state is precisely what it is for. A Kubernetes controller whose
desired state is not in Kubernetes is not a Kubernetes controller. So the scope
of "Argo on our stack" is secrets → KMS. The cache is the bus and stays; the
revocation list gets a bound, not a store; Applications stay where the operator
pattern wants them.

## ZAP: staged, not half-landed

The gRPC surface is 106 files importing `google.golang.org/grpc`, 17 proto
service definitions, and 27 server/client constructions. `zap-proto/go` is not
a dependency here yet.

The part that decides the sequencing: gRPC here is not only internal plumbing.
`pkg/apiclient` is the contract every `argocd` CLI in existence speaks, and
grpc-web serves the UI. That is a public compatibility surface. Changing it
breaks clients that are not ours to upgrade.

So the internal transport and the public contract are two different jobs with
two different risk profiles, and only the internal one is reachable from
HIP-0106 without breaking users. This is the plane everything else ships
through; a partially-migrated transport here is not a partial win.

## Base images

`Dockerfile` pulls four bases from `docker.io` (ubuntu 26.04, golang 1.26.5
twice, node 24.17.0). The registry rule's destination stands; the swap is
sequenced behind evidence, not exempted:

- `ghcr.io/hanzoai/node` publishes only `18.19.0-alpine`; the UI build needs
  24.x. Verified by anonymous pull against ghcr.
- `ghcr.io/hanzoai/ubuntu` and `ghcr.io/hanzoai/golang` return 403 — they
  exist, but their tags cannot be enumerated without `read:packages`, so
  nothing can be pinned to a digest.

Swapping blind breaks the build. The bases needed are node 24.17.0, golang
1.26.5, and ubuntu 26.04 published to `ghcr.io/hanzoai/*`.

## The `gitops-engine` submodule, and how cloud embeds it

`gitops-engine/` is its own Go module with its own `go.mod`, wired into the
parent through `replace github.com/argoproj/argo-cd/gitops-engine[...] =>
./gitops-engine`. Because it is independently importable, it is also the seam
the Hanzo cloud money binary uses: `github.com/hanzoai/cloud` `clients/deploy`
serves `/v1/deploy` — the fleet view over the operator `App` CRs — by running
the engine's read path (`pkg/health`, `pkg/diff`, `pkg/sync`) **in-process**.
There is no separate `argocd` process on that side.

The load-bearing decision, true on every lineage: **the submodule keeps its
`argoproj` module path; it is not renamed to `github.com/hanzoai/deploy/...`**.
The parent imports the engine under the `argoproj` path in hundreds of files
and wires it with its own `replace ... => ./gitops-engine`; renaming the
submodule would break the parent's `go build ./...` and force a repo-wide
import rewrite for no gain. The parent module rebrand (see Lineages) stops at
the parent — `hanzo/v3.4.5` carries `gitops-engine` at
`github.com/argoproj/argo-cd/gitops-engine`, and the 3.6 line at
`.../gitops-engine/v3`, both unrenamed. The fork is Hanzo-owned by repository,
patched here as needed; ownership is not the module string.

Cloud consumes it under that same `argoproj` path and points the path at this
fork with a **consumer-side filesystem replace** onto the sibling checkout — no
rename, no import rewrite. That wiring is the not-yet-landed `clients/deploy`
P2b step: cloud's `go.mod` carries no `gitops-engine` require or replace today,
and the read path is still the field-strip diff. The engine's read path funnels
through `pkg/utils/kube`, which drags the `k8s.io/kubernetes` staging tree, so
whichever `gitops-engine` version cloud embeds it must pin that k8s stack to
its own (`controller-runtime`-compatible) versions so the money binary's k8s
does not move. Those pins are the consumer's concern and live in cloud, against
the engine version cloud actually embeds — they are not restated here, where
they would rot against it.

End state: tag the `gitops-engine/` submodule on this repo and swap cloud's
filesystem replace for a pinned pseudo-version — same import path, no code
change. This section consolidates the `feat/hanzo-gitops-engine` note onto the
release lineage; that branch was a doc-only branch off the unbranded `master`
mirror and is retired.
