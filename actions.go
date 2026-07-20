// actions.go — the two GitOps write actions.
//
//	POST /v1/deploy/{name}/rollback — pin the CR image tag to a prior clean semver.
//	  It REUSES the P1 release seam (cloud.OnServiceRelease → clients/paas
//	  releaseService), so the clean-semver gate + idempotent spec.image patch live
//	  in exactly ONE place; the operator reconciles the rollout.
//	POST /v1/deploy/{name}/sync — request an operator reconcile now by touching the
//	  CR (an annotation bump the operator's watch observes). Today the CR is the
//	  desired source, so sync = nudge-reconcile; when git.hanzo.ai is the source it
//	  becomes apply-desired-from-git, same endpoint.
package deploy

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/hanzoai/cloud"
	"github.com/hanzoai/cloud/clients/paas"
	"github.com/zap-proto/zip"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

const syncAnnotation = "gitops.hanzo.ai/sync-requested-at"

// rollback pins {name}'s Service CR image to a prior clean-semver tag, delegating
// the CR patch to the P1 release seam. Body: {"tag":"vX.Y.Z"}.
func rollback(s *cloud.Service[state], c *zip.Ctx) error {
	if err := ready(s); err != nil {
		return err
	}
	name := reqName(c)
	if !appNameRE.MatchString(name) {
		return zip.ErrBadRequest("name must be a DNS-1123 label")
	}
	var body struct {
		Tag string `json:"tag"`
	}
	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return zip.Errorf(http.StatusBadRequest, "invalid JSON body: %v", err)
	}
	tag := strings.TrimSpace(body.Tag)
	if !paas.IsSemverTag(tag) {
		return zip.ErrBadRequest("'tag' must be a clean semver (vX.Y.Z) — the prior release to pin")
	}
	// The release seam owns the CR patch (one way). It requires the paas control
	// plane co-resident; absent it, fail closed with the honest reason.
	if !cloud.ServiceReleaserRegistered() {
		return zip.Errorf(http.StatusServiceUnavailable, "release plane not available (paas subsystem not co-resident)")
	}

	// Resolve the CR to derive its repository (rollback keeps the repo; only the tag
	// moves) and to 404 an unknown app before firing the release.
	ns, err := resolveNamespace(s, c, name)
	if err != nil {
		return err
	}
	cr, _, err := getAppCR(s, c.Context(), ns, name)
	if err != nil {
		return k8sErr(s, "get", err)
	}
	repository, _, _ := unstructured.NestedString(cr.Object, "spec", "image", "repository")
	if repository == "" {
		return zip.Errorf(http.StatusUnprocessableEntity, "application %q has no spec.image.repository to roll back", name)
	}
	image := repository + ":" + tag

	if rerr := cloud.OnServiceRelease(c.Context(), cloud.ServiceReleaseEvent{Service: name, Image: image}); rerr != nil {
		return zip.Errorf(http.StatusBadGateway, "rollback failed: %v", rerr)
	}
	s.Log.Info("gitops rollback applied via release seam", "app", name, "namespace", ns, "tag", tag, "actor", c.User(), "requestID", c.RequestID())

	// Re-read for an honest post-rollback row (running version lags until the
	// operator rolls the Deployment).
	updated, _, gerr := getAppCR(s, c.Context(), ns, name)
	if gerr != nil {
		updated = cr
	}
	running := runningVersions(s, c.Context(), ns)
	return c.JSON(http.StatusOK, map[string]any{
		"rolledBack":  true,
		"target":      ns + "/" + name,
		"tag":         tag,
		"application": observeApplication(updated, ns, running[name]),
	})
}

// sync requests an operator reconcile of {name}'s CR now by bumping a sync
// annotation the operator's watch observes. Returns the honest requested state.
func sync(s *cloud.Service[state], c *zip.Ctx) error {
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
	// Patch the SAME kind the CR lives under (App forward, Service in transition).
	_, gvr, err := getAppCR(s, c.Context(), ns, name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return zip.ErrNotFound("application not found")
		}
		return k8sErr(s, "get", err)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	patch, _ := json.Marshal(map[string]any{"metadata": map[string]any{"annotations": map[string]any{syncAnnotation: now}}})
	if _, err := s.State.dyn.Resource(gvr).Namespace(ns).Patch(c.Context(), name, k8stypes.MergePatchType, patch, metav1.PatchOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			return zip.ErrNotFound("application not found")
		}
		return k8sErr(s, "patch", err)
	}
	s.Log.Info("gitops sync requested", "app", name, "namespace", ns, "at", now, "actor", c.User(), "requestID", c.RequestID())
	return c.JSON(http.StatusOK, map[string]any{
		"synced":      true,
		"target":      ns + "/" + name,
		"requestedAt": now,
		"note":        "operator reconcile requested; desired source is the Service CR (git.hanzo.ai manifest sync is the follow-on)",
	})
}
