// resource.go — GET /v1/deploy/{name}/resource/{ref}: one tree node's live
// manifest plus a desired-vs-live diff.
//
// {ref} is the canonical "group:kind:namespace:name" token the tree emits on each
// node, so the console round-trips it back verbatim. The kind must be in the
// closed registry (kindGVR) and the namespace a platform namespace, and the object
// must belong to {name}'s tree (it IS the CR, or carries an ownerRef/label tying it
// to the app) — so the endpoint can never be steered at an arbitrary cluster
// object. Secrets are not in the registry, so their manifests are never returned.
//
// desiredSource: today "last-applied" (the object's kubectl last-applied-config
// annotation) or "none". When git.hanzo.ai becomes the manifest source of truth
// (RegisterPushBuilder → commit → engine sync), desiredSource becomes "git" with
// the SAME diff shape. P2b replaces the field-strip diff with gitops-engine
// pkg/diff (three-way) for exact ArgoCD parity.
package deploy

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/hanzoai/cloud"
	"github.com/zap-proto/zip"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const lastAppliedAnnotation = "kubectl.kubernetes.io/last-applied-configuration"

// parseRef parses a "group:kind:namespace:name" token into its ResourceRef + GVR.
// Group may be empty (core). The kind must be in the closed registry; anything
// else is refused, so a ref can only address a known operator-owned kind.
func parseRef(ref string) (ResourceRef, schema.GroupVersionResource, error) {
	parts := strings.SplitN(ref, ":", 4)
	if len(parts) != 4 {
		return ResourceRef{}, schema.GroupVersionResource{}, zip.ErrBadRequest("ref must be group:kind:namespace:name")
	}
	group, kind, ns, name := parts[0], parts[1], parts[2], parts[3]
	gvr, ok := kindGVR[group+"/"+kind]
	if !ok {
		return ResourceRef{}, schema.GroupVersionResource{}, zip.ErrBadRequest("unsupported resource kind " + group + "/" + kind)
	}
	if _, ok := nsEnv[ns]; !ok {
		return ResourceRef{}, schema.GroupVersionResource{}, zip.ErrBadRequest("namespace must be a platform namespace")
	}
	if !appNameRE.MatchString(name) {
		return ResourceRef{}, schema.GroupVersionResource{}, zip.ErrBadRequest("resource name must be a DNS-1123 label")
	}
	return makeRef(group, gvr.Version, kind, ns, name), gvr, nil
}

// appResource returns the live manifest + diff for one node of {name}'s tree.
func appResource(s *cloud.Service[state], c *zip.Ctx) error {
	if err := ready(s); err != nil {
		return err
	}
	app := reqName(c)
	if !appNameRE.MatchString(app) {
		return zip.ErrBadRequest("name must be a DNS-1123 label")
	}
	ref, gvr, perr := parseRef(c.Param("ref"))
	if perr != nil {
		return perr
	}
	// Resolve the app CR first (for the membership check + namespace consistency).
	crNS, err := resolveNamespace(s, c, app)
	if err != nil {
		return err
	}
	if ref.Namespace != crNS {
		return zip.ErrNotFound("resource is not in this application's namespace")
	}
	cr, _, err := getAppCR(s, c.Context(), crNS, app)
	if err != nil {
		return k8sErr(s, "get", err)
	}
	live, err := s.State.dyn.Resource(gvr).Namespace(ref.Namespace).Get(c.Context(), ref.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return zip.ErrNotFound("resource not found")
		}
		return k8sErr(s, "get", err)
	}
	if !belongsToApp(live, cr, app) {
		return zip.ErrNotFound("resource is not part of this application")
	}

	health, hmsg := resourceHealth(live)
	desiredSource, modified, desired := computeDiff(live)
	return c.JSON(http.StatusOK, map[string]any{
		"ref":           ref,
		"health":        health,
		"healthMessage": hmsg,
		"liveManifest":  live.Object,
		"desiredSource": desiredSource,
		"diff": map[string]any{
			"modified":        modified,
			"desiredManifest": desired,
		},
	})
}

// belongsToApp reports whether live is part of app's tree: it IS the CR, or it
// carries an ownerRef to the CR, or its name equals app, or a standard app label
// ties it to app. The membership boundary that keeps this endpoint scoped.
func belongsToApp(live, cr *unstructured.Unstructured, app string) bool {
	if live.GetUID() == cr.GetUID() && live.GetUID() != "" {
		return true
	}
	if ownedBy(live, string(cr.GetUID())) {
		return true
	}
	if live.GetName() == app {
		return true
	}
	labels := live.GetLabels()
	return labels["app.kubernetes.io/instance"] == app || labels["app.kubernetes.io/part-of"] == app
}

// computeDiff derives (desiredSource, modified, desired) for a live object from its
// last-applied-configuration annotation. modified is true when the normalized live
// object differs from the normalized desired (server-set noise stripped from
// both). Absent the annotation, the source is "none" and modified is false (no
// desired to compare) — honest, never a fabricated diff.
func computeDiff(live *unstructured.Unstructured) (source string, modified bool, desired map[string]any) {
	raw := live.GetAnnotations()[lastAppliedAnnotation]
	if strings.TrimSpace(raw) == "" {
		return "none", false, nil
	}
	var d map[string]any
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		return "none", false, nil
	}
	modified = !jsonEqual(normalizeForDiff(live.Object), normalizeForDiff(d))
	return deployDesiredTODO, modified, d
}

// normalizeForDiff strips server-set / volatile fields so a diff reflects only
// intent: status, metadata.managedFields/resourceVersion/uid/generation/
// creationTimestamp, and the last-applied annotation itself.
func normalizeForDiff(in map[string]any) map[string]any {
	out := deepCopyMap(in)
	delete(out, "status")
	if md, ok := out["metadata"].(map[string]any); ok {
		// Strip every field the server/operator sets that the last-applied intent
		// never carries (ownerReferences included — the operator owns those), so the
		// coarse two-way compare reflects intent, not server bookkeeping. The precise
		// three-way merge arrives with the gitops-engine diff (P2b).
		for _, k := range []string{"managedFields", "resourceVersion", "uid", "generation", "creationTimestamp", "selfLink", "ownerReferences"} {
			delete(md, k)
		}
		if ann, ok := md["annotations"].(map[string]any); ok {
			delete(ann, lastAppliedAnnotation)
			if len(ann) == 0 {
				delete(md, "annotations")
			}
		}
	}
	return out
}

// jsonEqual compares two maps by canonical JSON (encoding/json sorts string keys),
// so field order never produces a false diff.
func jsonEqual(a, b map[string]any) bool {
	ab, _ := json.Marshal(a)
	bb, _ := json.Marshal(b)
	return string(ab) == string(bb)
}

func deepCopyMap(in map[string]any) map[string]any {
	b, err := json.Marshal(in)
	if err != nil {
		return map[string]any{}
	}
	var out map[string]any
	_ = json.Unmarshal(b, &out)
	return out
}
