## Failed to parse new settings

### error converting YAML to JSON

YAML syntax is incorrect.

**incorrect:**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cd-notifications-cm
data:
  service.slack: |
    token: $slack-token
    icon: :rocket:
```

**correct:**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cd-notifications-cm
data:
  service.slack: |
    token: $slack-token
    icon: ":rocket:" # <- diff here
```

### service type 'xxxx' is not supported

Check the `cd-notifications` controller version. For example, the Teams integration support started in `v1.1.0`.

## Failed to notify recipient

### notification service 'xxxx' is not supported

You have not defined `xxxx` in `cd-notifications-cm` or parsing failed.

### GitHub.repoURL (\u003cno value\u003e) does not have a / using the configuration

Likely caused by an Application with [multiple sources](../../user-guide/multiple_sources.md):

```yaml
spec:
  sources:  # <- multiple sources
  - repoURL: https://github.com/exampleOrg/first.git
    path: sources/example
  - repoURL: https://github.com/exampleOrg/second.git
    targetRevision: "{{branch}}"
```

The standard notification template only supports a single source (`{{.app.spec.source.repoURL}}`). Use an index to specify the source in the array:

```yaml
template.example: |
  github:
    repoURLPath: "{{ (index .app.spec.sources 0).repoURL }}"
```

### Error message `POST https://api.github.com/repos/xxxx/yyyy/statuses/: 404 Not Found`

This case is similar to the previous one, you have multiple sources in the Application manifest. 
Default `revisionPath` template `{{.app.status.operationState.syncResult.revision}}` is for an Application with single source.

Multi-source applications report application statuses in an array:

```yaml
status:
  operationState:
    syncResult:
      revisions:
        - 38cfa22edf9148caabfecb288bfb47dc4352dfc6
        - 38cfa22edf9148caabfecb288bfb47dc4352dfc6
Quick fix for this is to use `index` function to get the first revision:
```yaml
template.example: |
  github:
    revisionPath: "{{index .app.status.operationState.syncResult.revisions 0}}"
```

## config referenced xxx, but key does not exist in secret

- If you are using a custom secret, check that the secret is in the same namespace
- You have added the label: `app.kubernetes.io/part-of: cd` to the secret
- You have tried restarting `cd-notifications` controller

### Example:
Secret:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cd-slackbot
  namespace: <the namespace where cd is installed>
  labels:
    app.kubernetes.io/part-of: cd
type: Opaque
data:
  slack-token: <base64encryptedtoken>
```
ConfigMap
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cd-notifications-cm
data:
  service.slack: |
    token: $cd-slackbot:slack-token
```
