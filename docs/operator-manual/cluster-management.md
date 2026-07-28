# Cluster Management

This guide is for operators looking to manage clusters on the CLI. If you want to use Kubernetes resources for this, check out [Declarative Setup](./declarative-setup.md#clusters).

Not all commands are described here, see the [cd cluster Command Reference](../user-guide/commands/cd_cluster.md) for all available commands.

## Adding a cluster

Run `cd cluster add context-name`.

If you're unsure about the context names, run `kubectl config get-contexts` to get them all listed.

This will connect to the cluster and install the necessary resources for Hanzo CD to connect to it.
Note that you will need privileged access to the cluster.

## Skipping cluster reconciliation

You can stop the controller from reconciling a cluster without removing it by annotating its secret:

```bash
kubectl -n cd annotate secret <cluster-secret-name> cd.hanzo.ai/skip-reconcile=true
```

The cluster will still appear in `cd cluster list` but the controller will skip reconciliation
for all apps targeting it. To resume, remove the annotation:

```bash
kubectl -n cd annotate secret <cluster-secret-name> cd.hanzo.ai/skip-reconcile-
```

See [Declarative Setup - Skipping Cluster Reconciliation](./declarative-setup.md#skipping-cluster-reconciliation) for details.

## Removing a cluster

Run `cd cluster rm context-name`.

This removes the cluster with the specified name.

> [!NOTE]
> **in-cluster cannot be removed**
>
> The `in-cluster` cluster cannot be removed with this. If you want to disable the `in-cluster` configuration, you need to update your `cd-cm` ConfigMap. Set [`cluster.inClusterEnabled`](./cd-cm-yaml.md) to `"false"`
