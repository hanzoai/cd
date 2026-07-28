## Additional configuration methods

Additional configuration methods for configuring commands `cd-server`, `cd-repo-server` and `cd-application-controller`.


### Synopsis

The commands can also be configured by setting the respective flag of the available options in `cd-cmd-params-cm.yaml`. Each component has a specific prefix associated with it.

```
cd-server                 --> server
cd-repo-server            --> reposerver
cd-application-controller --> controller
```

The flags that do not have a prefix are shared across multiple components. One such flag is `repo.server`
The list of flags that are available can be found in [cd-cmd-params-cm.yaml](../cd-cmd-params-cm.yaml) 


### Example

To set `logformat` of `cd-application-controller`, add below entry to the config map `cd-cmd-params-cm.yaml`.

```
data:
    controller.log.format: "json"
```

