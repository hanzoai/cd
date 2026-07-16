# Hanzo CD

A fork of Argo CD (Apache-2.0). See NOTICE for attribution and the list of
changes. Product name is Hanzo CD; the Go module is
`github.com/hanzoai/deploy/v3`; images publish to `ghcr.io/hanzoai/deploy`.

## Lineages

Two branches build images, and they are NOT interchangeable.

| branch | VERSION | what it is |
|---|---|---|
| `hanzo/v3.4.5` | 3.4.5 | the release lineage production runs |
| `master` | 3.6.0 | tracks upstream master |

`v3.4.5` is **not an ancestor of master** (`git merge-base --is-ancestor
v3.4.5 master` exits 1). Master is 823 commits ahead of v3.4.5, and v3.4.5
carries 98 commits master lacks — release-branch backports. So an image built
from master is a different codebase, not a rebuild.

The first cutover of `hanzo-cd` off `quay.io/argoproj/argocd:v3.4.5` must come
from `hanzo/v3.4.5`. It changes the registry and nothing else. Moving
production onto the master lineage is a separate change with its own gate.

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
not a cache inside a process. It is the shared cache **between four separate
processes**, and the pub/sub channel they use to invalidate each other.

`util/cache.CacheClient` is already a clean seam with two implementations,
`redisCache` and `InMemoryCache`. It is tempting to conclude the swap is free.
It is not, and the trap is specific:

    func (i *InMemoryCache) OnUpdated(...) error   { return nil }
    func (i *InMemoryCache) NotifyUpdated(...) error { return nil }

`InMemoryCache` satisfies the interface by making cross-process invalidation a
silent no-op. Dropping it in under the current four-process split does not
degrade performance — it serves stale state with no error anywhere. That is a
wrong-state bug wearing a cache's clothes.

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
