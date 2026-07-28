# `cd app delete` Command Reference

## cd app delete

Delete an application

```
cd app delete APPNAME [flags]
```

### Examples

```
  # Delete an app
  cd app delete my-app

  # Delete multiple apps
  cd app delete my-app other-app

  # Delete apps by label
  cd app delete -l app.kubernetes.io/instance=my-app
  cd app delete -l app.kubernetes.io/instance!=my-app
  cd app delete -l app.kubernetes.io/instance
  cd app delete -l '!app.kubernetes.io/instance'
  cd app delete -l 'app.kubernetes.io/instance notin (my-app,other-app)'
```

### Options

```
  -N, --app-namespace string        Namespace where the application will be deleted from
      --cascade                     Perform a cascaded deletion of all application resources (default true)
  -h, --help                        help for delete
  -p, --propagation-policy string   Specify propagation policy for deletion of application's resources. One of: foreground|background (default "foreground")
  -l, --selector string             Delete all apps with matching label. Supports '=', '==', '!=', in, notin, exists & not exists. Matching apps must satisfy all of the specified label constraints.
      --wait                        Wait until deletion of the application(s) completes
  -y, --yes                         Turn off prompting to confirm cascaded deletion of application resources
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

