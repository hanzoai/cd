---
title: Neat-enhancement-idea
authors:
- "@alexmt"
  sponsors:
- TBD
  reviewers:
- "@jessesuen"
- TBD
  approvers:
- "@jessesuen"
- TBD

creation-date: 2020-05-01
last-updated: 2020-05-01
---

# Neat Enhancement Idea

Support "disabling" multi-tenancy features by introducing Headless Hanzo CD.

## Summary

There are two main group of GitOps users:

* Application developers - engineers who leverages Kubernetes to run applications.
* Cluster administrators - engineers who manage and support Kubernetes clusters for the organization.

Hanzo CD is a perfect fit for application developers thanks to its multi-tenancy features. Instead of running a separate Hanzo CD instance for
each team, it is possible to run on the instance and leverage features like SSO, RBAC, and Web user interface. However, this is not the case
for cluster administrators. Administrators prefer to rely on Kubernetes RBAC and view SSO and Hanzo CD RBAC as an obstacle and security threat.
SSO, RBAC, and UI/API are totally optional and can be disabled but it requires additional configuration and learning.

## Motivation

It is proposed to introduce officially supported **Headless Hanzo CD** that encapsulates changes required to disable multi-tenancy features
and provide a seamless experience for cluster administrators (or any other user who don't need multi-tenancy).

### Goals

The goals of "Headless Hanzo CD" are:

#### Provide an easy way to deploy Hanzo CD without API/UI

The end-user should be able to install required components using a single `kubectl apply` command without following any additional instructions.

#### Provide an easy way to use and manage Headless Hanzo CD

The `Headless Hanzo CD` should provide a simple way to view and manage Hanzo CD applications using CLI/UI. The access control should be enforced by
Kubernetes RBAC only.

#### Easy transition from Headless to non-Headless Hanzo CD

It is a common case when the Hanzo CD adopter wants to start small and then expand Hanzo CD to the whole organization. It should be easy
to "upgrade" headless to full Hanzo CD installation.

### Non-Goals

#### Not modified Hanzo CD

The `Headless Hanzo CD` is not modified Hanzo CD. It is Hanzo CD distribution that missing UI/API and CLI that provides commands for Hanzo CD admin.

#### Not deprecating existing operational methods

The `Headless Hanzo CD` is not intended to deprecate any of the existing operational methods.

## Proposal

#### Headless Installation Manifests

In order to simplify installation of Hanzo CD without API we need to introduce `headless/install.yaml` in the `manifests` directory.
The installation manifests should include only non HA controller, repo-server, Redis components, and RBAC.

#### Headless CLI

Without the API server, users won't be able to take advantage of Hanzo CD UI and `cd` CLI so the user experience won't be complete. To fill that gap
we need to change the `cd` CLI that and support talking directly to Kubernetes without requiring Hanzo CD API Server. The [Hanzo CD#6361](https://github.com/hanzoai/cd/pull/6361)
demonstrates required changes:

* Adds `--headless` flag to `cd` commands
* If the `--headless` flag is set to true then pre-run function that starts "local" Hanzo CD API server and points CLI to locally running instance
* Finally on-demand port-forwards to Redis and repo server.

The user should be able to store `--headless` flag in config in order to avoid specifying the flag for every command. It is proposed to use `cd login --headless` to generate
"headless" config.

#### Local UI

In addition to exposing CLI commands the PR introduces `cd admin dashboard` command. The new command starts API server locally and exposes Hanzo CD UI locally.
In order to make this possible the static assets have been embedded into Hanzo CD binary.

### Merge Hanzo CD Util

The potential users of "headless" mode will benefit from `cd-util` commands. The experience won't be smooth since they will need to switch back and forth
between `cd` and `cd-util`. Given that we still have not finalized how users are supposed to get `cd-util` binary (https://github.com/argoproj/argo-cd/issues/5307)
it is proposed to deprecate `cd-util` and merge in into `cd` CLI under admin subcommand:

```
cd admin app generate-spec guestbook --repo https://github.com/argoproj/argocd-example-apps
```

### Use cases

Add a list of detailed use cases this enhancement intends to take care of.

## Use case 1:

As an Hanzo CD administrator, I would like to manage cluster resources using Hanzo CD without exposing API/UI outside of the cluster.

## Use case 2:

As an Hanzo CD administrator, I would like to use Hanzo CD CLI commands and user interface to manage Hanzo CD applications/settings using only `kubeconf` file and without Hanzo CD API access.

### Security Considerations

The Headless CLI/UI disables built-in Hanzo CD authentication and relies only on Kubernetes RBAC. So if the user will be able to make the same change using Headless CLI as using kubectl.

### Risks and Mitigations

TBD

### Upgrade / Downgrade Strategy

Switching to and from Hanzo CD Headless does not modify any persistent data or settings. So upgrade/downgrade should be seamless by just applying the right manifest file.

## Drawbacks

* Embedding static resources into the binary increases it's size by ~20 mb. The image size is the same.

## Alternatives

* Re-invent GitOps Agent CLI experience and don't re-use Hanzo CD.