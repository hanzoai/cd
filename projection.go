// projection.go — the App-CR → ArgoCD `Application` READ PROJECTION.
//
// The CTO decision (have-both): serve the full ArgoCD React UI, but feed it a
// projection of our operator `App` CRs shaped as ArgoCD `Application`s. There is
// NO stored Application/AppProject CRD — each App CR IS projected on the fly, its
// resource tree + health synthesized from the SAME readers the native
// /v1/deploy routes use (listAppCRs/getAppCR/buildTree/resourceHealth), with no
// repo-server and no redis. App CRs stay the single source of truth; the
// Application shape exists only at this API layer.
//
// These `argo*` types are the MINIMAL ArgoCD v1alpha1 JSON the React app renders
// (list + detail + tree). Distinct from the native `Application` (applications.go)
// which backs the native /v1/deploy/applications surface — this backs the
// ArgoCD-UI-compatible /v1/deploy/api/v1/* surface.
package deploy

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ── ArgoCD v1alpha1 JSON (minimal, UI-render-complete) ───────────────────────

type argoListMeta struct {
	ResourceVersion string `json:"resourceVersion"`
}

type argoMeta struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	UID               string            `json:"uid,omitempty"`
	CreationTimestamp string            `json:"creationTimestamp,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
}

type argoSource struct {
	RepoURL        string `json:"repoURL"`
	Path           string `json:"path"`
	TargetRevision string `json:"targetRevision"`
}

type argoDestination struct {
	Server    string `json:"server"`
	Namespace string `json:"namespace"`
}

type argoSpec struct {
	Source      argoSource      `json:"source"`
	Destination argoDestination `json:"destination"`
	Project     string          `json:"project"`
}

type argoHealth struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type argoSyncStatus struct {
	Status   string `json:"status"`
	Revision string `json:"revision,omitempty"`
}

type argoResourceStatus struct {
	Group     string      `json:"group,omitempty"`
	Version   string      `json:"version,omitempty"`
	Kind      string      `json:"kind"`
	Namespace string      `json:"namespace,omitempty"`
	Name      string      `json:"name"`
	Status    string      `json:"status,omitempty"`
	Health    *argoHealth `json:"health,omitempty"`
}

type argoSummary struct {
	Images []string `json:"images,omitempty"`
}

type argoStatus struct {
	Sync         argoSyncStatus       `json:"sync"`
	Health       argoHealth           `json:"health"`
	Resources    []argoResourceStatus `json:"resources"`
	Summary      argoSummary          `json:"summary"`
	ReconciledAt string               `json:"reconciledAt,omitempty"`
}

type argoApp struct {
	APIVersion string     `json:"apiVersion"`
	Kind       string     `json:"kind"`
	Metadata   argoMeta   `json:"metadata"`
	Spec       argoSpec   `json:"spec"`
	Status     argoStatus `json:"status"`
}

type argoAppList struct {
	APIVersion string       `json:"apiVersion"`
	Kind       string       `json:"kind"`
	Metadata   argoListMeta `json:"metadata"`
	Items      []argoApp    `json:"items"`
}

// ── resource tree (ArgoCD ApplicationTree) ───────────────────────────────────

type argoResourceRef struct {
	Group     string `json:"group,omitempty"`
	Version   string `json:"version,omitempty"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name"`
	UID       string `json:"uid,omitempty"`
}

type argoInfoItem struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type argoNode struct {
	argoResourceRef
	ParentRefs      []argoResourceRef `json:"parentRefs,omitempty"`
	Info            []argoInfoItem    `json:"info,omitempty"`
	Health          *argoHealth       `json:"health,omitempty"`
	ResourceVersion string            `json:"resourceVersion,omitempty"`
	CreatedAt       string            `json:"createdAt,omitempty"`
	Images          []string          `json:"images,omitempty"`
}

type argoTree struct {
	Nodes         []argoNode `json:"nodes"`
	OrphanedNodes []argoNode `json:"orphanedNodes"`
	Hosts         []any      `json:"hosts"`
}

// ── projection ───────────────────────────────────────────────────────────────

// argoHealthFrom maps the native lowercase health vocab (resourceHealth) to the
// Capitalized ArgoCD vocab the UI renders.
func argoHealthFrom(native string) string {
	switch native {
	case HealthHealthy:
		return "Healthy"
	case HealthProgressing:
		return "Progressing"
	case HealthDegraded:
		return "Degraded"
	case HealthSuspended:
		return "Suspended"
	case HealthMissing:
		return "Missing"
	default:
		return "Unknown"
	}
}

// argoSyncFrom maps the native sync verdict to the Capitalized ArgoCD vocab.
func argoSyncFrom(native string) string {
	switch native {
	case SyncSynced:
		return "Synced"
	case SyncOutOfSync:
		return "OutOfSync"
	default:
		return "Unknown"
	}
}

// deployManifestRepo is the desired-state source the projection reports as the
// Application's git source — the manifest repo the engine syncs. Display-only.
const deployManifestRepo = "https://git.hanzo.ai/hanzoai/universe"

// projectApp maps ONE operator App CR (+ its running image tag) to an ArgoCD
// Application. Reuses observeApplication's native derivation, then reshapes to
// the v1alpha1 JSON — one source of truth (the App CR), two shapes.
func projectApp(cr *unstructured.Unstructured, ns, runningTag string) argoApp {
	native := observeApplication(cr, ns, runningTag)
	repository, _, _ := unstructured.NestedString(cr.Object, "spec", "image", "repository")
	tag := native.Version
	image := repository
	if tag != "" {
		image = repository + ":" + tag
	}
	return argoApp{
		APIVersion: "argoproj.io/v1alpha1",
		Kind:       "Application",
		Metadata: argoMeta{
			Name:              native.Name,
			Namespace:         ns,
			UID:               string(cr.GetUID()),
			CreationTimestamp: cr.GetCreationTimestamp().Format("2006-01-02T15:04:05Z07:00"),
			Labels:            map[string]string{"argocd.argoproj.io/instance": native.Name, "hanzo.ai/env": native.Env},
		},
		Spec: argoSpec{
			Source:      argoSource{RepoURL: deployManifestRepo, Path: "infra/k8s/operator/crs", TargetRevision: "main"},
			Destination: argoDestination{Server: "https://kubernetes.default.svc", Namespace: ns},
			Project:     "default",
		},
		Status: argoStatus{
			Sync:      argoSyncStatus{Status: argoSyncFrom(native.Sync), Revision: native.Version},
			Health:    argoHealth{Status: argoHealthFrom(native.Health), Message: native.HealthMessage},
			Resources: []argoResourceStatus{},
			Summary:   argoSummary{Images: nonEmpty(image)},
		},
	}
}

// projectTree reshapes the native buildTree []Node into an ArgoCD ApplicationTree.
func projectTree(nodes []Node) argoTree {
	out := argoTree{Nodes: make([]argoNode, 0, len(nodes)), OrphanedNodes: []argoNode{}, Hosts: []any{}}
	for i := range nodes {
		n := &nodes[i]
		an := argoNode{
			argoResourceRef: argoResourceRef{
				Group: n.Group, Version: n.Version, Kind: n.Kind,
				Namespace: n.Namespace, Name: n.Name, UID: n.UID,
			},
			ResourceVersion: "",
			CreatedAt:       n.CreatedAt,
			Health:          &argoHealth{Status: argoHealthFrom(n.Health), Message: n.HealthMessage},
		}
		for _, p := range n.ParentRefs {
			an.ParentRefs = append(an.ParentRefs, argoResourceRef{
				Group: p.Group, Version: p.Version, Kind: p.Kind, Namespace: p.Namespace, Name: p.Name,
			})
		}
		if n.Version != "" {
			an.Info = append(an.Info, argoInfoItem{Name: "Image Tag", Value: n.Version})
		}
		out.Nodes = append(out.Nodes, an)
	}
	return out
}

func nonEmpty(s string) []string {
	if s == "" {
		return nil
	}
	return []string{s}
}
