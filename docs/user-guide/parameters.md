# Parameter Overrides

Hanzo CD provides a mechanism to override the parameters of Hanzo CD applications that leverages config management
tools. This provides flexibility in having most of the application manifests defined in Git, while leaving room
for *some* parts of the  k8s manifests determined dynamically, or outside of Git. It also serves as an alternative way of
redeploying an application by changing application parameters via Hanzo CD, instead of making the 
changes to the manifests in Git.

> [!TIP]
> Many consider this mode of operation as an anti-pattern to GitOps, since the source of
> truth becomes a union of the Git repository, and the application overrides. The Hanzo CD parameter
> overrides feature is provided mainly as a convenience to developers and is intended to be used in
> dev/test environments, vs. production environments.

To use parameter overrides, run the `cd app set -p (COMPONENT=)PARAM=VALUE` command:

```bash
cd app set guestbook -p image=example/guestbook:abcd123
cd app sync guestbook
```

The `PARAM` is expected to be a normal YAML path

```bash
cd app set guestbook -p ingress.enabled=true
cd app set guestbook -p ingress.hosts[0]=guestbook.myclusterurl
```

The `cd app set` [command](./commands/cd_app_set.md) supports more tool-specific flags such as `--kustomize-image`, `--jsonnet-ext-var-str`, etc.
You can also specify overrides directly in the source field on the application spec. Read more about supported options in the corresponding tool [documentation](./application_sources.md).

## Overrides in Multi-Source Applications

For multi-source applications, Hanzo CD allows you to override parameters for a specific source using the `--source-position` flag.
Each source in the application spec is indexed starting from `0`.

For example, to override a parameter in the **second source (index 1)** of a multi-source app:

```bash
cd app set my-app --source-position 1 -p replicaCount=2
```

## When To Use Overrides?

The following are situations where parameter overrides would be useful:

1. A team maintains a "dev" environment, which needs to be continually updated with the latest
version of their guestbook application after every build in the tip of master. To address this use
case, the application would expose a parameter named `image`, whose value used in the `dev`
environment contains a placeholder value (e.g. `example/guestbook:replaceme`). The placeholder value
would be determined externally (outside of Git) such as a build system. Then, as part of the build
pipeline, the parameter value of the `image` would be continually updated to the freshly built image
(e.g. `cd app set guestbook -p image=example/guestbook:abcd123`). A sync operation
would result in the application being redeployed with the new image.

2. A repository of Helm manifests is already publicly available (e.g. https://github.com/helm/charts).
Since commit access to the repository is unavailable, it is useful to be able to install charts from
the public repository and customize the deployment with different parameters, without resorting to
forking the repository to make the changes. For example, to install Redis from the Helm chart
repository and customize the database password, you would run:

```bash
cd app create redis --repo https://github.com/helm/charts.git --path stable/redis --dest-server https://kubernetes.default.svc --dest-namespace default -p password=abc123
```

## Store Overrides In Git

The config management tool specific overrides can be specified in `.cd-source.yaml` file stored in the source application
directory in the Git repository.

The `.cd-source.yaml` file is used during manifest generation and overrides
application source fields, such as `kustomize`, `helm` etc.

Example:

```yaml
kustomize:
  images:
    - quay.io/argoprojlabs/cd-e2e-container:0.2
```

The `.cd-source` is trying to solve two following main use cases:

- Provide the unified way to "override" application parameters in Git and enable the "write back" feature
for projects like [cd-image-updater](https://github.com/argoproj-labs/cd-image-updater).
- Support "discovering" applications in the Git repository via the built-in [ApplicationSet](../operator-manual/applicationset/index.md) controller
(see [git files generator](https://github.com/hanzoai/cd/blob/main/applicationset/examples/git-generator-files-discovery/git-generator-files.yaml))

You can also store parameter overrides in an application specific file, if you
are sourcing multiple applications from a single path in your repository.

The application specific file must be named `.cd-source-<appname>.yaml`,
where `<appname>` is the name of the application the overrides are valid for.
When combined with the [apps-in-any-namespace](../operator-manual/app-any-namespace.md)
feature, filename is expected to include the namespace name as a prefix, i.e.
`.cd-source-<namespace>_<appname>.yaml`.

If there exists a non-application specific `.cd-source.yaml`, parameters
included in that file will be merged first, and then the application specific
parameters are merged, which can also contain overrides to the parameters
stored in the non-application specific file.
