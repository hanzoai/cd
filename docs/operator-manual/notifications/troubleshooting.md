`cd admin notifications` is a CLI command group that helps to configure the controller
settings and troubleshoot issues. Full command details are available in the [command reference](../../user-guide/commands/cd_admin_notifications.md).

## Global flags
The following global flags are available for all sub-commands:

* `--config-map` - path to the file containing `cd-notifications-cm` ConfigMap. If not specified
then the command loads `cd-notification-cm` ConfigMap using the local Kubernetes config file.
* `--secret` - path to the file containing `cd-notifications-secret` ConfigMap. If not
specified then the command loads `cd-notification-secret` Secret using the local Kubernetes config file.
Additionally, you can specify `:empty` to use empty secret with no notification service settings. 

**Examples:**

* Get a list of triggers configured in the local config map:

    ```bash
    cd admin notifications trigger get \
      --config-map ./cd-notifications-cm.yaml --secret :empty
    ```

* Trigger notification using in-cluster config map and secret:

    ```bash
    cd admin notifications template notify \
      app-sync-succeeded guestbook --recipient slack:cd admin notifications
    ```

## Kustomize

If you are managing `cd-notifications` config using Kustomize you can pipe whole `kustomize build` output
into stdin using `--config-map -` flag:

```bash
kustomize build ./cd-notifications | \
  cd-notifications \
  template notify app-sync-succeeded guestbook --recipient grafana:cd \
  --config-map -
```

## How to get it

### On your laptop

You can download the `cd` CLI from the GitHub [release](https://github.com/hanzoai/cd/releases)
attachments.

The binary is available in the `ghcr.io/hanzoai/cd` image. Use the `docker run` and volume mount
to execute binary on any platform. 

**Example:**

```bash
docker run --rm -it -w /src -v $(pwd):/src \
  ghcr.io/hanzoai/cd:<version> \
  /app/cd admin notifications trigger get \
  --config-map ./cd-notifications-cm.yaml --secret :empty
```

### In your cluster

SSH into the running `cd-notifications-controller` pod and use `kubectl exec` command to validate in-cluster
configuration.

**Example**
```bash
kubectl exec -it cd-notifications-controller-<pod-hash> \
  /usr/local/bin/cd admin notifications trigger get
```

## Commands

The following commands may help debug issues with notifications:

* [`cd admin notifications template get`](../../user-guide/commands/cd_admin_notifications_template_get.md)
* [`cd admin notifications template notify`](../../user-guide/commands/cd_admin_notifications_template_notify.md)
* [`cd admin notifications trigger get`](../../user-guide/commands/cd_admin_notifications_trigger_get.md)
* [`cd admin notifications trigger run`](../../user-guide/commands/cd_admin_notifications_trigger_run.md)

## Errors

{!docs/operator-manual/notifications/troubleshooting-errors.md!}
