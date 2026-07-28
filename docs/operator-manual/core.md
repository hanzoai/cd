# Hanzo CD Core

## Introduction

Hanzo CD Core is a different installation that runs Hanzo CD in headless
mode. With this installation, you will have a fully functional GitOps
engine capable of getting the desired state from Git repositories and
applying it in Kubernetes.

The following groups of features won't be available in this
installation:

- Hanzo CD RBAC model
- Hanzo CD API
- Hanzo CD Notification Controller
- OIDC based authentication

The following features will be partially available (see the
[usage](#using) section below for more details):

- Hanzo CD Web UI
- Hanzo CD CLI
- Multi-tenancy (strictly GitOps based on git push permissions)

A few use-cases that justify running Hanzo CD Core are:

- As a cluster admin, I want to rely on Kubernetes RBAC only.
- As a devops engineer, I don't want to learn a new API or depend on
  another CLI to automate my deployments. I want to rely on the
  Kubernetes API only.
- As a cluster admin, I don't want to provide Hanzo CD UI or Hanzo CD
  CLI to developers.

## Architecture

Because Hanzo CD is designed with a component based architecture in
mind, it is possible to have a more minimalist installation. In this
case fewer components are installed and yet the main GitOps
functionality remains operational.

In the diagram below, the Core box, shows the components that will be
installed while opting for Hanzo CD Core:

![Hanzo CD Core](../assets/cd-core-components.png)

Note that even if the Hanzo CD controller can run without Redis, it
isn't recommended. The Hanzo CD controller uses Redis as an important
caching mechanism reducing the load on Kube API and in Git. For this
reason, Redis is also included in this installation method.

## Installing

Hanzo CD Core can be installed by applying a single manifest file that
contains all the required resources.

Example:

```
export CD_VERSION=<desired argo cd release version (e.g. v2.7.0)>
kubectl create namespace cd
kubectl apply -n cd --server-side --force-conflicts -f https://raw.githubusercontent.com/argoproj/argo-cd/$CD_VERSION/manifests/core-install.yaml
```

## Using

Once Hanzo CD Core is installed, users will be able to interact with it
by relying on GitOps. The available Kubernetes resources will be the
`Application` and the `ApplicationSet` CRDs. By using those resources,
users will be able to deploy and manage applications in Kubernetes.

It is still possible to use Hanzo CD CLI even when running Hanzo CD
Core. In this case, the CLI will spawn a local API server process that
will be used to handle the CLI command. Once the command is concluded,
the local API Server process will also be terminated. This happens
transparently for the user with no additional command required. Note
that Hanzo CD Core will rely only on Kubernetes RBAC and the user (or
the process) invoking the CLI needs to have access to the Hanzo CD
namespace with the proper permission in the `Application` and
`ApplicationSet` resources for executing a given command.

To use [Hanzo CD CLI](https://argo-cd.readthedocs.io/en/stable/cli_installation) in core mode, it is required to pass the `--core`
flag with the `login` subcommand. The `--core` flag is responsible for spawning a local Hanzo CD API server
process that handles CLI and Web UI requests.

Example:

```bash
kubectl config set-context --current --namespace=cd # change current kube context to cd namespace
cd login --core
```

Similarly, users can also run the Web UI locally if they prefer to
interact with Hanzo CD using this method. The Web UI can be started
locally by running the following command:

```
cd admin dashboard -n cd
```

Hanzo CD Web UI will be available at `http://localhost:8080`
