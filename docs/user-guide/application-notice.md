# Application Notice

Hanzo CD can display a per-application notice as a contextual banner on the
application details page and as a hover icon on the application list and tile
views. This is useful for surfacing deprecation warnings, scheduled
maintenance, or operational guidance scoped to a single Application without
affecting unrelated apps.

The notice is configured entirely through annotations on the Application
resource — no Hanzo CD restart or ConfigMap change is required.

> [!NOTE]
> The instance-level banner configured via `ui.bannercontent` in `cd-cm`
> remains unchanged and continues to apply to every UI page. The
> per-application notice described here is independent and stacks above the
> details-page content.

## Annotations

All notice annotations use the `notice.apps.hanzo.ai/` prefix.

| Annotation                                  | Required | Default | Description                                                                                              |
|---------------------------------------------|----------|---------|----------------------------------------------------------------------------------------------------------|
| `notice.apps.hanzo.ai/content`         | Yes      | —       | Plain-text notice message. Truncated at 500 characters.                                                  |
| `notice.apps.hanzo.ai/severity`        | No       | `info`  | One of `info`, `warning`, `critical`. Drives banner color and the default icon. Unknown values fall back to `info`. |
| `notice.apps.hanzo.ai/url`             | No       | —       | If set, the banner text and icon link to this URL. Must be `http://` or `https://`. Other schemes are rejected. |
| `notice.apps.hanzo.ai/permanent`       | No       | `false` | When `"true"`, the banner cannot be dismissed by users. Mirrors `ui.bannerpermanent` semantics.          |
| `notice.apps.hanzo.ai/icon`            | No       | severity-derived | Override the icon. Must be one of the allowlisted FontAwesome classes; unknown values fall back to the severity default. |
| `notice.apps.hanzo.ai/scope`           | No       | `both`  | One of `banner`, `icon`, `both`. Use to suppress either surface.                                         |

### Severity defaults

| Severity   | Default icon                |
|------------|-----------------------------|
| `info`     | `fa-info-circle`            |
| `warning`  | `fa-exclamation-triangle`   |
| `critical` | `fa-exclamation-circle`     |

### Allowlisted icons

Only these FontAwesome class fragments may be set via the `icon` annotation:

`fa-info-circle`, `fa-exclamation-triangle`, `fa-exclamation-circle`,
`fa-bell`, `fa-wrench`, `fa-clock`, `fa-bullhorn`, `fa-life-ring`,
`fa-shield-alt`.

## Examples

Minimal notice — info severity, dismissible, banner + icon:

```yaml
apiVersion: apps.hanzo.ai/v1alpha1
kind: Application
metadata:
  name: guestbook
  namespace: cd
  annotations:
    notice.apps.hanzo.ai/content: "Scheduled DB maintenance on 2026-06-01 02:00 UTC"
spec:
  # ...
```

Deprecation notice with a runbook link, marked permanent:

```yaml
metadata:
  annotations:
    notice.apps.hanzo.ai/content:   "Deprecated — migrate to v2 by 2026-08-01"
    notice.apps.hanzo.ai/severity:  "warning"
    notice.apps.hanzo.ai/url:       "https://wiki.example.com/v2-migration"
    notice.apps.hanzo.ai/permanent: "true"
```

List-icon-only hint (no page banner):

```yaml
metadata:
  annotations:
    notice.apps.hanzo.ai/content:  "Owned by team-payments"
    notice.apps.hanzo.ai/severity: "info"
    notice.apps.hanzo.ai/scope:    "icon"
    notice.apps.hanzo.ai/icon:     "fa-life-ring"
```

## Where it appears

- The banner is shown on the application details page, between the health
  and sync status panel and the application's resource tree.
- The icon is shown next to the application name in both the tile and list
  views of the applications page; hovering it reveals the notice content.
- By default the banner can be dismissed for the current user. Setting
  `notice.apps.hanzo.ai/permanent` to `"true"` removes the dismiss
  control so the banner stays visible. The list/tile icon is always shown
  while the notice exists.
