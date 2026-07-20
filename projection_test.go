package deploy

import (
	"encoding/json"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func projFixture(name, ns, repo, tag, phase string, replicas, ready int64) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "hanzo.ai/v1", "kind": "App",
		"metadata": map[string]any{"name": name, "namespace": ns, "uid": "uid-" + name},
		"spec":     map[string]any{"role": "service", "image": map[string]any{"repository": repo, "tag": tag}},
		"status":   map[string]any{"phase": phase, "replicas": replicas, "readyReplicas": ready, "availableReplicas": ready},
	}}
}

// TestProjectApp_ShapeIsUIRenderable asserts the projection produces the exact
// minimal ArgoCD Application JSON the React UI needs (per the api-server
// contract): the fields the list tiles + detail header destructure must all be
// present and correctly typed, or the SPA throws.
func TestProjectApp_ShapeIsUIRenderable(t *testing.T) {
	app := projectApp(projFixture("cloud", "hanzo", "ghcr.io/hanzoai/cloud", "v1.2.3", "Running", 1, 1), "hanzo", "v1.2.3")

	if app.APIVersion != "argoproj.io/v1alpha1" || app.Kind != "Application" {
		t.Fatalf("wrong TypeMeta: %s/%s", app.APIVersion, app.Kind)
	}
	if app.Metadata.Name != "cloud" || app.Metadata.Namespace != "hanzo" {
		t.Fatalf("wrong metadata: %+v", app.Metadata)
	}
	if app.Spec.Source.RepoURL == "" || app.Spec.Destination.Server == "" {
		t.Fatal("spec.source.repoURL and spec.destination.server are required by the UI")
	}
	// Running + declared==running ⇒ Healthy + Synced (Capitalized ArgoCD vocab).
	if app.Status.Health.Status != "Healthy" {
		t.Fatalf("health = %q, want Healthy", app.Status.Health.Status)
	}
	if app.Status.Sync.Status != "Synced" {
		t.Fatalf("sync = %q, want Synced", app.Status.Sync.Status)
	}

	// The UI's parseAppFields requires status.resources to be an array and
	// status.summary to be an object — assert they marshal as such.
	b, err := json.Marshal(app)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	status := m["status"].(map[string]any)
	if _, ok := status["resources"].([]any); !ok {
		t.Fatalf("status.resources must be a JSON array, got %T", status["resources"])
	}
	if _, ok := status["summary"].(map[string]any); !ok {
		t.Fatalf("status.summary must be a JSON object, got %T", status["summary"])
	}
	if m["spec"].(map[string]any)["source"].(map[string]any)["repoURL"] == "" {
		t.Fatal("spec.source.repoURL empty")
	}
}

// TestProjectApp_HealthAndSyncVocab covers the native→ArgoCD vocab mapping.
func TestProjectApp_HealthAndSyncVocab(t *testing.T) {
	// 1 desired / 0 ready ⇒ Degraded; declared != running ⇒ OutOfSync.
	app := projectApp(projFixture("chat", "hanzo", "ghcr.io/hanzoai/chat", "v2", "Creating", 1, 0), "hanzo", "v1")
	if app.Status.Health.Status != "Degraded" {
		t.Fatalf("health = %q, want Degraded", app.Status.Health.Status)
	}
	if app.Status.Sync.Status != "OutOfSync" {
		t.Fatalf("sync = %q, want OutOfSync", app.Status.Sync.Status)
	}
}

// TestProjectTree_ShapeIsUIRenderable asserts the ApplicationTree projection.
func TestProjectTree_ShapeIsUIRenderable(t *testing.T) {
	nodes := []Node{
		{ResourceRef: ResourceRef{Group: "apps", Version: "v1", Kind: "Deployment", Namespace: "hanzo", Name: "cloud"}, UID: "d1", Health: HealthHealthy, Version: "v1.2.3"},
		{ResourceRef: ResourceRef{Version: "v1", Kind: "Pod", Namespace: "hanzo", Name: "cloud-xyz"}, UID: "p1", Health: HealthHealthy, ParentRefs: []ResourceRef{{Group: "apps", Version: "v1", Kind: "ReplicaSet", Namespace: "hanzo", Name: "cloud-abc"}}},
	}
	tree := projectTree(nodes)
	if len(tree.Nodes) != 2 {
		t.Fatalf("nodes = %d, want 2", len(tree.Nodes))
	}
	b, _ := json.Marshal(tree)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	if _, ok := m["nodes"].([]any); !ok {
		t.Fatal("tree.nodes must be a JSON array")
	}
	if _, ok := m["orphanedNodes"]; !ok {
		t.Fatal("tree.orphanedNodes must be present")
	}
	// Node health mapped to Capitalized vocab; image surfaced as an info item.
	n0 := m["nodes"].([]any)[0].(map[string]any)
	if n0["health"].(map[string]any)["status"] != "Healthy" {
		t.Fatalf("node health = %v", n0["health"])
	}
	if n0["kind"] != "Deployment" || n0["name"] != "cloud" {
		t.Fatalf("node ref wrong: %v", n0)
	}
}

// TestBaseHRefRewrite asserts the SPA index gets the /v1/deploy/ui/ base href so
// the UI's /api/v1 calls + router basename resolve under the prefix.
func TestBaseHRefRewrite(t *testing.T) {
	in := []byte(`<!doctype html><html><head><base href="/"><title>x</title></head><body></body></html>`)
	out := baseHRefRe.ReplaceAll(in, []byte(`<base href="`+dashPrefix+`/">`))
	want := `<base href="/v1/deploy/ui/">`
	if string(out) == string(in) {
		t.Fatal("base href was not rewritten")
	}
	if !projContains(string(out), want) {
		t.Fatalf("rewritten index missing %q: %s", want, out)
	}
}

func projContains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
