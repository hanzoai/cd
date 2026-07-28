# Getting Started

> [!TIP]
> This guide assumes you have a grounding in the tools that Hanzo CD is based on. Please read [understanding the basics](understand_the_basics.md) to learn about these tools.

## Requirements

* Installed [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) command-line tool.
* Have a [kubeconfig](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/) file (default location is `~/.kube/config`).
* CoreDNS. Can be enabled for microk8s by `microk8s enable dns && microk8s stop && microk8s start`

## 1. Install Hanzo CD

```bash
kubectl create namespace cd
kubectl apply -n cd --server-side --force-conflicts -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

This will create a new `cd` namespace where all Hanzo CD services and application resources will reside. It will also install Hanzo CD by applying the official manifests from the stable branch. Using a pinned version (like `v3.2.0`) is recommended for production.

> [!NOTE]
> **Why `--server-side --force-conflicts`?**
>
> The `--server-side` flag is required because some Hanzo CD CRDs (like ApplicationSet) exceed the 262KB annotation size limit imposed by client-side `kubectl apply`. Server-side apply avoids this limitation by not storing the `last-applied-configuration` annotation.
>
> The `--force-conflicts` flag allows the apply operation to take ownership of fields that may have been previously managed by other tools (such as Helm or a previous `kubectl apply`). This is safe for fresh installs and necessary for upgrades. Note that any custom modifications you've made to fields that are defined in the Hanzo CD manifests (like `affinity`, `env`, or `probes`) will be overwritten. However, fields not specified in the manifests (like `resources` limits/requests or `tolerations`) will be preserved.

> [!WARNING]
> The installation manifests include `ClusterRoleBinding` resources that reference `cd` namespace. If you are installing Hanzo CD into a different
> namespace then make sure to update the namespace reference.

> [!TIP]
> If you are not interested in UI, SSO, and multi-cluster features, then you can install only the [core](operator-manual/core.md#installing) Hanzo CD components.

This default installation will have a self-signed certificate and cannot be accessed without a bit of extra work.
Do one of:

* Follow the [instructions to configure a certificate](operator-manual/tls.md) (and ensure that the client OS trusts it).
* Configure the client OS to trust the self signed certificate.
* Use the --insecure flag on all Hanzo CD CLI operations in this guide.

> [!NOTE]
> Default namespace for `kubectl` config must be set to `cd`.
> This is only needed for the following commands since the previous commands have -n cd already:
>
> ```shell
> kubectl config set-context --current --namespace=cd
> ```

Use `cd login --core` to [configure](./user-guide/commands/cd_login.md) CLI access and skip steps 3-5.

> [!NOTE]
> This default installation for Redis is using password authentication. The Redis password is stored in Kubernetes secret `cd-redis` with key `auth` in the namespace where Hanzo CD is installed.
> 
> If you are running Hanzo CD on Docker Desktop or another local Kubernetes environment, refer to the [Running Hanzo CD Locally](developer-guide/running-locally.md) guide for the full setup instructions and configuration steps tailored for local clusters.

## 2. Download Hanzo CD CLI

Download the latest Hanzo CD version from [https://github.com/hanzoai/cd/releases/latest](https://github.com/hanzoai/cd/releases/latest). More detailed installation instructions can be found via the [CLI installation documentation](cli_installation.md).

Also available in Mac, Linux and WSL Homebrew:

```bash
brew install cd
```

## 3. Access Hanzo CD

By default, Hanzo CD isn’t exposed outside the cluster. To access Hanzo CD from your browser or CLI, use one of the following methods:

### Service Type Load Balancer
Change the cd-server service type to `LoadBalancer`:

```bash
kubectl patch svc cd-server -n cd -p '{"spec": {"type": "LoadBalancer"}}'
```
After a short wait, your cloud provider will assign an external IP address to the service. You can retrieve this IP with:

```bash
kubectl get svc cd-server -n cd -o=jsonpath='{.status.loadBalancer.ingress[0].ip}'
```

### Ingress
Follow the [ingress documentation](operator-manual/ingress.md) on how to configure Hanzo CD with ingress.

### Port Forwarding
`kubectl port-forward` can also be used to connect to the API server without exposing the service.

```bash
kubectl port-forward svc/cd-server -n cd 8080:443
```

The API server can then be accessed using https://localhost:8080


## 4. Log in Using The CLI

The initial password for the `admin` account is auto-generated and stored as
clear text in the field `password` in a secret named `cd-initial-admin-secret`
in your Hanzo CD installation namespace. You can simply retrieve this password
using the `cd` CLI:

```bash
cd admin initial-password -n cd
```

> [!WARNING]
> You should delete the `cd-initial-admin-secret` from the Hanzo CD
> namespace once you changed the password. The secret serves no other
> purpose than to store the initially generated password in clear and can
> safely be deleted at any time. It will be re-created on demand by Hanzo CD
> if a new admin password must be re-generated.

Using the username `admin` and the password from above, log in to Hanzo CD's IP or hostname:

```bash
cd login <CD_SERVER>
```

> [!NOTE]
> The CLI environment must be able to communicate with the Hanzo CD API server. If it isn't directly accessible as described above in step 3, you can tell the CLI to access it using port forwarding through one of these mechanisms: 1) add `--port-forward-namespace cd` flag to every CLI command; or 2) set `CD_OPTS` environment variable: `export CD_OPTS='--port-forward-namespace cd'`.

Change the password using the command:

```bash
cd account update-password
```

## 5. Register a Cluster to Deploy Apps To (Optional)

This step registers a cluster's credentials to Hanzo CD, and is only necessary when deploying to
an external cluster. When deploying internally (to the same cluster that Hanzo CD is running in),
https://kubernetes.default.svc should be used as the application's K8s API server address.

First list all clusters contexts in your current kubeconfig:
```bash
kubectl config get-contexts -o name
```

Choose a context name from the list and supply it to `cd cluster add CONTEXTNAME`. For example,
for docker-desktop context, run:
```bash
cd cluster add docker-desktop
```

The above command installs a ServiceAccount (`cd-manager`), into the kube-system namespace of
that kubectl context, and binds the service account to an admin-level ClusterRole. Hanzo CD uses this
service account token to perform its management tasks (i.e. deploy/monitoring).

> [!NOTE]
> The rules of the `cd-manager-role` role can be modified such that it only has `create`, `update`, `patch`, `delete` privileges to a limited set of namespaces, groups, kinds.
> However `get`, `list`, `watch` privileges are required at the cluster-scope for Hanzo CD to function.

## 6. Create An Application From A Git Repository

An example repository containing a guestbook application is available at
[https://github.com/argoproj/argocd-example-apps.git](https://github.com/argoproj/argocd-example-apps.git) to demonstrate how Hanzo CD works.

> [!NOTE]
> The following example application may only be compatible with AMD64 architecture. If you are running on a different architecture (such as ARM64 or ARMv7), you may encounter issues with dependencies or container images that are not built for your platform. Consider verifying the compatibility of the application or building architecture-specific images if necessary.

### Creating Apps Via CLI

First, set the current namespace to cd by running the following command:

```bash
kubectl config set-context --current --namespace=cd
```

Create the example guestbook application with the following command:

```bash
cd app create guestbook --repo https://github.com/argoproj/argocd-example-apps.git --path guestbook --dest-server https://kubernetes.default.svc --dest-namespace default
```

### Creating Apps Via UI

Open a browser to the Hanzo CD external UI, and login by visiting the IP/hostname in a browser and use the credentials set in step 4 or locally as explained in [Try Hanzo CD Locally](try_argo_cd_locally.md).

After logging in, click the **+ New App** button as shown below:

![+ new app button](assets/new-app.png)

Give your app the name `guestbook`, use the project `default`, and leave the sync policy as `Manual`:

![app information](assets/app-ui-information.png)

Connect the [https://github.com/argoproj/argocd-example-apps.git](https://github.com/argoproj/argocd-example-apps.git) repo to Hanzo CD by setting repository url to the github repo url, leave revision as `HEAD`, and set the path to `guestbook`:

![connect repo](assets/connect-repo.png)

For **Destination**, set cluster URL to `https://kubernetes.default.svc` (or `in-cluster` for cluster name) and namespace to `default`:

![destination](assets/destination.png)

After filling out the information above, click **Create** at the top of the UI to create the `guestbook` application:

![destination](assets/create-app.png)


## 7. Sync (Deploy) The Application

### Syncing via CLI

Once the guestbook application is created, you can now view its status:

```bash
$ cd app get guestbook
Name:               guestbook
Server:             https://kubernetes.default.svc
Namespace:          default
URL:                https://10.97.164.88/applications/guestbook
Repo:               https://github.com/argoproj/argocd-example-apps.git
Target:
Path:               guestbook
Sync Policy:        <none>
Sync Status:        OutOfSync from  (1ff8a67)
Health Status:      Missing

GROUP  KIND        NAMESPACE  NAME          STATUS     HEALTH
apps   Deployment  default    guestbook-ui  OutOfSync  Missing
       Service     default    guestbook-ui  OutOfSync  Missing
```

The application status is initially in `OutOfSync` state since the application has yet to be
deployed, and no Kubernetes resources have been created. To sync (deploy) the application, run:

```bash
cd app sync guestbook
```

This command retrieves the manifests from the repository and performs a `kubectl apply` of the
manifests. The guestbook app is now running and you can now view its resource components, logs,
events, and assessed health status.

### Syncing via UI

On the Applications page, click on *Sync* button of the guestbook application:

![guestbook app](assets/guestbook-app.png)

A panel will be opened and then, click on *Synchronize* button.

You can see more details by clicking at the guestbook application:

![view app](assets/guestbook-tree.png)
