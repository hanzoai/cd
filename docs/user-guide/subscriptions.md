The subscription to Hanzo CD application events can be defined using `notifications.apps.hanzo.ai/subscribe.<trigger>.<service>: <recipient>` annotation.
For example, the following annotation subscribes two Slack channels to notifications about every successful synchronization of the Hanzo CD application:

```yaml
apiVersion: apps.hanzo.ai/v1alpha1
kind: Application
metadata:
  annotations:
    notifications.apps.hanzo.ai/subscribe.on-sync-succeeded.slack: my-channel1;my-channel2
```

Annotation key consists of following parts:

* `on-sync-succeeded` - trigger name
* `slack` - notification service name
* `my-channel1;my-channel2` - a semicolon separated list of recipients

You can create subscriptions for all applications of the Hanzo CD project by adding the same annotation to AppProject CRD:

```yaml
apiVersion: apps.hanzo.ai/v1alpha1
kind: AppProject
metadata:
  annotations:
    notifications.apps.hanzo.ai/subscribe.on-sync-succeeded.slack: my-channel1;my-channel2
```

## Default Subscriptions

The subscriptions might be configured globally in the `cd-notifications-cm` ConfigMap using `subscriptions` field. The default subscriptions
are applied to all applications. The trigger and applications might be configured using the
`triggers` and `selector` fields:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cd-notifications-cm
data:
  # Contains centrally managed global application subscriptions
  subscriptions: |
    # subscription for on-sync-status-unknown trigger notifications
    - recipients:
      - slack:test2
      - email:test@gmail.com
      triggers:
      - on-sync-status-unknown
    # subscription restricted to applications with matching labels only
    - recipients:
      - slack:test3
      selector: test=true
      triggers:
      - on-sync-status-unknown
```

If you want to use webhook in subscriptions, you need to store the custom name to recipients.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cd-notifications-cm
data:
  service.webhook.<webhook-name>: |
    (snip)
  subscriptions: |
    - recipients:
      - <webhook-name>
      triggers:
      - on-sync-status-unknown
```
