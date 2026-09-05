# Try Hanzo CD Locally

> [!TIP]
> This guide assumes you have a grounding in the tools that Hanzo CD is based on. Please read [understanding the basics](understand_the_basics.md) to learn about these tools.


Follow these steps to install `Kind` for local development and set it up with Hanzo CD.

To run an Hanzo CD development environment review the [developer guide for running locally](./developer-guide/running-locally.md).

## Install Kind

Install Kind following their quick-start [instructions](https://kind.sigs.k8s.io/docs/user/quick-start#installation).

##  Create a Kind Cluster
Once Kind is installed, create a new Kubernetes cluster with:
```bash
kind create cluster --name cd-cluster
```
This will create a local Kubernetes cluster named `cd-cluster`.

## Set Up kubectl to Use the Kind Cluster
After creating the cluster, set `kubectl` to use your new `kind` cluster:
```bash
kubectl cluster-info --context kind-cd-cluster
```
This command verifies that `kubectl` is pointed to the right cluster.

## Install Hanzo CD on the Cluster
You can now install Hanzo CD on your `kind` cluster. First, apply the Hanzo CD manifest to create the necessary resources:
```bash
kubectl create namespace cd
kubectl apply -n cd --server-side --force-conflicts -f https://raw.githubusercontent.com/hanzoai/cd/main/manifests/install.yaml
```

> [!NOTE]
> The `--server-side --force-conflicts` flags are required because some Hanzo CD CRDs exceed the size limit for client-side apply. See the [getting started guide](getting_started.md) for more details.

## Expose Hanzo CD API Server
By default, Hanzo CD's API server is not exposed outside the cluster. You need to expose it to access the UI locally. For development purposes, you can use Kubectl 'port-forward'.
```bash
kubectl port-forward svc/cd-server -n cd 8080:443
```
This will forward port 8080 on your local machine to the Hanzo CD API server’s port 443 inside the Kubernetes cluster.

## Access Hanzo CD UI
Now, you can open your browser and navigate to http://localhost:8080 to access the Hanzo CD UI.

### Log in to Hanzo CD
To log in to the Hanzo CD UI, you'll need the default admin password. You can retrieve it from the Kubernetes cluster:
```bash
kubectl -n cd get secret cd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d
```
Use the admin username and the retrieved password to log in.

You can now move on to step #2 in the [Getting Started Guide](getting_started.md).
