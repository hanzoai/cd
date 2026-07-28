# `cd repo add` Command Reference

## cd repo add

Add git, oci or helm repository connection parameters

```
cd repo add REPOURL [flags]
```

### Examples

```
  # Add a Git repository via SSH using a private key for authentication, ignoring the server's host key:
  cd repo add git@git.example.com:repos/repo --insecure-ignore-host-key --ssh-private-key-path ~/id_rsa

  # Add a Git repository via SSH on a non-default port - need to use ssh:// style URLs here
  cd repo add ssh://git@git.example.com:2222/repos/repo --ssh-private-key-path ~/id_rsa

  # Add a Git repository via SSH using socks5 proxy with no proxy credentials
  cd repo add ssh://git@github.com/argoproj/argocd-example-apps --ssh-private-key-path ~/id_rsa --proxy socks5://your.proxy.server.ip:1080

  # Add a Git repository via SSH using socks5 proxy with proxy credentials
  cd repo add ssh://git@github.com/argoproj/argocd-example-apps --ssh-private-key-path ~/id_rsa --proxy socks5://username:password@your.proxy.server.ip:1080

  # Add a private Git repository via HTTPS using username/password and TLS client certificates:
  cd repo add https://git.example.com/repos/repo --username git --password secret --tls-client-cert-path ~/mycert.crt --tls-client-cert-key-path ~/mycert.key

  # Add a private Git repository via HTTPS using username/password without verifying the server's TLS certificate
  cd repo add https://git.example.com/repos/repo --username git --password secret --insecure-skip-server-verification

  # Add a public Helm repository named 'stable' via HTTPS
  cd repo add https://charts.helm.sh/stable --type helm --name stable  

  # Add a private Helm repository named 'stable' via HTTPS
  cd repo add https://charts.helm.sh/stable --type helm --name stable --username test --password test

  # Add a private Helm OCI-based repository named 'stable' via HTTPS
  cd repo add helm-oci-registry.cn-zhangjiakou.cr.aliyuncs.com --type helm --name stable --enable-oci --username test --password test
  
  # Add a private HTTPS OCI repository named 'stable'
  cd repo add oci://helm-oci-registry.cn-zhangjiakou.cr.aliyuncs.com --type oci --name stable --username test --password test
  
  # Add a private OCI repository named 'stable' without verifying the server's TLS certificate
  cd repo add oci://helm-oci-registry.cn-zhangjiakou.cr.aliyuncs.com --type oci --name stable --username test --password test --insecure-skip-server-verification
  
  # Add a private HTTP OCI repository named 'stable'
  cd repo add oci://helm-oci-registry.cn-zhangjiakou.cr.aliyuncs.com --type oci --name stable --username test --password test --insecure-oci-force-http

  # Add a private Git repository on GitHub.com via GitHub App. github-app-installation-id is optional, if not provided, the installation id will be fetched from the GitHub API.
  cd repo add https://git.example.com/repos/repo --github-app-id 1 --github-app-installation-id 2 --github-app-private-key-path test.private-key.pem

  # Add a private Git repository on GitHub Enterprise via GitHub App. github-app-installation-id is optional, if not provided, the installation id will be fetched from the GitHub API.
  cd repo add https://ghe.example.com/repos/repo --github-app-id 1 --github-app-installation-id 2 --github-app-private-key-path test.private-key.pem --github-app-enterprise-base-url https://ghe.example.com/api/v3

  # Add a private Git repository on Google Cloud Sources via GCP service account credentials
  cd repo add https://source.developers.google.com/p/my-google-cloud-project/r/my-repo --gcp-service-account-key-path service-account-key.json

  # Add a private Git repository on Azure Devops via Azure Service Principal credentials
  cd repo add https://dev.azure.com/my-devops-organization/my-devops-project/_git/my-devops-repo --azure-service-principal-client-id 12345678-1234-1234-1234-123456789012 --azure-service-principal-client-secret test --azure-service-principal-tenant-id 12345678-1234-1234-1234-123456789012

  # Add a private Git repository on Azure Devops via Azure Service Principal credentials when not using default Azure public cloud
  cd repo add https://dev.azure.com/my-devops-organization/my-devops-project/_git/my-devops-repo --azure-service-principal-client-id 12345678-1234-1234-1234-123456789012 --azure-service-principal-client-secret test --azure-service-principal-tenant-id 12345678-1234-1234-1234-123456789012 --azure-active-directory-endpoint https://login.microsoftonline.de

```

### Options

```
      --azure-active-directory-endpoint string         Active Directory endpoint when not using default Azure public cloud (e.g. https://login.microsoftonline.de)
      --azure-service-principal-client-id string       client id of the Azure Service Principal
      --azure-service-principal-client-secret string   client secret of the Azure Service Principal
      --azure-service-principal-tenant-id string       tenant id of the Azure Service Principal
      --bearer-token string                            bearer token to the Git BitBucket Data Center repository
      --depth int                                      Specify a custom depth for git clone operations. Unless specified, a full clone is performed using the depth of 0
      --enable-lfs                                     enable git-lfs (Large File Support) on this repository
      --enable-oci                                     enable helm-oci (Helm OCI-Based Repository) (only valid for helm type repositories)
      --force-http-basic-auth                          whether to force use of basic auth when connecting repository via HTTP
      --gcp-service-account-key-path string            service account key for the Google Cloud Platform
      --github-app-enterprise-base-url string          base url to use when using GitHub Enterprise (e.g. https://ghe.example.com/api/v3
      --github-app-id int                              id of the GitHub Application
      --github-app-installation-id int                 installation id of the GitHub Application (optional, will be auto-discovered if not provided)
      --github-app-private-key-path string             private key of the GitHub Application
  -h, --help                                           help for add
      --insecure-ignore-host-key                       disables SSH strict host key checking (deprecated, use --insecure-skip-server-verification instead)
      --insecure-oci-force-http                        Use http when accessing an OCI repository
      --insecure-skip-server-verification              disables server certificate and host key checks
      --name string                                    name of the repository, mandatory for repositories of type helm
      --no-proxy string                                don't access these targets via proxy
      --password string                                password to the repository
      --project string                                 project of the repository
      --proxy string                                   use proxy to access repository
      --ssh-private-key-path string                    path to the private ssh key (e.g. ~/.ssh/id_rsa)
      --tls-client-cert-key-path string                path to the TLS client cert's key (must be PEM format)
      --tls-client-cert-path string                    path to the TLS client cert (must be PEM format)
      --type string                                    type of the repository, "git", "oci" or "helm" (default "git")
      --upsert                                         Override an existing repository with the same name even if the spec differs
      --use-azure-workload-identity                    whether to use azure workload identity for authentication
      --username string                                username to the repository
      --webhook-manifest-cache-warm-disabled           disable manifest cache warming during webhook processing for this repository (recommended for large monorepos with plain YAML manifests)
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

* [cd repo](cd_repo.md)	 - Manage repository connection parameters

