# Jsonnet

Any file matching `*.jsonnet` in a directory app is treated as a Jsonnet file. Hanzo CD evaluates the Jsonnet and is able to parse a generated object or array.

## Build Environment

Jsonnet apps have access to the [standard build environment](build-environment.md) via substitution into *TLAs* and *external variables*.
It is also possible to add a shared library (e.g. `vendor` folder) relative to the repository root.

E.g. via the CLI:

```bash
cd app create APPNAME \
  --jsonnet-ext-var-str 'app=${CD_APP_NAME}' \
  --jsonnet-tla-str 'ns=${CD_APP_NAMESPACE}' \
  --jsonnet-libs 'vendor'
```

Or by declarative syntax:

```yaml
  directory:
    jsonnet:
      extVars:
        - name: app
          value: $CD_APP_NAME
      tlas:
        - name: ns
          value: $CD_APP_NAMESPACE
      libs:
        - vendor
```
