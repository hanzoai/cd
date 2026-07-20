// engine.go embeds the argo gitops-engine (github.com/hanzoai/deploy/
// gitops-engine, the fork's independently-importable submodule) in-process, so
// the cloud binary reconciles git → cluster the way the retired argocd
// application-controller did — three-way merge (server-side apply), scoped
// prune, drift-correction, health, sync status — with NO separate argocd
// process and NO redis. It is the write/reconcile half of /v1/deploy; the
// existing routes are the read/visualize half.
//
// The apply-set is scoped by a tracking LABEL (deploy.hanzo.ai/instance): the
// isManaged predicate that drives pruning returns true ONLY for live objects
// carrying THIS instance's label, so a prune can never delete an App CR (or any
// object) this plane did not create. This is the prune-safety boundary for the
// 60+ live App CRs — the exact property universe-crs enforced with prune:false,
// kept here and made explicit.
//
// Enablement is opt-in and fail-safe (DEPLOY_ENGINE_ENABLED, default off): the
// first deploy of this binary is inert for the reconcile path, so it ships
// dark and is turned on deliberately after the shadow proof — mirroring the
// operator's gate discipline and the argocd shadow-then-flip cutover.
package deploy

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"

	"github.com/hanzoai/deploy/gitops-engine/pkg/cache"
	"github.com/hanzoai/deploy/gitops-engine/pkg/engine"
	enginehealth "github.com/hanzoai/deploy/gitops-engine/pkg/health"
	enginesync "github.com/hanzoai/deploy/gitops-engine/pkg/sync"
	synccommon "github.com/hanzoai/deploy/gitops-engine/pkg/sync/common"
	"github.com/hanzoai/deploy/gitops-engine/pkg/utils/kube"
)

// engineTrackingLabel scopes an apply-set. Every object the engine declares
// carries deploy.hanzo.ai/instance=<instance>; pruning only ever considers live
// objects with THIS value.
const engineTrackingLabel = "deploy.hanzo.ai/instance"

// engineFieldManager is the server-side-apply manager for git-sourced applies —
// distinct from the operator's own `hanzo-operator` manager, so a delivery apply
// is attributable and never silently fights a per-Kind reconcile.
const engineFieldManager = "hanzo-deploy"

// engineResInfo is cached per live resource; `tracked` records whether the
// object carries THIS instance's tracking label (the prune predicate).
type engineResInfo struct{ tracked bool }

// reconciler embeds the argo gitops-engine over one apply-set (one git source,
// one tracking-label instance) across a fixed set of namespaces.
type reconciler struct {
	instance   string
	namespaces []string
	log        logr.Logger
	cache      cache.ClusterCache
	engine     engine.GitOpsEngine
}

// newReconciler builds an engine-backed reconciler. `namespaces` is the platform
// tier the engine watches (empty = all); `instance` tags this apply-set.
func newReconciler(cfg *rest.Config, namespaces []string, instance string, log logr.Logger) *reconciler {
	cc := cache.NewClusterCache(cfg,
		cache.SetNamespaces(namespaces),
		cache.SetLogr(log),
		cache.SetPopulateResourceInfoHandler(func(un *unstructured.Unstructured, _ bool) (any, bool) {
			v := un.GetLabels()[engineTrackingLabel]
			return &engineResInfo{tracked: v == instance}, v != ""
		}),
	)
	return &reconciler{
		instance:   instance,
		namespaces: namespaces,
		log:        log,
		cache:      cc,
		engine:     engine.NewEngine(cfg, cc, engine.WithLogr(log)),
	}
}

// run starts the cluster informer cache; the returned StopFunc tears it down.
func (r *reconciler) run() (engine.StopFunc, error) { return r.engine.Run() }

// stamp puts the tracking label on every desired object so an applied object is
// cached as managed by THIS instance.
func (r *reconciler) stamp(objs []*unstructured.Unstructured) {
	for _, o := range objs {
		l := o.GetLabels()
		if l == nil {
			l = map[string]string{}
		}
		l[engineTrackingLabel] = r.instance
		o.SetLabels(l)
	}
}

// PruneFuse bounds how much a single reconcile may delete — the circuit breaker
// against a silent empty/partial render sweeping the fleet. Both limits are
// checked; either one trips the fuse. Zero disables that check.
type PruneFuse struct {
	MaxDeletions int     // absolute cap on objects pruned in one reconcile
	MaxRatio     float64 // cap as a fraction of the managed set (0..1)
}

// isProtectedKind is the data-anchor exclusion: PersistentVolumeClaim (delete =
// irreversible data loss) and KMSSecret (kms.hanzo.ai) are NEVER prune
// candidates, even when absent from the desired set. They are still applied
// (target side); they are only removed from the prune decision.
func isProtectedKind(group, kind string) bool {
	if group == "" && kind == "PersistentVolumeClaim" {
		return true
	}
	if kind == "KMSSecret" {
		return true
	}
	return false
}

// managed is the prune predicate: an object is a prune candidate only if it
// carries THIS instance's tracking label AND is not a protected data anchor.
func (r *reconciler) managed(res *cache.Resource) bool {
	ri, ok := res.Info.(*engineResInfo)
	if !ok || !ri.tracked {
		return false
	}
	k := res.ResourceKey()
	return !isProtectedKind(k.Group, k.Kind)
}

// reconcile syncs `target` → cluster at `revision`. With prune, a live object
// carrying THIS instance's tracking label but absent from `target` is deleted —
// UNLESS it is a protected data anchor (PVC/KMSSecret) or the PruneFuse trips.
// An UNTRACKED object is never touched.
//
// Prune safety (RED HIGH-1), all enforced here:
//   - refuse an EMPTY desired set (a silent target=[] would sweep everything);
//   - pre-flight DRY-RUN sizes the prune set before any deletion;
//   - the PruneFuse caps the prune set by count and by ratio;
//   - protected data anchors (PVC/KMSSecret) are excluded from prune entirely.
func (r *reconciler) reconcile(ctx context.Context, target []*unstructured.Unstructured, revision, defaultNS string, prune bool, fuse PruneFuse) ([]synccommon.ResourceSyncResult, error) {
	// (i) Never reconcile nothing — an empty render must not delete the fleet.
	if len(target) == 0 {
		return nil, fmt.Errorf("prune fuse: refusing to reconcile an empty desired set (render produced 0 objects)")
	}
	r.stamp(target)

	if prune {
		// (ii)/(iii) Size the prune set with a dry-run BEFORE any deletion, then
		// apply the fuse. A partial render that would prune most of the fleet is
		// refused here, before a single object is removed.
		dry, err := r.engine.Sync(ctx, target, r.managed, revision, defaultNS,
			enginesync.WithOperationSettings(true /*dryRun*/, true /*prune*/, false, false),
			enginesync.WithLogr(r.log))
		if err != nil {
			return nil, fmt.Errorf("prune fuse dry-run: %w", err)
		}
		pruneN, managedN := 0, 0
		for _, rr := range dry {
			managedN++
			if rr.Status == synccommon.ResultCodePruned {
				pruneN++
			}
		}
		if fuse.MaxDeletions > 0 && pruneN > fuse.MaxDeletions {
			return nil, fmt.Errorf("prune fuse tripped: reconcile would prune %d object(s) (> max %d); refusing — fix the git source or raise DEPLOY_ENGINE_PRUNE_MAX", pruneN, fuse.MaxDeletions)
		}
		if fuse.MaxRatio > 0 && managedN > 0 && float64(pruneN)/float64(managedN) > fuse.MaxRatio {
			return nil, fmt.Errorf("prune fuse tripped: reconcile would prune %d/%d managed (> %.0f%%); refusing", pruneN, managedN, fuse.MaxRatio*100)
		}
	}

	return r.engine.Sync(ctx, target, r.managed, revision, defaultNS,
		enginesync.WithPrune(prune),
		enginesync.WithPruneConfirmed(prune), // prune only after the fuse above confirms it
		enginesync.WithServerSideApply(true),
		enginesync.WithServerSideApplyManager(engineFieldManager),
		enginesync.WithLogr(r.log),
	)
}

// resourceHealth assesses one live object with the engine's built-in per-GVK
// checks — the SAME health library the ArgoCD dashboard uses.
func engineResourceHealth(un *unstructured.Unstructured) (*enginehealth.HealthStatus, error) {
	return enginehealth.GetResourceHealth(un, nil)
}

// gitSource is the desired-state source: a shallow clone of repo@ref, from which
// `path` is rendered into the manifest set. It shells the `git` CLI — the same
// mechanism clients/git and the operator use — so there is ONE git strategy and
// no vendored transport. The revision is the cloned HEAD sha.
type gitSource struct {
	repo string // https URL or local path
	ref  string // branch/tag/sha
	path string // repo-relative dir of CR manifests
}

// render shallow-clones and parses the source into (objects, revision). The
// clone dir is removed before return.
func (g gitSource) render(ctx context.Context) ([]*unstructured.Unstructured, string, error) {
	dir, err := os.MkdirTemp("", "deploy-src-")
	if err != nil {
		return nil, "", fmt.Errorf("workdir: %w", err)
	}
	defer os.RemoveAll(dir)

	clone := exec.CommandContext(ctx, "git", "clone", "--depth", "1", "--single-branch", "--branch", g.ref, g.repo, dir)
	clone.Env = hardenedGitEnv()
	if out, err := clone.CombinedOutput(); err != nil {
		return nil, "", fmt.Errorf("git clone: %v: %s", err, strings.TrimSpace(string(out)))
	}

	rev := exec.CommandContext(ctx, "git", "-C", dir, "rev-parse", "--short", "HEAD")
	rev.Env = hardenedGitEnv()
	revBytes, err := rev.Output()
	if err != nil {
		return nil, "", fmt.Errorf("git rev-parse: %w", err)
	}
	revision := strings.TrimSpace(string(revBytes))

	objs, err := parseManifestDir(filepath.Join(dir, g.path))
	if err != nil {
		return nil, "", err
	}
	return objs, revision, nil
}

// parseManifestDir walks dir RECURSIVELY and splits every *.yaml/*.yml/*.json
// into typed objects, skipping kustomization inputs. Recursive on purpose (RED
// HIGH-1 (v)): a non-recursive read silently drops manifests in subdirectories,
// which — combined with prune — would delete the objects those nested files
// declare. Walking every subdir means the desired set is complete, so prune
// never mistakes a nested-but-present object for a removed one.
func parseManifestDir(dir string) ([]*unstructured.Unstructured, error) {
	var objs []*unstructured.Unstructured
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if name == "kustomization.yaml" || name == "kustomization.yml" {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".yaml" && ext != ".yml" && ext != ".json" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		items, err := kube.SplitYAML(data)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		objs = append(objs, items...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk manifest dir %s: %w", dir, err)
	}
	return objs, nil
}

// hardenedGitEnv is the minimal, credential-free git environment: no interactive
// prompt, no ambient user/system config, no inherited secrets.
func hardenedGitEnv() []string {
	return []string{
		"GIT_TERMINAL_PROMPT=0",
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
		"HOME=" + os.TempDir(),
		"PATH=" + os.Getenv("PATH"),
	}
}
