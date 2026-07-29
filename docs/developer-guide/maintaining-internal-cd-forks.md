# Maintaining Internal Hanzo CD Forks

Most Hanzo CD contributors don't need this section to contribute to Hanzo CD. In most cases, the [Regular Developer Guide](index.md) is sufficient.

This section will help companies that need to publish custom Hanzo CD images or publish custom Hanzo CD releases from their forks.
Such companies need the below documentation in addition to the [Regular Developer Guide](index.md).
This section will also help Hanzo CD maintainers to test the release process in a test environment.

## Understanding where and which upstream images are published

Official upstream release tags (`vX.Y.Z*`) publish their multi-platform images and the corresponding provenance attestations—to `ghcr.io/hanzoai/cd` (or whatever registry a fork configures via `IMAGE_*` variables).  
Upstream master builds continue to refresh the `latest` tag in the same primary registry, while also pushing commit-tagged images (and their provenance) to `ghcr.io/hanzoai/cd/cd` so `cd.apps.apps.hanzo.ai` can pin exact SHAs.  
Forks inherit the same behavior but target their customized registries/namespaces and do not deploy to `cd.apps.apps.hanzo.ai`.

## Publishing custom images from forked master branches

Fork builds can publish their own containers once workflow variables point at your registry/namespace instead of `hanzoai`.

### Configuring GitHub Actions variables
Adjust the variables below to match your setup (overriding `IMAGE_NAMESPACE` is required, because it flips the workflows out of “upstream” mode):

- `IMAGE_NAMESPACE` – defaults to `hanzoai` (overriding required)
- `IMAGE_REPOSITORY` – defaults to `cd` (may need overriding)
- `GHCR_NAMESPACE` – defaults to `${{ github.repository }}`, which translates to `<YOUR_GITHUB_USERNAME>/<YOUR_FORK_REPO>`, rarely needs overriding)
- `GHCR_REPOSITORY` – defaults to `cd` (may need overriding)

These values produce the final image names:

- `quay.io/$IMAGE_NAMESPACE/$IMAGE_REPOSITORY`
- `ghcr.io/$GHCR_NAMESPACE/$GHCR_REPOSITORY`

Example: if your GitHub account is `my-user`, your fork is `my-argo-cd-fork`, and you want to push release images to `quay.io/my-quay-user/cd`, configure:

- `IMAGE_NAMESPACE = my-quay-user`
Your master build images will then be published to `quay.io/my-quay-user/cd:latest`, and the commit tagged images along with the attestations will be published under the Packages (GHCR) of your GitHub fork repo. 

### Configuring GitHub Actions secrets
Supply credentials for your primary registry so the workflow can push:

- `RELEASE_QUAY_USERNAME`
- `RELEASE_QUAY_TOKEN`

## Enabling fork releases

Forks can run the full release workflow by setting `ENABLE_FORK_RELEASES: true`, ensuring all upstream tags are fetched (the release tooling needs previous tags for changelog diffs), and reusing the same image variables/secrets listed above so release images push to your custom registry. After that, follow the standard [Release Process](releasing.md) with one critical adjustment:

> [!WARNING]
> When invoking `hack/trigger-release.sh`, point it at your fork remote (usually `origin`) rather than ~~upstream~~, otherwise the script may try to push official tags.  
> Example: `./hack/trigger-release.sh v2.7.2 origin`
