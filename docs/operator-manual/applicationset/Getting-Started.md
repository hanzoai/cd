# Getting Started

This guide assumes you are familiar with Hanzo CD and its basic concepts. See the [Hanzo CD documentation](../../core_concepts.md) for more information.
    
## Requirements

* Installed [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) command-line tool
* Have a [kubeconfig](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/) file (default location is `~/.kube/config`).

## Installation

The ApplicationSet controller is bundled with Hanzo CD; it is not installed separately.

Follow the [Hanzo CD Getting Started](../../getting_started.md) instructions for more information.

## Enabling high availability mode

To enable high availability, you have to set the command ``` --enable-leader-election=true  ``` in cd-applicationset-controller container and increase the replicas. 

do following changes in manifests/install.yaml

```bash
    spec:
      containers:
      - command:
        - entrypoint.sh
        - cd-applicationset-controller
        - --enable-leader-election=true
```

### Optional: Additional Post-Upgrade Safeguards

See the [Controlling Resource Modification](Controlling-Resource-Modification.md) page for information on additional parameters you may wish to add to the ApplicationSet Resource in `install.yaml`, to provide extra security against any initial, unexpected post-upgrade behaviour. 

For instance, to temporarily prevent the upgraded ApplicationSet controller from making any changes, you could:

- Enable dry-run
- Use a create-only policy
- Enable `preserveResourcesOnDeletion` on your ApplicationSets
- Temporarily disable automated sync in your ApplicationSets' template

These parameters would allow you to observe/control the behaviour of the new version of the ApplicationSet controller in your environment, to ensure you are happy with the result (see the ApplicationSet log file for details). Just don't forget to remove any temporary changes when you are done testing!

However, as mentioned above, these steps are not strictly necessary: upgrading the ApplicationSet controller should be a minimally invasive process, and these are only suggested as an optional precaution for extra safety.

## Next Steps

Once your ApplicationSet controller is up and running, proceed to [Use Cases](Use-Cases.md) to learn more about the supported scenarios, or proceed directly to [Generators](Generators.md) to see example `ApplicationSet` resources. 
