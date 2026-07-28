# `cd app get` Command Reference

## cd app get

Get application details

```
cd app get APPNAME [flags]
```

### Examples

```
  # Get basic details about the application "my-app" in wide format
  cd app get my-app -o wide
  
  # Get detailed information about the application "my-app" in YAML format
  cd app get my-app -o yaml
  
  # Get details of the application "my-app" in JSON format
  cd get my-app -o json
  
  # Get application details and include information about the current operation
  cd app get my-app --show-operation
  
  # Show application parameters and overrides
  cd app get my-app --show-params
  
  # Show application parameters and overrides for a source at position 1 under spec.sources of app my-app
  cd app get my-app --show-params --source-position 1
  
  # Show application parameters and overrides for a source named "test"
  cd app get my-app --show-params --source-name test
  
  # Refresh application data when retrieving
  cd app get my-app --refresh
  
  # Perform a hard refresh, including refreshing application data and target manifests cache
  cd app get my-app --hard-refresh
  
  # Get application details and display them in a tree format
  cd app get my-app --output tree
  
  # Get application details and display them in a detailed tree format
  cd app get my-app --output tree=detailed
```

### Options

```
  -N, --app-namespace string   Only get application from namespace
      --hard-refresh           Refresh application data as well as target manifests cache
  -h, --help                   help for get
  -o, --output string          Output format. One of: json|yaml|wide|tree (default "wide")
      --refresh                Refresh application data when retrieving
      --show-operation         Show application operation
      --show-params            Show application parameters and overrides
      --source-name string     Name of the source from the list of sources of the app.
      --source-position int    Position of the source from the list of sources of the app. Counting starts at 1. (default -1)
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

