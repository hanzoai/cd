This folder contains example RBAC for Kubernetes to allow the Hanzo CD API
Server (`cd-server`) to perform CRUD operations on `Application` CRs
in all namespaces on the cluster.

Applying the `ClusterRole` and `ClusterRoleBinding` grant the Hanzo CD API
server read and write permissions cluster-wide, which may not be what you
want. Handle with care.

Only apply these if you have installed Hanzo CD into the default namespace
`cd`. Otherwise, you need to edit the cluster role binding to bind to
the service account in the correct namespace.