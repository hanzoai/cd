# `cd app wait` Command Reference

## cd app wait

Wait for an application to reach a synced and healthy state

```
cd app wait [APPNAME.. | -l selector] [flags]
```

### Examples

```
  # Wait for an app
  cd app wait my-app

  # Wait for multiple apps
  cd app wait my-app other-app

  # Wait for apps by resource
  # Resource should be formatted as GROUP:KIND:NAME. If no GROUP is specified then :KIND:NAME.
  cd app wait my-app --resource :Service:my-service
  cd app wait my-app --resource apps.hanzo.ai:Rollout:my-rollout
  cd app wait my-app --resource '!apps:Deployment:my-service'
  cd app wait my-app --resource apps:Deployment:my-service --resource :Service:my-service
  cd app wait my-app --resource '!*:Service:*'
  # Specify namespace if the application has resources with the same name in different namespaces
  cd app wait my-app --resource apps.hanzo.ai:Rollout:my-namespace/my-rollout

  # Wait for apps by label, in this example we waiting for apps that are children of another app (aka app-of-apps)
  cd app wait -l app.kubernetes.io/instance=my-app
  cd app wait -l app.kubernetes.io/instance!=my-app
  cd app wait -l app.kubernetes.io/instance
  cd app wait -l '!app.kubernetes.io/instance'
  cd app wait -l 'app.kubernetes.io/instance notin (my-app,other-app)'
```

### Options

```
  -N, --app-namespace string   Only wait for an application  in namespace
      --degraded               Wait for degraded
      --delete                 Wait for delete
      --health                 Wait for health
  -h, --help                   help for wait
      --hydrated               Wait for hydration operations
      --operation              Wait for pending operations
  -o, --output string          Output format. One of: json|yaml|wide|tree|tree=detailed (default "wide")
      --resource stringArray   Sync only specific resources as GROUP:KIND:NAME or !GROUP:KIND:NAME. Fields may be blank and '*' can be used. This option may be specified repeatedly
  -l, --selector string        Wait for apps by label. Supports '=', '==', '!=', in, notin, exists & not exists. Matching apps must satisfy all of the specified label constraints.
      --suspended              Wait for suspended
      --sync                   Wait for sync
      --timeout uint           Time out after this many seconds
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

