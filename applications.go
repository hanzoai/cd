// applications.go — GET /v1/deploy/applications: the fleet list. Each operator
// Service CR is one Application row: its declared image version, the running
// version observed from the live Deployment, the reconciled health, and the sync
// verdict (declared == running ⇒ Synced, else OutOfSync). The console renders this
// as the ArgoCD application list.
package deploy

import (
	"context"
	"net/http"
	"sort"
	"strings"

	"github.com/hanzoai/cloud"
	"github.com/zap-proto/zip"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Sync codes — the ArgoCD sync vocabulary, lowercased for the wire.
const (
	SyncSynced    = "synced"
	SyncOutOfSync = "out-of-sync"
	SyncUnknown   = "unknown"
)

// Application is one fleet row. Shapes the exact fields the console list consumes.
type Application struct {
	Name           string   `json:"name"`
	Namespace      string   `json:"namespace"`
	Env            string   `json:"env"`            // main|test|dev
	Role           string   `json:"role,omitempty"` // spec.role after the kind collapse (App)
	Repository     string   `json:"repository"`
	Version        string   `json:"version"`        // declared: spec.image.tag
	RunningVersion string   `json:"runningVersion"` // observed from the Deployment
	Health         string   `json:"health"`         // healthy|progressing|degraded|suspended|missing|unknown
	HealthMessage  string   `json:"healthMessage,omitempty"`
	Sync           string   `json:"sync"` // synced|out-of-sync|unknown
	Phase          string   `json:"phase,omitempty"`
	Endpoints      []string `json:"endpoints"`
}

// listApplications returns every Service CR across the platform namespaces as an
// Application row, ordered (namespace, name). Optional narrowing: ?env=, ?health=,
// ?sync=, ?name=.
func listApplications(s *cloud.Service[state], c *zip.Ctx) error {
	if err := ready(s); err != nil {
		return err
	}
	var apps []Application
	for _, ns := range scanOrder() {
		crs, err := listAppCRs(s, c.Context(), ns)
		if err != nil {
			return k8sErr(s, "list", err)
		}
		running := runningVersions(s, c.Context(), ns)
		for i := range crs {
			cr := &crs[i]
			apps = append(apps, observeApplication(cr, ns, running[cr.GetName()]))
		}
	}
	sort.Slice(apps, func(i, j int) bool {
		if apps[i].Namespace != apps[j].Namespace {
			return apps[i].Namespace < apps[j].Namespace
		}
		return apps[i].Name < apps[j].Name
	})

	env := strings.TrimSpace(c.Query("env"))
	healthF := strings.TrimSpace(c.Query("health"))
	syncF := strings.TrimSpace(c.Query("sync"))
	nameF := strings.TrimSpace(c.Query("name"))
	out := make([]Application, 0, len(apps))
	summary := map[string]int{"total": 0, "healthy": 0, "degraded": 0, "outOfSync": 0}
	for _, a := range apps {
		if env != "" && a.Env != env {
			continue
		}
		if healthF != "" && a.Health != healthF {
			continue
		}
		if syncF != "" && a.Sync != syncF {
			continue
		}
		if nameF != "" && a.Name != nameF {
			continue
		}
		out = append(out, a)
		summary["total"]++
		if a.Health == HealthHealthy {
			summary["healthy"]++
		}
		if a.Health == HealthDegraded {
			summary["degraded"]++
		}
		if a.Sync == SyncOutOfSync {
			summary["outOfSync"]++
		}
	}
	return c.JSON(http.StatusOK, map[string]any{"applications": out, "summary": summary})
}

// observeApplication maps one Service CR (+ its running version) to an Application
// row. Declared version from the CR spec, running version from the Deployment,
// health + phase + endpoints from the operator-reconciled CR status, sync from
// declared-vs-running.
func observeApplication(cr *unstructured.Unstructured, ns, runningTag string) Application {
	repository, _, _ := unstructured.NestedString(cr.Object, "spec", "image", "repository")
	declaredTag, _, _ := unstructured.NestedString(cr.Object, "spec", "image", "tag")
	role, _, _ := unstructured.NestedString(cr.Object, "spec", "role")
	phase, _, _ := unstructured.NestedString(cr.Object, "status", "phase")
	health, msg := resourceHealth(cr)
	return Application{
		Name:           cr.GetName(),
		Namespace:      ns,
		Env:            nsEnv[ns],
		Role:           role,
		Repository:     repository,
		Version:        declaredTag,
		RunningVersion: runningTag,
		Health:         health,
		HealthMessage:  msg,
		Sync:           syncStatus(declaredTag, runningTag),
		Phase:          phase,
		Endpoints:      nestedStringSlice(cr.Object, "status", "endpoints"),
	}
}

// syncStatus is the sync verdict for the declared vs running image tag: equal ⇒
// Synced, both known and unequal ⇒ OutOfSync, either unknown ⇒ Unknown. It is the
// cluster-vs-desired signal today (the CR is the desired source); when git.hanzo.ai
// becomes the desired source it compares CR-in-cluster vs CR-in-git, same verdict.
func syncStatus(declared, running string) string {
	if declared == "" || running == "" {
		return SyncUnknown
	}
	if declared == running {
		return SyncSynced
	}
	return SyncOutOfSync
}

// runningVersions lists the Deployments in ns and returns name → running image
// tag. Best-effort: a list error yields an empty map so the board still renders
// declared/health/phase.
func runningVersions(s *cloud.Service[state], ctx context.Context, ns string) map[string]string {
	out := map[string]string{}
	list, err := s.State.dyn.Resource(deploymentsGVR).Namespace(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		s.Log.Warn("list deployments for running version failed", "namespace", ns, "err", err)
		return out
	}
	for i := range list.Items {
		out[list.Items[i].GetName()] = firstContainerTag(&list.Items[i])
	}
	return out
}

// ── image ref + slice helpers (pure) ────────────────────────────────────────

// firstContainerTag is the running tag from a Deployment's first container image.
func firstContainerTag(dep *unstructured.Unstructured) string {
	imgs := deploymentContainerImages(dep)
	if len(imgs) == 0 {
		return ""
	}
	return tagFromImageRef(imgs[0])
}

func deploymentContainerImages(dep *unstructured.Unstructured) []string {
	raw, ok, _ := unstructured.NestedSlice(dep.Object, "spec", "template", "spec", "containers")
	if !ok {
		return nil
	}
	imgs := make([]string, 0, len(raw))
	for _, c := range raw {
		if m, ok := c.(map[string]any); ok {
			if img, ok := m["image"].(string); ok && img != "" {
				imgs = append(imgs, img)
			}
		}
	}
	return imgs
}

// repoFromImageRef splits `ghcr.io/hanzoai/iam:v1` → `ghcr.io/hanzoai/iam`; a
// digest ref keeps the repo; a host:port segment is not read as a tag.
func repoFromImageRef(ref string) string {
	if at := strings.LastIndex(ref, "@"); at >= 0 {
		ref = ref[:at]
	}
	slash := strings.LastIndex(ref, "/")
	colon := strings.LastIndex(ref, ":")
	if colon > slash {
		return ref[:colon]
	}
	return ref
}

// tagFromImageRef splits `ghcr.io/hanzoai/iam:v1` → `v1`; a digest ref returns the
// digest; a bare repo returns "".
func tagFromImageRef(ref string) string {
	if at := strings.LastIndex(ref, "@"); at >= 0 {
		return ref[at+1:]
	}
	slash := strings.LastIndex(ref, "/")
	colon := strings.LastIndex(ref, ":")
	if colon > slash && colon < len(ref)-1 {
		return ref[colon+1:]
	}
	return ""
}

func nestedStringSlice(obj map[string]any, fields ...string) []string {
	raw, ok, _ := unstructured.NestedSlice(obj, fields...)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, e := range raw {
		if s, ok := e.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
