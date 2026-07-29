# Hanzo CD — Declarative Continuous Delivery for Kubernetes

Hanzo CD is the GitOps control plane for the Hanzo fleet: it reads desired state
from git and reconciles it into Kubernetes, continuously, with drift correction
and a sync/health verdict per application.

It is a hard fork of [Argo CD](https://github.com/argoproj/argo-cd), a CNCF
graduated project, and owes that project everything — see [NOTICE](NOTICE).
This repository does not track upstream and does not contribute back.

## What is different from upstream

**The API group is `apps.hanzo.ai`.** Applications, ApplicationSets and
AppProjects are served at `apps.hanzo.ai/v1alpha1`, and every annotation, label
and finalizer the product defines carries that prefix — `apps.hanzo.ai/hook`,
`apps.hanzo.ai/sync-wave`, `resources-finalizer.apps.hanzo.ai`, and so on.
Manifests written for upstream's `argocd.argoproj.io/*` annotations need those
keys renamed.

`argoproj.io` still appears throughout the tree, and correctly so: Argo
Rollouts, Argo Workflows and Argo Promoter are separate products whose custom
resources we merely observe. Their health checks under
`resource_customizations/` must keep naming them. See [AGENTS.md](AGENTS.md)
before changing any of these strings.

Components are named `hanzocd-*` and the CLI is `hanzocd`.

## Documentation

Upstream's [documentation](https://argo-cd.readthedocs.io/) still describes the
concepts and most of the CLI accurately. Substitute the API group and annotation
prefix above when following it.

## Contributing

Issues and pull requests belong here, not upstream. Read
[AGENTS.md](AGENTS.md) first — it covers the group split, the derive-don't-restate
rule, and how to establish a test baseline before assuming a red package is your
fault. CI is native (`git.hanzo.ai`, `cd.hanzo.ai`); there is no GitHub Actions
pipeline.

## License

Apache 2.0, inherited from upstream. See [LICENSE](LICENSE) and [NOTICE](NOTICE).
