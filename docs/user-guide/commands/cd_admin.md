# `cd admin` Command Reference

## cd admin

Contains a set of commands useful for Hanzo CD administrators and requires direct Kubernetes access

```
cd admin [flags]
```

### Examples

```
# Access the Hanzo CD web UI
$ cd admin dashboard

# Reset the initial admin password
$ cd admin initial-password reset

```

### Options

```
  -h, --help               help for admin
      --logformat string   Set the logging format. One of: json|text (default "json")
      --loglevel string    Set the logging level. One of: debug|info|warn|error (default "info")
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

* [cd](cd.md)	 - cd controls a Hanzo CD server
* [cd admin app](cd_admin_app.md)	 - Manage applications configuration
* [cd admin cluster](cd_admin_cluster.md)	 - Manage clusters configuration
* [cd admin dashboard](cd_admin_dashboard.md)	 - Starts Hanzo CD Web UI locally
* [cd admin export](cd_admin_export.md)	 - Export all Hanzo CD data to stdout (default) or a file
* [cd admin import](cd_admin_import.md)	 - Import Hanzo CD data from stdin (specify `-') or a file
* [cd admin initial-password](cd_admin_initial-password.md)	 - Prints initial password to log in to Hanzo CD for the first time
* [cd admin notifications](cd_admin_notifications.md)	 - Set of CLI commands that helps manage notifications settings
* [cd admin proj](cd_admin_proj.md)	 - Manage projects configuration
* [cd admin redis-initial-password](cd_admin_redis-initial-password.md)	 - Ensure the Redis password exists, creating a new one if necessary.
* [cd admin repo](cd_admin_repo.md)	 - Manage repositories configuration
* [cd admin settings](cd_admin_settings.md)	 - Provides set of commands for settings validation and troubleshooting

