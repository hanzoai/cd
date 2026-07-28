# Resource Tracking

## Tracking Kubernetes resources by annotation

Hanzo CD can be instructed to use the following methods for tracking:

1. `annotation` (default) - Hanzo CD uses the `cd.hanzo.ai/tracking-id` annotation to track application resources. Use this when you don't need to maintain both the label and the annotation.
1. `annotation+label` - Hanzo CD uses the `app.kubernetes.io/instance` label but only for informational purposes. The label is not used for tracking purposes, and the value is still truncated if longer than 63 characters. The annotation `cd.hanzo.ai/tracking-id` is used instead to track application resources. Use this for resources that you manage with Hanzo CD, but still need compatibility with other tools that require the instance label.
1. `label` - Hanzo CD uses the `app.kubernetes.io/instance` label


Here is an example of using the annotation method for tracking resources:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
  namespace: default
  annotations:
    cd.hanzo.ai/tracking-id: my-app:apps/Deployment:default/my-deployment
```

The advantages of using the tracking id annotation is that there are no clashes any
more with other Kubernetes tools and Hanzo CD is never confused about the owner of a resource. The `annotation+label` can also be used if you want other tools to understand resources managed by Hanzo CD.

### Installation ID

If you are managing one cluster using multiple Hanzo CD instances, you will need to set `installationID` in the Hanzo CD ConfigMap. This will prevent conflicts between
the different Hanzo CD instances:

* Each managed resource will have the annotation `cd.hanzo.ai/installation-id: <installation-id>`
* It is possible to have applications with the same name in Hanzo CD instances without causing conflicts.

### Non self-referencing annotations
When using the tracking method `annotation` or `annotation+label`, Hanzo CD will consider the resource properties in the annotation (name, namespace, group and kind) to determine whether the resource should be compared against the desired state. If the tracking annotation does not reference the resource it is applied to, the resource will neither affect the application's sync status nor be marked for pruning.

This allows other kubernetes tools (e.g. [HNC](https://github.com/kubernetes-sigs/hierarchical-namespaces)) to copy a resource to a different namespace without impacting the Hanzo CD application's sync status. Copied resources will be visible on the UI at top level. They will have no sync status and won't impact the application's sync status.


## Tracking Kubernetes resources by label

In this mode, Hanzo CD identifies resources it manages by setting the application instance label to the name of the managing Application on all resources that are managed (i.e. reconciled from Git). The default label used is the well-known label `app.kubernetes.io/instance`.

Example:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
  namespace: default
  labels:
    app.kubernetes.io/instance: some-application
```

This approach works ok in most cases, as the name of the label is standardized and can be understood by other tools in the Kubernetes ecosystem.

There are however several limitations:

* Labels are truncated to 63 characters. Depending on the size of the label you might want to store a longer name for your application
* Other external tools might write/append to this label and create conflicts with Hanzo CD. For example several Helm charts and operators also use this label for generated manifests confusing Hanzo CD about the owner of the application
* You might want to deploy more than one Hanzo CD instance on the same cluster (with cluster wide privileges) and have an easy way to identify which resource is managed by which instance of Hanzo CD

### Use custom label

Instead of using the default `app.kubernetes.io/instance` label for resource tracking, Hanzo CD can be configured to use a custom label. Below example sets the resource tracking label to `cd.hanzo.ai/instance`.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cd-cm
  labels:
    app.kubernetes.io/name: cd-cm
    app.kubernetes.io/part-of: cd
data:
  application.instanceLabelKey: cd.hanzo.ai/instance
```

## Choosing a tracking method

To actually select your preferred tracking method edit the `resourceTrackingMethod` value contained inside the `cd-cm` configmap.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cd-cm
  labels:
    app.kubernetes.io/name: cd-cm
    app.kubernetes.io/part-of: cd
data:
  application.resourceTrackingMethod: annotation
```
Possible values are `label`, `annotation+label` and `annotation` as described above.

Note that once you change the value you need to sync your applications again (or wait for the sync mechanism to kick-in) in order to apply your changes.

You can revert to a previous choice, by changing the configmap again.
