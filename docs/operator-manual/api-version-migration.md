# Migrating the CRD API version

`apps.hanzo.ai/v1alpha1` → `apps.hanzo.ai/v1`.

The current state is the correct intermediate one: **both versions served, `v1`
is storage, every resource migrated to `v1` storage.** Do not remove
`v1alpha1` until the running binary speaks `v1`.

## The order, and why it is this order

1. Add `v1` to `spec.versions`, `served: true`, `storage: false`.
   Non-destructive. `conversion.strategy: None` is correct because the schemas
   are identical — this is a rename, not a shape change.
2. Flip `storage: true` to `v1`.
3. Rewrite every resource so etcd stores it at `v1` (any no-op patch).
4. Prune `status.storedVersions` to `["v1"]`. Only safe after step 3.
5. **Deploy the binary that speaks `v1`.**
6. Only now remove `v1alpha1` from `spec.versions`.

## What happens if you do 6 before 5

This was done on 2026-07-28 and took the control plane down.

The running controller (`sha-7f2d7a9`, built from the `v1alpha1` Go types)
listed `*v1alpha1.Application` against an API that no longer served it:

```
failed to list *v1alpha1.Application: the server could not find the
requested resource (get applications.apps.hanzo.ai)
```

The controller CrashLooped, and 65 Applications plus the ApplicationSet were
lost. **The workloads survived** — 159 pods stayed Running and the
`resources-finalizer.apps.hanzo.ai` on 63 Applications never fired, which is
the only reason this was recoverable rather than a fleet deletion.

Recovery was: restore `v1alpha1` as served, restart the controller, re-apply
`applicationset-fleet.yaml` (regenerates 62) and the four standalone
Applications. Everything was reconstructible from git, which is the property
that made the blast radius survivable.

The lesson is narrow and worth stating plainly: the CRD migration steps were
all individually correct. What was never checked was **which version the
running consumer speaks**. Verify the thing that depends on the API, not only
the API being changed.
