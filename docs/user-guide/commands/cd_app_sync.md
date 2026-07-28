# `cd app sync` Command Reference

## cd app sync

Sync an application to its target state

```
cd app sync [APPNAME... | -l selector | --project project-name] [flags]
```

### Examples

```
  # Sync an app
  cd app sync my-app

  # Sync multiples apps
  cd app sync my-app other-app

  # Sync apps by label, in this example we sync apps that are children of another app (aka app-of-apps)
  cd app sync -l app.kubernetes.io/instance=my-app
  cd app sync -l app.kubernetes.io/instance!=my-app
  cd app sync -l app.kubernetes.io/instance
  cd app sync -l '!app.kubernetes.io/instance'
  cd app sync -l 'app.kubernetes.io/instance notin (my-app,other-app)'

  # Sync a multi-source application for specific revision of specific sources
  cd app sync my-app --revisions 0.0.1 --source-positions 1 --revisions 0.0.2 --source-positions 2
  cd app sync my-app --revisions 0.0.1 --source-names my-chart --revisions 0.0.2 --source-names my-values

  # Sync a specific resource
  # Resource should be formatted as GROUP:KIND:NAME. If no GROUP is specified then :KIND:NAME
  cd app sync my-app --resource :Service:my-service
  cd app sync my-app --resource apps.hanzo.ai:Rollout:my-rollout
  cd app sync my-app --resource '!apps:Deployment:my-service'
  cd app sync my-app --resource apps:Deployment:my-service --resource :Service:my-service
  cd app sync my-app --resource '!*:Service:*'
  # Specify namespace if the application has resources with the same name in different namespaces
  cd app sync my-app --resource apps.hanzo.ai:Rollout:my-namespace/my-rollout
```

### Options

```
  -N, --app-namespace string                              Only sync an application in namespace
      --apply-out-of-sync-only                            Sync only out-of-sync resources
      --assumeYes                                         Assume yes as answer for all user queries or prompts
      --async                                             Do not wait for application to sync before continuing
      --dry-run                                           Preview apply without affecting cluster
      --force                                             Use a force apply
  -h, --help                                              help for sync
      --ignore-normalizer-jq-execution-timeout duration   Set ignore normalizer JQ execution timeout (default 1s)
      --info stringArray                                  A list of key-value pairs during sync process. These infos will be persisted in app.
      --label stringArray                                 Sync only specific resources with a label. This option may be specified repeatedly.
      --local string                                      Path to a local directory. When this flag is present no git queries will be made
      --local-repo-root string                            Path to the repository root. Used together with --local allows setting the repository root (default "/")
  -o, --output string                                     Output format. One of: json|yaml|wide|tree|tree=detailed (default "wide")
      --preview-changes                                   Preview difference against the target and live state before syncing app and wait for user confirmation
      --project stringArray                               Sync apps that belong to the specified projects. This option may be specified repeatedly.
      --prune                                             Allow deleting unexpected resources
      --replace                                           Use a kubectl create/replace instead apply
      --resource stringArray                              Sync only specific resources as GROUP:KIND:NAME or !GROUP:KIND:NAME. Fields may be blank and '*' can be used. This option may be specified repeatedly
      --retry-backoff-duration duration                   Retry backoff base duration. Input needs to be a duration (e.g. 2m, 1h) (default 5s)
      --retry-backoff-factor int                          Factor multiplies the base duration after each failed retry (default 2)
      --retry-backoff-max-duration duration               Max retry backoff duration. Input needs to be a duration (e.g. 2m, 1h) (default 3m0s)
      --retry-limit int                                   Max number of allowed sync retries
      --retry-refresh                                     Indicates if the latest revision should be used on retry instead of the initial one
      --revision string                                   Sync to a specific revision. Preserves parameter overrides
      --revisions stringArray                             Show manifests at specific revisions for source position in source-positions
  -l, --selector string                                   Sync apps that match this label. Supports '=', '==', '!=', in, notin, exists & not exists. Matching apps must satisfy all of the specified label constraints.
      --server-side                                       Use server-side apply while syncing the application
      --server-side-diff-concurrency int                  Max concurrent batches for server-side diff. -1 = unlimited, 1 = sequential, 2+ = concurrent (0 = invalid) (default -1)
      --server-side-diff-max-batch-kb int                 Max batch size in KB for server-side diff. Smaller values are safer for proxies (default 250)
      --source-names stringArray                          List of source names. Default is an empty array.
      --source-positions int64Slice                       List of source positions. Default is empty array. Counting start at 1. (default [])
      --strategy string                                   Sync strategy (one of: apply|hook)
      --timeout uint                                      Time out after this many seconds
```

### Options inherited from parent commands

```
      --cd-context string           The name of the Hanzo CD server context to use
      --auth-token string               Authentication token; set this or the CD_AUTH_TOKEN environment variable
      --client-crt string               Client certificate file
      --client-crt-key string           Client certificate key file
      --config string                   Path to Hanzo CD config (default "/home/user/.config/cd/config")
      --controller-name string          Name of the Hanzo CD Application controller; set this or the CD_APPLICATION_CONTROLLER_NAME environment variable when the controller's name label differs from the default, for example when installing via the Helm chart (default "cd-application-controller")
      --core                            If set to true then CLI talks directly to Kubernetes instead of talking to Hanzo CD API server
      --grpc-web                        Enables gRPC-web protocol. Useful if Hanzo CD server is behind proxy which does not support HTTP2.
      --grpc-web-root-path string       Enables gRPC-web protocol. Useful if Hanzo CD server is behind proxy which does not support HTTP2. Set web root.
  -H, --header strings                  Sets additional header to all requests made by Hanzo CD CLI. (Can be repeated multiple times to add multiple headers, also supports comma separated headers)
      --http-retry-max int              Maximum number of retries to establish http connection to Hanzo CD server
      --insecure                        Skip server certificate and domain verification
      --kube-context string             Directs the command to the given kube-context
      --logformat string                Set the logging format. One of: json|text (default "json")
      --loglevel string                 Set the logging level. One of: debug|info|warn|error (default "info")
      --plaintext                       Disable TLS
      --port-forward                    Connect to a random cd-server port using port forwarding
      --port-forward-namespace string   Namespace name which should be used for port forwarding
      --prompts-enabled                 Force optional interactive prompts to be enabled or disabled, overriding local configuration. If not specified, the local configuration value will be used, which is false by default.
      --redis-compress string           Enable this if the application controller is configured with redis compression enabled. (possible values: gzip, none) (default "gzip")
      --redis-haproxy-name string       Name of the Redis HA Proxy; set this or the CD_KV_HAPROXY_NAME environment variable when the HA Proxy's name label differs from the default, for example when installing via the Helm chart (default "cd-redis-ha-haproxy")
      --redis-name string               Name of the Redis deployment; set this or the CD_KV_NAME environment variable when the Redis's name label differs from the default, for example when installing via the Helm chart (default "cd-redis")
      --repo-server-name string         Name of the Hanzo CD Repo server; set this or the CD_REPO_SERVER_NAME environment variable when the server's name label differs from the default, for example when installing via the Helm chart (default "cd-repo-server")
      --server string                   Hanzo CD server address
      --server-crt string               Server certificate file
      --server-name string              Name of the Hanzo CD API server; set this or the CD_SERVER_NAME environment variable when the server's name label differs from the default, for example when installing via the Helm chart (default "cd-server")
```

### SEE ALSO

* [cd app](cd_app.md)	 - Manage applications

