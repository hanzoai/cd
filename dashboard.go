// dashboard.go — serves the full ArgoCD React UI (Hanzo monochrome) at
// /v1/deploy/ui/, fed the App-CR projection (projection.go) through a
// reimplementation of the SUBSET of the ArgoCD api-server REST surface the UI
// needs. NO argocd api-server, NO repo-server, NO redis, NO stored
// Application/AppProject CRD — every response is synthesized from our operator
// App CRs + the embedded engine's readers.
//
// Layout (base href = /v1/deploy/ui/, so the UI's /api/v1 calls land under it):
//
//	GET  /v1/deploy/ui/                         → the SPA (index.html, base-href rewritten)
//	GET  /v1/deploy/ui/*                        → static assets, else SPA fallback
//	GET  /v1/deploy/ui/api/v1/settings          → AuthSettings (auth disabled; IAM gates at the edge)
//	GET  /v1/deploy/ui/api/v1/session/userinfo  → {loggedIn:true,...}
//	GET  /v1/deploy/ui/api/version              → VersionMessage
//	GET  /v1/deploy/ui/api/v1/account/can-i/*   → {"value":"yes"}
//	GET  /v1/deploy/ui/api/v1/applications          → ApplicationList (projected)
//	GET  /v1/deploy/ui/api/v1/applications/{name}    → Application (projected)
//	GET  /v1/deploy/ui/api/v1/applications/{name}/resource-tree → ApplicationTree
//	POST /v1/deploy/ui/api/v1/applications/{name}/sync     → request App-CR reconcile
//	POST /v1/deploy/ui/api/v1/applications/{name}/rollback → request App-CR reconcile
//
// Every route is SuperAdmin-gated (c.IsAdmin) — the same fail-closed predicate
// as the rest of /v1/deploy; the argocd UI's own auth is disabled because IAM
// owns identity at the edge. AppProject → IAM/Org (no argocd RBAC).
package deploy

import (
	"bytes"
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/hanzoai/cloud"
	"github.com/zap-proto/zip"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

// dashFS holds the built monochrome ArgoCD bundle. `all:` embeds the whole tree.
// At a plain `go build` this is the committed fallback shell (webui/dist/
// index.html); the image build runs the hanzoai/deploy rebrand/hanzo-monochrome
// `yarn build` and OVERWRITES webui/dist with the real bundle BEFORE `go build`
// (the Dockerfile deploy-ui stage / `make deploy-ui`), exactly as webui.go does
// for the console. One artifact, one binary.
//
//go:embed all:webui/dist
var dashFS embed.FS

const dashPrefix = "/v1/deploy/ui"

var baseHRefRe = regexp.MustCompile(`<base href="[^"]*">`)

// registerDashboardRoutes wires the ArgoCD-UI-compatible surface. Specific API
// routes are registered BEFORE the static wildcard so they win; the wildcard is
// the SPA fallback. Called from routes() (deploy.go) after the native routes.
func registerDashboardRoutes(app *zip.App, s *cloud.Service[state]) {
	// Bootstrap (the SPA awaits settings + userinfo before first render).
	app.Get(dashPrefix+"/api/v1/settings", guard(s, cloud.Handle(s, dashSettings)))
	app.Get(dashPrefix+"/api/v1/session/userinfo", guard(s, cloud.Handle(s, dashUserInfo)))
	app.Get(dashPrefix+"/api/version", guard(s, cloud.Handle(s, dashVersion)))
	app.Get(dashPrefix+"/api/v1/account/can-i/*", guard(s, cloud.Handle(s, dashCanI)))

	// Applications projection (read).
	app.Get(dashPrefix+"/api/v1/applications", guard(s, cloud.Handle(s, dashAppList)))
	app.Get(dashPrefix+"/api/v1/applications/:name", guard(s, cloud.Handle(s, dashApp)))
	app.Get(dashPrefix+"/api/v1/applications/:name/resource-tree", guard(s, cloud.Handle(s, dashResourceTree)))

	// Actions → App-CR reconcile ops.
	app.Post(dashPrefix+"/api/v1/applications/:name/sync", guard(s, cloud.Handle(s, dashSync)))
	app.Post(dashPrefix+"/api/v1/applications/:name/rollback", guard(s, cloud.Handle(s, dashSync)))

	// Static SPA — terminal wildcard, registered LAST.
	app.All(dashPrefix, guard(s, cloud.Handle(s, dashStatic)))
	app.All(dashPrefix+"/*", guard(s, cloud.Handle(s, dashStatic)))
}

// ── bootstrap ────────────────────────────────────────────────────────────────

func dashSettings(s *cloud.Service[state], c *zip.Ctx) error {
	return c.JSON(http.StatusOK, map[string]any{
		"url":                       "https://cd.hanzo.ai",
		"statusBadgeEnabled":        false,
		"statusBadgeRootUrl":        "",
		"oidcConfig":                nil,
		"dexConfig":                 map[string]any{"connectors": []any{}},
		"googleAnalytics":           map[string]any{"trackingID": "", "anonymizeUsers": true},
		"help":                      map[string]any{"chatUrl": "", "chatText": "", "binaryUrls": map[string]any{}},
		"plugins":                   []any{},
		"userLoginsDisabled":        true,
		"kustomizeVersions":         []any{},
		"uiCssURL":                  "",
		"uiBannerContent":           "",
		"execEnabled":               false,
		"appsInAnyNamespaceEnabled": false,
		"hydratorEnabled":           false,
		"syncWithReplaceAllowed":    false,
	})
}

func dashUserInfo(s *cloud.Service[state], c *zip.Ctx) error {
	// IAM authenticated the request at the edge; report the principal.
	user := c.User()
	if user == "" {
		user = "admin"
	}
	return c.JSON(http.StatusOK, map[string]any{
		"loggedIn": true,
		"username": user,
		"iss":      "argocd", // keep == argocd so the UI never triggers an SSO redirect
		"groups":   []string{},
	})
}

func dashVersion(s *cloud.Service[state], c *zip.Ctx) error {
	// PascalCase keys (VersionMessage wire shape).
	return c.JSON(http.StatusOK, map[string]any{
		"Version":   "hanzo-cd (projection)",
		"BuildDate": time.Now().UTC().Format(time.RFC3339),
		"GoVersion": "", "Compiler": "gc", "Platform": "linux/amd64",
	})
}

func dashCanI(s *cloud.Service[state], c *zip.Ctx) error {
	// Every route is already SuperAdmin-gated; report yes so buttons enable.
	return c.JSON(http.StatusOK, map[string]any{"value": "yes"})
}

// ── applications projection ──────────────────────────────────────────────────

func dashAppList(s *cloud.Service[state], c *zip.Ctx) error {
	if err := ready(s); err != nil {
		return err
	}
	list := argoAppList{APIVersion: "argoproj.io/v1alpha1", Kind: "ApplicationList", Metadata: argoListMeta{}, Items: []argoApp{}}
	for _, ns := range scanOrder() {
		crs, err := listAppCRs(s, c.Context(), ns)
		if err != nil {
			return k8sErr(s, "list", err)
		}
		running := runningVersions(s, c.Context(), ns)
		for i := range crs {
			list.Items = append(list.Items, projectApp(&crs[i], ns, running[crs[i].GetName()]))
		}
	}
	return c.JSON(http.StatusOK, list)
}

func dashApp(s *cloud.Service[state], c *zip.Ctx) error {
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
	running := runningVersions(s, c.Context(), ns)
	app := projectApp(cr, ns, running[name])
	// Detail view: populate status.resources from the reconciled tree.
	tree := projectTree(buildTree(s, c.Context(), ns, name, cr))
	for _, n := range tree.Nodes {
		app.Status.Resources = append(app.Status.Resources, argoResourceStatus{
			Group: n.Group, Version: n.Version, Kind: n.Kind, Namespace: n.Namespace,
			Name: n.Name, Status: app.Status.Sync.Status, Health: n.Health,
		})
	}
	return c.JSON(http.StatusOK, app)
}

func dashResourceTree(s *cloud.Service[state], c *zip.Ctx) error {
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
	return c.JSON(http.StatusOK, projectTree(buildTree(s, c.Context(), ns, name, cr)))
}

// dashSync requests an operator reconcile of the App CR (the sync + rollback UI
// actions both map to "reconcile this App now" — the App CR is the source of
// truth; rollback-by-revision is the image-pin follow-on). Returns the projected
// Application (the UI only checks for a non-error response).
func dashSync(s *cloud.Service[state], c *zip.Ctx) error {
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
	cr, gvr, err := getAppCR(s, c.Context(), ns, name)
	if err != nil {
		return k8sErr(s, "get", err)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	patch, _ := json.Marshal(map[string]any{"metadata": map[string]any{"annotations": map[string]any{syncAnnotation: now}}})
	if _, err := s.State.dyn.Resource(gvr).Namespace(ns).Patch(c.Context(), name, k8stypes.MergePatchType, patch, metav1.PatchOptions{}); err != nil {
		return k8sErr(s, "patch", err)
	}
	s.Log.Info("dashboard sync requested", "app", name, "namespace", ns, "actor", c.User())
	return c.JSON(http.StatusOK, projectApp(cr, ns, runningVersions(s, c.Context(), ns)[name]))
}

// ── static SPA ───────────────────────────────────────────────────────────────

// dashStatic serves the embedded UI: an existing asset by path, else the SPA
// index.html (base-href rewritten to /v1/deploy/ui/) for client-side routes.
func dashStatic(s *cloud.Service[state], c *zip.Ctx) error {
	sub, err := fs.Sub(dashFS, "webui/dist")
	if err != nil {
		return zip.Errorf(http.StatusInternalServerError, "dashboard assets: %v", err)
	}
	rel := strings.TrimPrefix(strings.TrimPrefix(c.Path(), dashPrefix), "/")
	if rel == "" {
		return serveDashIndex(c, sub)
	}
	f, err := sub.Open(rel)
	if err != nil {
		// Unknown path under the SPA root → client-side route → index.html.
		return serveDashIndex(c, sub)
	}
	defer f.Close()
	data, err := fs.ReadFile(sub, rel)
	if err != nil {
		return serveDashIndex(c, sub)
	}
	if ct := contentTypeFor(rel); ct != "" {
		c.SetHeader("Content-Type", ct)
	}
	return c.Bytes(http.StatusOK, data)
}

func serveDashIndex(c *zip.Ctx, sub fs.FS) error {
	data, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		return zip.Errorf(http.StatusInternalServerError, "dashboard index: %v", err)
	}
	// Rewrite <base href> so the SPA's /api/v1 + router basename resolve under
	// /v1/deploy/ui/ (the upstream server does the same via replaceBaseHRef).
	data = baseHRefRe.ReplaceAll(data, []byte(`<base href="`+dashPrefix+`/">`))
	if !bytes.Contains(data, []byte("<base href=")) {
		data = bytes.Replace(data, []byte("<head>"), []byte(`<head><base href="`+dashPrefix+`/">`), 1)
	}
	c.SetHeader("Content-Type", "text/html; charset=utf-8")
	return c.Bytes(http.StatusOK, data)
}

func contentTypeFor(name string) string {
	switch {
	case strings.HasSuffix(name, ".js"):
		return "application/javascript"
	case strings.HasSuffix(name, ".css"):
		return "text/css"
	case strings.HasSuffix(name, ".html"):
		return "text/html; charset=utf-8"
	case strings.HasSuffix(name, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(name, ".json"):
		return "application/json"
	case strings.HasSuffix(name, ".woff2"):
		return "font/woff2"
	case strings.HasSuffix(name, ".png"):
		return "image/png"
	case strings.HasSuffix(name, ".ico"):
		return "image/x-icon"
	}
	return ""
}
