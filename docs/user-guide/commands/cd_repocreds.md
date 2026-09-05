# `cd repocreds` Command Reference

## cd repocreds

Manage credential templates for repositories

```
cd repocreds [flags]
```

### Examples

```
  # Add credentials with user/pass authentication to use for all repositories under the specified URL
  cd repocreds add URL --username USERNAME --password PASSWORD
  
  # List all the configured repository credentials
  cd repocreds list
  
  # Remove credentials for the repositories with specified URL
  cd repocreds rm URL
```

### Options

```
      --cluster string             The name of the kubeconfig cluster to use
      --context string             The name of the kubeconfig context to use
  -h, --help                       help for repocreds
      --insecure-skip-tls-verify   If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string          Path to a kube config. Only required if out-of-cluster
  -n, --namespace string           If present, the namespace scope for this CLI request
      --password string            Password for basic authentication to the API server
      --proxy-url string           If provided, this URL will be used to connect via proxy
      --request-timeout string     The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
      --token string               Bearer token for authentication to the API server
      --user string                The name of the kubeconfig user to use
      --username string            Username for basic authentication to the API server
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

* [cd](cd.md)	 - cd controls a Hanzo CD server
* [cd repocreds add](cd_repocreds_add.md)	 - Add git repository connection parameters
* [cd repocreds list](cd_repocreds_list.md)	 - List configured repository credentials
* [cd repocreds rm](cd_repocreds_rm.md)	 - Remove repository credentials

