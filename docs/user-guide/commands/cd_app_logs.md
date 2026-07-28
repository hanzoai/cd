# `cd app logs` Command Reference

## cd app logs

Get logs of application pods

```
cd app logs APPNAME [flags]
```

### Examples

```
  # Get logs of pods associated with the application "my-app"
  cd app logs my-app
  
  # Get logs of pods associated with the application "my-app" in a specific resource group
  cd app logs my-app --group my-group
  
  # Get logs of pods associated with the application "my-app" in a specific resource kind
  cd app logs my-app --kind my-kind
  
  # Get logs of pods associated with the application "my-app" in a specific namespace
  cd app logs my-app --namespace my-namespace
  
  # Get logs of pods associated with the application "my-app" for a specific resource name
  cd app logs my-app --name my-resource
  
  # Stream logs in real-time for the application "my-app"
  cd app logs my-app -f
  
  # Get the last N lines of logs for the application "my-app"
  cd app logs my-app --tail 100
  
  # Get logs since a specified number of seconds ago
  cd app logs my-app --since-seconds 3600
  
  # Get logs until a specified time (format: "2023-10-10T15:30:00Z")
  cd app logs my-app --until-time "2023-10-10T15:30:00Z"
  
  # Filter logs to show only those containing a specific string
  cd app logs my-app --filter "error"
  
  # Filter logs to show only those containing a specific string and match case
  cd app logs my-app --filter "error" --match-case
  
  # Get logs for a specific container within the pods
  cd app logs my-app -c my-container
  
  # Get previously terminated container logs
  cd app logs my-app -p
```

### Options

```
  -N, --app-namespace string   Namespace of the application
  -c, --container string       Optional container name
      --filter string          Show logs contain this string
  -f, --follow                 Specify if the logs should be streamed
      --group string           Resource group
  -h, --help                   help for logs
      --kind string            Resource kind
  -m, --match-case             Specify if the filter should be case-sensitive
      --name string            Resource name
      --namespace string       Resource namespace
  -p, --previous               Specify if the previously terminated container logs should be returned
      --since-seconds int      A relative time in seconds before the current time from which to show logs
      --tail int               The number of lines from the end of the logs to show
      --until-time string      Show logs until this time
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

