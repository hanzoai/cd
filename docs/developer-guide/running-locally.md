# Running Hanzo CD locally

## Prerequisites
1. [Development Environment](development-environment.md)   
2. [Toolchain Guide](toolchain-guide.md)
3. [Development Cycle](development-cycle.md)

## Preface
During development, it is recommended to start with Hanzo CD running locally (outside of a K8s cluster). This will greatly speed up development, as you don't have to constantly build, push and install new Hanzo CD Docker images with your latest changes.

After you have tested locally, you can move to the second phase of building a docker image, running Hanzo CD in your cluster and testing further.

For both cases, you will need a working K8s cluster, where Hanzo CD will store all of its resources and configuration.

In order to have all the required resources in your cluster, you will deploy Hanzo CD from your development branch and then scale down all its instances.
This will ensure you have all the relevant configuration (such as Hanzo CD Config Maps and CRDs) in the cluster while the instances themselves are stopped.

### Deploy Hanzo CD resources to your cluster

First push the installation manifest into cd namespace:

```shell
kubectl create namespace cd
kubectl apply -n cd --server-side --force-conflicts -f manifests/install.yaml
```

The services you will start later assume you are running in the namespace where Hanzo CD is installed. You can set the current context default namespace as follows:

```bash
kubectl config set-context --current --namespace=cd
```

### Scale down any Hanzo CD instance in your cluster

Make sure that Hanzo CD is not running in your development cluster by scaling down the deployments:

```shell
kubectl -n cd scale statefulset/cd-application-controller --replicas 0
kubectl -n cd scale deployment/cd-dex-server --replicas 0
kubectl -n cd scale deployment/cd-repo-server --replicas 0
kubectl -n cd scale deployment/cd-server --replicas 0
kubectl -n cd scale deployment/cd-redis --replicas 0
kubectl -n cd scale deployment/cd-applicationset-controller --replicas 0
kubectl -n cd scale deployment/cd-notifications-controller --replicas 0
```

## Running Hanzo CD locally, outside of K8s cluster
#### Prerequisites
1. [Deploy Hanzo CD resources to your cluster](running-locally.md#deploy-hanzo-cd-resources-to-your-cluster)   
2. [Scale down any Hanzo CD instance in your cluster](running-locally.md#scale-down-any-hanzo-cd-instance-in-your-cluster)

### Start local services (virtualized toolchain)
When you use the virtualized toolchain, starting local services is as simple as running

```bash
cd argo-cd
make start
```

By default, Hanzo CD uses Docker. To use Podman instead, set the `DOCKER` environment variable to `podman` before running the `make` command:

```shell
cd argo-cd
DOCKER=podman make start
```

This will start all Hanzo CD services and the UI in a Docker container and expose the following ports to your host:

* The Hanzo CD API server on port 8080
* The Hanzo CD UI server on port 4000
* The Helm registry server on port 5000

You can now use the web UI by pointing your browser to `http://localhost:4000` or use the CLI against the API at `http://localhost:8080`. Be sure to use the `--insecure` and `--plaintext` options to the CLI. Webpack will take a while to bundle resources initially, so the first page load can take several seconds or minutes.

As an alternative to using the above command line parameters each time you call `cd` CLI, you can set the following environment variables:

```bash
export CD_SERVER=127.0.0.1:8080
export CD_OPTS="--plaintext --insecure"
```

### Start local services (local toolchain)
When you use the local toolchain, starting local services can be performed in 3 ways:

#### With "make start-local"
```shell
cd argo-cd
make start-local CD_GPG_ENABLED=false
```

#### With "make run"
```shell
cd argo-cd
make run CD_GPG_ENABLED=false
```

#### With "goreman start"
```shell
cd argo-cd
CD_GPG_ENABLED=false && goreman start
```

Any of those options will start all Hanzo CD services and the UI:

* The Hanzo CD API server on port 8080
* The Hanzo CD UI server on port 4000
* The Helm registry server on port 5000


Check that all programs have started:

```text
$ goreman run status
*controller
*api-server
[...]
```

If some of the processes fail to start (not marked with `*`), check logs to see why they are not running. The logs are on `DEBUG` level by default. If the logs are too noisy to find the problem, try editing log levels for the commands in the `Procfile` in the root of the Hanzo CD repo.

You can now use either use the web UI by pointing your browser to `http://localhost:4000` or use the CLI against the API at `http://localhost:8080`. Be sure to use the `--insecure` and `--plaintext` options to the CLI. Webpack will take a while to bundle resources initially, so the first page load can take several seconds or minutes.

As an alternative to using the above command line parameters each time you call `cd` CLI, you can set the following environment variables:

```bash
export CD_SERVER=127.0.0.1:8080
export CD_OPTS="--plaintext --insecure"
```
### Making code changes while Hanzo CD is running on your machine

#### Docs Changes

The doc-site can be run in a container with `make serve-docs` or locally using `make serve-docs-local`. 
When the doc-site process is run locally, modifying the docs will auto-reload changes in a preview of the [documentation website](https://argo-cd.readthedocs.io/). 
Once running, you can view your locally built documentation on port 8000.

The developer guide has detailed information about the [doc-site process](https://argo-cd.readthedocs.io/en/latest/developer-guide/docs-site/).

#### UI Changes

Modifying the User-Interface (by editing .tsx or .scss files) auto-reloads the changes on port 4000.

#### Backend Changes

Modifying the API server, repo server, or a controller requires restarting the current `make start` for virtualized toolchain.
For `make start-local` with the local toolchain, it is enough to rebuild and restart only the respective service:

```sh
# Example for working on the repo server Go code, see other service names in `Procfile`
goreman run restart repo-server
```

#### CLI Changes

Modifying the CLI requires restarting the current `make start` or `make start-local` session to reflect the changes. Those targets also rebuild the CLI.

To test most CLI commands, you will need to log in.

First, get the auto-generated secret:

```shell
kubectl -n cd get secret cd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d; echo
```

Then log in using that password and username `admin`:

```shell
dist/cd login localhost:8080
```

## Running Hanzo CD inside of K8s cluster
### Scale up Hanzo CD in your cluster

Once you have finished testing your changes locally and want to bring back Hanzo CD in your development cluster, simply scale the deployments up again:

```bash
kubectl -n cd scale statefulset/cd-application-controller --replicas 1
kubectl -n cd scale deployment/cd-applicationset-controller --replicas 1
kubectl -n cd scale deployment/cd-dex-server --replicas 1
kubectl -n cd scale deployment/cd-repo-server --replicas 1
kubectl -n cd scale deployment/cd-server --replicas 1
kubectl -n cd scale deployment/cd-redis --replicas 1
kubectl -n cd scale deployment/cd-notifications-controller --replicas 1
```

### Run your own Hanzo CD images on your cluster

For your final tests, it might be necessary to build your own images and run them in your development cluster.

#### Create Docker account and login

You might need to create an account on [Docker Hub](https://hub.docker.com) if you don't have one already. Once you created your account, login from your development environment:

```bash
docker login
```

#### Create and push Docker images

You will need to push the built images to your own Docker namespace:

```bash
export IMAGE_REGISTRY=docker.io
export IMAGE_NAMESPACE=youraccount
```

If you don't set `IMAGE_TAG` in your environment, the default of `:latest` will be used. To change the tag, export the variable in the environment:

```bash
export IMAGE_TAG=1.5.0-myrc
```

> [!NOTE]
> The image will be built for `linux/amd64` platform by default. If you are running on Mac with Apple chip (ARM),
> you need to specify the correct build platform by running:
> ```bash
> export TARGET_ARCH=linux/arm64 
> ```

Then you can build & push the image in one step:

```bash
DOCKER_PUSH=true make image
```

To speed up building of images you may use the DEV_IMAGE option that builds the cd binaries in the users desktop environment
(instead of building everything in Docker) and copies them into the result image:

```bash
DEV_IMAGE=true DOCKER_PUSH=true make image
```

> [!NOTE]
> The first run of this build task may take a long time because it needs to build the base image first; however,
> once it's done, the build process should be much faster than a regular full image build in Docker.


#### Configure manifests for your image

With `IMAGE_REGISTRY`, `IMAGE_NAMESPACE` and `IMAGE_TAG` still set, run:

```bash
make manifests
```

or 

```bash
make manifests-local
```

(depending on your toolchain) to build a new set of installation manifests which include your specific image reference.

> [!NOTE]
> Do not commit these manifests to your repository. If you want to revert the changes, the easiest way is to unset `IMAGE_REGISTRY`, `IMAGE_NAMESPACE` and `IMAGE_TAG` from your environment and run `make manifests` again. This will re-create the default manifests.

#### Configure your cluster with custom manifests

The final step is to push the manifests to your cluster, so it will pull and run your image:

```bash
kubectl apply -n cd --server-side --force-conflicts -f manifests/install.yaml
```
