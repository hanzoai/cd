# `cd proj role` Command Reference

## cd proj role

Manage a project's roles

```
cd proj role [flags]
```

### Options

```
  -h, --help   help for role
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
* [cd proj role add-group](cd_proj_role_add-group.md)	 - Add a group claim to a project role
* [cd proj role add-policy](cd_proj_role_add-policy.md)	 - Add a policy to a project role
* [cd proj role create](cd_proj_role_create.md)	 - Create a project role
* [cd proj role create-token](cd_proj_role_create-token.md)	 - Create a project token
* [cd proj role delete](cd_proj_role_delete.md)	 - Delete a project role
* [cd proj role delete-token](cd_proj_role_delete-token.md)	 - Delete a project token
* [cd proj role get](cd_proj_role_get.md)	 - Get the details of a specific role
* [cd proj role list](cd_proj_role_list.md)	 - List all the roles in a project
* [cd proj role list-tokens](cd_proj_role_list-tokens.md)	 - List tokens for a given role.
* [cd proj role remove-group](cd_proj_role_remove-group.md)	 - Remove a group claim from a role within a project
* [cd proj role remove-policy](cd_proj_role_remove-policy.md)	 - Remove a policy from a role within a project

