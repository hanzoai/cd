# Environment Variables

The following environment variables can be used with `argocd` CLI:

| Environment Variable                 | Description                                                                                                                                                                                               |
| ------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `CD_SERVER`                      | the address of the Argo CD server without `https://` prefix <br> (instead of specifying `--server` for every command) <br> eg. `CD_SERVER=argocd.example.com` if served through an ingress with DNS   |
| `CD_AUTH_TOKEN`                  | the Argo CD `apiKey` for your Argo CD user to be able to authenticate                                                                                                                                     |
| `CD_OPTS`                        | command-line options to pass to `argocd` CLI <br> eg. `CD_OPTS="--grpc-web"`                                                                                                                          |
| `CD_CONFIG_DIR`                  | sets the directory hosting `argocd` CLI config, e.g., `~/.config/argocd/config`. (if ENV var not provided, defaults to `$XDG_CONFIG_HOME/argocd`, or `~/.config/argocd`, or if exists legacy `~/.argocd`) |
| `CD_SERVER_NAME`                 | the Argo CD API Server name (default "argocd-server")                                                                                                                                                     |
| `CD_REPO_SERVER_NAME`            | the Argo CD Repository Server name (default "argocd-repo-server")                                                                                                                                         |
| `CD_APPLICATION_CONTROLLER_NAME` | the Argo CD Application Controller name (default "argocd-application-controller")                                                                                                                         |
| `CD_KV_NAME`                  | the Argo CD Redis name (default "argocd-redis")                                                                                                                                                           |
| `CD_KV_HAPROXY_NAME`          | the Argo CD Redis HA Proxy name (default "argocd-redis-ha-haproxy")                                                                                                                                       |
| `CD_KV_KEY_PREFIX`            | the Argo CD Redis keys prefix (default "")
|
| `CD_GRPC_KEEP_ALIVE_MIN`         | defines the GRPCKeepAliveEnforcementMinimum, used in the grpc.KeepaliveEnforcementPolicy. Expects a "Duration" format (default `10s`).                                                                    |
