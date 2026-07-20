// health.go — per-resource health, the ArgoCD `pkg/health` vocabulary made native
// (Healthy / Progressing / Degraded / Suspended / Missing / Unknown). It is a PURE
// function over one live object, so the list + tree derive an honest per-node
// health with no cluster round-trip beyond the object already read.
//
// P2b swaps the internals for github.com/argoproj/gitops-engine pkg/health
// (health.GetResourceHealth) for exact ArgoCD parity; the CODES emitted here are
// already those strings, so the wire contract the console consumes does not change.
package deploy

import (
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Health codes — the ArgoCD health vocabulary, lowercased for the wire.
const (
	HealthHealthy     = "healthy"
	HealthProgressing = "progressing"
	HealthDegraded    = "degraded"
	HealthSuspended   = "suspended"
	HealthMissing     = "missing"
	HealthUnknown     = "unknown"
)

// resourceHealth derives (code, message) for one live object by kind. Unknown
// kinds default to Healthy (a ConfigMap/PDB/HPA existing IS its healthy state);
// workloads read their reconciled status. Never fabricates — an absent status
// signal yields Progressing/Unknown, never a green.
func resourceHealth(obj *unstructured.Unstructured) (code, message string) {
	if obj == nil {
		return HealthMissing, "resource not found"
	}
	switch obj.GetKind() {
	case "App": // hanzo.ai/v1 App (forward kind)
		return appCRHealth(obj)
	case "Service":
		if obj.GetAPIVersion() == "hanzo.ai/v1" { // transition kind (pre-collapse)
			return appCRHealth(obj)
		}
		return coreServiceHealth(obj)
	case "Deployment":
		return deploymentHealth(obj)
	case "ReplicaSet":
		return replicaSetHealth(obj)
	case "Pod":
		return podHealth(obj)
	case "Ingress":
		return ingressHealth(obj)
	default:
		return HealthHealthy, ""
	}
}

// appCRHealth reads the operator-reconciled App/Service CR status: phase + the
// replica counts the operator publishes. Mirrors clients/paas healthFromStatus but
// in the ArgoCD vocabulary.
func appCRHealth(obj *unstructured.Unstructured) (string, string) {
	phase, _, _ := unstructured.NestedString(obj.Object, "status", "phase")
	desired, hasDesired := nestedInt(obj.Object, "status", "replicas")
	ready, _ := nestedInt(obj.Object, "status", "readyReplicas")
	if !hasDesired {
		if avail, ok := nestedInt(obj.Object, "status", "availableReplicas"); ok {
			if avail > 0 {
				return HealthHealthy, phaseMsg(phase, "available")
			}
			return HealthDegraded, phaseMsg(phase, "no available replicas")
		}
		if phase == "" {
			return HealthProgressing, "awaiting operator reconcile"
		}
		return HealthProgressing, phaseMsg(phase, "reconciling")
	}
	if desired == 0 {
		return HealthSuspended, "scaled to zero"
	}
	if ready >= desired {
		return HealthHealthy, phaseMsg(phase, "all replicas ready")
	}
	if ready > 0 {
		return HealthProgressing, phaseMsg(phase, "rolling out")
	}
	return HealthDegraded, phaseMsg(phase, "no replicas ready")
}

// deploymentHealth ports the ArgoCD Deployment rule: a ProgressDeadlineExceeded
// condition is Degraded; an unfinished rollout (observedGeneration behind, or
// updated/available < desired) is Progressing; otherwise Healthy.
func deploymentHealth(obj *unstructured.Unstructured) (string, string) {
	gen, _ := nestedInt(obj.Object, "metadata", "generation")
	observed, _ := nestedInt(obj.Object, "status", "observedGeneration")
	if observed < gen {
		return HealthProgressing, "waiting for rollout to be observed"
	}
	for _, cond := range conditions(obj) {
		if cond["type"] == "Progressing" && cond["status"] == "False" && cond["reason"] == "ProgressDeadlineExceeded" {
			return HealthDegraded, str(cond["message"], "rollout exceeded its progress deadline")
		}
	}
	desired := 1
	if v, ok := nestedInt(obj.Object, "spec", "replicas"); ok {
		desired = v
	}
	updated, _ := nestedInt(obj.Object, "status", "updatedReplicas")
	available, _ := nestedInt(obj.Object, "status", "availableReplicas")
	if updated < desired {
		return HealthProgressing, "waiting for the updated replicas to roll out"
	}
	if available < desired {
		return HealthProgressing, "waiting for replicas to become available"
	}
	return HealthHealthy, "rollout complete"
}

// replicaSetHealth: readyReplicas >= replicas → Healthy, else Progressing.
func replicaSetHealth(obj *unstructured.Unstructured) (string, string) {
	desired := 1
	if v, ok := nestedInt(obj.Object, "spec", "replicas"); ok {
		desired = v
	}
	ready, _ := nestedInt(obj.Object, "status", "readyReplicas")
	if ready >= desired {
		return HealthHealthy, ""
	}
	return HealthProgressing, "waiting for replicas to become ready"
}

// podHealth: phase drives the code, with a CrashLoopBackOff/terminated-error
// container overriding to Degraded (the signal that matters on a dashboard).
func podHealth(obj *unstructured.Unstructured) (string, string) {
	phase, _, _ := unstructured.NestedString(obj.Object, "status", "phase")
	if reason, bad := crashingContainer(obj); bad {
		return HealthDegraded, reason
	}
	switch phase {
	case "Running":
		if podContainersReady(obj) {
			return HealthHealthy, "running"
		}
		return HealthProgressing, "containers not ready"
	case "Succeeded":
		return HealthHealthy, "completed"
	case "Failed":
		return HealthDegraded, str(mustString(obj, "status", "reason"), "pod failed")
	case "Pending", "":
		return HealthProgressing, "pending"
	default:
		return HealthProgressing, phase
	}
}

// coreServiceHealth: a LoadBalancer awaiting an ingress address is Progressing;
// every other Service is Healthy the moment it exists.
func coreServiceHealth(obj *unstructured.Unstructured) (string, string) {
	if t, _, _ := unstructured.NestedString(obj.Object, "spec", "type"); t == "LoadBalancer" {
		if ing, _, _ := unstructured.NestedSlice(obj.Object, "status", "loadBalancer", "ingress"); len(ing) == 0 {
			return HealthProgressing, "waiting for the load-balancer address"
		}
	}
	return HealthHealthy, ""
}

// ingressHealth: an address published on status.loadBalancer.ingress → Healthy,
// else Progressing (cert/route still coming up).
func ingressHealth(obj *unstructured.Unstructured) (string, string) {
	if ing, _, _ := unstructured.NestedSlice(obj.Object, "status", "loadBalancer", "ingress"); len(ing) > 0 {
		return HealthHealthy, ""
	}
	return HealthProgressing, "waiting for the ingress address"
}

// ── pure helpers ────────────────────────────────────────────────────────────

func conditions(obj *unstructured.Unstructured) []map[string]string {
	raw, _, _ := unstructured.NestedSlice(obj.Object, "status", "conditions")
	out := make([]map[string]string, 0, len(raw))
	for _, c := range raw {
		cm, ok := c.(map[string]any)
		if !ok {
			continue
		}
		m := map[string]string{}
		for k, v := range cm {
			if s, ok := v.(string); ok {
				m[k] = s
			}
		}
		out = append(out, m)
	}
	return out
}

// crashingContainer reports the first container stuck in CrashLoopBackOff or a
// non-zero terminated state — the honest "Degraded" signal for a pod.
func crashingContainer(obj *unstructured.Unstructured) (string, bool) {
	statuses, _, _ := unstructured.NestedSlice(obj.Object, "status", "containerStatuses")
	for _, cs := range statuses {
		m, ok := cs.(map[string]any)
		if !ok {
			continue
		}
		waiting, _, _ := unstructured.NestedString(m, "state", "waiting", "reason")
		if waiting == "CrashLoopBackOff" || waiting == "ImagePullBackOff" || waiting == "ErrImagePull" {
			name, _, _ := unstructured.NestedString(m, "name")
			return name + ": " + waiting, true
		}
	}
	return "", false
}

// podContainersReady reports whether every container's Ready condition is true.
func podContainersReady(obj *unstructured.Unstructured) bool {
	statuses, _, _ := unstructured.NestedSlice(obj.Object, "status", "containerStatuses")
	if len(statuses) == 0 {
		return false
	}
	for _, cs := range statuses {
		m, ok := cs.(map[string]any)
		if !ok {
			return false
		}
		if ready, _ := m["ready"].(bool); !ready {
			return false
		}
	}
	return true
}

// nestedInt reads an int-valued nested key, tolerating the int64/float64 the k8s
// decoder produces.
func nestedInt(obj map[string]any, fields ...string) (int, bool) {
	v, ok, err := unstructured.NestedFieldNoCopy(obj, fields...)
	if err != nil || !ok {
		return 0, false
	}
	switch n := v.(type) {
	case int64:
		return int(n), true
	case int:
		return n, true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}

func mustString(obj *unstructured.Unstructured, fields ...string) string {
	s, _, _ := unstructured.NestedString(obj.Object, fields...)
	return s
}

func str(v any, dflt string) string {
	if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
		return s
	}
	return dflt
}

func phaseMsg(phase, detail string) string {
	if phase == "" {
		return detail
	}
	return phase + ": " + detail
}
