---
title: Hanzo CD CLI Plugin
authors:
  - "@christianh814"
  - "@alexmt"
  - "@nitishfy"
  
sponsors:
  - TBD
reviewers:
  - TBD
approvers:
  - TBD

creation-date: 2024-08-21
---

# Hanzo CD CLI Plugin Support

Support for `kubectl`-like plugins for the `cd` CLI.

## Summary

Enhance the Hanzo CD `cd` CLI client to support the ability to provide other CLI tools as "plugins" in a similar way that the Kubernetes `kubectl` CLI client provides. This will shift `cd` to be more of a "building block" or "interface" for interacting with Hanzo CD and plugins can be a means to develop more complex workflows while leaving the baseline cd use-case small. 

## Motivation

Currently, the `cd` CLI client is-getting/already-is pretty bloated. The solution has either been written-in the support for certain solutions (like ApplicationSets) or having a separate CLI tool in Argo Project Labs (like [Autopilot](https://github.com/hanzoai-labs/cd-autopilot) and [Vault Plugin](https://github.com/hanzoai-labs/cd-vault-plugin)).

Having a plugin system makes sense, as we add more and more features down the line (OCI, Hydrator, GitOps Promoter, etc). In this way, we can keep the `cd` CLI "lean" but also extensible and keep things "out of tree". As mentioned before, other tools that can benefit from this besides the Hanzo CD OCI cli is a tool like `cd-autopilot`. Also, there is a potential for other Argo Project tools to have plugins, like a Rollouts plugin or an `cd-image-updater` plugin for Hanzo CD.

The idea initially came up during a discussion about [adding CLI support for the upcoming OCI integration](https://cloud-native.slack.com/archives/C06Q17QJPJR/p1721227926706059?thread_ts=1720731234.357799&cid=C06Q17QJPJR).

## Goals

The goal is to provide a plugin mechanism without changing the current behavior of `cd`'s subcommands and options. 

## Non Goals

Make any guarantees for any public `cd` plugins provided by any third party. 

## Proposal

Similarly to how `kubectl` plugins are handled, The `cd` CLI tool will look in the end user's `$PATH` for any binaries that start with `cd-` and execute that binary. For example if I had a binary called `cd-mytool` in my `$PATH`, I could call it by running `cd mytool`. Support for tab completion should also be taken into account.

## Outstanding Questions

Things to consider:

* Should it act exactly like `kubectl` and just look in `$PATH` or be more stringent and have users store plugins in `~/.config/cd/plugins`? Similar to [Tekton plugins](https://tekton.dev/vault/cli-main/tkn-plugins/#location)
* Is there a way to integrate this with [Krew](https://krew.sigs.k8s.io/) for installing plugins?
* Should we let each plugin manage its own configuration settings or make plugins use `~/.config/cd/config` and provide a new field called `.pluginConfigs`? For example the `cd-mytool` plugin's config will be under `.pluginConfigs.mytool` Should we even care/have an opinion?
* Should we provide any guidelines to submit a plugin? Do we only "accept" plugins that are in hanzoai-labs?

