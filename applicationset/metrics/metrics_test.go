package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/yaml"

	"github.com/hanzoai/cd/applicationset/utils"
	"github.com/hanzoai/cd/pkg/apis/application/v1alpha1"
	metricsutil "github.com/hanzoai/cd/util/metrics"
)

var (
	applicationsetNamespaces = []string{"cd", "test-namespace1"}

	filter = func(appset *v1alpha1.ApplicationSet) bool {
		return utils.IsNamespaceAllowed(applicationsetNamespaces, appset.Namespace)
	}

	collectedLabels = []string{"included/test"}
)

const fakeAppsetList = `
apiVersion: apps.hanzo.ai/v1alpha1
kind: ApplicationSet
metadata:
  name: test1
  namespace: cd
  labels:
    included/test: test
    not-included.label/test: test
spec:
  generators:
  - git:
      directories:
      - path: test/*
      repoURL: https://github.com/test/test.git
      revision: HEAD
  template:
    metadata:
      name: '{{.path.basename}}'
    spec:
      destination:
        namespace: '{{.path.basename}}'
        server: https://kubernetes.default.svc
      project: default
      source:
        path: '{{.path.path}}'
        repoURL: https://github.com/test/test.git
        targetRevision: HEAD
status:
  resources:
  - group: apps.hanzo.ai
    health:
      status: Missing
    kind: Application
    name: test-app1
    namespace: cd
    status: OutOfSync
    version: v1alpha1
  - group: apps.hanzo.ai
    health:
      status: Missing
    kind: Application
    name: test-app2
    namespace: cd
    status: OutOfSync
    version: v1alpha1
  conditions:
  - lastTransitionTime: "2024-01-01T00:00:00Z"
    message: Successfully generated parameters for all Applications
    reason: ApplicationSetUpToDate
    status: "False"
    type: ErrorOccurred
  - lastTransitionTime: "2024-01-01T00:00:00Z"
    message: Successfully generated parameters for all Applications
    reason: ParametersGenerated
    status: "True"
    type: ParametersGenerated
  - lastTransitionTime: "2024-01-01T00:00:00Z"
    message: ApplicationSet up to date
    reason: ApplicationSetUpToDate
    status: "True"
    type: ResourcesUpToDate
---
apiVersion: apps.hanzo.ai/v1alpha1
kind: ApplicationSet
metadata:
  name: test2
  namespace: cd
  labels:
    not-included.label/test: test
spec:
  generators:
  - git:
      directories:
      - path: test/*
      repoURL: https://github.com/test/test.git
      revision: HEAD
  template:
    metadata:
      name: '{{.path.basename}}'
    spec:
      destination:
        namespace: '{{.path.basename}}'
        server: https://kubernetes.default.svc
      project: default
      source:
        path: '{{.path.path}}'
        repoURL: https://github.com/test/test.git
        targetRevision: HEAD
---
apiVersion: apps.hanzo.ai/v1alpha1
kind: ApplicationSet
metadata:
  name: should-be-filtered-out
  namespace: not-allowed
spec:
  generators:
  - git:
      directories:
      - path: test/*
      repoURL: https://github.com/test/test.git
      revision: HEAD
  template:
    metadata:
      name: '{{.path.basename}}'
    spec:
      destination:
        namespace: '{{.path.basename}}'
        server: https://kubernetes.default.svc
      project: default
      source:
        path: '{{.path.path}}'
        repoURL: https://github.com/test/test.git
        targetRevision: HEAD
`

func newFakeAppsets(fakeAppsetYAML string) []v1alpha1.ApplicationSet {
	var results []v1alpha1.ApplicationSet

	appsetRawYamls := strings.SplitSeq(fakeAppsetYAML, "---")

	for appsetRawYaml := range appsetRawYamls {
		var appset v1alpha1.ApplicationSet
		err := yaml.Unmarshal([]byte(appsetRawYaml), &appset)
		if err != nil {
			panic(err)
		}

		results = append(results, appset)
	}

	return results
}

func TestApplicationsetCollector(t *testing.T) {
	appsetList := newFakeAppsets(fakeAppsetList)
	client := initializeClient(appsetList)
	metrics.Registry = prometheus.NewRegistry()

	appsetCollector := newAppsetCollector(utils.NewAppsetLister(client), collectedLabels, filter)

	metrics.Registry.MustRegister(appsetCollector)
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/metrics", http.NoBody)
	require.NoError(t, err)
	rr := httptest.NewRecorder()
	handler := promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{})
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	// Test correct appset_info and owned applications
	assert.Contains(t, rr.Body.String(), `
cd_appset_info{name="test1",namespace="cd",resource_update_status="ApplicationSetUpToDate"} 1
`)
	assert.Contains(t, rr.Body.String(), `
cd_appset_owned_applications{name="test1",namespace="cd"} 2
`)
	// Test labels collection - should not include labels not included in the list of collected labels and include the ones that do.
	assert.Contains(t, rr.Body.String(), `
cd_appset_labels{label_included_test="test",name="test1",namespace="cd"} 1
`)
	assert.NotContains(t, rr.Body.String(), normalizeLabel("not-included.label/test"))
	// If collected label is not present on the applicationset the value should be empty
	assert.Contains(t, rr.Body.String(), `
cd_appset_labels{label_included_test="",name="test2",namespace="cd"} 1
`)
	// If ResourcesUpToDate condition is not present on the applicationset the status should be reported as 'Unknown'
	assert.Contains(t, rr.Body.String(), `
cd_appset_info{name="test2",namespace="cd",resource_update_status="Unknown"} 1
`)
	// If there are no resources on the applicationset the owned application gague should return 0
	assert.Contains(t, rr.Body.String(), `
cd_appset_owned_applications{name="test2",namespace="cd"} 0
`)
	// Test that filter is working
	assert.NotContains(t, rr.Body.String(), `name="should-be-filtered-out"`)
}

func TestObserveReconcile(t *testing.T) {
	appsetList := newFakeAppsets(fakeAppsetList)
	client := initializeClient(appsetList)
	metrics.Registry = prometheus.NewRegistry()

	appsetMetrics := NewApplicationsetMetrics(utils.NewAppsetLister(client), collectedLabels, filter)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/metrics", http.NoBody)
	require.NoError(t, err)
	rr := httptest.NewRecorder()
	handler := promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{})
	appsetMetrics.ObserveReconcile(&appsetList[0], 5*time.Second)
	handler.ServeHTTP(rr, req)
	assert.Contains(t, rr.Body.String(), `
cd_appset_reconcile_sum{name="test1",namespace="cd"} 5
`)
	// If there are no resources on the applicationset the owned application gague should return 0
	assert.Contains(t, rr.Body.String(), `
cd_appset_reconcile_count{name="test1",namespace="cd"} 1
`)
}

func initializeClient(appsets []v1alpha1.ApplicationSet) ctrlclient.WithWatch {
	scheme := runtime.NewScheme()
	err := v1alpha1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}

	var clientObjects []ctrlclient.Object

	for _, appset := range appsets {
		clientObjects = append(clientObjects, appset.DeepCopy())
	}

	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(clientObjects...).Build()
}

func normalizeLabel(label string) string {
	return metricsutil.NormalizeLabels("label", []string{label})[0]
}
