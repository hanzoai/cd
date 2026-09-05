# Disaster Recovery

You can use `cd admin` to import and export all Hanzo CD data.

Make sure you have `~/.kube/config` pointing to your Hanzo CD cluster.

Figure out what version of Hanzo CD you're running:

```bash
cd version | grep server
# ...
export VERSION=v1.0.1
```

Export to a backup:

```bash
docker run -v ~/.kube:/home/cd/.kube --rm ghcr.io/hanzoai/cd:$VERSION cd admin export > backup.yaml
```

Import from a backup:

```bash
docker run -i -v ~/.kube:/home/cd/.kube --rm ghcr.io/hanzoai/cd:$VERSION cd admin import - < backup.yaml
```

> [!NOTE]
> If you are running Hanzo CD on a namespace different than default remember to pass the namespace parameter (-n <namespace>). 'cd admin export' will not fail if you run it in the wrong namespace.
