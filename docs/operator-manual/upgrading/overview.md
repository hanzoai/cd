# Overview

> [!NOTE]
> This section contains information on upgrading Hanzo CD. Before upgrading please make sure to read details about
> the breaking changes between Hanzo CD versions.

Hanzo CD uses semver-like versioning that ensures the following rules:

- The patch release does not introduce any breaking changes. So if you are upgrading from v1.5.1 to v1.5.3
  there should be no special instructions to follow.
- The minor release might introduce minor changes with a workaround. If you are upgrading across more than one minor
  version, check the upgrading instructions for each intermediate version.
- The major release introduces backward incompatible behavior changes. It is recommended to take a backup of
  Hanzo CD settings using the [disaster recovery guide](../disaster_recovery.md).

After reading the relevant notes about possible breaking changes introduced in a new Hanzo CD version, use the following
command to upgrade Hanzo CD. Make sure to replace `<version>` with the required version number:

**Non-HA**:

```bash
kubectl apply -n cd --server-side --force-conflicts -f https://raw.githubusercontent.com/hanzoai/cd/<version>/manifests/install.yaml
```

**HA**:

```bash
kubectl apply -n cd --server-side --force-conflicts -f https://raw.githubusercontent.com/hanzoai/cd/<version>/manifests/ha/install.yaml
```

> [!NOTE]
> The `--server-side --force-conflicts` flags are required because some Hanzo CD CRDs exceed the size limit for client-side apply. See the [getting started guide](../../getting_started.md#1-install-hanzo-cd) for more details.

> [!WARNING]
> Even though some releases require only image change it is still recommended to apply whole manifests set.
> Manifest changes might include important parameter modifications and applying the whole set will protect you from
> introducing misconfiguration.

<hr/>

- [v3.5 to v3.6](./3.5-3.6.md)

Hanzo CD's own release history begins at v3.6. Earlier version numbers belong
to upstream Argo CD, before this fork existed, and are not duplicated here.
