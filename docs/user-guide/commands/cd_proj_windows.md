# `cd proj windows` Command Reference

## cd proj windows

Manage a project's sync windows

```
cd proj windows [flags]
```

### Examples

```

#Add a sync window to a project
cd proj windows add my-project \
--schedule "0 0 * * 1-5" \
--duration 3600 \
--prune

#Delete a sync window from a project
cd proj windows delete <project-name> <window-id>

#List project sync windows
cd proj windows list <project-name>
```

### Options

```
  -h, --help   help for windows
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

* [cd proj](cd_proj.md)	 - Manage projects
* [cd proj windows add](cd_proj_windows_add.md)	 - Add a sync window to a project
* [cd proj windows delete](cd_proj_windows_delete.md)	 - Delete a sync window from a project. Requires ID which can be found by running "cd proj windows list PROJECT"
* [cd proj windows disable-manual-sync](cd_proj_windows_disable-manual-sync.md)	 - Disable manual sync for a sync window
* [cd proj windows disable-sync-overrun](cd_proj_windows_disable-sync-overrun.md)	 - Disable sync overrun for a sync window
* [cd proj windows enable-manual-sync](cd_proj_windows_enable-manual-sync.md)	 - Enable manual sync for a sync window
* [cd proj windows enable-sync-overrun](cd_proj_windows_enable-sync-overrun.md)	 - Enable sync overrun for a sync window
* [cd proj windows list](cd_proj_windows_list.md)	 - List project sync windows
* [cd proj windows update](cd_proj_windows_update.md)	 - Update a project sync window

