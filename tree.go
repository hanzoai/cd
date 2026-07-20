// tree.go — GET /v1/deploy/{name}/tree: the owned-resource tree for one
// Application, the ArgoCD ApplicationTree shape (a FLAT node list with parentRefs
// edges; the console renders the DAG). The root is the Service CR; depth-1 nodes
// are the operator-owned Deployment/Service/Ingress/HPA/PDB/ConfigMap; depth-2 are
// the Deployment's ReplicaSets and their Pods. Ownership is by ownerReferences.uid
// with a name-equals-app fallback (some operator-rendered children are named after
// the CR). Secret objects are never included — the tree cannot leak env.
//
// P2b swaps buildTree's cluster walk for github.com/argoproj/gitops-engine
// pkg/cache (ClusterCache.GetManagedLiveObjs / hierarchy) for a watch-backed tree;
// the Node shape the console consumes does not change.
package deploy

import (
	"context"
	"net/http"

	"github.com/hanzoai/cloud"
	"github.com/zap-proto/zip"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ResourceRef identifies one node — the round-trip token the resource endpoint
// parses. Ref is the canonical "group:kind:namespace:name" string.
type ResourceRef struct {
	Group     string `json:"group"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Ref       string `json:"ref"`
}

// Node is one resource in the tree: its ref, ownerRef parents, and derived
// health/sync (+ image version for a workload). ArgoCD ResourceNode shape.
type Node struct {
	ResourceRef
	UID           string        `json:"uid,omitempty"`
	CreatedAt     string        `json:"createdAt,omitempty"`
	Health        string        `json:"health"`
	HealthMessage string        `json:"healthMessage,omitempty"`
	Sync          string        `json:"sync,omitempty"`
	Version       string        `json:"version,omitempty"` // image tag for a workload node
	ParentRefs    []ResourceRef `json:"parentRefs,omitempty"`
}

// appTree returns the flat node tree for one Application.
func appTree(s *cloud.Service[state], c *zip.Ctx) error {
	if err := ready(s); err != nil {
		return err
	}
	name := reqName(c)
	if !appNameRE.MatchString(name) {
		return zip.ErrBadRequest("name must be a DNS-1123 label")
	}
	ns, err := resolveNamespace(s, c, name)
	if err != nil {
		return err
	}
	cr, _, err := getAppCR(s, c.Context(), ns, name)
	if err != nil {
		return k8sErr(s, "get", err)
	}
	nodes := buildTree(s, c.Context(), ns, name, cr)
	return c.JSON(http.StatusOK, map[string]any{
		"application": observeApplication(cr, ns, ""),
		"nodes":       nodes,
	})
}

// depth1GVRs are the operator-owned kinds scanned directly under the Service CR.
var depth1GVRs = []schema.GroupVersionResource{
	deploymentsGVR, coreSvcGVR, ingressGVR, hpaGVR, pdbGVR, configMapsGVR,
}

// buildTree walks the owned-resource hierarchy for the CR and returns a flat node
// list (root first). Depth-1 = objects owned by the CR (ownerRef uid, or named ==
// app); depth-2 = ReplicaSets owned by those Deployments and Pods owned by those
// ReplicaSets (or matching a Deployment's selector). Best-effort per GVR: an
// RBAC/list error on one kind is logged and skipped, never fatal — the tree still
// renders the reachable kinds.
func buildTree(s *cloud.Service[state], ctx context.Context, ns, app string, cr *unstructured.Unstructured) []Node {
	root := toNode(cr, ns)
	nodes := []Node{root}
	crUID := string(cr.GetUID())

	var deployUIDs []string
	deploySelectors := map[string]map[string]string{} // deployment name → matchLabels
	for _, gvr := range depth1GVRs {
		list, err := s.State.dyn.Resource(gvr).Namespace(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			s.Log.Warn("tree: list kind failed (skipping)", "resource", gvr.Resource, "namespace", ns, "err", err)
			continue
		}
		for i := range list.Items {
			obj := &list.Items[i]
			if !ownedBy(obj, crUID) && obj.GetName() != app {
				continue
			}
			n := toNode(obj, ns)
			n.ParentRefs = []ResourceRef{root.ResourceRef}
			nodes = append(nodes, n)
			if gvr == deploymentsGVR {
				deployUIDs = append(deployUIDs, string(obj.GetUID()))
				if sel, ok, _ := unstructured.NestedStringMap(obj.Object, "spec", "selector", "matchLabels"); ok {
					deploySelectors[obj.GetName()] = sel
				}
			}
		}
	}

	// Depth-2: ReplicaSets owned by the app's Deployments, then Pods owned by those
	// ReplicaSets (uid map) or matching a Deployment selector (fallback).
	var rsUIDs []string
	rsRefByUID := map[string]ResourceRef{}
	if rsList, err := s.State.dyn.Resource(replicaSetsGVR).Namespace(ns).List(ctx, metav1.ListOptions{}); err == nil {
		for i := range rsList.Items {
			rs := &rsList.Items[i]
			if !ownedByAny(rs, deployUIDs) {
				continue
			}
			n := toNode(rs, ns)
			n.ParentRefs = []ResourceRef{deploymentParentRef(ns, rs)}
			nodes = append(nodes, n)
			rsUIDs = append(rsUIDs, string(rs.GetUID()))
			rsRefByUID[string(rs.GetUID())] = n.ResourceRef
		}
	}
	if podList, err := s.State.dyn.Resource(podsGVR).Namespace(ns).List(ctx, metav1.ListOptions{}); err == nil {
		for i := range podList.Items {
			pod := &podList.Items[i]
			parent, owned := podParent(pod, rsUIDs, rsRefByUID, deploySelectors, ns)
			if !owned {
				continue
			}
			n := toNode(pod, ns)
			n.ParentRefs = []ResourceRef{parent}
			nodes = append(nodes, n)
		}
	}
	return nodes
}

// toNode maps a live object to a tree Node: its ref, uid, creation time, derived
// health, and (for a workload) its image tag. Sync is set on the root from the
// declared-vs-running comparison; child sync is left empty (the operator owns the
// child, so there is no independent desired to compare).
func toNode(obj *unstructured.Unstructured, ns string) Node {
	gvk := obj.GroupVersionKind()
	health, msg := resourceHealth(obj)
	n := Node{
		ResourceRef:   makeRef(gvk.Group, gvk.Version, gvk.Kind, ns, obj.GetName()),
		UID:           string(obj.GetUID()),
		Health:        health,
		HealthMessage: msg,
		Version:       workloadTag(obj),
	}
	if ts := obj.GetCreationTimestamp(); !ts.IsZero() {
		n.CreatedAt = ts.UTC().Format("2006-01-02T15:04:05Z")
	}
	if obj.GetAPIVersion() == "hanzo.ai/v1" && (obj.GetKind() == "App" || obj.GetKind() == "Service") {
		declared, _, _ := unstructured.NestedString(obj.Object, "spec", "image", "tag")
		n.Version = declared
		// Root sync is filled by the caller-facing list; here we leave it to the
		// application row (observeApplication) which has the running tag.
	}
	return n
}

// workloadTag returns the first-container image tag for a Deployment/ReplicaSet
// (spec.template.spec.containers) or Pod (spec.containers) — the running version
// shown on the node — else "".
func workloadTag(obj *unstructured.Unstructured) string {
	switch obj.GetKind() {
	case "Deployment", "ReplicaSet":
		return firstContainerTag(obj)
	case "Pod":
		if raw, ok, _ := unstructured.NestedSlice(obj.Object, "spec", "containers"); ok {
			for _, c := range raw {
				if m, ok := c.(map[string]any); ok {
					if img, ok := m["image"].(string); ok && img != "" {
						return tagFromImageRef(img)
					}
				}
			}
		}
	}
	return ""
}

// makeRef builds a ResourceRef with the canonical "group:kind:namespace:name" token.
func makeRef(group, version, kind, ns, name string) ResourceRef {
	return ResourceRef{
		Group: group, Version: version, Kind: kind, Namespace: ns, Name: name,
		Ref: group + ":" + kind + ":" + ns + ":" + name,
	}
}

// ownedBy reports whether obj carries an ownerReference to uid.
func ownedBy(obj *unstructured.Unstructured, uid string) bool {
	if uid == "" {
		return false
	}
	for _, o := range obj.GetOwnerReferences() {
		if string(o.UID) == uid {
			return true
		}
	}
	return false
}

func ownedByAny(obj *unstructured.Unstructured, uids []string) bool {
	for _, u := range uids {
		if ownedBy(obj, u) {
			return true
		}
	}
	return false
}

// deploymentParentRef derives a ReplicaSet's Deployment parent ref from its
// ownerReferences (the Deployment that owns it).
func deploymentParentRef(ns string, rs *unstructured.Unstructured) ResourceRef {
	for _, o := range rs.GetOwnerReferences() {
		if o.Kind == "Deployment" {
			return makeRef("apps", "v1", "Deployment", ns, o.Name)
		}
	}
	return makeRef("apps", "v1", "Deployment", ns, "")
}

// podParent resolves a Pod's parent: the owning ReplicaSet (by uid) when present,
// else the Deployment whose selector matches the Pod's labels (fallback). Returns
// (parentRef, owned).
func podParent(pod *unstructured.Unstructured, rsUIDs []string, rsRefByUID map[string]ResourceRef, deploySelectors map[string]map[string]string, ns string) (ResourceRef, bool) {
	for _, o := range pod.GetOwnerReferences() {
		if o.Kind == "ReplicaSet" {
			if ref, ok := rsRefByUID[string(o.UID)]; ok {
				return ref, true
			}
		}
	}
	labels := pod.GetLabels()
	for dep, sel := range deploySelectors {
		if labelsMatch(labels, sel) {
			return makeRef("apps", "v1", "Deployment", ns, dep), true
		}
	}
	return ResourceRef{}, false
}

func labelsMatch(labels, selector map[string]string) bool {
	if len(selector) == 0 {
		return false
	}
	for k, v := range selector {
		if labels[k] != v {
			return false
		}
	}
	return true
}
