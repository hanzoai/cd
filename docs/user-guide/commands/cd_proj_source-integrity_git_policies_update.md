# `cd proj source-integrity git policies update` Command Reference

## cd proj source-integrity git policies update

Update a git source integrity policy

```
cd proj source-integrity git policies update PROJECT POLICY_ID [flags]
```

### Examples

```
  # Update policy at index to set specific repo URLs, removing the old ones
  cd proj source-integrity git policies update PROJECT POLICY_ID \
  --repo-url 'https://github.com/foo/*'
  
  # Update policy at index to add and remove repo URLs
  cd proj source-integrity git policies update PROJECT POLICY_ID \
  --add-repo-url 'https://github.com/bar/*' \
  --delete-repo-url 'https://github.com/foo/*'
  
  # Update policy GPG mode and keys
  cd proj source-integrity git policies update PROJECT POLICY_ID \
  --gpg-mode strict \
  --add-gpg-key D56C4FCA57A46444
```

### Options

```
      --add-gpg-key strings       Add GPG key ID
      --add-repo-url strings      Add repository URL pattern
      --delete-gpg-key strings    Delete GPG key ID
      --delete-repo-url strings   Delete repository URL pattern
      --gpg-key strings           Set GPG key ID (replaces existing)
      --gpg-mode string           Set GPG verification mode (strict, head, or none)
  -h, --help                      help for update
      --repo-url strings          Set repository URL pattern (replaces existing)
  -y, --yes                       Skip explicit confirmation
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

* [cd proj source-integrity git policies](cd_proj_source-integrity_git_policies.md)	 - Manage git source integrity policies

