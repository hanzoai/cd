# `cd appset` Command Reference

## cd appset

Manage ApplicationSets

```
cd appset [flags]
```

### Examples

```
  # Get an ApplicationSet.
  cd appset get APPSETNAME
  
  # List all the ApplicationSets
  cd appset list
  
  # Create an ApplicationSet from a YAML stored in a file or at given URL
  cd appset create <filename or URL> (<filename or URL>...)
  
  # Delete an ApplicationSet
  cd appset delete APPSETNAME (APPSETNAME...)
  
  # Namespace precedence for --appset-namespace (-N):
  # - get/delete: if the argument is namespace/name, that namespace wins; -N is ignored.
  # - create/generate: metadata.namespace in the YAML wins when set; -N applies only when the manifest omits namespace.
```

### Options

```
      --cluster string             The name of the kubeconfig cluster to use
      --context string             The name of the kubeconfig context to use
  -h, --help                       help for appset
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

* [cd](cd.md)	 - cd controls an Hanzo CD server
* [cd appset create](cd_appset_create.md)	 - Create one or more ApplicationSets
* [cd appset delete](cd_appset_delete.md)	 - Delete one or more ApplicationSets
* [cd appset generate](cd_appset_generate.md)	 - Generate apps of ApplicationSet rendered templates
* [cd appset get](cd_appset_get.md)	 - Get ApplicationSet details
* [cd appset list](cd_appset_list.md)	 - List ApplicationSets

