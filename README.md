[![Release](https://img.shields.io/github/v/release/hanzoai/cd?label=hanzo-cd)](https://github.com/hanzoai/cd/releases/latest)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/hanzoai/cd/badge)](https://scorecard.dev/viewer/?uri=github.com/hanzoai/cd)

# Hanzo CD — declarative continuous delivery for Kubernetes

Git holds the desired state; the controller applies it and reports drift.

Hanzo CD is a hard fork of [Argo CD](https://github.com/argoproj/argo-cd)
(Apache-2.0 — see [NOTICE](NOTICE)). The engine is theirs. The packaging, API
group and platform integration are ours. **Their community, docs, support
channels and roadmap are not ours**, so this file does not borrow them — the
previous version of this README was upstream's with the names replaced, which
had the effect of putting our name on other people's conference talks and blog
posts.

## What differs from upstream

| | |
|---|---|
| API group | `apps.hanzo.ai` — `Application`, `ApplicationSet`, `AppProject` |
| annotations | `cd.hanzo.ai/*` |
| binaries | `cd`, `cd-server`, `cd-repo-server`, `cd-application-controller`, `cd-applicationset-controller`, `cd-commit-server`, `cd-cmp-server`, `cd-dex`, `cd-git-ask-pass`, `cd-k8s-auth` |
| module | `github.com/hanzoai/cd` |
| image | `ghcr.io/hanzoai/cd` |
| SSO | Hanzo IAM (hanzo.id) over OIDC — dex builds but is not deployed |
| notifications | removed; we use [hanzoai/notify](https://github.com/hanzoai/notify) |

## Where it runs

**[cd.hanzo.ai](https://cd.hanzo.ai)**, namespace `hanzo-cd` on hanzo-k8s.

It moves git → CR for the fleet; the Hanzo operator keeps CR → workload. The two
never write the same field: Hanzo CD owns CR **spec** from git, the operator owns
CR **status** and the child workload. Disjoint by construction, which is what
lets both run without fighting over field ownership.

Deployment layout, the AppProject fences, and the one-observer/three-enforcer
split live in
[`hanzoai/universe` → `infra/k8s/hanzo-cd/`](https://github.com/hanzoai/universe/tree/main/infra/k8s/hanzo-cd).

## Build

```sh
CGO_ENABLED=0 go build ./...      # the gate hanzo.yml runs on every PR
```

Images build on [platform.hanzo.ai](https://platform.hanzo.ai) from the root
`Dockerfile`; `hanzo.yml` declares the target.

**Releases are tagged.** A branch build publishes `sha-<short>` and will never
advance a service already pinned to a release, so shipping means cutting a
`vX.Y.Z` tag — not merging to main.

## License

Apache-2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).
