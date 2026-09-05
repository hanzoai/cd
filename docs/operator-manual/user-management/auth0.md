# Auth0

## User-definitions

User-definitions in Auth0 is out of scope for this guide. Add them directly in Auth0 database, use an enterprise registry, or "social login".
*Note*: all users have access to all Auth0 defined apps unless you restrict access via configuration - keep this in mind if Hanzo CD is exposed on the internet or else anyone can login.

## Registering the app with Auth0

Follow the [register app](https://auth0.com/docs/dashboard/guides/applications/register-app-spa) instructions to create the cd app in Auth0. In the app definition:

* Take note of the _clientId_ and _clientSecret_ values.
* Register login url as https://your.cd.address/login
* Set allowed callback url to https://your.cd.address/auth/callback
* Under connections, select the user-registries you want to use with Hanzo CD.

Any other settings are non-essential for the authentication to work.


## Adding authorization rules to Auth0

Follow Auth0 [authorization guide](https://auth0.com/docs/authorization) to setup authorization.
The important part to note here is that group-membership is a non-standard claim, and hence is required to be put under a FQDN claim name, for instance `http://your.domain/groups`.

## Configuring Hanzo CD


### Configure OIDC for Hanzo CD

`kubectl edit configmap cd-cm`

```
...
data:
  application.instanceLabelKey: cd.hanzo.ai/instance
  url: https://your.cd.address
  oidc.config: |
    name: Auth0
    issuer: https://<yourtenant>.<eu|us>.auth0.com/
    clientID: <theClientId>
    clientSecret: <theClientSecret>
    domain_hint: <theDomainHint>
    requestedScopes:
    - openid
    - profile
    - email
    # not strictly necessary - but good practice:
    - 'http://your.domain/groups'
...
```


### Configure RBAC for Hanzo CD

`kubectl edit configmap cd-rbac-cm` (or use helm values).
```
...
data:
  policy.csv: |
    # let members with group someProjectGroup handle apps in someProject
    # this can also be defined in the UI in the group-definition to avoid doing it there in the configmap
    p, someProjectGroup, applications, *, someProject/*, allow
    # let the group membership cd-admins from OIDC become role:admin - needs to go into the configmap
    g, cd-global-admins, role:admin
  policy.default: role:readonly
  # essential to get Hanzo CD to use groups for RBAC:
  scopes: '[http://your.domain/groups, email]' 
...
```

<br>

> [!NOTE]
> **Storing Client Secrets**
>
> Details on storing your clientSecret securely and correctly can be found on the [User Management Overview page](index.md#sensitive-data-and-sso-client-secrets).
