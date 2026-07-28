# `cd admin proj generate-spec` Command Reference

## cd admin proj generate-spec

Generate declarative config for a project

```
cd admin proj generate-spec PROJECT [flags]
```

### Examples

```
  # Generate a YAML configuration for a project named "myproject"
  cd admin proj generate-spec myproject
  
  # Generate a JSON configuration for a project named "anotherproject" and specify an output file
  cd admin proj generate-spec anotherproject --output json --file config.json
  
  # Generate a YAML configuration for a project named "someproject" and write it back to the input file
  cd admin proj generate-spec someproject --inline
```

### Options

```
      --allow-cluster-resource stringArray      List of allowed cluster level resources, optionally with group and name (e.g. ClusterRole, apiextensions.k8s.io/CustomResourceDefinition, /Namespace/team1-*)
      --allow-namespaced-resource stringArray   List of allowed namespaced resources
      --deny-cluster-resource stringArray       List of denied cluster level resources, optionally with group and name (e.g. ClusterRole, apiextensions.k8s.io/CustomResourceDefinition, /Namespace/kube-*)
      --deny-namespaced-resource stringArray    List of denied namespaced resources
      --description string                      Project description
  -d, --dest stringArray                        Permitted destination server and namespace (e.g. https://192.168.99.100:8443,default)
      --dest-service-accounts stringArray       Destination server, namespace and target service account (e.g. https://192.168.99.100:8443,default,default-sa)
  -f, --file string                             Filename or URL to Kubernetes manifests for the project
  -h, --help                                    help for generate-spec
  -i, --inline                                  If set then generated resource is written back to the file specified in --file flag
      --orphaned-resources                      Enables orphaned resources monitoring
      --orphaned-resources-warn                 Specifies if applications should have a warning condition when orphaned resources detected
  -o, --output string                           Output format. One of: json|yaml (default "yaml")
      --signature-keys strings                  GnuPG public key IDs for commit signature verification
      --source-namespaces strings               List of source namespaces for applications
  -s, --src stringArray                         Permitted source repository URL
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

* [cd admin proj](cd_admin_proj.md)	 - Manage projects configuration

