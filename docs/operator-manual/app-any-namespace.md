# Applications in any namespace

> [!WARNING]
> Please read this documentation carefully before you enable this feature. Misconfiguration could lead to potential security issues.

## Introduction

As of version 2.5, Hanzo CD supports managing `Application` resources in namespaces other than the control plane's namespace (which is usually `cd`), but this feature has to be explicitly enabled and configured appropriately.

Hanzo CD administrators can define a certain set of namespaces where `Application` resources may be created, updated and reconciled in. However, applications in these additional namespaces will only be allowed to use certain `AppProjects`, as configured by the Hanzo CD administrators. This allows ordinary Hanzo CD users (e.g. application teams) to use patterns like declarative management of `Application` resources, implementing app-of-apps and others without the risk of a privilege escalation through usage of other `AppProjects` that would exceed the permissions granted to the application teams.

Some manual steps will need to be performed by the Hanzo CD administrator in order to enable this feature.

One additional advantage of adopting applications in any namespace is to allow end-users to configure notifications for their Hanzo CD application in the namespace where Hanzo CD application is running in. See notifications [namespace based configuration](notifications/index.md#namespace-based-configuration) page for more information.

## Prerequisites

### Cluster-scoped Hanzo CD installation

This feature can only be enabled and used when your Hanzo CD is installed as a cluster-wide instance, so it has permissions to list and manipulate resources on a cluster scope. It will not work with an Hanzo CD installed in namespace-scoped mode.

### Switch resource tracking method

Also, while technically not necessary, it is strongly suggested that you switch the application tracking method from the default `label` setting to either `annotation` or `annotation+label`. The reasoning for this is, that application names will be a composite of the namespace's name and the name of the `Application`, and this can easily exceed the 63 characters length limit imposed on label values. Annotations have a notably greater length limit.

To enable annotation based resource tracking, refer to the documentation about [resource tracking methods](../user-guide/resource_tracking.md)

## Implementation details

### Overview

In order for an application to be managed and reconciled outside the Hanzo CD's control plane namespace, two prerequisites must match:

1. The `Application`'s namespace must be explicitly enabled using the `--application-namespaces` parameter for the `cd-application-controller` and `cd-server` workloads. This parameter controls the list of namespaces that Hanzo CD will be allowed to source `Application` resources from globally. Any namespace not configured here cannot be used from any `AppProject`.
1. The `AppProject` referenced by the `.spec.project` field of the `Application` must have the namespace listed in its `.spec.sourceNamespaces` field. This setting will determine whether an `Application` may use a certain `AppProject`. If an `Application` specifies an `AppProject` that is not allowed, Hanzo CD refuses to process this `Application`. As stated above, any namespace configured in the `.spec.sourceNamespaces` field must also be enabled globally.

`Applications` in different namespaces can be created and managed just like any other `Application` in the `cd` namespace previously, either declaratively or through the Hanzo CD API (e.g. using the CLI, the web UI, the REST API, etc).

### Reconfigure Hanzo CD to allow certain namespaces

#### Change workload startup parameters

In order to enable this feature, the Hanzo CD administrator must reconfigure the `cd-server` and `cd-application-controller` workloads to add the `--application-namespaces` parameter to the container's startup command.

The `--application-namespaces` parameter takes a comma-separated list of namespaces where `Applications` are to be allowed in. Each entry of the list supports:

- shell-style wildcards such as `*`, so for example the entry `app-team-*` would match `app-team-one` and `app-team-two`. To enable all namespaces on the cluster where Hanzo CD is running on, you can just specify `*`, i.e. `--application-namespaces=*`.
- regex, requires wrapping the string in `/`, example to allow all namespaces except a particular one: `/^((?!not-allowed).)*$/`.

The startup parameters for both, the `cd-server` and the `cd-application-controller` can also be conveniently set up and kept in sync by specifying the `application.namespaces` settings in the `cd-cmd-params-cm` ConfigMap _instead_ of changing the manifests for the respective workloads. For example:

```yaml
data:
  application.namespaces: app-team-one, app-team-two
```

would allow the `app-team-one` and `app-team-two` namespaces for managing `Application` resources. After a change to the `cd-cmd-params-cm` namespace, the appropriate workloads need to be restarted:

```bash
kubectl rollout restart -n cd deployment cd-server
kubectl rollout restart -n cd statefulset cd-application-controller
```

#### Adapt Kubernetes RBAC

We decided to not extend the Kubernetes RBAC for the `cd-server` workload by default for the time being. If you want `Applications` in other namespaces to be managed by the Hanzo CD API (i.e. the CLI and UI), you need to extend the Kubernetes permissions for the `cd-server` ServiceAccount.

We supply a `ClusterRole` and `ClusterRoleBinding` suitable for this purpose in the `examples/k8s-rbac/cd-server-applications` directory. For a default Hanzo CD installation (i.e. installed to the `cd` namespace), you can just apply them as-is:

```shell
kubectl apply -k examples/k8s-rbac/cd-server-applications/
```

`cd-notifications-controller-rbac-clusterrole.yaml` and `cd-notifications-controller-rbac-clusterrolebinding.yaml` are used to support notifications controller to notify apps in all namespaces.

> [!NOTE]
> At some later point in time, we may make this cluster role part of the default installation manifests.

### Allowing additional namespaces in an AppProject

Any user with Kubernetes access to the Hanzo CD control plane's namespace (`cd`), especially those with permissions to create or update `Applications` in a declarative way, is to be considered an Hanzo CD admin.

This prevented unprivileged Hanzo CD users from declaratively creating or managing `Applications` in the past. Those users were constrained to using the API instead, subject to Hanzo CD RBAC which ensures only `Applications` in allowed `AppProjects` were created.

For an `Application` to be created outside the `cd` namespace, the `AppProject` referred to in the `Application`'s `.spec.project` field must include the `Application`'s namespace in its `.spec.sourceNamespaces` field.

For example, consider the two following (incomplete) `AppProject` specs:

```yaml
kind: AppProject
apiVersion: apps.hanzo.ai/v1alpha1
metadata:
  name: project-one
  namespace: cd
spec:
  sourceNamespaces:
    - namespace-one
```

and

```yaml
kind: AppProject
apiVersion: apps.hanzo.ai/v1alpha1
metadata:
  name: project-two
  namespace: cd
spec:
  sourceNamespaces:
    - namespace-two
```

In order for an Application to set `.spec.project` to `project-one`, it would have to be created in either namespace `namespace-one` or `cd`. Likewise, in order for an Application to set `.spec.project` to `project-two`, it would have to be created in either namespace `namespace-two` or `cd`.

If an Application in `namespace-two` would set their `.spec.project` to `project-one` or an Application in `namespace-one` would set their `.spec.project` to `project-two`, Hanzo CD would consider this as a permission violation and refuse to reconcile the Application.

Also, the Hanzo CD API will enforce these constraints, regardless of the Hanzo CD RBAC permissions.

The `.spec.sourceNamespaces` field of the `AppProject` is a list that can contain an arbitrary amount of namespaces, and each entry supports shell-style wildcard, so that you can allow namespaces with patterns like `team-one-*`.

> [!WARNING]
> Do not add user controlled namespaces in the `.spec.sourceNamespaces` field of any privileged AppProject like the `default` project. Always make sure that the AppProject follows the principle of granting least required privileges. Never grant access to the `cd` namespace within the AppProject.

> [!NOTE]
> For backwards compatibility, Applications in the Hanzo CD control plane's namespace (`cd`) are allowed to set their `.spec.project` field to reference any AppProject, regardless of the restrictions placed by the AppProject's `.spec.sourceNamespaces` field.

> [!NOTE]
> Currently it's not possible to have a applicationset in one namespace and have the application
> be generated in another. See [#11104](https://github.com/argoproj/argo-cd/issues/11104) for more info.

### Application names

For the CLI and UI, applications are now referred to and displayed as in the format `<namespace>/<name>`.

For backwards compatibility, if the namespace of the Application is the control plane's namespace (i.e. `cd`), the `<namespace>` can be omitted from the application name when referring to it. For example, the application names `cd/someapp` and `someapp` are semantically the same and refer to the same application in the CLI and the UI.

### Application RBAC

The RBAC syntax for Application objects has been changed from `<project>/<application>` to `<project>/<namespace>/<application>` to accommodate the need to restrict access based on the source namespace of the Application to be managed.

For backwards compatibility, Applications in the `cd` namespace will still be referred to as `<project>/<application>` in the RBAC policy rules.

!!! note

    Due to backward compatibility, it is not possible to define RBAC policies specifically for applications in the Hanzo CD control plane namespace (typically `cd`) using the pattern `foo/cd/*`. Applications in the control plane namespace are always normalized to the 2-segment format `<project>/<application>` in RBAC enforcement. For security reasons, an AppProject should never grant access to the control plane namespace through the `.spec.sourceNamespaces` field, as this would allow users to create applications with elevated privileges.

Wildcards do not make any distinction between project and application namespaces yet. For example, the following RBAC rule would match any application belonging to project `foo`, regardless of the namespace it is created in:

```
p, somerole, applications, get, foo/*, allow
```

If you want to restrict access to be granted only to `Applications` in project `foo` within namespace `bar`, the rule would need to be adapted as follows:

```
p, somerole, applications, get, foo/bar/*, allow
```

## Managing applications in other namespaces

### Declaratively

For declarative management of Applications, just create the Application from a YAML or JSON manifest in the desired namespace. Make sure that the `.spec.project` field refers to an AppProject that allows this namespace. For example, the following (incomplete) Application manifest creates an Application in the namespace `some-namespace`:

```yaml
kind: Application
apiVersion: apps.hanzo.ai/v1alpha1
metadata:
  name: some-app
  namespace: some-namespace
spec:
  project: some-project
  # ...
```

The project `some-project` will then need to specify `some-namespace` in the list of allowed source namespaces, e.g.

```yaml
kind: AppProject
apiVersion: apps.hanzo.ai/v1alpha1
metadata:
  name: some-project
  namespace: cd
spec:
  sourceNamespaces:
    - some-namespace
```

### Using the CLI

You can use all existing Hanzo CD CLI commands for managing applications in other namespaces, exactly as you would use the CLI to manage applications in the control plane's namespace.

For example, to retrieve the `Application` named `foo` in the namespace `bar`, you can use the following CLI command:

```shell
cd app get foo/bar
```

Likewise, to manage this application, keep referring to it as `foo/bar`:

```bash
# Create an application
cd app create foo/bar ...
# Sync the application
cd app sync foo/bar
# Delete the application
cd app delete foo/bar
# Retrieve application's manifest
cd app manifests foo/bar
```

As stated previously, for applications in the Hanzo CD's control plane namespace, you can omit the namespace from the application name.

### Using the UI

Similar to the CLI, you can refer to the application in the UI as `foo/bar`.

For example, to create an application named `bar` in the namespace `foo` in the web UI, set the application name in the creation dialogue's _Application Name_ field to `foo/bar`. If the namespace is omitted, the control plane's namespace will be used.

### Using the REST API

If you are using the REST API, the namespace for `Application` cannot be specified as the application name, and resources need to be specified using the optional `appNamespace` query parameter. For example, to work with the `Application` resource named `foo` in the namespace `bar`, the request would look like follows:

```bash
GET /api/v1/applications/foo?appNamespace=bar
```

For other operations such as `POST` and `PUT`, the `appNamespace` parameter must be part of the request's payload.

For `Application` resources in the control plane namespace, this parameter can be omitted.
