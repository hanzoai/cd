# `cd proj windows add` Command Reference

## cd proj windows add

Add a sync window to a project

```
cd proj windows add PROJECT [flags]
```

### Examples

```

#Add a 1 hour allow sync window
cd proj windows add PROJECT \
    --kind allow \
    --schedule "0 22 * * *" \
    --duration 1h \
    --applications "*" \
    --description "Ticket 123"

#Add a deny sync window with the ability to manually sync and sync overrun.
cd proj windows add PROJECT \
    --kind deny \
    --schedule "30 10 * * *" \
    --duration 30m \
    --applications "prod-\\*,website" \
    --namespaces "default,\\*-prod" \
    --clusters "prod,staging" \
    --manual-sync \
    --sync-overrun \
    --description "Ticket 123"
```

### Options

```
      --applications strings   Applications that the schedule will be applied to. Comma separated, wildcards supported (e.g. --applications prod-\*,website)
      --clusters strings       Clusters that the schedule will be applied to. Comma separated, wildcards supported (e.g. --clusters prod,staging)
      --description string     Sync window description
      --duration string        Sync window duration. (e.g. --duration 1h)
  -h, --help                   help for add
  -k, --kind string            Sync window kind, either allow or deny
      --manual-sync            Allow manual syncs for both deny and allow windows
      --namespaces strings     Namespaces that the schedule will be applied to. Comma separated, wildcards supported (e.g. --namespaces default,\*-prod)
      --schedule string        Sync window schedule in cron format. (e.g. --schedule "0 22 * * *")
      --sync-overrun           Allow syncs to continue: for deny windows, syncs that started before the window; for allow windows, syncs that started during the window
      --time-zone string       Time zone of the sync window (default "UTC")
      --use-and-operator       Use AND operator for matching applications, namespaces and clusters instead of the default OR operator
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

* [cd proj windows](cd_proj_windows.md)	 - Manage a project's sync windows

