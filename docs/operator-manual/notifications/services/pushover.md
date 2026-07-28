# Pushover

1. Create an app at [pushover.net](https://pushover.net/apps/build).
2. Store the API key in `<secret-name>` Secret and define the secret name in `cd-notifications-cm` ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cd-notifications-cm
data:
  service.pushover: |
    token: $pushover-token
```

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: <secret-name>
stringData:
  pushover-token: avtc41pn13asmra6zaiyf7dh6cgx97
```

3. Add your user key to your Application resource:

```yaml
apiVersion: apps.hanzo.ai/v1alpha1
kind: Application
metadata:
  annotations:
    notifications.apps.hanzo.ai/subscribe.on-sync-succeeded.pushover: uumy8u4owy7bgkapp6mc5mvhfsvpcd
```