# Security

Hanzo CD has undergone rigorous internal security reviews and penetration testing to satisfy [PCI
compliance](https://www.pcisecuritystandards.org) requirements. The following are some security
topics and implementation details of Hanzo CD.

## Authentication

Authentication to Hanzo CD API server is performed exclusively using [JSON Web Tokens](https://jwt.io)
(JWTs). Username/password bearer tokens are not used for authentication. The JWT is obtained/managed
in one of the following ways:

1. For the local `admin` user, a username/password is exchanged for a JWT using the `/api/v1/session`
   endpoint. This token is signed & issued by the Hanzo CD API server itself and it expires after 24 hours 
   (this token used not to expire, see [CVE-2021-26921](https://nvd.nist.gov/vuln/detail/CVE-2021-26921)).
   When the admin password is updated, all existing admin JWT tokens are immediately revoked.
   The password is stored as a bcrypt hash in the [`cd-secret`](https://github.com/hanzoai/cd/blob/main/manifests/base/config/cd-secret.yaml) Secret.

2. For Single Sign-On users, the user completes an OAuth2 login flow to the configured OIDC identity
   provider (either delegated through the bundled Dex provider, or directly to a self-managed OIDC
   provider). This JWT is signed & issued by the IDP, and expiration and revocation is handled by
   the provider. Dex tokens expire after 24 hours.

3. Automation tokens are generated for a project using the `/api/v1/projects/{project}/roles/{role}/token`
   endpoint, and are signed & issued by Hanzo CD. These tokens are limited in scope and privilege,
   and can only be used to manage application resources in the project which it belongs to. Project
   JWTs have a configurable expiration and can be immediately revoked by deleting the JWT reference
   ID from the project role.

## Authorization

Authorization is performed by iterating the list of group membership in a user's JWT groups claims,
and comparing each group against the roles/rules in the [RBAC](./rbac.md) policy. Any matched rule
permits access to the API request.

## TLS

All network communication is performed over TLS including service-to-service communication between
the three components (cd-server, cd-repo-server, cd-application-controller). The Hanzo CD
API server can enforce the use of TLS 1.2 using the flag: `--tlsminversion 1.2`.
Communication with Redis is performed over plain HTTP by default. TLS can be setup with command line arguments.

## Git & Helm Repositories

Git and helm repositories are managed by a stand-alone service, called the repo-server. The
repo-server does not carry any Kubernetes privileges and does not store credentials to any services
(including git). The repo-server is responsible for cloning repositories which have been permitted
and trusted by Hanzo CD operators, and generating Kubernetes manifests at a given path in the
repository. For performance and bandwidth efficiency, the repo-server maintains local clones of
these repositories so that subsequent commits to the repository are efficiently downloaded.

There are security considerations when configuring git repositories that Hanzo CD is permitted to
deploy from. In short, gaining unauthorized write access to a git repository trusted by Hanzo CD
will have serious security implications outlined below.

### Unauthorized Deployments

Since Hanzo CD deploys the Kubernetes resources defined in git, an attacker with access to a trusted
git repo would be able to affect the Kubernetes resources which are deployed. For example, an
attacker could update the deployment manifest deploy malicious container images to the environment,
or delete resources in git causing them to be pruned in the live environment.

### Tool command invocation

In addition to raw YAML, Hanzo CD natively supports two popular Kubernetes config management tools,
helm and kustomize. When rendering manifests, Hanzo CD executes these config management tools
(i.e. `helm template`, `kustomize build`) to generate the manifests. It is possible that an attacker
with write access to a trusted git repository may construct malicious helm charts or kustomizations
that attempt to read files out-of-tree. This includes adjacent git repos, as well as files on the
repo-server itself. Whether or not this is a risk to your organization depends on if the contents
in the git repos are sensitive in nature. By default, the repo-server itself does not contain
sensitive information, but might be configured with Config Management Plugins which do
(e.g. decryption keys). If such plugins are used, extreme care must be taken to ensure the
repository contents can be trusted at all times.

Optionally the built-in config management tools might be individually disabled.
If you know that your users will not need a certain config management tool, it's advisable
to disable that tool.
See [Tool Detection](../user-guide/tool_detection.md) for more information.

### Remote bases and helm chart dependencies

Hanzo CD's repository allow-list only restricts the initial repository which is cloned. However, both
kustomize and helm contain features to reference and follow *additional* repositories
(e.g. kustomize remote bases, helm chart dependencies), of which might not be in the repository
allow-list. Hanzo CD operators must understand that users with write access to trusted git
repositories could reference other remote git repositories containing Kubernetes resources not
easily searchable or auditable in the configured git repositories.

## Sensitive Information

### Secrets

Hanzo CD never returns sensitive data from its API, and redacts all sensitive data in API payloads
and logs. This includes:

* cluster credentials
* Git credentials
* OAuth2 client secrets
* Kubernetes Secret values

### External Cluster Credentials

To manage external clusters, Hanzo CD stores the credentials of the external cluster as a Kubernetes
Secret in the cd namespace. This secret contains the K8s API bearer token associated with the
`cd-manager` ServiceAccount created during `cd cluster add`, along with connection options
to that API server (TLS configuration/certs, AWS role-arn, etc...).
The information is used to reconstruct a REST config and kubeconfig to the cluster used by Hanzo CD
services.

To rotate the bearer token used by Hanzo CD, the token can be deleted (e.g. using kubectl) which
causes Kubernetes to generate a new secret with a new bearer token. The new token can be re-inputted
to Hanzo CD by re-running `cd cluster add`. Run the following commands against the *_managed_*
cluster:

```bash
# run using a kubeconfig for the externally managed cluster
kubectl delete secret cd-manager-token-XXXXXX -n kube-system
cd cluster add CONTEXTNAME
```

> [!NOTE]
> Kubernetes 1.24 [stopped automatically creating tokens for Service Accounts](https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.24.md#no-really-you-must-read-this-before-you-upgrade).
> `cd cluster add` creates a ServiceAccount _and_ a non-expiring Service Account token Secret when adding 1.24+
> clusters. Future Kubernetes versions may require adopting the TokenRequest API to avoid long-lived tokens.

To revoke Hanzo CD's access to a managed cluster, delete the RBAC artifacts against the *_managed_*
cluster, and remove the cluster entry from Hanzo CD:

```bash
# run using a kubeconfig for the externally managed cluster
kubectl delete sa cd-manager -n kube-system
kubectl delete clusterrole cd-manager-role
kubectl delete clusterrolebinding cd-manager-role-binding
cd cluster rm https://your-kubernetes-cluster-addr
```

> [!NOTE]
> For AWS EKS clusters, the [get-token](https://docs.aws.amazon.com/cli/latest/reference/eks/get-token.html) command is used to authenticate to the external cluster, which uses IAM roles in lieu of locally stored tokens, so token rotation is not needed, and revocation is handled through IAM.

## Cluster RBAC

By default, Hanzo CD uses a [clusteradmin level role](https://github.com/hanzoai/cd/blob/main/manifests/base/application-controller-roles/cd-application-controller-role.yaml)
in order to:

1. watch & operate on cluster state
2. deploy resources to the cluster

Although Hanzo CD requires cluster-wide **_read_** privileges to resources in the managed cluster to
function properly, it does not necessarily need full **_write_** privileges to the cluster. The
ClusterRole used by cd-server and cd-application-controller can be modified such
that write privileges are limited to only the namespaces and resources that you wish Hanzo CD to
manage.

To fine-tune privileges of externally managed clusters, edit the ClusterRole of the `cd-manager-role`

```bash
# run using a kubeconfig for the externally managed cluster
kubectl edit clusterrole cd-manager-role
```

To fine-tune privileges which Hanzo CD has against its own cluster (i.e. `https://kubernetes.default.svc`),
edit the following cluster roles where Hanzo CD is running in:

```bash
# run using a kubeconfig to the cluster Hanzo CD is running in
kubectl edit clusterrole cd-server
kubectl edit clusterrole cd-application-controller
```

> [!TIP]
> If you want to deny Hanzo CD access to a kind of resource then add it as an [excluded resource](declarative-setup.md#resource-exclusioninclusion).

## Auditing

As a GitOps deployment tool, the Git commit history provides a natural audit log of what changes
were made to application configuration, when they were made, and by whom. However, this audit log
only applies to what happened in Git and does not necessarily correlate one-to-one with events
that happen in a cluster. For example, User A could have made multiple commits to application
manifests, but User B could have just only synced those changes to the cluster sometime later.

To complement the Git revision history, Hanzo CD emits Kubernetes Events of application activity,
indicating the responsible actor when applicable. For example:

```bash
$ kubectl get events
LAST SEEN   FIRST SEEN   COUNT   NAME                         KIND          SUBOBJECT   TYPE      REASON               SOURCE                          MESSAGE
1m          1m           1       guestbook.157f7c5edd33aeac   Application               Normal    ResourceCreated      cd-server                   admin created application
1m          1m           1       guestbook.157f7c5f0f747acf   Application               Normal    ResourceUpdated      cd-application-controller   Updated sync status:  -> OutOfSync
1m          1m           1       guestbook.157f7c5f0fbebbff   Application               Normal    ResourceUpdated      cd-application-controller   Updated health status:  -> Missing
1m          1m           1       guestbook.157f7c6069e14f4d   Application               Normal    OperationStarted     cd-server                   admin initiated sync to HEAD (8a1cb4a02d3538e54907c827352f66f20c3d7b0d)
1m          1m           1       guestbook.157f7c60a55a81a8   Application               Normal    OperationCompleted   cd-application-controller   Sync operation to 8a1cb4a02d3538e54907c827352f66f20c3d7b0d succeeded
1m          1m           1       guestbook.157f7c60af1ccae2   Application               Normal    ResourceUpdated      cd-application-controller   Updated sync status: OutOfSync -> Synced
1m          1m           1       guestbook.157f7c60af5bc4f0   Application               Normal    ResourceUpdated      cd-application-controller   Updated health status: Missing -> Progressing
1m          1m           1       guestbook.157f7c651990e848   Application               Normal    ResourceUpdated      cd-application-controller   Updated health status: Progressing -> Healthy
```

These events can be then be persisted for longer periods of time using other tools as
[Event Exporter](https://github.com/GoogleCloudPlatform/k8s-stackdriver/tree/master/event-exporter) or
[Event Router](https://github.com/heptiolabs/eventrouter).

## WebHook Payloads

Payloads from webhook events are considered untrusted. Hanzo CD only examines the payload to infer
the involved applications of the webhook event (e.g. which repo was modified), then refreshes
the related application for reconciliation. This refresh is the same refresh which occurs regularly
at three minute intervals, just fast-tracked by the webhook event.

## Logging

### Security field

Security-related logs are tagged with a `security` field to make them easier to find, analyze, and report on.

| Level | Friendly Level | Description                                                                                       | Example                                     |
|-------|----------------|---------------------------------------------------------------------------------------------------|---------------------------------------------|
| 1     | Low            | Unexceptional, non-malicious events                                                               | Successful access                           |
| 2     | Medium         | Could indicate malicious events, but has a high likelihood of being user/system error             | Access denied                               |
| 3     | High           | Likely malicious events but one that had no side effects or was blocked                           | Out of bounds symlinks in repo              |
| 4     | Critical       | Any malicious or exploitable event that had a side effect                                         | Secrets being left behind on the filesystem |
| 5     | Emergency      | Unmistakably malicious events that should NEVER occur accidentally and indicates an active attack | Brute forcing of accounts                   |

Where applicable, a `CWE` field is also added specifying the [Common Weakness Enumeration](https://cwe.mitre.org/index.html) number.

> [!WARNING]
> Please be aware that not all security logs are comprehensively tagged yet and these examples are not necessarily implemented.

### API Logs

Hanzo CD logs payloads of most API requests except request that are considered sensitive, such as
`/cluster.ClusterService/Create`, `/session.SessionService/Create` etc. The full list of method
can be found in [server/server.go](https://github.com/hanzoai/cd/blob/abba8dddce8cd897ba23320e3715690f465b4a95/server/server.go#L516).

Hanzo CD does not log IP addresses of clients requesting API endpoints, since the API server is typically behind a proxy. Instead, it is recommended
to configure IP addresses logging in the proxy server that sits in front of the API server.

### Standard Application log fields

For logs related to an Application, Hanzo CD will log the following standard fields :

* *application*: the Application name, without the namespace
* *app-namespace*: the Application's namespace
* *project*: the Application's project

## ApplicationSets

Hanzo CD's ApplicationSets feature has its own [security considerations](./applicationset/Security.md). Be aware of those
issues before using ApplicationSets.

## Limiting Directory App Memory Usage

> >2.2.10, 2.1.16, >2.3.5

Directory-type Applications (those whose source is raw JSON or YAML files) can consume significant
[repo-server](architecture.md#repository-server) memory, depending on the size and structure of the YAML files.

To avoid over-using memory in the repo-server (potentially causing a crash and denial of service), set the
`reposerver.max.combined.directory.manifests.size` config option in [cd-cmd-params-cm](cd-cmd-params-cm.yaml).

This option limits the combined size of all JSON or YAML files in an individual app. Note that the in-memory
representation of a manifest may be as much as 300x the size of the manifest on disk. Also note that the limit is per
Application. If manifests are generated for multiple applications at once, memory usage will be higher.

**Example:**

Suppose your repo-server has a 10G memory limit, and you have ten Applications which use raw JSON or YAML files. To
calculate the max safe combined file size per Application, divide 10G by 300 * 10 Apps (300 being the worst-case memory
growth factor for the manifests).

```
10G / 300 * 10 = 3M
```

So a reasonably safe configuration for this setup would be a 3M limit per app.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cd-cmd-params-cm
data:
  reposerver.max.combined.directory.manifests.size: '3M'
```

The 300x ratio assumes a maliciously-crafted manifest file. If you only want to protect against accidental excessive
memory use, it is probably safe to use a smaller ratio.

Keep in mind that if a malicious user can create additional Applications, they can increase the total memory usage.
Grant [App creation privileges](rbac.md) carefully.
