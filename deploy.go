// Package gitops mounts the native GitOps control plane at /v1/deploy — the
// ArgoCD-grade deploy dashboard for the operator-managed fleet, made native to
// the cloud binary and parallel to /v1/git (the native git server).
//
// Each operator hanzo.ai/v1 App CR IS a GitOps Application: the desired state
// declared for one workload, which the Hanzo operator reconciles into a
// Deployment + Service + Ingress (+ HPA/PDB/Pods). This plane OBSERVES that
// reconciliation the way ArgoCD observes a synced Application —
//
//	GET  /v1/deploy/applications        — the fleet list: name, declared version,
//	                                      health, sync, per app.
//	GET  /v1/deploy/{name}/tree         — the owned-resource tree (ownerRef edges)
//	                                      with per-node health + sync.
//	GET  /v1/deploy/{name}/resource/{ref} — one node's live manifest + a
//	                                      desired-vs-live diff.
//	GET  /v1/deploy/{name}/logs         — the app's current pod logs.
//	POST /v1/deploy/{name}/rollback     — pin the CR image tag to a prior semver
//	                                      (the operator reconciles the rollout).
//	POST /v1/deploy/{name}/sync         — request an operator reconcile now.
//
// SECURITY — every route is SUPERADMIN ONLY, fail-closed, on the SAME predicate
// the rest of cloud uses (c.IsAdmin()): the plane reads and mutates SYSTEM Service
// CRs across the whole fleet, so a tenant must never reach it. Secret objects are
// never surfaced (no node, no manifest) so the tree can never leak materialized
// env. The user-facing per-org PaaS is /v1/platform; this is the platform-operator
// console the admin dashboard consumes.
//
// GitOps note (the follow-on seam): today the CR is the desired-state source and a
// rollback/rollout PATCHES it directly (P1's RegisterServiceReleaser), so deploys
// work now. The end-state is true GitOps on OUR native git — RegisterPushBuilder
// commits the CR image-tag change to the manifest repo on git.hanzo.ai
// (github.com/hanzoai/git) and this engine syncs that repo → cluster with
// self-heal. The desired-vs-live diff below is already structured for that: it
// reads a desired source that is "cluster last-applied" now and becomes the
// git.hanzo.ai manifest later, with no shape change. See deployDesiredTODO.
package deploy

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/hanzoai/cloud"
	"github.com/zap-proto/zip"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// appsCRGVR is the operator App CR (apps.hanzo.ai) — the one workload kind this
// plane reads. Each App IS a GitOps Application: the desired state for one
// workload, which the operator reconciles into a Deployment + Service + Ingress
// (+ HPA/PDB/Pods). Group hanzo.ai disambiguates it from the core/v1 Service (a
// CHILD it reconciles), which is why a resource ref always carries its group.
var appsCRGVR = schema.GroupVersionResource{Group: "hanzo.ai", Version: "v1", Resource: "apps"}

// childGVRs are the operator-owned workload objects the tree walks at depth 1
// (owned by the Service CR) and their descendants (ReplicaSet → Pod). Secrets are
// DELIBERATELY absent: the tree never surfaces materialized env. Order is the
// display order.
var (
	deploymentsGVR = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	replicaSetsGVR = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "replicasets"}
	podsGVR        = schema.GroupVersionResource{Version: "v1", Resource: "pods"}
	coreSvcGVR     = schema.GroupVersionResource{Version: "v1", Resource: "services"}
	ingressGVR     = schema.GroupVersionResource{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"}
	hpaGVR         = schema.GroupVersionResource{Group: "autoscaling", Version: "v2", Resource: "horizontalpodautoscalers"}
	pdbGVR         = schema.GroupVersionResource{Group: "policy", Version: "v1", Resource: "poddisruptionbudgets"}
	configMapsGVR  = schema.GroupVersionResource{Version: "v1", Resource: "configmaps"}
)

// kindGVR resolves a resource ref's (group, kind) to its GVR — the fixed, closed
// registry of kinds this plane reads. A ref for any kind NOT here is refused, so
// the resource endpoint can never be steered at an arbitrary cluster object.
// Keyed by "group/Kind" (group "" for the core API group).
var kindGVR = map[string]schema.GroupVersionResource{
	"hanzo.ai/App":                        appsCRGVR,
	"apps/Deployment":                     deploymentsGVR,
	"apps/ReplicaSet":                     replicaSetsGVR,
	"/Pod":                                podsGVR,
	"/Service":                            coreSvcGVR,
	"/ConfigMap":                          configMapsGVR,
	"networking.k8s.io/Ingress":           ingressGVR,
	"autoscaling/HorizontalPodAutoscaler": hpaGVR,
	"policy/PodDisruptionBudget":          pdbGVR,
}

// nsEnv maps each scanned platform namespace to its lifecycle env, mirroring
// clients/paas (main first). Only these namespaces are read — the plane never
// reaches beyond the platform tier.
var nsEnv = map[string]string{"hanzo": "main", "hanzo-testnet": "test", "hanzo-devnet": "dev"}

// scanOrder is the platform namespaces in stable env order (production first), so
// a bare app name resolves to main before test/dev.
func scanOrder() []string { return []string{"hanzo", "hanzo-testnet", "hanzo-devnet"} }

// appNameRE constrains the {name} path segment to a DNS-1123 label (every Service
// CR metadata.name satisfies this) — the injection guard for the CR a route reads.
var appNameRE = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

const userAgent = "hanzo-cloud-deploy"

// deployDesiredTODO documents the desired-state source seam: "last-applied" (the
// kubectl last-applied-configuration annotation on the live object) today; the
// git.hanzo.ai manifest repo once RegisterPushBuilder commits CR changes there.
// The diff shape does not change when the source flips.
const deployDesiredTODO = "last-applied"

// state is gitops's own data; shared deps live in the embedded cloud.Base.
type state struct {
	dyn       dynamic.Interface // nil when no kubeconfig resolved (fail-closed)
	clientset kubernetes.Interface
	initErr   string
}

// Mount wires /v1/deploy/* onto app. Every handler gates on c.IsAdmin() first.
func Mount(app *zip.App, deps cloud.Deps) error {
	return cloud.Mount(app, deps, "deploy", build, routes)
}

// build resolves the in-process k8s clients (fail-closed: when no kubeconfig
// resolves the subsystem still mounts and every endpoint 503s honestly).
func build(b cloud.Base) (state, error) {
	var st state
	dyn, cs, err := newClients()
	if err != nil {
		st.initErr = err.Error()
		b.Log.Warn("kubernetes client unavailable; /v1/deploy endpoints will fail closed", "err", err)
	} else {
		st.dyn, st.clientset = dyn, cs
	}
	b.Log.Info("deploy control plane mounted", "prefix", "/v1/deploy", "k8s", st.dyn != nil, "brand", b.Brand, "env", b.Env)
	return st, nil
}

// routes registers the /v1/deploy/* surface. Every observing/mutating route is
// SuperAdmin-gated; the health probe is public (real k8s reachability).
func routes(app *zip.App, s *cloud.Service[state]) {
	app.Get("/v1/deploy/applications", guard(s, cloud.Handle(s, listApplications)))
	app.Get("/v1/deploy/health", cloud.Handle(s, health))
	app.Get("/v1/deploy/:name/tree", guard(s, cloud.Handle(s, appTree)))
	app.Get("/v1/deploy/:name/resource/:ref", guard(s, cloud.Handle(s, appResource)))
	app.Get("/v1/deploy/:name/logs", guard(s, cloud.Handle(s, appLogs)))
	app.Post("/v1/deploy/:name/rollback", guard(s, cloud.Handle(s, rollback)))
	app.Post("/v1/deploy/:name/sync", guard(s, cloud.Handle(s, sync)))
	// Engine (write) routes — the embedded gitops-engine reconcile that replaces
	// universe-crs. Gated by DEPLOY_ENGINE_ENABLED; see engine_mount.go.
	registerEngineRoutes(app, s)
	// ArgoCD monochrome dashboard (App-CR projection) at /v1/deploy/ui/*.
	registerDashboardRoutes(app, s)
}

// guard wraps a handler with the SuperAdmin gate (fail-closed: a non-SuperAdmin is
// refused 403 before any cluster object is read or mutated), matching clients/paas.
func guard(s *cloud.Service[state], h zip.Handler) zip.Handler {
	return func(c *zip.Ctx) error {
		if !c.IsAdmin() {
			return zip.ErrForbidden("SuperAdmin required")
		}
		return h(c)
	}
}

// health is a REAL probe: the API server is reachable AND the App CRD is served.
// 200 only when both hold; 503 + the real reason otherwise. Not admin-gated —
// liveness must be probe-able without a JWT.
func health(s *cloud.Service[state], c *zip.Ctx) error {
	res := map[string]any{"service": "deploy", "status": "ok"}
	if s.State.dyn == nil {
		res["status"], res["k8s"], res["error"] = "degraded", false, s.State.initErr
		return c.JSON(http.StatusServiceUnavailable, res)
	}
	if _, err := s.State.dyn.Resource(appsCRGVR).Namespace("hanzo").List(c.Context(), metav1.ListOptions{Limit: 1}); err != nil {
		res["status"], res["k8s"], res["crd"], res["error"] = "degraded", true, false, err.Error()
		return c.JSON(http.StatusServiceUnavailable, res)
	}
	res["k8s"], res["crd"] = true, true
	return c.JSON(http.StatusOK, res)
}

// ready fails closed when no cluster client resolved (503 + the real reason).
func ready(s *cloud.Service[state]) error {
	if s.State.dyn == nil {
		return zip.Errorf(http.StatusServiceUnavailable, "deploy: kubernetes client not configured: %s", s.State.initErr)
	}
	return nil
}

// reqName reads and normalizes the {name} path segment.
func reqName(c *zip.Ctx) string { return regexpLower(c.Param("name")) }

func regexpLower(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b >= 'A' && b <= 'Z' {
			b += 'a' - 'A'
		}
		if b == ' ' || b == '\t' || b == '\n' || b == '\r' {
			continue
		}
		out = append(out, b)
	}
	return string(out)
}

// resolveNamespace finds the platform namespace an App CR lives in, scanning in
// env order (main first). Returns a clean 404 when found in none.
func resolveNamespace(s *cloud.Service[state], c *zip.Ctx, name string) (string, error) {
	for _, ns := range scanOrder() {
		if _, _, err := getAppCR(s, c.Context(), ns, name); err == nil {
			return ns, nil
		} else if !apierrors.IsNotFound(err) {
			return "", k8sErr(s, "get", err)
		}
	}
	return "", zip.ErrNotFound("application " + name + " not found in the platform namespaces")
}

// getAppCR gets an App CR by name from ns. Returns the object and its GVR so a
// mutation (sync/rollback) patches the App CR it read. A miss is an IsNotFound
// error.
func getAppCR(s *cloud.Service[state], ctx context.Context, ns, name string) (*unstructured.Unstructured, schema.GroupVersionResource, error) {
	obj, err := s.State.dyn.Resource(appsCRGVR).Namespace(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, schema.GroupVersionResource{}, err
	}
	return obj, appsCRGVR, nil
}

// listAppCRs lists every App CR in ns.
func listAppCRs(s *cloud.Service[state], ctx context.Context, ns string) ([]unstructured.Unstructured, error) {
	list, err := s.State.dyn.Resource(appsCRGVR).Namespace(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return list.Items, nil
}

// k8sErr maps a raw API error to an honest gateway error, naming the missing RBAC
// so an operator knows exactly what to grant the cloud service account.
func k8sErr(s *cloud.Service[state], op string, err error) error {
	s.Log.Error("k8s op failed", "op", op, "err", err)
	if apierrors.IsForbidden(err) {
		return zip.Errorf(http.StatusBadGateway,
			"%s: kubernetes RBAC denied (cloud service account needs %s across the platform namespaces): %v", op, op, err)
	}
	return zip.Errorf(http.StatusBadGateway, "%s failed: %v", op, err)
}

// newClients builds the dynamic + typed clients from the in-cluster service
// account, falling back to KUBECONFIG for local/dev — identical construction to
// clients/paas + clients/platform.
func newClients() (dynamic.Interface, kubernetes.Interface, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{})
		cfg, err = cc.ClientConfig()
		if err != nil {
			return nil, nil, fmt.Errorf("no in-cluster config and no kubeconfig: %w", err)
		}
	}
	cfg.UserAgent = userAgent
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("dynamic client: %w", err)
	}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("clientset: %w", err)
	}
	return dyn, cs, nil
}
