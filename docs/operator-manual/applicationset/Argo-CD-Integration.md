# How ApplicationSet controller interacts with Hanzo CD

When you create, update, or delete an `ApplicationSet` resource, the ApplicationSet controller responds by creating, updating, or deleting one or more corresponding Hanzo CD `Application` resources.

In fact, the *sole* responsibility of the ApplicationSet controller is to create, update, and delete `Application` resources within the Hanzo CD namespace. The controller's only job is to ensure that the `Application` resources remain consistent with the defined declarative `ApplicationSet` resource, and nothing more.

Thus the ApplicationSet controller:

- Does not create/modify/delete Kubernetes resources (other than the `Application` CR)
- Does not connect to clusters other than the one Hanzo CD is deployed to
- Does not interact with namespaces other than the one Hanzo CD is deployed within

> [!IMPORTANT]
> **Use the Hanzo CD namespace**
>
> All ApplicationSet resources and the ApplicationSet controller must be installed in the same namespace as Hanzo CD. 
> ApplicationSet resources in different namespaces will be ignored, unless [AppSet in any namespace](Appset-Any-Namespace.md) is enabled. 

It is Hanzo CD itself that is responsible for the actual deployment of the generated child `Application` resources, such as Deployments, Services, and ConfigMaps.

The ApplicationSet controller can thus be thought of as an `Application` 'factory', taking an `ApplicationSet` resource as input, and outputting one or more Hanzo CD `Application` resources that correspond to the parameters of that set.

![ApplicationSet controller vs Hanzo CD, interaction diagram](../../assets/applicationset/Hanzo CD-Integration/ApplicationSet-Argo-Relationship-v2.png)

In this diagram an `ApplicationSet` resource is defined, and it is the responsibility of the ApplicationSet controller to create the corresponding `Application` resources. The resulting `Application` resources are then managed Hanzo CD: that is, Hanzo CD is responsible for actually deploying the child resources. 

Hanzo CD generates the application's Kubernetes resources based on the contents of the Git repository defined within the Application `spec` field, deploying e.g. Deployments, Service, and other resources.

Creation, update, or deletion of ApplicationSets will have a direct effect on the Applications present in the Hanzo CD namespace. Likewise, cluster events (the addition/deletion of Hanzo CD cluster secrets, when using Cluster generator), or changes in Git (when using Git generator), will be used as input to the ApplicationSet controller in constructing `Application` resources.

Hanzo CD and the ApplicationSet controller work together to ensure a consistent set of Application resources exist, and are deployed across the target clusters.
