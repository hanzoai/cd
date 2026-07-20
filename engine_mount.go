// engine_mount.go wires the embedded gitops-engine (engine.go) into the
// /v1/deploy surface: a SuperAdmin-gated, one-shot reconcile endpoint that
// renders the configured git source and syncs it → cluster. This is the write
// half of /v1/deploy that replaces the retired universe-crs Application — the
// operator still renders each App CR into workloads (the domain half).
//
// Fail-safe: the whole path is gated by DEPLOY_ENGINE_ENABLED (default off), so
// the first deploy of this binary is inert and the engine is turned on
// deliberately after the shadow proof.
package deploy

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	"github.com/hanzoai/cloud"
	"github.com/zap-proto/zip"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	synccommon "github.com/hanzoai/deploy/gitops-engine/pkg/sync/common"
)

// Engine config — all optional; defaults target the live universe manifest repo
// (the exact source universe-crs syncs). Configure only what must vary.
func engineEnabled() bool { return os.Getenv("DEPLOY_ENGINE_ENABLED") == "true" }
func enginePrune() bool   { return os.Getenv("DEPLOY_ENGINE_PRUNE") == "true" }

// pruneFuse bounds a single reconcile's deletions (RED HIGH-1). Conservative
// defaults: at most 10 objects OR 20% of the managed set, whichever is smaller,
// unless explicitly raised. A silent empty/partial render trips the fuse instead
// of sweeping the fleet.
func pruneFuse() PruneFuse {
	return PruneFuse{
		MaxDeletions: envInt("DEPLOY_ENGINE_PRUNE_MAX", 10),
		MaxRatio:     envFloat("DEPLOY_ENGINE_PRUNE_MAX_RATIO", 0.20),
	}
}

func envInt(k string, d int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return d
}

func envFloat(k string, d float64) float64 {
	if v := os.Getenv(k); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return d
}
func engineRepo() string      { return envOr("DEPLOY_ENGINE_REPO", "https://github.com/hanzoai/universe") }
func engineRef() string       { return envOr("DEPLOY_ENGINE_REF", "main") }
func enginePath() string      { return envOr("DEPLOY_ENGINE_PATH", "infra/k8s/operator/crs") }
func engineInstance() string  { return envOr("DEPLOY_ENGINE_INSTANCE", "universe") }
func engineDefaultNS() string { return envOr("DEPLOY_ENGINE_NAMESPACE", "hanzo") }

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

// registerEngineRoutes adds the engine (write) routes alongside the existing
// read/visualize routes. Called from routes() in deploy.go.
func registerEngineRoutes(app *zip.App, s *cloud.Service[state]) {
	app.Post("/v1/deploy/reconcile", guard(s, cloud.Handle(s, engineReconcile)))
}

// engineReconcile is POST /v1/deploy/reconcile — a SuperAdmin-gated, one-shot
// engine sync: render the configured git source and reconcile it → cluster via
// the embedded gitops-engine (three-way server-side apply, scoped prune,
// per-resource health). The write half that replaces universe-crs.
func engineReconcile(s *cloud.Service[state], c *zip.Ctx) error {
	if !engineEnabled() {
		return zip.Errorf(http.StatusServiceUnavailable, "deploy engine disabled (set DEPLOY_ENGINE_ENABLED=true)")
	}
	cfg, err := engineRestConfig()
	if err != nil {
		return zip.Errorf(http.StatusServiceUnavailable, "engine: kube config: %v", err)
	}
	ctx := c.Context()

	rec := newReconciler(cfg, []string{engineDefaultNS()}, engineInstance(), logr.Discard())
	stop, err := rec.run()
	if err != nil {
		return zip.Errorf(http.StatusBadGateway, "engine start: %v", err)
	}
	defer stop()
	// Let the informer cache warm before the first sync so live state is known.
	time.Sleep(2 * time.Second)

	objs, revision, err := gitSource{repo: engineRepo(), ref: engineRef(), path: enginePath()}.render(ctx)
	if err != nil {
		return zip.Errorf(http.StatusBadGateway, "engine: render git source: %v", err)
	}

	results, err := rec.reconcile(ctx, objs, revision, engineDefaultNS(), enginePrune(), pruneFuse())
	if err != nil {
		return zip.Errorf(http.StatusBadGateway, "engine: sync: %v", err)
	}

	synced, pruned, failed := 0, 0, 0
	items := make([]map[string]any, 0, len(results))
	for _, rr := range results {
		switch rr.Status {
		case synccommon.ResultCodeSynced:
			synced++
		case synccommon.ResultCodePruned:
			pruned++
		case synccommon.ResultCodeSyncFailed:
			failed++
		}
		items = append(items, map[string]any{
			"resource": rr.ResourceKey.String(),
			"status":   string(rr.Status),
			"message":  rr.Message,
		})
	}
	s.Log.Info("deploy engine reconcile", "revision", revision, "objects", len(objs),
		"synced", synced, "pruned", pruned, "failed", failed, "prune", enginePrune())

	return c.JSON(http.StatusOK, map[string]any{
		"revision": revision,
		"source":   map[string]any{"repo": engineRepo(), "ref": engineRef(), "path": enginePath()},
		"instance": engineInstance(),
		"prune":    enginePrune(),
		"declared": len(objs),
		"synced":   synced,
		"pruned":   pruned,
		"failed":   failed,
		"results":  items,
	})
}

// engineRestConfig builds a rest.Config from the in-cluster service account,
// falling back to KUBECONFIG for local/dev — the SAME construction as
// newClients() (deploy.go), so the engine talks to the same cluster.
func engineRestConfig() (*rest.Config, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{})
		cfg, err = cc.ClientConfig()
		if err != nil {
			return nil, err
		}
	}
	cfg.UserAgent = userAgent
	return cfg, nil
}
