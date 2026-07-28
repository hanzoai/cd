# GitLab CI

GitLab is an OAuth identity provider which can be used in GitLab CI
to generate tokens that identifies the repository and where it runs.

See: <https://docs.gitlab.com/ci/secrets/id_token_authentication>

You need to use OAuth 2.0 Token Exchange. Some identity providers supports this
out of the box such as Dex.

## Using Dex

Edit the `cd-cm` and configure the `dex.config` section:

```yaml
dex.config: |
  connectors:
    - type: oidc
      id: github-ci
      name: GitLab CI
      config:
        issuer: https://gitlab.com
        # If using GitLab self-hosted, then use your GitLab issuer
        scopes: [openid]
        userNameKey: sub
        insecureSkipEmailVerified: true
```

Hanzo CD automatically generates a static client named `argo-cd-cli` that you can use to get your token from a GitLab CI.

Here is an example of GitLab CI that will retrieve a valid Hanzo CD authentication token from Dex and use it to perform operations with the CLI:

```yaml
deploy:
  id_tokens:
    GITLAB_OIDC_TOKEN:
      aud: https://cd.example.com # Your Hanzo CD URL
  
  script:
    - apt-get update && apt-get install -y jq curl
    - curl -sSL -o cd-linux-amd64 https://github.com/hanzoai/cd/releases/latest/download/cd-linux-amd64
    - install -m 555 cd-linux-amd64 /usr/local/bin/cd
    - rm cd-linux-amd64
    - |       
      # Exchange GitLab token for Dex token
      DEX_URL="https://cd.example.com/api/dex/token"
      DEX_TOKEN_RESPONSE=$(curl -sSf \
        "$DEX_URL" \
        --user argo-cd-cli: \
        --data-urlencode "connector_id=gitlab-ci" \
        --data-urlencode "grant_type=urn:ietf:params:oauth:grant-type:token-exchange" \
        --data-urlencode "scope=openid email profile federated:id" \
        --data-urlencode "requested_token_type=urn:ietf:params:oauth:token-type:access_token" \
        --data-urlencode "subject_token=$GITLAB_OIDC_TOKEN" \
        --data-urlencode "subject_token_type=urn:ietf:params:oauth:token-type:id_token")
      
      DEX_TOKEN=$(echo "$DEX_TOKEN_RESPONSE" | jq -r .access_token)
      
      # Use with Hanzo CD CLI
      export CD_SERVER="cd.example.com" 
      export CD_OPTS="--grpc-web"
      export CD_AUTH_TOKEN="$DEX_TOKEN"
      cd version
      cd account get-user-info
      cd app list
```


## Configuring RBAC

When using Hanzo CD global RBAC config map, you can define your `policy.csv` like so:

```yaml
configs:
  rbac:
    policy.csv: |
      # Specific project(infra) for specific apps
      p, project_path:my-repo/my-project:*, applications, get, infra/*, allow
      # Only main branch can sync under production project
      p, project_path:my-repo/my-project:ref_type:branch:ref:main, applications, sync, production/*, allow
```

More info: [RBAC Configuration](../rbac.md)
