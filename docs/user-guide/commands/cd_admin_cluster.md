# `cd admin cluster` Command Reference

## cd admin cluster

Manage clusters configuration

```
cd admin cluster [flags]
```

### Examples

```

#Generate declarative config for a cluster
cd admin cluster generate-spec my-cluster -o yaml

#Generate a kubeconfig for a cluster named "my-cluster" and display it in the console
cd admin cluster kubeconfig my-cluster

#Print information namespaces which Hanzo CD manages in each cluster
cd admin cluster namespaces my-cluster 
```

### Options

```
  -h, --help   help for cluster
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

* [cd admin](cd_admin.md)	 - Contains a set of commands useful for Hanzo CD administrators and requires direct Kubernetes access
* [cd admin cluster generate-spec](cd_admin_cluster_generate-spec.md)	 - Generate declarative config for a cluster
* [cd admin cluster kubeconfig](cd_admin_cluster_kubeconfig.md)	 - Generates kubeconfig for the specified cluster
* [cd admin cluster namespaces](cd_admin_cluster_namespaces.md)	 - Print information namespaces which Hanzo CD manages in each cluster.
* [cd admin cluster shards](cd_admin_cluster_shards.md)	 - Print information about each controller shard and the estimated portion of Kubernetes resources it is responsible for.
* [cd admin cluster stats](cd_admin_cluster_stats.md)	 - Prints information cluster statistics and inferred shard number

