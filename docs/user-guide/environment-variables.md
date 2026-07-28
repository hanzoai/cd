# Environment Variables

The following environment variables can be used with `cd` CLI:

| Environment Variable                 | Description                                                                                                                                                                                               |
| ------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `CD_SERVER`                      | the address of the Hanzo CD server without `https://` prefix <br> (instead of specifying `--server` for every command) <br> eg. `CD_SERVER=cd.example.com` if served through an ingress with DNS   |
| `CD_AUTH_TOKEN`                  | the Hanzo CD `apiKey` for your Hanzo CD user to be able to authenticate                                                                                                                                     |
| `CD_OPTS`                        | command-line options to pass to `cd` CLI <br> eg. `CD_OPTS="--grpc-web"`                                                                                                                          |
| `CD_CONFIG_DIR`                  | sets the directory hosting `cd` CLI config, e.g., `~/.config/cd/config`. (if ENV var not provided, defaults to `$XDG_CONFIG_HOME/cd`, or `~/.config/cd`, or if exists legacy `~/.cd`) |
| `CD_SERVER_NAME`                 | the Hanzo CD API Server name (default "cd-server")                                                                                                                                                     |
| `CD_REPO_SERVER_NAME`            | the Hanzo CD Repository Server name (default "cd-repo-server")                                                                                                                                         |
| `CD_APPLICATION_CONTROLLER_NAME` | the Hanzo CD Application Controller name (default "cd-application-controller")                                                                                                                         |
| `CD_KV_NAME`                  | the Hanzo CD Redis name (default "cd-redis")                                                                                                                                                           |
| `CD_KV_HAPROXY_NAME`          | the Hanzo CD Redis HA Proxy name (default "cd-redis-ha-haproxy")                                                                                                                                       |
| `CD_KV_KEY_PREFIX`            | the Hanzo CD Redis keys prefix (default "")
|
| `CD_GRPC_KEEP_ALIVE_MIN`         | defines the GRPCKeepAliveEnforcementMinimum, used in the grpc.KeepaliveEnforcementPolicy. Expects a "Duration" format (default `10s`).                                                                    |
