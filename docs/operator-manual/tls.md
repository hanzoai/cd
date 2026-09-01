# TLS configuration

> [!TIP]
> Need repo-server mutual TLS between components? See [Mutual TLS (mTLS) for repo-server](./mtls.md).

Hanzo CD provides three inbound TLS endpoints that can be configured:

* The user-facing endpoint of the `cd-server` workload, which serves the UI
  and the API
* The endpoint of the `cd-repo-server`, which is accessed by `cd-server`
  and `cd-application-controller` workloads to request repository
  operations.
* The endpoint of the `cd-dex-server`, which is accessed by `cd-server`
  to handle OIDC authentication.

By default, and without further configuration, these endpoints will be
set up to use an automatically generated, self-signed certificate. However,
most users will want to explicitly configure the certificates for these TLS
endpoints, possibly using automated means such as `cert-manager` or using
their own dedicated Certificate Authority.

## TLS Configuration Quick Reference

### Certificate Configuration Overview

| Component | Secret Name | Hot Reload | Default Cert | Required SAN Entries |
|-----------|-------------|------------|---------------|---------------------|
| `cd-server` | `cd-server-tls` | ✅ Yes | Self-signed | External hostname (e.g., `cd.example.com`) |
| `cd-repo-server` | `cd-repo-server-tls` | ❌ Restart required | Self-signed | `DNS:cd-repo-server`, `DNS:cd-repo-server.cd.svc` |
| `cd-dex-server` | `cd-dex-server-tls` | ❌ Restart required | Self-signed | `DNS:cd-dex-server`, `DNS:cd-dex-server.cd.svc` |


| Connection | Recommended Parameter | Legacy Parameter (deprecated) | Plain Text Parameter | Default Behavior |
|------------|----------------------|-------------------------------|---------------------|------------------|
| `cd-server` → `cd-repo-server` | `--repo-server-ca-cert-path` | `--repo-server-strict-tls` | `--repo-server-plaintext` | Non-validating TLS |
| `cd-server` → `cd-dex-server` | — | `--dex-server-strict-tls` | `--dex-server-plaintext` | Non-validating TLS |
| `cd-application-controller` → `cd-repo-server` | `--repo-server-ca-cert-path` | `--repo-server-strict-tls` | `--repo-server-plaintext` | Non-validating TLS |
| `cd-applicationset-controller` → `cd-repo-server` | `--repo-server-ca-cert-path` | `--repo-server-strict-tls` | `--repo-server-plaintext` | Non-validating TLS |
| `cd-notifications-controller` → `cd-repo-server` | `--cd-repo-server-ca-cert-path` | `--cd-repo-server-strict-tls` | `--cd-repo-server-plaintext` | Non-validating TLS |

### Certificate Priority (cd-server only)

1. `cd-server-tls` secret (recommended)
2. `cd-secret` secret (deprecated) 
3. Auto-generated self-signed certificate

## Configuring TLS for cd-server

### Inbound TLS options for cd-server

You can configure certain TLS options for the `cd-server` workload by
setting command line parameters. The following parameters are available:

|Parameter|Default|Description|
|---------|-------|-----------|
|`--insecure`|`false`|Disables TLS completely|
|`--tlsminversion`|`1.2`|The minimum TLS version to be offered to clients|
|`--tlsmaxversion`|`1.3`|The maximum TLS version to be offered to clients|
|`--tlsciphers`|`TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:TLS_RSA_WITH_AES_256_GCM_SHA384`|A colon separated list of TLS cipher suites to be offered to clients|

### TLS certificates used by cd-server

There are two ways to configure the TLS certificates used by `cd-server`:

* Setting the `tls.crt` and `tls.key` keys in the `cd-server-tls` secret
  to hold PEM data of the certificate and the corresponding private key. The
  `cd-server-tls` secret may be of type `tls`, but does not have to be.
* Setting the `tls.crt` and `tls.key` keys in the `cd-secret` secret to
  hold PEM data of the certificate and the corresponding private key. This
  method is considered deprecated and only exists for purposes of backwards
  compatibility. Changing `cd-secret` should not be used to override the
  TLS certificate anymore.

Hanzo CD decides which TLS certificate to use for the endpoint of
`cd-server` as follows:

* If the `cd-server-tls` secret exists and contains a valid key pair in the
  `tls.crt` and `tls.key` keys, this will be used for the certificate of the
  endpoint of `cd-server`.
* Otherwise, if the `cd-secret` secret contains a valid key pair in the
 `tls.crt` and `tls.key` keys, this will be used as the certificate for the
  endpoint of `cd-server`.
* If no `tls.crt` and `tls.key` keys are found in neither of the two mentioned
  secrets, Hanzo CD will generate a self-signed certificate and persist it in
  the `cd-secret` secret.

The `cd-server-tls` secret contains only information for TLS configuration
to be used by `cd-server` and is safe to be managed via third-party tools
such as `cert-manager` or `SealedSecrets`

To create this secret manually from an existing key pair, you can use `kubectl`:

```shell
kubectl create -n cd secret tls cd-server-tls \
  --cert=/path/to/cert.pem \
  --key=/path/to/key.pem
```

Hanzo CD will pick up changes to the `cd-server-tls` secret automatically
and will not require restarting to use a renewed certificate.

## Configuring inbound TLS for cd-repo-server

### Inbound TLS options for cd-repo-server

You can configure certain TLS options for the `cd-repo-server` workload by
setting command line parameters. The following parameters are available:

|Parameter|Default|Description|
|---------|-------|-----------|
|`--disable-tls`|`false`|Disables TLS completely|
|`--tlsminversion`|`1.2`|The minimum TLS version to be offered to clients|
|`--tlsmaxversion`|`1.3`|The maximum TLS version to be offered to clients|
|`--tlsciphers`|`TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:TLS_RSA_WITH_AES_256_GCM_SHA384`|A colon-separated list of TLS cipher suites to be offered to clients|

### Inbound TLS certificates used by cd-repo-server

To configure the TLS certificate used by the `cd-repo-server` workload,
create a secret named `cd-repo-server-tls` in the namespace where Hanzo CD
is running in with the certificate's key pair stored in `tls.crt` and
`tls.key` keys. If this secret does not exist, `cd-repo-server` will
generate and use a self-signed certificate.

To create this secret, you can use `kubectl`:

```shell
kubectl create -n cd secret tls cd-repo-server-tls \
  --cert=/path/to/cert.pem \
  --key=/path/to/key.pem
```

If the certificate is self-signed, you will also need to add `ca.crt` to the secret
with the contents of your CA certificate.

Please note, that as opposed to `cd-server`, the `cd-repo-server` is
not able to pick up changes to this secret automatically. If you create (or
update) this secret, the `cd-repo-server` pods need to be restarted.

Also note, that the certificate should be issued with the correct SAN entries
for the `cd-repo-server`, containing at least the entries for
`DNS:cd-repo-server` and `DNS:cd-repo-server.argo-cd.svc` depending
on how your workloads connect to the repository server.

## Configuring inbound TLS for cd-dex-server

### Inbound TLS options for cd-dex-server

You can configure certain TLS options for the `cd-dex-server` workload by
setting command line parameters. The following parameters are available:

|Parameter|Default|Description|
|---------|-------|-----------|
|`--disable-tls`|`false`|Disables TLS completely|

### Inbound TLS certificates used by cd-dex-server

To configure the TLS certificate used by the `cd-dex-server` workload,
create a secret named `cd-dex-server-tls` in the namespace where Hanzo CD
is running in with the certificate's key pair stored in `tls.crt` and
`tls.key` keys. If this secret does not exist, `cd-dex-server` will
generate and use a self-signed certificate.

To create this secret, you can use `kubectl`:

```shell
kubectl create -n cd secret tls cd-dex-server-tls \
  --cert=/path/to/cert.pem \
  --key=/path/to/key.pem
```

If the certificate is self-signed, you will also need to add `ca.crt` to the secret
with the contents of your CA certificate.

Please note, that as opposed to `cd-server`, the `cd-dex-server` is
not able to pick up changes to this secret automatically. If you create (or
update) this secret, the `cd-dex-server` pods need to be restarted.

Also note, that the certificate should be issued with the correct SAN entries
for the `cd-dex-server`, containing at least the entries for
`DNS:cd-dex-server` and `DNS:cd-dex-server.argo-cd.svc` depending
on how your workloads connect to the repository server.

## Configuring TLS between Hanzo CD components

### Configuring TLS to cd-repo-server

The components `cd-server`, `cd-application-controller`, `cd-notifications-controller`, 
and `cd-applicationset-controller` communicate with the `cd-repo-server` 
using a gRPC API over TLS. By default, `cd-repo-server` generates a non-persistent, 
self-signed certificate to use for its gRPC endpoint on startup. Because the 
`cd-repo-server` has no means to connect to the K8s control plane API, this certificate 
is not available to outside consumers for verification. These components will use a 
non-validating connection to the `cd-repo-server` for this reason.

To change this behavior to be more secure by having these components validate the TLS certificate of the
`cd-repo-server` endpoint, the following steps need to be performed:

* Create a persistent TLS certificate to be used by `cd-repo-server`, as
  shown above
* Restart the `cd-repo-server` pod(s)
* Modify the pod startup parameters for `cd-server`, `cd-application-controller`,
  and `cd-applicationset-controller` to include the
  `--repo-server-ca-cert-path` parameter pointing to the CA certificate file.
* Modify the pod startup parameters for `cd-notifications-controller` to include the
  `--cd-repo-server-ca-cert-path` parameter pointing to the CA certificate file.

The `cd-server`, `cd-application-controller`, `cd-notifications-controller`,
and `cd-applicationset-controller` workloads will now
validate the TLS certificate of the `cd-repo-server` using the provided CA certificate.

> [!NOTE]
> **Legacy path vs. recommended path**
>
> `--repo-server-strict-tls` (and `--cd-repo-server-strict-tls` for the notifications controller)
> is the **legacy path**: when set, the component auto-discovers the repo-server certificate from
> the `cd-repo-server-tls` Kubernetes secret. This flag is **deprecated** and may be removed
> in a future release.
>
> `--repo-server-ca-cert-path` (and `--cd-repo-server-ca-cert-path` for the notifications controller)
> is the **recommended explicit path**: you provide the path to a CA certificate file directly.
> This is required for mTLS setups and gives you full control over which CA is trusted.
> See [Mutual TLS (mTLS) for repo-server](./mtls.md) for details.

> [!NOTE]
> **Certificate expiry**
>
> Please make sure that the certificate has a proper lifetime. Remember, 
> when replacing certificates, all workloads must be restarted to pick up
> the certificate and work properly.

### Configuring TLS to cd-dex-server

`cd-server` communicates with the `cd-dex-server` using an HTTPS API
over TLS. By default, `cd-dex-server` generates a non-persistent, self-signed 
certificate for its HTTPS endpoint on startup. Because `cd-dex-server` 
has no means to connect to the K8s control plane API, this certificate is not 
available to outside consumers for verification. `cd-server` will use a 
non-validating connection to `cd-dex-server` for this reason.

To change this behavior to be more secure by having the `cd-server` validate 
the TLS certificate of the `cd-dex-server` endpoint, the following steps need
to be performed:

* Create a persistent TLS certificate to be used by `cd-dex-server`, as
  shown above
* Restart the `cd-dex-server` pod(s)
* Modify the pod startup parameters for `cd-server` to include the 
`--dex-server-strict-tls` parameter.

The `cd-server` workload will now validate the TLS certificate of the
`cd-dex-server` by using the certificate stored in the `cd-dex-server-tls`
secret.

> [!NOTE]
> **Certificate expiry**
>
> Please make sure that the certificate has a proper lifetime. Remember, 
> when replacing certificates, all workloads must be restarted to pick up
> the certificate and work properly.

### Disabling TLS to cd-repo-server

In some scenarios where mTLS through sidecar proxies is involved (e.g.
in a service mesh), you may want to configure the connections between the
`cd-server`, `cd-application-controller`, `cd-notifications-controller`, 
and `cd-applicationset-controller` to `cd-repo-server`
to not use TLS at all.

In this case, you will need to:

* Configure `cd-repo-server` with TLS on the gRPC API disabled by specifying
  the `--disable-tls` parameter to the pod container's startup arguments.
  Also, consider restricting listening addresses to the loopback interface by specifying
  `--listen 127.0.0.1` parameter, so that the insecure endpoint is not exposed on
  the pod's network interfaces, but still available to the sidecar container.
* Configure `cd-server`, `cd-application-controller`, 
  and `cd-applicationset-controller` to not use TLS
  for connections to the `cd-repo-server` by specifying the parameter
  `--repo-server-plaintext` to the pod container's startup arguments
* Modify the pod startup parameters for `cd-notifications-controller` to include the
  `--cd-repo-server-plaintext` parameter
* Configure `cd-server` and `cd-application-controller` to connect to
  the sidecar instead of directly to the `cd-repo-server` service by
  specifying its address via the `--repo-server <address>` parameter

After this change, `cd-server`, `cd-application-controller`, `cd-notifications-controller`, 
and `cd-applicationset-controller` will
use a plain text connection to the sidecar proxy, which will handle all aspects
of TLS to `cd-repo-server`'s TLS sidecar proxy.

### Disabling TLS to cd-dex-server

In some scenarios where mTLS through sidecar proxies is involved (e.g.
in a service mesh), you may want to configure the connections between
`cd-server` to `cd-dex-server` to not use TLS at all.

In this case, you will need to:

* Configure `cd-dex-server` with TLS on the HTTPS API disabled by specifying
  the `--disable-tls` parameter to the pod container's startup arguments
* Configure `cd-server` to not use TLS for connections to `cd-dex-server` 
  by specifying the parameter `--dex-server-plaintext` to the pod container's startup
  arguments
* Configure `cd-server` to connect to the sidecar instead of directly to the 
  `cd-dex-server` service by specifying its address via the `--dex-server <address>`
  parameter

After this change, `cd-server` will use a plain text connection to the sidecar 
proxy, that will handle all aspects of TLS to the `cd-dex-server`'s TLS sidecar proxy.

