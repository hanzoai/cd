package cd

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/hanzoai/cd/gitops-engine/pkg/utils/kube"
	"github.com/hanzoai/cd/gitops-engine/pkg/utils/kube/kubetest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"

	"github.com/hanzoai/cd/gitops-engine/pkg/sync/common"

	appv1 "github.com/hanzoai/cd/pkg/apis/application/v1alpha1"
	appclientset "github.com/hanzoai/cd/pkg/client/clientset/versioned/fake"
	"github.com/hanzoai/cd/pkg/client/informers/externalversions/application/v1alpha1"
	applisters "github.com/hanzoai/cd/pkg/client/listers/application/v1alpha1"
	"github.com/hanzoai/cd/reposerver/apiclient"
	"github.com/hanzoai/cd/reposerver/apiclient/mocks"
	"github.com/hanzoai/cd/test"
	"github.com/hanzoai/cd/util/db"
	dbmocks "github.com/hanzoai/cd/util/db/mocks"
	"github.com/hanzoai/cd/util/settings"
)

func TestRefreshApp(t *testing.T) {
	t.Parallel()
	var testApp appv1.Application
	testApp.Name = "test-app"
	testApp.Namespace = "default"
	appClientset := appclientset.NewSimpleClientset(&testApp)
	appIf := appClientset.AppsV1alpha1().Applications("default")
	ht := appv1.HydrateTypeNormal
	_, err := RefreshApp(appIf, "test-app", appv1.RefreshTypeNormal, &ht)
	require.NoError(t, err)
	// For some reason, the fake Application interface doesn't reflect the patch status after Patch(),
	// so can't verify it was set in unit tests.
	// _, ok := newApp.Annotations[common.AnnotationKeyRefresh]
	// assert.True(t, ok)
}

func TestGetAppProjectWithNoProjDefined(t *testing.T) {
	t.Parallel()
	projName := "default"
	namespace := "default"

	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cd-cm",
			Namespace: test.FakeArgoCDNamespace,
			Labels: map[string]string{
				"app.kubernetes.io/part-of": "cd",
			},
		},
	}

	testProj := &appv1.AppProject{
		ObjectMeta: metav1.ObjectMeta{Name: projName, Namespace: namespace},
	}

	var testApp appv1.Application
	testApp.Name = "test-app"
	testApp.Namespace = namespace
	appClientset := appclientset.NewSimpleClientset(testProj)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	indexers := cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}
	informer := v1alpha1.NewAppProjectInformer(appClientset, namespace, 0, indexers)
	go informer.Run(ctx.Done())
	cache.WaitForCacheSync(ctx.Done(), informer.HasSynced)

	kubeClient := fake.NewClientset(&cm)
	settingsMgr := settings.NewSettingsManager(t.Context(), kubeClient, test.FakeArgoCDNamespace)
	appDB := db.NewDB("default", settingsMgr, kubeClient)
	proj, err := GetAppProject(ctx, &testApp, applisters.NewAppProjectLister(informer.GetIndexer()), namespace, settingsMgr, appDB)
	require.NoError(t, err)
	assert.Equal(t, proj.Name, projName)
}

func TestIncludeResource(t *testing.T) {
	t.Parallel()
	// Resource filters format - GROUP:KIND:NAMESPACE/NAME or GROUP:KIND:NAME
	var (
		blankValues = appv1.SyncOperationResource{Group: "", Kind: "", Name: "", Namespace: "", Exclude: false}
		// *:*:*
		includeAllResources = appv1.SyncOperationResource{Group: "*", Kind: "*", Name: "*", Namespace: "", Exclude: false}
		// !*:*:*
		excludeAllResources = appv1.SyncOperationResource{Group: "*", Kind: "*", Name: "*", Namespace: "", Exclude: true}
		// *:Service:*
		includeAllServiceResources = appv1.SyncOperationResource{Group: "*", Kind: "Service", Name: "*", Namespace: "", Exclude: false}
		// !*:Service:*
		excludeAllServiceResources = appv1.SyncOperationResource{Group: "*", Kind: "Service", Name: "*", Namespace: "", Exclude: true}
		// apps:ReplicaSet:backend
		includeAllReplicaSetResource = appv1.SyncOperationResource{Group: "apps", Kind: "ReplicaSet", Name: "*", Namespace: "", Exclude: false}
		// apps:ReplicaSet:backend
		includeReplicaSetResource = appv1.SyncOperationResource{Group: "apps", Kind: "ReplicaSet", Name: "backend", Namespace: "", Exclude: false}
		// !apps:ReplicaSet:backend
		excludeReplicaSetResource = appv1.SyncOperationResource{Group: "apps", Kind: "ReplicaSet", Name: "backend", Namespace: "", Exclude: true}
	)
	tests := []struct {
		testName              string
		name                  string
		namespace             string
		gvk                   schema.GroupVersionKind
		syncOperationResource []*appv1.SyncOperationResource
		expectedResult        bool
	}{
		//--resource apps:ReplicaSet:backend --resource *:Service:*
		{
			testName:              "Include ReplicaSet backend resource and all service resources",
			name:                  "backend",
			namespace:             "default",
			gvk:                   schema.GroupVersionKind{Group: "apps", Kind: "ReplicaSet"},
			syncOperationResource: []*appv1.SyncOperationResource{&includeAllServiceResources, &includeReplicaSetResource},
			expectedResult:        true,
		},
		//--resource apps:ReplicaSet:backend --resource *:Service:*
		{
			testName:              "Include ReplicaSet backend resource and all service resources",
			name:                  "main-page-down",
			namespace:             "default",
			gvk:                   schema.GroupVersionKind{Group: "batch", Kind: "Job"},
			syncOperationResource: []*appv1.SyncOperationResource{&includeAllServiceResources, &includeReplicaSetResource},
			expectedResult:        false,
		},
		//--resource apps:ReplicaSet:backend --resource !*:Service:*
		{
			testName:              "Include ReplicaSet backend resource and exclude all service resources",
			name:                  "main-page-down",
			namespace:             "default",
			gvk:                   schema.GroupVersionKind{Group: "batch", Kind: "Job"},
			syncOperationResource: []*appv1.SyncOperationResource{&excludeAllServiceResources, &includeReplicaSetResource},
			expectedResult:        false,
		},
		// --resource !apps:ReplicaSet:backend --resource !*:Service:*
		{
			testName:              "Exclude ReplicaSet backend resource and all service resources",
			name:                  "main-page-down",
			namespace:             "default",
			gvk:                   schema.GroupVersionKind{Group: "batch", Kind: "Job"},
			syncOperationResource: []*appv1.SyncOperationResource{&excludeReplicaSetResource, &excludeAllServiceResources},
			expectedResult:        true,
		},
		// --resource !apps:ReplicaSet:backend
		{
			testName:              "Exclude ReplicaSet backend resource",
			name:                  "backend",
			namespace:             "default",
			gvk:                   schema.GroupVersionKind{Group: "apps", Kind: "ReplicaSet"},
			syncOperationResource: []*appv1.SyncOperationResource{&excludeReplicaSetResource},
			expectedResult:        false,
		},
		// --resource !apps:ReplicaSet:backend --resource !*:Service:*
		{
			testName:              "Exclude ReplicaSet backend resource and all service resources(dummy condition)",
			name:                  "backend",
			namespace:             "default",
			gvk:                   schema.GroupVersionKind{Group: "apps", Kind: "ReplicaSet"},
			syncOperationResource: []*appv1.SyncOperationResource{&excludeReplicaSetResource, &excludeAllServiceResources},
			expectedResult:        false,
		},
		// --resource apps:ReplicaSet:backend
		{
			testName:              "Include ReplicaSet backend resource",
			name:                  "backend",
			namespace:             "default",
			gvk:                   schema.GroupVersionKind{Group: "apps", Kind: "ReplicaSet"},
			syncOperationResource: []*appv1.SyncOperationResource{&includeReplicaSetResource},
			expectedResult:        true,
		},
		// --resource !*:Service:*
		{
			testName:              "Exclude Service resources",
			name:                  "backend",
			namespace:             "default",
			gvk:                   schema.GroupVersionKind{Group: "", Kind: "Service"},
			syncOperationResource: []*appv1.SyncOperationResource{&excludeAllServiceResources},
			expectedResult:        false,
		},
		// --resource *:Service:*
		{
			testName:              "Include Service resources",
			name:                  "backend",
			namespace:             "default",
			gvk:                   schema.GroupVersionKind{Group: "", Kind: "Service"},
			syncOperationResource: []*appv1.SyncOperationResource{&includeAllServiceResources},
			expectedResult:        true,
		},
		// --resource apps:ReplicaSet:* --resource !apps:ReplicaSet:backend
		{
			testName:              "Include & Exclude ReplicaSet resources",
			name:                  "backend",
			namespace:             "default",
			gvk:                   schema.GroupVersionKind{Group: "apps", Kind: "ReplicaSet"},
			syncOperationResource: []*appv1.SyncOperationResource{&includeAllReplicaSetResource, &excludeReplicaSetResource},
			expectedResult:        false,
		},
		// --resource !*:*:*
		{
			testName:              "Exclude all resources",
			name:                  "backend",
			namespace:             "default",
			gvk:                   schema.GroupVersionKind{Group: "", Kind: "Service"},
			syncOperationResource: []*appv1.SyncOperationResource{&excludeAllResources},
			expectedResult:        false,
		},
		// --resource *:*:*
		{
			testName:              "Include all resources",
			name:                  "backend",
			namespace:             "default",
			gvk:                   schema.GroupVersionKind{Group: "", Kind: "Service"},
			syncOperationResource: []*appv1.SyncOperationResource{&includeAllResources},
			expectedResult:        true,
		},
		{
			testName:              "No Filters",
			name:                  "backend",
			namespace:             "default",
			gvk:                   schema.GroupVersionKind{Group: "", Kind: "Service"},
			syncOperationResource: []*appv1.SyncOperationResource{&blankValues},
			expectedResult:        false,
		},
		{
			testName:       "Default values",
			expectedResult: true,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			t.Parallel()
			isResourceIncluded := IncludeResource(test.name, test.namespace, test.gvk, test.syncOperationResource)
			assert.Equal(t, test.expectedResult, isResourceIncluded)
		})
	}
}

func TestContainsSyncResource(t *testing.T) {
	t.Parallel()
	var (
		blankUnstructured unstructured.Unstructured
		blankResource     appv1.SyncOperationResource
		helloResource     = appv1.SyncOperationResource{Name: "hello"}
	)
	tables := []struct {
		u        *unstructured.Unstructured
		rr       []appv1.SyncOperationResource
		expected bool
	}{
		{&blankUnstructured, []appv1.SyncOperationResource{}, false},
		{&blankUnstructured, []appv1.SyncOperationResource{blankResource}, true},
		{&blankUnstructured, []appv1.SyncOperationResource{helloResource}, false},
	}

	for _, table := range tables {
		out := ContainsSyncResource(table.u.GetName(), table.u.GetNamespace(), table.u.GroupVersionKind(), table.rr)
		assert.Equal(t, table.expected, out, "Expected %t for slice %+v contains resource %+v; instead got %t", table.expected, table.rr, table.u, out)
	}
}

// TestNilOutZerValueAppSources verifies we will nil out app source specs when they are their zero-value
func TestNilOutZerValueAppSources(t *testing.T) {
	t.Parallel()
	var spec *appv1.ApplicationSpec
	spec = NormalizeApplicationSpec(&appv1.ApplicationSpec{Source: &appv1.ApplicationSource{Kustomize: &appv1.ApplicationSourceKustomize{NamePrefix: "foo"}}})
	assert.NotNil(t, spec.GetSource().Kustomize)
	spec = NormalizeApplicationSpec(&appv1.ApplicationSpec{Source: &appv1.ApplicationSource{Kustomize: &appv1.ApplicationSourceKustomize{NamePrefix: ""}}})
	source := spec.GetSource()
	assert.Nil(t, source.Kustomize)

	spec = NormalizeApplicationSpec(&appv1.ApplicationSpec{Source: &appv1.ApplicationSource{Kustomize: &appv1.ApplicationSourceKustomize{NameSuffix: "foo"}}})
	assert.NotNil(t, spec.GetSource().Kustomize)
	spec = NormalizeApplicationSpec(&appv1.ApplicationSpec{Source: &appv1.ApplicationSource{Kustomize: &appv1.ApplicationSourceKustomize{NameSuffix: ""}}})
	source = spec.GetSource()
	assert.Nil(t, source.Kustomize)

	spec = NormalizeApplicationSpec(&appv1.ApplicationSpec{Source: &appv1.ApplicationSource{Helm: &appv1.ApplicationSourceHelm{ValueFiles: []string{"values.yaml"}}}})
	assert.NotNil(t, spec.GetSource().Helm)
	spec = NormalizeApplicationSpec(&appv1.ApplicationSpec{Source: &appv1.ApplicationSource{Helm: &appv1.ApplicationSourceHelm{ValueFiles: []string{}}}})
	assert.Nil(t, spec.GetSource().Helm)

	spec = NormalizeApplicationSpec(&appv1.ApplicationSpec{Source: &appv1.ApplicationSource{Directory: &appv1.ApplicationSourceDirectory{Recurse: true}}})
	assert.NotNil(t, spec.GetSource().Directory)
	spec = NormalizeApplicationSpec(&appv1.ApplicationSpec{Source: &appv1.ApplicationSource{Directory: &appv1.ApplicationSourceDirectory{Recurse: false}}})
	assert.Nil(t, spec.GetSource().Directory)
}

func TestValidatePermissionsEmptyDestination(t *testing.T) {
	t.Parallel()
	conditions, err := ValidatePermissions(t.Context(), &appv1.ApplicationSpec{
		Source: &appv1.ApplicationSource{RepoURL: "https://github.com/hanzoai/cd", Path: "."},
	}, &appv1.AppProject{
		Spec: appv1.AppProjectSpec{
			SourceRepos:  []string{"*"},
			Destinations: []appv1.ApplicationDestination{{Server: "*", Namespace: "*"}},
		},
	}, nil)
	require.NoError(t, err)
	assert.ElementsMatch(t, conditions, []appv1.ApplicationCondition{{Type: appv1.ApplicationConditionInvalidSpecError, Message: "Destination server missing from app spec"}})
}

func TestValidateChartWithoutRevision(t *testing.T) {
	t.Parallel()
	appSpec := &appv1.ApplicationSpec{
		Source: &appv1.ApplicationSource{RepoURL: "https://charts.helm.sh/incubator/", Chart: "myChart", TargetRevision: ""},
		Destination: appv1.ApplicationDestination{
			Server: "https://kubernetes.default.svc", Namespace: "default",
		},
	}
	cluster := &appv1.Cluster{Server: "https://kubernetes.default.svc"}
	db := &dbmocks.DB{}
	ctx := t.Context()
	db.EXPECT().GetCluster(ctx, appSpec.Destination.Server).Return(cluster, nil).Maybe()

	conditions, err := ValidatePermissions(ctx, appSpec, &appv1.AppProject{
		Spec: appv1.AppProjectSpec{
			SourceRepos:  []string{"*"},
			Destinations: []appv1.ApplicationDestination{{Server: "*", Namespace: "*"}},
		},
	}, db)
	require.NoError(t, err)
	assert.Len(t, conditions, 1)
	assert.Equal(t, appv1.ApplicationConditionInvalidSpecError, conditions[0].Type)
	assert.Equal(t, "spec.source.targetRevision is required if the manifest source is a helm chart", conditions[0].Message)
}

func TestAPIResourcesToStrings(t *testing.T) {
	t.Parallel()
	resources := []kube.APIResourceInfo{{
		GroupVersionResource: schema.GroupVersionResource{Group: "apps", Version: "v1beta1"},
		GroupKind:            schema.GroupKind{Kind: "Deployment"},
	}, {
		GroupVersionResource: schema.GroupVersionResource{Group: "apps", Version: "v1beta2"},
		GroupKind:            schema.GroupKind{Kind: "Deployment"},
	}, {
		GroupVersionResource: schema.GroupVersionResource{Group: "extensions", Version: "v1beta1"},
		GroupKind:            schema.GroupKind{Kind: "Deployment"},
	}}

	assert.ElementsMatch(t, []string{"apps/v1beta1", "apps/v1beta2", "extensions/v1beta1"}, APIResourcesToStrings(resources, false))
	assert.ElementsMatch(t, []string{
		"apps/v1beta1", "apps/v1beta1/Deployment", "apps/v1beta2", "apps/v1beta2/Deployment", "extensions/v1beta1", "extensions/v1beta1/Deployment",
	},
		APIResourcesToStrings(resources, true))
}

func TestValidateRepo(t *testing.T) {
	t.Parallel()
	repoPath, err := filepath.Abs("./../..")
	require.NoError(t, err)

	apiResources := []kube.APIResourceInfo{{
		GroupVersionResource: schema.GroupVersionResource{Group: "apps", Version: "v1beta1"},
		GroupKind:            schema.GroupKind{Kind: "Deployment"},
	}, {
		GroupVersionResource: schema.GroupVersionResource{Group: "apps", Version: "v1beta2"},
		GroupKind:            schema.GroupKind{Kind: "Deployment"},
	}}
	kubeVersion := "v1.16"
	kustomizeOptions := &appv1.KustomizeOptions{BuildOptions: ""}
	repo := &appv1.Repository{Repo: "file://" + repoPath}
	cluster := &appv1.Cluster{Server: "sample server"}
	app := &appv1.Application{
		Spec: appv1.ApplicationSpec{
			Source: &appv1.ApplicationSource{
				RepoURL: repo.Repo,
			},
			Destination: appv1.ApplicationDestination{
				Server:    cluster.Server,
				Namespace: "default",
			},
		},
	}

	proj := &appv1.AppProject{
		Spec: appv1.AppProjectSpec{
			SourceRepos: []string{"*"},
		},
	}

	helmRepos := []*appv1.Repository{{Repo: "sample helm repo"}}

	repoClient := &mocks.RepoServerServiceClient{}
	source := app.Spec.GetSource()
	repoClient.EXPECT().GetAppDetails(mock.Anything, &apiclient.RepoServerAppDetailsQuery{
		Repo:             repo,
		Source:           &source,
		Repos:            helmRepos,
		KustomizeOptions: kustomizeOptions,
		HelmOptions:      &appv1.HelmOptions{ValuesFileSchemes: []string{"https", "http"}},
		NoRevisionCache:  true,
	}).Return(&apiclient.RepoAppDetailsResponse{}, nil).Maybe()

	repo.Type = "git"
	repoClient.EXPECT().TestRepository(mock.Anything, &apiclient.TestRepositoryRequest{
		Repo: repo,
	}).Return(&apiclient.TestRepositoryResponse{
		VerifiedRepository: true,
	}, nil).Maybe()

	repoClientSet := &mocks.Clientset{RepoServerServiceClient: repoClient}

	db := &dbmocks.DB{}

	db.EXPECT().GetRepository(mock.Anything, app.Spec.Source.RepoURL, "").Return(repo, nil).Maybe()
	db.EXPECT().ListHelmRepositories(mock.Anything).Return(helmRepos, nil).Maybe()
	db.EXPECT().ListOCIRepositories(mock.Anything).Return([]*appv1.Repository{}, nil).Maybe()
	db.EXPECT().GetCluster(mock.Anything, app.Spec.Destination.Server).Return(cluster, nil).Maybe()
	db.EXPECT().GetAllHelmRepositoryCredentials(mock.Anything).Return(nil, nil).Maybe()
	db.EXPECT().GetAllOCIRepositoryCredentials(mock.Anything).Return([]*appv1.RepoCreds{}, nil).Maybe()

	var receivedRequest *apiclient.ManifestRequest

	repoClient.EXPECT().GenerateManifest(mock.Anything, mock.MatchedBy(func(req *apiclient.ManifestRequest) bool {
		receivedRequest = req
		return true
	})).Return(nil, nil).Maybe()

	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cd-cm",
			Namespace: test.FakeArgoCDNamespace,
			Labels: map[string]string{
				"app.kubernetes.io/part-of": "cd",
			},
		},
		Data: map[string]string{
			"globalProjects": `
 - projectName: default-x
   labelSelector:
     matchExpressions:
      - key: is-x
        operator: Exists
 - projectName: default-non-x
   labelSelector:
     matchExpressions:
      - key: is-x
        operator: DoesNotExist
`,
		},
	}

	kubeClient := fake.NewClientset(&cm)
	settingsMgr := settings.NewSettingsManager(t.Context(), kubeClient, test.FakeArgoCDNamespace)

	conditions, err := ValidateRepo(t.Context(), app, repoClientSet, db, &kubetest.MockKubectlCmd{Version: kubeVersion, APIResources: apiResources}, proj, settingsMgr)

	require.NoError(t, err)
	assert.Empty(t, conditions)
	assert.ElementsMatch(t, []string{"apps/v1beta1", "apps/v1beta1/Deployment", "apps/v1beta2", "apps/v1beta2/Deployment"}, receivedRequest.ApiVersions)
	assert.Equal(t, kubeVersion, receivedRequest.KubeVersion)
	assert.Equal(t, app.Spec.Destination.Namespace, receivedRequest.Namespace)
	assert.Equal(t, &source, receivedRequest.ApplicationSource)
	assert.Equal(t, kustomizeOptions, receivedRequest.KustomizeOptions)
}

func TestValidateRepo_SourceHydrator(t *testing.T) {
	t.Parallel()
	repoPath, err := filepath.Abs("./../..")
	require.NoError(t, err)

	apiResources := []kube.APIResourceInfo{{
		GroupVersionResource: schema.GroupVersionResource{Group: "apps", Version: "v1beta1"},
		GroupKind:            schema.GroupKind{Kind: "Deployment"},
	}, {
		GroupVersionResource: schema.GroupVersionResource{Group: "apps", Version: "v1beta2"},
		GroupKind:            schema.GroupKind{Kind: "Deployment"},
	}}
	kubeVersion := "v1.16"

	repoURL := "file://" + repoPath
	repo := &appv1.Repository{Repo: repoURL, Type: "git"}
	cluster := &appv1.Cluster{Server: "sample server"}

	app := &appv1.Application{
		Spec: appv1.ApplicationSpec{
			Destination: appv1.ApplicationDestination{
				Server:    cluster.Server,
				Namespace: "default",
			},
			SourceHydrator: &appv1.SourceHydrator{
				DrySource: appv1.DrySource{
					RepoURL:        repoURL,
					TargetRevision: "HEAD",
					Path:           "guestbook",
				},
				SyncSource: appv1.SyncSource{
					TargetBranch: "env/test",
					Path:         "guestbook",
				},
			},
		},
	}

	proj := &appv1.AppProject{
		Spec: appv1.AppProjectSpec{
			SourceRepos: []string{"*"},
		},
	}

	repoClient := &mocks.RepoServerServiceClient{}

	syncSource := app.Spec.GetSource()
	repoClient.EXPECT().TestRepository(mock.Anything, mock.MatchedBy(func(req *apiclient.TestRepositoryRequest) bool {
		return req.Repo.Repo == syncSource.RepoURL
	})).Return(&apiclient.TestRepositoryResponse{VerifiedRepository: true}, nil)

	var receivedRequest *apiclient.ManifestRequest
	repoClient.EXPECT().GenerateManifest(mock.Anything, mock.MatchedBy(func(req *apiclient.ManifestRequest) bool {
		receivedRequest = req
		return true
	})).Return(nil, nil)

	repoClientSet := &mocks.Clientset{RepoServerServiceClient: repoClient}

	db := &dbmocks.DB{}
	db.EXPECT().GetRepository(mock.Anything, repoURL, "").Return(repo, nil).Maybe()
	db.EXPECT().ListHelmRepositories(mock.Anything).Return(nil, nil)
	db.EXPECT().ListOCIRepositories(mock.Anything).Return(nil, nil)
	db.EXPECT().GetCluster(mock.Anything, cluster.Server).Return(cluster, nil)
	db.EXPECT().GetAllHelmRepositoryCredentials(mock.Anything).Return(nil, nil)
	db.EXPECT().GetAllOCIRepositoryCredentials(mock.Anything).Return(nil, nil)

	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cd-cm",
			Namespace: test.FakeArgoCDNamespace,
			Labels:    map[string]string{"app.kubernetes.io/part-of": "cd"},
		},
	}
	kubeClient := fake.NewClientset(&cm)
	settingsMgr := settings.NewSettingsManager(t.Context(), kubeClient, test.FakeArgoCDNamespace)

	conditions, err := ValidateRepo(t.Context(), app, repoClientSet, db, &kubetest.MockKubectlCmd{Version: kubeVersion, APIResources: apiResources}, proj, settingsMgr)

	require.NoError(t, err)
	assert.Empty(t, conditions)
	require.NotNil(t, receivedRequest)

	drySource := app.Spec.SourceHydrator.GetDrySource()
	assert.Equal(t, &drySource, receivedRequest.ApplicationSource)
	assert.Equal(t, "HEAD", receivedRequest.Revision)
	assert.Empty(t, receivedRequest.KubeVersion, "KubeVersion must be empty for dry sources to match hydrator cache keys")
	assert.Nil(t, receivedRequest.ApiVersions, "ApiVersions must be nil for dry sources to match hydrator cache keys")
}

func TestFormatAppConditions(t *testing.T) {
	t.Parallel()
	conditions := []appv1.ApplicationCondition{
		{
			Type:    EventReasonOperationCompleted,
			Message: "Foo",
		},
		{
			Type:    EventReasonResourceCreated,
			Message: "Bar",
		},
	}

	t.Run("Single Condition", func(t *testing.T) {
		t.Parallel()
		res := FormatAppConditions(conditions[0:1])
		assert.NotEmpty(t, res)
		assert.Equal(t, EventReasonOperationCompleted+": Foo", res)
	})

	t.Run("Multiple Conditions", func(t *testing.T) {
		t.Parallel()
		res := FormatAppConditions(conditions)
		assert.NotEmpty(t, res)
		assert.Equal(t, fmt.Sprintf("%s: Foo;%s: Bar", EventReasonOperationCompleted, EventReasonResourceCreated), res)
	})

	t.Run("Empty Conditions", func(t *testing.T) {
		t.Parallel()
		res := FormatAppConditions([]appv1.ApplicationCondition{})
		assert.Empty(t, res)
	})
}

func TestFilterByProjects(t *testing.T) {
	t.Parallel()
	apps := []appv1.Application{
		{
			Spec: appv1.ApplicationSpec{
				Project: "fooproj",
			},
		},
		{
			Spec: appv1.ApplicationSpec{
				Project: "barproj",
			},
		},
	}

	t.Run("No apps in single project", func(t *testing.T) {
		t.Parallel()
		res := FilterByProjects(apps, []string{"foobarproj"})
		assert.Empty(t, res)
	})

	t.Run("Single app in single project", func(t *testing.T) {
		t.Parallel()
		res := FilterByProjects(apps, []string{"fooproj"})
		assert.Len(t, res, 1)
	})

	t.Run("Single app in multiple project", func(t *testing.T) {
		t.Parallel()
		res := FilterByProjects(apps, []string{"fooproj", "foobarproj"})
		assert.Len(t, res, 1)
	})

	t.Run("Multiple apps in multiple project", func(t *testing.T) {
		t.Parallel()
		res := FilterByProjects(apps, []string{"fooproj", "barproj"})
		assert.Len(t, res, 2)
	})
}

func TestFilterByProjectsP(t *testing.T) {
	t.Parallel()
	apps := []*appv1.Application{
		{
			Spec: appv1.ApplicationSpec{
				Project: "fooproj",
			},
		},
		{
			Spec: appv1.ApplicationSpec{
				Project: "barproj",
			},
		},
	}

	t.Run("No apps in single project", func(t *testing.T) {
		t.Parallel()
		res := FilterByProjectsP(apps, []string{"foobarproj"})
		assert.Empty(t, res)
	})

	t.Run("Single app in single project", func(t *testing.T) {
		t.Parallel()
		res := FilterByProjectsP(apps, []string{"fooproj"})
		assert.Len(t, res, 1)
	})

	t.Run("Single app in multiple project", func(t *testing.T) {
		t.Parallel()
		res := FilterByProjectsP(apps, []string{"fooproj", "foobarproj"})
		assert.Len(t, res, 1)
	})

	t.Run("Multiple apps in multiple project", func(t *testing.T) {
		t.Parallel()
		res := FilterByProjectsP(apps, []string{"fooproj", "barproj"})
		assert.Len(t, res, 2)
	})
}

func TestFilterAppSetsByProjects(t *testing.T) {
	t.Parallel()
	appsets := []appv1.ApplicationSet{
		{
			Spec: appv1.ApplicationSetSpec{
				Template: appv1.ApplicationSetTemplate{
					Spec: appv1.ApplicationSpec{
						Project: "fooproj",
					},
				},
			},
		},
		{
			Spec: appv1.ApplicationSetSpec{
				Template: appv1.ApplicationSetTemplate{
					Spec: appv1.ApplicationSpec{
						Project: "barproj",
					},
				},
			},
		},
	}

	t.Run("No apps in single project", func(t *testing.T) {
		t.Parallel()
		res := FilterAppSetsByProjects(appsets, []string{"foobarproj"})
		assert.Empty(t, res)
	})

	t.Run("Single app in single project", func(t *testing.T) {
		t.Parallel()
		res := FilterAppSetsByProjects(appsets, []string{"fooproj"})
		assert.Len(t, res, 1)
	})

	t.Run("Single app in multiple project", func(t *testing.T) {
		t.Parallel()
		res := FilterAppSetsByProjects(appsets, []string{"fooproj", "foobarproj"})
		assert.Len(t, res, 1)
	})

	t.Run("Multiple apps in multiple project", func(t *testing.T) {
		t.Parallel()
		res := FilterAppSetsByProjects(appsets, []string{"fooproj", "barproj"})
		assert.Len(t, res, 2)
	})
}

func TestFilterByRepo(t *testing.T) {
	t.Parallel()
	apps := []appv1.Application{
		{
			Spec: appv1.ApplicationSpec{
				Source: &appv1.ApplicationSource{
					RepoURL: "git@github.com:owner/repo.git",
				},
			},
		},
		{
			Spec: appv1.ApplicationSpec{
				Source: &appv1.ApplicationSource{
					RepoURL: "git@github.com:owner/otherrepo.git",
				},
			},
		},
	}

	t.Run("Empty filter", func(t *testing.T) {
		t.Parallel()
		res := FilterByRepo(apps, "")
		assert.Len(t, res, 2)
	})

	t.Run("Match", func(t *testing.T) {
		t.Parallel()
		res := FilterByRepo(apps, "git@github.com:owner/repo.git")
		assert.Len(t, res, 1)
	})

	t.Run("No match", func(t *testing.T) {
		t.Parallel()
		res := FilterByRepo(apps, "git@github.com:owner/willnotmatch.git")
		assert.Empty(t, res)
	})
}

func TestFilterByRepoP(t *testing.T) {
	t.Parallel()
	apps := []*appv1.Application{
		{
			Spec: appv1.ApplicationSpec{
				Source: &appv1.ApplicationSource{
					RepoURL: "git@github.com:owner/repo.git",
				},
			},
		},
		{
			Spec: appv1.ApplicationSpec{
				Source: &appv1.ApplicationSource{
					RepoURL: "git@github.com:owner/otherrepo.git",
				},
			},
		},
	}

	t.Run("Empty filter", func(t *testing.T) {
		t.Parallel()
		res := FilterByRepoP(apps, "")
		assert.Len(t, res, 2)
	})

	t.Run("Match", func(t *testing.T) {
		t.Parallel()
		res := FilterByRepoP(apps, "git@github.com:owner/repo.git")
		assert.Len(t, res, 1)
	})

	t.Run("No match", func(t *testing.T) {
		t.Parallel()
		res := FilterByRepoP(apps, "git@github.com:owner/willnotmatch.git")
		assert.Empty(t, res)
	})
}

func TestFilterByPath(t *testing.T) {
	t.Parallel()
	apps := []appv1.Application{
		{
			Spec: appv1.ApplicationSpec{
				Source: &appv1.ApplicationSource{
					Path: "example/app/foo",
				},
			},
		},
		{
			Spec: appv1.ApplicationSpec{
				Source: &appv1.ApplicationSource{
					Path: "example/app/existent",
				},
			},
		},
	}

	t.Run("Empty filter", func(t *testing.T) {
		t.Parallel()
		res := FilterByPath(apps, "")
		assert.Len(t, res, 2)
	})

	t.Run("Found one match", func(t *testing.T) {
		t.Parallel()
		res := FilterByPath(apps, "example/app/foo")
		assert.Len(t, res, 1)
	})

	t.Run("No match found", func(t *testing.T) {
		t.Parallel()
		res := FilterByPath(apps, "example/app/non-existent")
		assert.Empty(t, res)
	})
}

func TestValidatePermissions(t *testing.T) {
	t.Parallel()
	t.Run("Empty Repo URL result in condition", func(t *testing.T) {
		t.Parallel()
		spec := appv1.ApplicationSpec{
			Source: &appv1.ApplicationSource{
				RepoURL: "",
			},
		}
		proj := appv1.AppProject{}
		db := &dbmocks.DB{}
		conditions, err := ValidatePermissions(t.Context(), &spec, &proj, db)
		require.NoError(t, err)
		assert.Len(t, conditions, 1)
		assert.Equal(t, appv1.ApplicationConditionInvalidSpecError, conditions[0].Type)
		assert.Contains(t, conditions[0].Message, "are required")
	})

	t.Run("Incomplete Path/Chart combo result in condition", func(t *testing.T) {
		t.Parallel()
		spec := appv1.ApplicationSpec{
			Source: &appv1.ApplicationSource{
				RepoURL: "http://some/where",
				Path:    "",
				Chart:   "",
			},
		}
		proj := appv1.AppProject{}
		db := &dbmocks.DB{}
		conditions, err := ValidatePermissions(t.Context(), &spec, &proj, db)
		require.NoError(t, err)
		assert.Len(t, conditions, 1)
		assert.Equal(t, appv1.ApplicationConditionInvalidSpecError, conditions[0].Type)
		assert.Contains(t, conditions[0].Message, "are required")
	})

	t.Run("Helm chart requires targetRevision", func(t *testing.T) {
		t.Parallel()
		spec := appv1.ApplicationSpec{
			Source: &appv1.ApplicationSource{
				RepoURL: "http://some/where",
				Path:    "",
				Chart:   "somechart",
			},
		}
		proj := appv1.AppProject{}
		db := &dbmocks.DB{}
		conditions, err := ValidatePermissions(t.Context(), &spec, &proj, db)
		require.NoError(t, err)
		assert.Len(t, conditions, 1)
		assert.Equal(t, appv1.ApplicationConditionInvalidSpecError, conditions[0].Type)
		assert.Contains(t, conditions[0].Message, "is required if the manifest source is a helm chart")
	})

	t.Run("Application source is not permitted in project", func(t *testing.T) {
		t.Parallel()
		spec := appv1.ApplicationSpec{
			Source: &appv1.ApplicationSource{
				RepoURL:        "http://some/where",
				Path:           "",
				Chart:          "somechart",
				TargetRevision: "1.4.1",
			},
			Destination: appv1.ApplicationDestination{
				Server:    "https://127.0.0.1:6443",
				Namespace: "testns",
			},
		}
		proj := appv1.AppProject{
			Spec: appv1.AppProjectSpec{
				Destinations: []appv1.ApplicationDestination{
					{
						Server:    "*",
						Namespace: "*",
					},
				},
				SourceRepos: []string{"http://some/where/else"},
			},
		}
		cluster := &appv1.Cluster{Server: "https://127.0.0.1:6443", Name: "test"}
		db := &dbmocks.DB{}
		db.EXPECT().GetCluster(mock.Anything, spec.Destination.Server).Return(cluster, nil).Maybe()
		conditions, err := ValidatePermissions(t.Context(), &spec, &proj, db)
		require.NoError(t, err)
		assert.Len(t, conditions, 1)
		assert.Contains(t, conditions[0].Message, "application repo http://some/where is not permitted")
	})

	t.Run("Application destination is not permitted in project", func(t *testing.T) {
		t.Parallel()
		spec := appv1.ApplicationSpec{
			Source: &appv1.ApplicationSource{
				RepoURL:        "http://some/where",
				Path:           "",
				Chart:          "somechart",
				TargetRevision: "1.4.1",
			},
			Destination: appv1.ApplicationDestination{
				Server:    "https://127.0.0.1:6443",
				Namespace: "testns",
			},
		}
		proj := appv1.AppProject{
			Spec: appv1.AppProjectSpec{
				Destinations: []appv1.ApplicationDestination{
					{
						Server:    "*",
						Namespace: "default",
					},
				},
				SourceRepos: []string{"http://some/where"},
			},
		}
		cluster := &appv1.Cluster{Server: "https://127.0.0.1:6443", Name: "test"}
		db := &dbmocks.DB{}
		db.EXPECT().GetCluster(mock.Anything, spec.Destination.Server).Return(cluster, nil).Maybe()
		conditions, err := ValidatePermissions(t.Context(), &spec, &proj, db)
		require.NoError(t, err)
		assert.Len(t, conditions, 1)
		assert.Contains(t, conditions[0].Message, "application destination")
	})

	t.Run("Destination cluster does not exist", func(t *testing.T) {
		t.Parallel()
		spec := appv1.ApplicationSpec{
			Source: &appv1.ApplicationSource{
				RepoURL:        "http://some/where",
				Path:           "",
				Chart:          "somechart",
				TargetRevision: "1.4.1",
			},
			Destination: appv1.ApplicationDestination{
				Server:    "https://127.0.0.1:6443",
				Namespace: "default",
			},
		}
		proj := appv1.AppProject{
			Spec: appv1.AppProjectSpec{
				Destinations: []appv1.ApplicationDestination{
					{
						Server:    "*",
						Namespace: "default",
					},
				},
				SourceRepos: []string{"http://some/where"},
			},
		}
		db := &dbmocks.DB{}
		db.EXPECT().GetCluster(mock.Anything, spec.Destination.Server).Return(nil, status.Errorf(codes.NotFound, "Cluster does not exist")).Maybe()
		conditions, err := ValidatePermissions(t.Context(), &spec, &proj, db)
		require.NoError(t, err)
		assert.Len(t, conditions, 1)
		assert.Contains(t, conditions[0].Message, "Cluster does not exist")
	})

	t.Run("Destination cluster name does not exist", func(t *testing.T) {
		t.Parallel()
		spec := appv1.ApplicationSpec{
			Source: &appv1.ApplicationSource{
				RepoURL:        "http://some/where",
				Path:           "",
				Chart:          "somechart",
				TargetRevision: "1.4.1",
			},
			Destination: appv1.ApplicationDestination{
				Name:      "does-not-exist",
				Namespace: "default",
			},
		}
		proj := appv1.AppProject{
			Spec: appv1.AppProjectSpec{
				Destinations: []appv1.ApplicationDestination{
					{
						Server:    "*",
						Namespace: "default",
					},
				},
				SourceRepos: []string{"http://some/where"},
			},
		}
		db := &dbmocks.DB{}
		db.EXPECT().GetClusterServersByName(mock.Anything, "does-not-exist").Return(nil, nil).Maybe()
		conditions, err := ValidatePermissions(t.Context(), &spec, &proj, db)
		require.NoError(t, err)
		assert.Len(t, conditions, 1)
		assert.Contains(t, conditions[0].Message, "there are no clusters with this name: does-not-exist")
	})

	t.Run("Cannot get cluster info from DB", func(t *testing.T) {
		t.Parallel()
		spec := appv1.ApplicationSpec{
			Source: &appv1.ApplicationSource{
				RepoURL:        "http://some/where",
				Path:           "",
				Chart:          "somechart",
				TargetRevision: "1.4.1",
			},
			Destination: appv1.ApplicationDestination{
				Server:    "https://127.0.0.1:6443",
				Namespace: "default",
			},
		}
		proj := appv1.AppProject{
			Spec: appv1.AppProjectSpec{
				Destinations: []appv1.ApplicationDestination{
					{
						Server:    "*",
						Namespace: "default",
					},
				},
				SourceRepos: []string{"http://some/where"},
			},
		}
		db := &dbmocks.DB{}
		db.EXPECT().GetCluster(mock.Anything, spec.Destination.Server).Return(nil, errors.New("Unknown error occurred")).Maybe()
		conditions, err := ValidatePermissions(t.Context(), &spec, &proj, db)
		require.NoError(t, err)
		assert.Len(t, conditions, 1)
		assert.Contains(t, conditions[0].Message, "Unknown error occurred")
	})

	t.Run("Destination cluster name resolves to valid server", func(t *testing.T) {
		t.Parallel()
		spec := appv1.ApplicationSpec{
			Source: &appv1.ApplicationSource{
				RepoURL:        "http://some/where",
				Path:           "",
				Chart:          "somechart",
				TargetRevision: "1.4.1",
			},
			Destination: appv1.ApplicationDestination{
				Name:      "does-exist",
				Namespace: "default",
			},
		}
		proj := appv1.AppProject{
			Spec: appv1.AppProjectSpec{
				Destinations: []appv1.ApplicationDestination{
					{
						Server:    "*",
						Namespace: "default",
					},
				},
				SourceRepos: []string{"http://some/where"},
			},
		}
		db := &dbmocks.DB{}
		cluster := appv1.Cluster{
			Name:   "does-exist",
			Server: "https://127.0.0.1:6443",
		}
		db.EXPECT().GetClusterServersByName(mock.Anything, "does-exist").Return([]string{"https://127.0.0.1:6443"}, nil).Maybe()
		db.EXPECT().GetCluster(mock.Anything, "https://127.0.0.1:6443").Return(&cluster, nil).Maybe()
		conditions, err := ValidatePermissions(t.Context(), &spec, &proj, db)
		require.NoError(t, err)
		assert.Empty(t, conditions)
	})
}

func TestValidatePermissions_SourceHydratorSyncSourceRepo(t *testing.T) {
	t.Parallel()
	t.Run("Sync source repo not permitted in project", func(t *testing.T) {
		t.Parallel()
		spec := appv1.ApplicationSpec{
			SourceHydrator: &appv1.SourceHydrator{
				DrySource: appv1.DrySource{
					RepoURL:        "https://example.com/dry-repo",
					TargetRevision: "main",
					Path:           "dry",
				},
				SyncSource: appv1.SyncSource{
					RepoURL:      "https://example.com/sync-repo",
					TargetBranch: "main",
					Path:         "hydrated",
				},
			},
			Destination: appv1.ApplicationDestination{
				Server:    "https://127.0.0.1:6443",
				Namespace: "default",
			},
		}
		proj := appv1.AppProject{
			ObjectMeta: metav1.ObjectMeta{Name: "default"},
			Spec: appv1.AppProjectSpec{
				Destinations: []appv1.ApplicationDestination{{
					Server:    "*",
					Namespace: "*",
				}},
				SourceRepos: []string{"https://example.com/dry-repo"},
			},
		}
		cluster := &appv1.Cluster{Server: "https://127.0.0.1:6443", Name: "test"}
		db := &dbmocks.DB{}
		db.EXPECT().GetCluster(mock.Anything, spec.Destination.Server).Return(cluster, nil).Maybe()

		conditions, err := ValidatePermissions(t.Context(), &spec, &proj, db)
		require.NoError(t, err)
		require.Len(t, conditions, 1)
		assert.Contains(t, conditions[0].Message, "sync source repo https://example.com/sync-repo is not permitted in project")
	})

	t.Run("HydrateTo requires targetBranch", func(t *testing.T) {
		t.Parallel()
		spec := appv1.ApplicationSpec{
			SourceHydrator: &appv1.SourceHydrator{
				DrySource: appv1.DrySource{
					RepoURL:        "https://example.com/dry-repo",
					TargetRevision: "main",
					Path:           "dry",
				},
				SyncSource: appv1.SyncSource{
					TargetBranch: "main",
					Path:         "hydrated",
				},
				HydrateTo: &appv1.HydrateTo{},
			},
		}
		proj := appv1.AppProject{}
		db := &dbmocks.DB{}

		conditions, err := ValidatePermissions(t.Context(), &spec, &proj, db)
		require.NoError(t, err)
		require.Len(t, conditions, 1)
		assert.Contains(t, conditions[0].Message, "spec.sourceHydrator.hydrateTo.targetBranch is required")
	})
}

func TestSetAppOperations(t *testing.T) {
	t.Parallel()
	t.Run("Application not existing", func(t *testing.T) {
		t.Parallel()
		appIf := appclientset.NewSimpleClientset().AppsV1alpha1().Applications("default")
		app, err := SetAppOperation(appIf, "someapp", &appv1.Operation{Sync: &appv1.SyncOperation{Revision: "aaa"}})
		require.Error(t, err)
		assert.Nil(t, app)
	})

	t.Run("Operation already in progress", func(t *testing.T) {
		t.Parallel()
		a := appv1.Application{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "someapp",
				Namespace: "default",
			},
			Operation: &appv1.Operation{Sync: &appv1.SyncOperation{Revision: "aaa"}},
		}
		appIf := appclientset.NewSimpleClientset(&a).AppsV1alpha1().Applications("default")
		app, err := SetAppOperation(appIf, "someapp", &appv1.Operation{Sync: &appv1.SyncOperation{Revision: "aaa"}})
		require.ErrorContains(t, err, "operation is already in progress")
		assert.Nil(t, app)
	})

	t.Run("Operation unspecified", func(t *testing.T) {
		t.Parallel()
		a := appv1.Application{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "someapp",
				Namespace: "default",
			},
		}
		appIf := appclientset.NewSimpleClientset(&a).AppsV1alpha1().Applications("default")
		app, err := SetAppOperation(appIf, "someapp", &appv1.Operation{Sync: nil})
		require.ErrorContains(t, err, "Operation unspecified")
		assert.Nil(t, app)
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		a := appv1.Application{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "someapp",
				Namespace: "default",
			},
		}
		appIf := appclientset.NewSimpleClientset(&a).AppsV1alpha1().Applications("default")
		app, err := SetAppOperation(appIf, "someapp", &appv1.Operation{Sync: &appv1.SyncOperation{Revision: "aaa"}})
		require.NoError(t, err)
		assert.NotNil(t, app)
	})
}

func TestGetDestinationCluster(t *testing.T) {
	t.Parallel()
	t.Run("Validate destination with server url", func(t *testing.T) {
		t.Parallel()
		dest := appv1.ApplicationDestination{
			Server:    "https://127.0.0.1:6443",
			Namespace: "default",
		}

		expectedCluster := &appv1.Cluster{Server: "https://127.0.0.1:6443"}
		db := &dbmocks.DB{}
		db.EXPECT().GetCluster(mock.Anything, "https://127.0.0.1:6443").Return(expectedCluster, nil).Maybe()

		destCluster, err := GetDestinationCluster(t.Context(), dest, db)
		require.NoError(t, err)
		require.NotNil(t, expectedCluster)
		assert.Equal(t, expectedCluster, destCluster)
	})

	t.Run("Validate destination with server name", func(t *testing.T) {
		t.Parallel()
		dest := appv1.ApplicationDestination{
			Name: "minikube",
		}

		db := &dbmocks.DB{}
		db.EXPECT().GetClusterServersByName(mock.Anything, "minikube").Return([]string{"https://127.0.0.1:6443"}, nil).Maybe()
		db.EXPECT().GetCluster(mock.Anything, "https://127.0.0.1:6443").Return(&appv1.Cluster{Server: "https://127.0.0.1:6443", Name: "minikube"}, nil).Maybe()

		destCluster, err := GetDestinationCluster(t.Context(), dest, db)
		require.NoError(t, err)
		assert.Equal(t, "https://127.0.0.1:6443", destCluster.Server)
	})

	t.Run("Error when having both server url and name", func(t *testing.T) {
		t.Parallel()
		dest := appv1.ApplicationDestination{
			Server:    "https://127.0.0.1:6443",
			Name:      "minikube",
			Namespace: "default",
		}

		_, err := GetDestinationCluster(t.Context(), dest, nil)
		assert.EqualError(t, err, "application destination can't have both name and server defined: minikube https://127.0.0.1:6443")
	})

	t.Run("GetClusterServersByName fails", func(t *testing.T) {
		t.Parallel()
		dest := appv1.ApplicationDestination{
			Name: "minikube",
		}

		db := &dbmocks.DB{}
		db.EXPECT().GetClusterServersByName(mock.Anything, "minikube").Return(nil, errors.New("an error occurred"))

		_, err := GetDestinationCluster(t.Context(), dest, db)
		require.ErrorContains(t, err, "an error occurred")
	})

	t.Run("Destination cluster does not exist", func(t *testing.T) {
		t.Parallel()
		dest := appv1.ApplicationDestination{
			Name: "minikube",
		}

		db := &dbmocks.DB{}
		db.EXPECT().GetClusterServersByName(mock.Anything, "minikube").Return(nil, nil)

		_, err := GetDestinationCluster(t.Context(), dest, db)
		assert.EqualError(t, err, "there are no clusters with this name: minikube")
	})

	t.Run("Validate too many clusters with the same name", func(t *testing.T) {
		t.Parallel()
		dest := appv1.ApplicationDestination{
			Name: "dind",
		}

		db := &dbmocks.DB{}
		db.EXPECT().GetClusterServersByName(mock.Anything, "dind").Return([]string{"https://127.0.0.1:2443", "https://127.0.0.1:8443"}, nil)

		_, err := GetDestinationCluster(t.Context(), dest, db)
		assert.EqualError(t, err, "there are 2 clusters with the same name: [https://127.0.0.1:2443 https://127.0.0.1:8443]")
	})
}

func TestFilterByName(t *testing.T) {
	t.Parallel()
	apps := []appv1.Application{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo",
			},
			Spec: appv1.ApplicationSpec{
				Project: "fooproj",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "bar",
			},
			Spec: appv1.ApplicationSpec{
				Project: "barproj",
			},
		},
	}

	t.Run("Name is empty string", func(t *testing.T) {
		t.Parallel()
		res, err := FilterByName(apps, "")
		require.NoError(t, err)
		assert.Len(t, res, 2)
	})

	t.Run("Single app by name", func(t *testing.T) {
		t.Parallel()
		res, err := FilterByName(apps, "foo")
		require.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("No such app", func(t *testing.T) {
		t.Parallel()
		res, err := FilterByName(apps, "foobar")
		require.Error(t, err)
		assert.Empty(t, res)
	})
}

func TestFilterByNameP(t *testing.T) {
	t.Parallel()
	apps := []*appv1.Application{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo",
			},
			Spec: appv1.ApplicationSpec{
				Project: "fooproj",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "bar",
			},
			Spec: appv1.ApplicationSpec{
				Project: "barproj",
			},
		},
	}

	t.Run("Name is empty string", func(t *testing.T) {
		t.Parallel()
		res := FilterByNameP(apps, "")
		assert.Len(t, res, 2)
	})

	t.Run("Single app by name", func(t *testing.T) {
		t.Parallel()
		res := FilterByNameP(apps, "foo")
		assert.Len(t, res, 1)
	})

	t.Run("No such app", func(t *testing.T) {
		t.Parallel()
		res := FilterByNameP(apps, "foobar")
		assert.Empty(t, res)
	})
}

func TestFilterByCluster(t *testing.T) {
	t.Parallel()
	apps := []appv1.Application{
		{
			Spec: appv1.ApplicationSpec{
				Destination: appv1.ApplicationDestination{
					Server: "https://cluster-1.example.com",
				},
			},
		},
		{
			Spec: appv1.ApplicationSpec{
				Destination: appv1.ApplicationDestination{
					Name: "in-cluster",
				},
			},
		},
	}

	tests := []struct {
		name     string
		cluster  string
		expected int
	}{
		{
			name:     "Empty filter returns all",
			cluster:  "",
			expected: 2,
		},
		{
			name:     "Match by Server",
			cluster:  "https://cluster-1.example.com",
			expected: 1,
		},
		{
			name:     "Match by Name",
			cluster:  "in-cluster",
			expected: 1,
		},
		{
			name:     "No match found",
			cluster:  "https://cluster-2.example.com",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res := FilterByCluster(apps, tt.cluster)
			assert.Len(t, res, tt.expected)
		})
	}
}

func TestGetGlobalProjects(t *testing.T) {
	t.Parallel()
	t.Run("Multiple global projects", func(t *testing.T) {
		t.Parallel()
		namespace := "default"

		cm := corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cd-cm",
				Namespace: test.FakeArgoCDNamespace,
				Labels: map[string]string{
					"app.kubernetes.io/part-of": "cd",
				},
			},
			Data: map[string]string{
				"globalProjects": `
 - projectName: default-x
   labelSelector:
     matchExpressions:
      - key: is-x
        operator: Exists
 - projectName: default-non-x
   labelSelector:
     matchExpressions:
      - key: is-x
        operator: DoesNotExist
`,
			},
		}

		defaultX := &appv1.AppProject{
			ObjectMeta: metav1.ObjectMeta{Name: "default-x", Namespace: namespace},
			Spec: appv1.AppProjectSpec{
				ClusterResourceWhitelist: []appv1.ClusterResourceRestrictionItem{
					{Group: "*", Kind: "*"},
				},
				ClusterResourceBlacklist: []appv1.ClusterResourceRestrictionItem{
					{Kind: "Volume"},
				},
			},
		}

		defaultNonX := &appv1.AppProject{
			ObjectMeta: metav1.ObjectMeta{Name: "default-non-x", Namespace: namespace},
			Spec: appv1.AppProjectSpec{
				ClusterResourceBlacklist: []appv1.ClusterResourceRestrictionItem{
					{Group: "*", Kind: "*"},
				},
			},
		}

		isX := &appv1.AppProject{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "is-x",
				Namespace: namespace,
				Labels: map[string]string{
					"is-x": "yep",
				},
			},
		}

		isNoX := &appv1.AppProject{
			ObjectMeta: metav1.ObjectMeta{Name: "is-no-x", Namespace: namespace},
		}

		projClientset := appclientset.NewSimpleClientset(defaultX, defaultNonX, isX, isNoX)
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		indexers := cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}
		informer := v1alpha1.NewAppProjectInformer(projClientset, namespace, 0, indexers)
		go informer.Run(ctx.Done())
		cache.WaitForCacheSync(ctx.Done(), informer.HasSynced)

		kubeClient := fake.NewSimpleClientset(&cm)
		settingsMgr := settings.NewSettingsManager(t.Context(), kubeClient, test.FakeArgoCDNamespace)

		projLister := applisters.NewAppProjectLister(informer.GetIndexer())

		xGlobalProjects := GetGlobalProjects(isX, projLister, settingsMgr)
		assert.Len(t, xGlobalProjects, 1)
		assert.Equal(t, "default-x", xGlobalProjects[0].Name)

		nonXGlobalProjects := GetGlobalProjects(isNoX, projLister, settingsMgr)
		assert.Len(t, nonXGlobalProjects, 1)
		assert.Equal(t, "default-non-x", nonXGlobalProjects[0].Name)
	})
}

func Test_mergeVirtualProject(t *testing.T) {
	proj := &appv1.AppProject{
		Spec: appv1.AppProjectSpec{
			ClusterResourceBlacklist: []appv1.ClusterResourceRestrictionItem{{Group: "", Kind: "Namespace"}},
			ClusterResourceWhitelist: []appv1.ClusterResourceRestrictionItem{{Group: "rbac.authorization.k8s.io", Kind: "ClusterRole"}},
			DestinationServiceAccounts: []appv1.ApplicationDestinationServiceAccount{
				{
					Server:                "test",
					Namespace:             "test",
					DefaultServiceAccount: "custom-serviceaccount",
				},
			},
			Destinations:               []appv1.ApplicationDestination{{Server: "test", Namespace: "test"}},
			NamespaceResourceBlacklist: []metav1.GroupKind{{Group: "", Kind: "Service"}},
			NamespaceResourceWhitelist: []metav1.GroupKind{{Group: "", Kind: "Pod"}},
			SourceRepos:                []string{"http://some/where"},
			SyncWindows: appv1.SyncWindows{
				{
					Kind:         "allow",
					Schedule:     "* * * * *",
					Duration:     "24h",
					Clusters:     []string{"test"},
					Namespaces:   []string{"test"},
					Applications: []string{"test"},
				},
			},
		},
	}

	globalProj := &appv1.AppProject{
		Spec: appv1.AppProjectSpec{
			ClusterResourceBlacklist: []appv1.ClusterResourceRestrictionItem{{Group: "*", Kind: "*"}},
			ClusterResourceWhitelist: []appv1.ClusterResourceRestrictionItem{{Group: "*", Kind: "*"}},
			DestinationServiceAccounts: []appv1.ApplicationDestinationServiceAccount{
				{
					Server:                "*",
					Namespace:             "*",
					DefaultServiceAccount: "default",
				},
			},
			Destinations:               []appv1.ApplicationDestination{{Server: "*", Namespace: "*"}},
			NamespaceResourceBlacklist: []metav1.GroupKind{{Group: "*", Kind: "*"}},
			NamespaceResourceWhitelist: []metav1.GroupKind{{Group: "*", Kind: "*"}},
			SourceRepos:                []string{"http://some/where/else"},
			SyncWindows: appv1.SyncWindows{
				{
					Kind:         "deny",
					Schedule:     "0 0 * * *",
					Duration:     "24h",
					Clusters:     []string{"*"},
					Namespaces:   []string{"*"},
					Applications: []string{"*"},
				},
			},
		},
	}

	expected := &appv1.AppProject{
		Spec: appv1.AppProjectSpec{
			ClusterResourceBlacklist: []appv1.ClusterResourceRestrictionItem{{Group: "", Kind: "Namespace"}, {Group: "*", Kind: "*"}},
			ClusterResourceWhitelist: []appv1.ClusterResourceRestrictionItem{{Group: "rbac.authorization.k8s.io", Kind: "ClusterRole"}, {Group: "*", Kind: "*"}},
			DestinationServiceAccounts: []appv1.ApplicationDestinationServiceAccount{
				{
					Server:                "test",
					Namespace:             "test",
					DefaultServiceAccount: "custom-serviceaccount",
				},
				{
					Server:                "*",
					Namespace:             "*",
					DefaultServiceAccount: "default",
				},
			},
			Destinations:               []appv1.ApplicationDestination{{Server: "test", Namespace: "test"}, {Server: "*", Namespace: "*"}},
			NamespaceResourceBlacklist: []metav1.GroupKind{{Group: "", Kind: "Service"}, {Group: "*", Kind: "*"}},
			NamespaceResourceWhitelist: []metav1.GroupKind{{Group: "", Kind: "Pod"}, {Group: "*", Kind: "*"}},
			SourceRepos:                []string{"http://some/where", "http://some/where/else"},
			SyncWindows: appv1.SyncWindows{
				{
					Kind:         "allow",
					Schedule:     "* * * * *",
					Duration:     "24h",
					Clusters:     []string{"test"},
					Namespaces:   []string{"test"},
					Applications: []string{"test"},
				},
				{
					Kind:         "deny",
					Schedule:     "0 0 * * *",
					Duration:     "24h",
					Clusters:     []string{"*"},
					Namespaces:   []string{"*"},
					Applications: []string{"*"},
				},
			},
		},
	}
	mergeVirtualProject(proj, globalProj)
	assert.Equal(t, expected, proj)
}

func Test_GetDifferentPathsBetweenStructs(t *testing.T) {
	r1 := appv1.Repository{}
	r2 := appv1.Repository{
		Name: "SomeName",
	}

	difference, _ := GetDifferentPathsBetweenStructs(r1, r2)
	assert.Equal(t, []string{"Name"}, difference)
}

func Test_GenerateSpecIsDifferentErrorMessageWithNoDiff(t *testing.T) {
	r1 := appv1.Repository{}
	r2 := appv1.Repository{}

	msg := GenerateSpecIsDifferentErrorMessage("application", r1, r2)
	assert.Equal(t, "existing application spec is different; use upsert flag to force update", msg)
}

func Test_GenerateSpecIsDifferentErrorMessageWithDiff(t *testing.T) {
	r1 := appv1.Repository{}
	r2 := appv1.Repository{
		Name: "test",
	}

	msg := GenerateSpecIsDifferentErrorMessage("repo", r1, r2)
	assert.Equal(t, "existing repo spec is different; use upsert flag to force update; difference in keys \"Name\"", msg)
}

func Test_ParseAppQualifiedName(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name       string
		input      string
		implicitNs string
		appName    string
		appNs      string
	}{
		{"Full qualified without implicit NS", "namespace/name", "", "name", "namespace"},
		{"Non qualified without implicit NS", "name", "", "name", ""},
		{"Full qualified with implicit NS", "namespace/name", "namespace2", "name", "namespace"},
		{"Non qualified with implicit NS", "name", "namespace2", "name", "namespace2"},
		{"Invalid without implicit NS", "namespace_name", "", "namespace_name", ""},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			appName, appNs := ParseFromQualifiedName(tt.input, tt.implicitNs)
			assert.Equal(t, tt.appName, appName)
			assert.Equal(t, tt.appNs, appNs)
		})
	}
}

func Test_ParseAppInstanceName(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name       string
		input      string
		implicitNs string
		appName    string
		appNs      string
	}{
		{"Full qualified without implicit NS", "namespace_name", "", "name", "namespace"},
		{"Non qualified without implicit NS", "name", "", "name", ""},
		{"Full qualified with implicit NS", "namespace_name", "namespace2", "name", "namespace"},
		{"Non qualified with implicit NS", "name", "namespace2", "name", "namespace2"},
		{"Invalid without implicit NS", "namespace/name", "", "namespace/name", ""},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			appName, appNs := ParseInstanceName(tt.input, tt.implicitNs)
			assert.Equal(t, tt.appName, appName)
			assert.Equal(t, tt.appNs, appNs)
		})
	}
}

func Test_AppInstanceName(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name         string
		appName      string
		appNamespace string
		defaultNs    string
		result       string
	}{
		{"defaultns different as appns", "appname", "appns", "defaultns", "appns_appname"},
		{"defaultns same as appns", "appname", "appns", "appns", "appname"},
		{"defaultns set and appns not given", "appname", "", "appns", "appname"},
		{"neither defaultns nor appns set", "appname", "", "appns", "appname"},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := AppInstanceName(tt.appName, tt.appNamespace, tt.defaultNs)
			assert.Equal(t, tt.result, result)
		})
	}
}

func Test_AppInstanceNameFromQualified(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name      string
		appName   string
		defaultNs string
		result    string
	}{
		{"Qualified name with namespace not being defaultns", "appns/appname", "defaultns", "appns_appname"},
		{"Qualified name with namespace being defaultns", "defaultns/appname", "defaultns", "appname"},
		{"Qualified name without namespace", "appname", "defaultns", "appname"},
		{"Qualified name without namespace and defaultns", "appname", "", "appname"},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := InstanceNameFromQualified(tt.appName, tt.defaultNs)
			assert.Equal(t, tt.result, result)
		})
	}
}

func Test_GetRefSources(t *testing.T) {
	t.Parallel()
	repoPath, err := filepath.Abs("./../..")
	require.NoError(t, err)

	getMultiSourceAppSpec := func(sources appv1.ApplicationSources) *appv1.ApplicationSpec {
		return &appv1.ApplicationSpec{
			Sources: sources,
		}
	}

	repo := &appv1.Repository{Repo: "file://" + repoPath}

	t.Run("target ref exists", func(t *testing.T) {
		t.Parallel()
		appSpec := getMultiSourceAppSpec(appv1.ApplicationSources{
			{RepoURL: "file://" + repoPath, Ref: "source-1_2"},
			{RepoURL: "file://" + repoPath},
		})

		refSources, err := GetRefSources(t.Context(), appSpec.Sources, appSpec.Project, func(_ context.Context, _ string, _ string) (*appv1.Repository, error) {
			return repo, nil
		}, []string{})

		expectedRefSource := appv1.RefTargetRevisionMapping{
			"$source-1_2": &appv1.RefTarget{
				Repo: *repo,
			},
		}
		require.NoError(t, err)
		assert.Len(t, refSources, 1)
		assert.Equal(t, expectedRefSource, refSources)
	})

	t.Run("target ref does not exist", func(t *testing.T) {
		t.Parallel()
		appSpec := getMultiSourceAppSpec(appv1.ApplicationSources{
			{RepoURL: "file://does-not-exist", Ref: "source1"},
			{RepoURL: "file://" + repoPath},
		})

		refSources, err := GetRefSources(t.Context(), appSpec.Sources, appSpec.Project, func(_ context.Context, _ string, _ string) (*appv1.Repository, error) {
			return nil, errors.New("repo does not exist")
		}, []string{})

		require.Error(t, err)
		assert.Empty(t, refSources)
	})

	t.Run("invalid ref", func(t *testing.T) {
		t.Parallel()
		appSpec := getMultiSourceAppSpec(appv1.ApplicationSources{
			{RepoURL: "file://does-not-exist", Ref: "%invalid-name%"},
			{RepoURL: "file://" + repoPath},
		})

		refSources, err := GetRefSources(t.Context(), appSpec.Sources, appSpec.Project, func(_ context.Context, _ string, _ string) (*appv1.Repository, error) {
			return nil, err
		}, []string{})

		require.Error(t, err)
		assert.Empty(t, refSources)
	})

	t.Run("duplicate ref keys", func(t *testing.T) {
		t.Parallel()
		appSpec := getMultiSourceAppSpec(appv1.ApplicationSources{
			{RepoURL: "file://does-not-exist", Ref: "source1"},
			{RepoURL: "file://does-not-exist", Ref: "source1"},
		})

		refSources, err := GetRefSources(t.Context(), appSpec.Sources, appSpec.Project, func(_ context.Context, _ string, _ string) (*appv1.Repository, error) {
			return nil, err
		}, []string{})

		require.Error(t, err)
		assert.Empty(t, refSources)
	})

	t.Run("single source with ref populates refSources", func(t *testing.T) {
		t.Parallel()
		appSpec := getMultiSourceAppSpec(appv1.ApplicationSources{
			{RepoURL: "file://" + repoPath, TargetRevision: "HEAD", Ref: "gitops"},
		})

		refSources, err := GetRefSources(t.Context(), appSpec.Sources, appSpec.Project, func(_ context.Context, _ string, _ string) (*appv1.Repository, error) {
			return repo, nil
		}, []string{})

		require.NoError(t, err)
		expected := appv1.RefTargetRevisionMapping{
			"$gitops": &appv1.RefTarget{
				Repo:           *repo,
				TargetRevision: "HEAD",
			},
		}
		assert.Equal(t, expected, refSources)
	})
}

func TestValidatePermissionsMultipleSources(t *testing.T) {
	t.Parallel()
	t.Run("Empty Repo URL result in condition", func(t *testing.T) {
		t.Parallel()
		spec := appv1.ApplicationSpec{
			Sources: appv1.ApplicationSources{
				{RepoURL: ""},
			},
		}

		proj := appv1.AppProject{}
		db := &dbmocks.DB{}
		conditions, err := ValidatePermissions(t.Context(), &spec, &proj, db)
		require.NoError(t, err)
		assert.Len(t, conditions, 1)
		assert.Equal(t, appv1.ApplicationConditionInvalidSpecError, conditions[0].Type)
		assert.Contains(t, conditions[0].Message, "are required")
	})

	t.Run("Incomplete Path/Chart/Ref combo result in condition", func(t *testing.T) {
		t.Parallel()
		spec := appv1.ApplicationSpec{
			Sources: appv1.ApplicationSources{
				{
					RepoURL: "http://some/where",
					Path:    "",
					Chart:   "",
					Ref:     "",
				},
			},
		}
		proj := appv1.AppProject{}
		db := &dbmocks.DB{}
		conditions, err := ValidatePermissions(t.Context(), &spec, &proj, db)
		require.NoError(t, err)
		assert.Len(t, conditions, 1)
		assert.Equal(t, appv1.ApplicationConditionInvalidSpecError, conditions[0].Type)
		assert.Contains(t, conditions[0].Message, "are required")
	})

	t.Run("One of the Application sources is not permitted in project", func(t *testing.T) {
		t.Parallel()
		spec := appv1.ApplicationSpec{
			Sources: appv1.ApplicationSources{
				{
					RepoURL:        "http://some/where",
					Path:           "",
					Chart:          "somechart",
					TargetRevision: "1.4.1",
				},
			},
			Destination: appv1.ApplicationDestination{
				Server:    "https://127.0.0.1:6443",
				Namespace: "testns",
			},
		}
		proj := appv1.AppProject{
			Spec: appv1.AppProjectSpec{
				Destinations: []appv1.ApplicationDestination{
					{
						Server:    "*",
						Namespace: "*",
					},
				},
				SourceRepos: []string{"http://some/where/else"},
			},
		}
		cluster := &appv1.Cluster{Server: "https://127.0.0.1:6443", Name: "test"}
		db := &dbmocks.DB{}
		db.EXPECT().GetCluster(mock.Anything, spec.Destination.Server).Return(cluster, nil)
		conditions, err := ValidatePermissions(t.Context(), &spec, &proj, db)
		require.NoError(t, err)
		assert.Len(t, conditions, 1)
		assert.Contains(t, conditions[0].Message, "application repo http://some/where is not permitted")
	})

	t.Run("Source with a Ref field and missing Path/Chart field", func(t *testing.T) {
		t.Parallel()
		spec := appv1.ApplicationSpec{
			Sources: appv1.ApplicationSources{
				{
					RepoURL: "http://some/where",
					Path:    "",
					Chart:   "",
					Ref:     "somechart",
				},
			},
			Destination: appv1.ApplicationDestination{
				Name:      "does-exist",
				Namespace: "default",
			},
		}
		proj := appv1.AppProject{
			Spec: appv1.AppProjectSpec{
				Destinations: []appv1.ApplicationDestination{
					{
						Server:    "*",
						Namespace: "default",
					},
				},
				SourceRepos: []string{"http://some/where"},
			},
		}
		db := &dbmocks.DB{}
		cluster := appv1.Cluster{
			Name:   "does-exist",
			Server: "https://127.0.0.1:6443",
		}
		db.EXPECT().GetClusterServersByName(mock.Anything, "does-exist").Return([]string{"https://127.0.0.1:6443"}, nil).Maybe()
		db.EXPECT().GetCluster(mock.Anything, "https://127.0.0.1:6443").Return(&cluster, nil).Maybe()
		conditions, err := ValidatePermissions(t.Context(), &spec, &proj, db)
		require.NoError(t, err)
		assert.Empty(t, conditions)
	})
}

func TestAugmentSyncMsg(t *testing.T) {
	t.Parallel()
	mockAPIResourcesFn := func() ([]kube.APIResourceInfo, error) {
		return []kube.APIResourceInfo{
			{
				GroupKind: schema.GroupKind{
					Group: "apps",
					Kind:  "Deployment",
				},
				GroupVersionResource: schema.GroupVersionResource{
					Group:   "apps",
					Version: "v1",
				},
			},
			{
				GroupKind: schema.GroupKind{
					Group: "networking.k8s.io",
					Kind:  "Ingress",
				},
				GroupVersionResource: schema.GroupVersionResource{
					Group:   "networking.k8s.io",
					Version: "v1",
				},
			},
		}, nil
	}

	testcases := []struct {
		name            string
		msg             string
		expectedMessage string
		res             common.ResourceSyncResult
		mockFn          func() ([]kube.APIResourceInfo, error)
		errMsg          string
	}{
		{
			name: "match specific k8s error",
			msg:  "the server could not find the requested resource",
			res: common.ResourceSyncResult{
				ResourceKey: kube.ResourceKey{
					Name:      "deployment-resource",
					Namespace: "test-namespace",
					Kind:      "Deployment",
					Group:     "apps",
				},
				Version: "v1beta1",
			},
			expectedMessage: "The Kubernetes API could not find version \"v1beta1\" of apps/Deployment for requested resource test-namespace/deployment-resource. Version \"v1\" of apps/Deployment is installed on the destination cluster.",
			mockFn:          mockAPIResourcesFn,
		},
		{
			name: "any random k8s msg",
			msg:  "random message from k8s",
			res: common.ResourceSyncResult{
				ResourceKey: kube.ResourceKey{
					Name:      "deployment-resource",
					Namespace: "test-namespace",
					Kind:      "Deployment",
					Group:     "apps",
				},
				Version: "v1beta1",
			},
			expectedMessage: "random message from k8s",
			mockFn:          mockAPIResourcesFn,
		},
		{
			name: "resource doesn't exist in the target cluster",
			res: common.ResourceSyncResult{
				ResourceKey: kube.ResourceKey{
					Name:      "persistent-volume-resource",
					Namespace: "test-namespace",
					Kind:      "PersistentVolume",
					Group:     "",
				},
				Version: "v1",
			},
			msg:             "the server could not find the requested resource",
			expectedMessage: "The Kubernetes API could not find /PersistentVolume for requested resource test-namespace/persistent-volume-resource. Make sure the \"PersistentVolume\" CRD is installed on the destination cluster.",
			mockFn:          mockAPIResourcesFn,
		},
		{
			name: "API Resource returns error",
			res: common.ResourceSyncResult{
				ResourceKey: kube.ResourceKey{
					Name:      "persistent-volume-resource",
					Namespace: "test-namespace",
					Kind:      "PersistentVolume",
					Group:     "",
				},
				Version: "v1",
			},
			msg:             "the server could not find the requested resource",
			expectedMessage: "the server could not find the requested resource",
			mockFn: func() ([]kube.APIResourceInfo, error) {
				return nil, errors.New("failed to fetch resource of given kind %s from the target cluster")
			},
			errMsg: "failed to get API resource info for group \"\" and kind \"PersistentVolume\": failed to get API resource info: failed to fetch resource of given kind %s from the target cluster",
		},
		{
			name: "old Ingress type returns error suggesting new Ingress type",
			res: common.ResourceSyncResult{
				ResourceKey: kube.ResourceKey{
					Name:      "ingress-resource",
					Namespace: "test-namespace",
					Kind:      "Ingress",
					Group:     "extensions",
				},
				Version: "v1beta1",
			},
			msg:             "the server could not find the requested resource",
			expectedMessage: "The Kubernetes API could not find version \"v1beta1\" of extensions/Ingress for requested resource test-namespace/ingress-resource. Version \"v1\" of networking.k8s.io/Ingress is installed on the destination cluster.",
			mockFn:          mockAPIResourcesFn,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.res.Message = tt.msg
			msg, err := AugmentSyncMsg(tt.res, tt.mockFn)
			if tt.errMsg != "" {
				assert.EqualError(t, err, tt.errMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedMessage, msg)
			}
		})
	}
}

func TestGetAppEventLabels(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		cmInEventLabelKeys  string
		cmExEventLabelKeys  string
		appLabels           map[string]string
		projLabels          map[string]string
		expectedEventLabels map[string]string
	}{
		{
			name:                "no label keys in cm - no event labels",
			cmInEventLabelKeys:  "",
			appLabels:           map[string]string{"team": "A", "tier": "frontend"},
			projLabels:          map[string]string{"environment": "dev"},
			expectedEventLabels: nil,
		},
		{
			name:                "label keys in cm, no labels on app & proj - no event labels",
			cmInEventLabelKeys:  "team, environment",
			appLabels:           nil,
			projLabels:          nil,
			expectedEventLabels: nil,
		},
		{
			name:                "labels on app, no labels on proj - event labels matched on app only",
			cmInEventLabelKeys:  "team, environment",
			appLabels:           map[string]string{"team": "A", "tier": "frontend"},
			projLabels:          nil,
			expectedEventLabels: map[string]string{"team": "A"},
		},
		{
			name:                "no labels on app, labels on proj - event labels matched on proj only",
			cmInEventLabelKeys:  "team, environment",
			appLabels:           nil,
			projLabels:          map[string]string{"environment": "dev"},
			expectedEventLabels: map[string]string{"environment": "dev"},
		},
		{
			name:                "labels on app & proj with conflicts - event labels matched on both app & proj and app labels prioritized on conflict",
			cmInEventLabelKeys:  "team, environment",
			appLabels:           map[string]string{"team": "A", "environment": "stage", "tier": "frontend"},
			projLabels:          map[string]string{"environment": "dev"},
			expectedEventLabels: map[string]string{"team": "A", "environment": "stage"},
		},
		{
			name:                "wildcard support - matched all labels",
			cmInEventLabelKeys:  "*",
			appLabels:           map[string]string{"team": "A", "tier": "frontend"},
			projLabels:          map[string]string{"environment": "dev"},
			expectedEventLabels: map[string]string{"team": "A", "tier": "frontend", "environment": "dev"},
		},
		{
			name:                "exclude event labels",
			cmInEventLabelKeys:  "example.com/team,tier,env*",
			cmExEventLabelKeys:  "tie*",
			appLabels:           map[string]string{"example.com/team": "A", "tier": "frontend"},
			projLabels:          map[string]string{"environment": "dev"},
			expectedEventLabels: map[string]string{"example.com/team": "A", "environment": "dev"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cm := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cd-cm",
					Namespace: test.FakeArgoCDNamespace,
					Labels: map[string]string{
						"app.kubernetes.io/part-of": "cd",
					},
				},
				Data: map[string]string{
					"resource.includeEventLabelKeys": tt.cmInEventLabelKeys,
					"resource.excludeEventLabelKeys": tt.cmExEventLabelKeys,
				},
			}

			proj := &appv1.AppProject{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default",
					Namespace: test.FakeArgoCDNamespace,
					Labels:    tt.projLabels,
				},
			}

			var app appv1.Application
			app.Name = "test-app"
			app.Namespace = test.FakeArgoCDNamespace
			app.Labels = tt.appLabels
			appClientset := appclientset.NewSimpleClientset(proj)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			indexers := cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}
			informer := v1alpha1.NewAppProjectInformer(appClientset, test.FakeArgoCDNamespace, 0, indexers)
			go informer.Run(ctx.Done())
			cache.WaitForCacheSync(ctx.Done(), informer.HasSynced)

			kubeClient := fake.NewSimpleClientset(&cm)
			settingsMgr := settings.NewSettingsManager(t.Context(), kubeClient, test.FakeArgoCDNamespace)
			appDB := db.NewDB("default", settingsMgr, kubeClient)

			eventLabels := GetAppEventLabels(ctx, &app, applisters.NewAppProjectLister(informer.GetIndexer()), test.FakeArgoCDNamespace, settingsMgr, appDB)
			assert.Len(t, eventLabels, len(tt.expectedEventLabels))
			for ek, ev := range tt.expectedEventLabels {
				v, found := eventLabels[ek]
				assert.True(t, found)
				assert.Equal(t, ev, v)
			}
		})
	}
}

func TestValidateManagedByURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		app        *appv1.Application
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "Valid HTTPS URL",
			app: &appv1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						appv1.AnnotationKeyManagedByURL: "https://cd.example.com",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid HTTP URL",
			app: &appv1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						appv1.AnnotationKeyManagedByURL: "http://cd.example.com",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid localhost HTTPS URL",
			app: &appv1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						appv1.AnnotationKeyManagedByURL: "https://localhost:8081",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid localhost HTTP URL",
			app: &appv1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						appv1.AnnotationKeyManagedByURL: "http://localhost:8081",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid 127.0.0.1 URL",
			app: &appv1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						appv1.AnnotationKeyManagedByURL: "http://127.0.0.1:8081",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "No annotations",
			app: &appv1.Application{
				ObjectMeta: metav1.ObjectMeta{},
			},
			wantErr: false,
		},
		{
			name: "Empty managed-by-url",
			app: &appv1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						appv1.AnnotationKeyManagedByURL: "",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Missing managed-by-url annotation",
			app: &appv1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"other.annotation": "value",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid protocol - javascript",
			app: &appv1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						appv1.AnnotationKeyManagedByURL: "javascript:alert('xss')",
					},
				},
			},
			wantErr:    true,
			wantErrMsg: "invalid managed-by URL: URL must include http or https protocol",
		},
		{
			name: "Invalid protocol - data",
			app: &appv1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						appv1.AnnotationKeyManagedByURL: "data:text/html,<script>alert('xss')</script>",
					},
				},
			},
			wantErr:    true,
			wantErrMsg: "invalid managed-by URL: URL must include http or https protocol",
		},
		{
			name: "Invalid protocol - file",
			app: &appv1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						appv1.AnnotationKeyManagedByURL: "file:///etc/passwd",
					},
				},
			},
			wantErr:    true,
			wantErrMsg: "invalid managed-by URL: URL must include http or https protocol",
		},
		{
			name: "Invalid URL format",
			app: &appv1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						appv1.AnnotationKeyManagedByURL: "not-a-url",
					},
				},
			},
			wantErr:    true,
			wantErrMsg: "invalid managed-by URL: URL must include http or https protocol",
		},
		{
			name: "URL with path and query",
			app: &appv1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						appv1.AnnotationKeyManagedByURL: "https://cd.example.com/applications?namespace=default",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			conditions := ValidateManagedByURL(tt.app)

			if tt.wantErr {
				require.Len(t, conditions, 1, "Expected exactly one validation condition")
				assert.Equal(t, appv1.ApplicationConditionInvalidSpecError, conditions[0].Type)
				assert.Contains(t, conditions[0].Message, tt.wantErrMsg)
			} else {
				assert.Empty(t, conditions, "Expected no validation conditions for valid URL")
			}
		})
	}
}

func Test_GetSyncedRefSources(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		refSources      appv1.RefTargetRevisionMapping
		sources         appv1.ApplicationSources
		syncedRevisions []string
		result          appv1.RefTargetRevisionMapping
	}{
		{
			name: "multi ref sources",
			refSources: appv1.RefTargetRevisionMapping{
				"$values": &appv1.RefTarget{
					Repo:           appv1.Repository{Repo: "https://github.com/cd"},
					TargetRevision: "main-1",
					Chart:          "chart",
				},
				"$values_1": &appv1.RefTarget{
					Repo:           appv1.Repository{Repo: "https://github.com/cd-1"},
					TargetRevision: "main-2",
					Chart:          "chart",
				},
			},
			sources: appv1.ApplicationSources{
				{RepoURL: "https://helm.registry", TargetRevision: "0.0.1", Chart: "my-chart", Helm: &appv1.ApplicationSourceHelm{ValueFiles: []string{"$values/path"}}},
				{RepoURL: "https://github.com/cd", TargetRevision: "main-1", Ref: "values"},
				{RepoURL: "https://github.com/cd-1", TargetRevision: "main-2", Ref: "values_1"},
			},
			syncedRevisions: []string{"0.0.1", "resolved-main-1", "resolved-main-2"},
			result: appv1.RefTargetRevisionMapping{
				"$values": &appv1.RefTarget{
					Repo:           appv1.Repository{Repo: "https://github.com/cd"},
					TargetRevision: "resolved-main-1",
					Chart:          "chart",
				},
				"$values_1": &appv1.RefTarget{
					Repo:           appv1.Repository{Repo: "https://github.com/cd-1"},
					TargetRevision: "resolved-main-2",
					Chart:          "chart",
				},
			},
		},
		{
			name: "ref source",
			refSources: appv1.RefTargetRevisionMapping{
				"$values": &appv1.RefTarget{
					Repo:           appv1.Repository{Repo: "https://github.com/cd"},
					TargetRevision: "main-1",
					Chart:          "chart",
				},
			},
			sources: appv1.ApplicationSources{
				{RepoURL: "https://helm.registry", TargetRevision: "0.0.1", Chart: "my-chart", Helm: &appv1.ApplicationSourceHelm{ValueFiles: []string{"$values/path"}}},
				{RepoURL: "https://github.com/cd", TargetRevision: "main-1", Ref: "values"},
			},
			syncedRevisions: []string{"0.0.1", "resolved-main-1"},
			result: appv1.RefTargetRevisionMapping{
				"$values": &appv1.RefTarget{
					Repo:           appv1.Repository{Repo: "https://github.com/cd"},
					TargetRevision: "resolved-main-1",
					Chart:          "chart",
				},
			},
		},
		{
			name:       "empty ref source",
			refSources: appv1.RefTargetRevisionMapping{},
			sources: appv1.ApplicationSources{
				{RepoURL: "https://helm.registry", TargetRevision: "0.0.1", Chart: "my-chart"},
			},
			syncedRevisions: []string{"0.0.1"},
			result:          appv1.RefTargetRevisionMapping{},
		},
		{
			name:            "empty sources",
			refSources:      appv1.RefTargetRevisionMapping{},
			sources:         appv1.ApplicationSources{},
			syncedRevisions: []string{},
			result:          appv1.RefTargetRevisionMapping{},
		},
		{
			name: "no synced revisions",
			refSources: appv1.RefTargetRevisionMapping{
				"$values": &appv1.RefTarget{
					Repo:           appv1.Repository{Repo: "https://github.com/cd"},
					TargetRevision: "main-1",
					Chart:          "chart",
				},
			},
			sources: appv1.ApplicationSources{
				{RepoURL: "https://helm.registry", TargetRevision: "0.0.1", Chart: "my-chart", Helm: &appv1.ApplicationSourceHelm{ValueFiles: []string{"$values/path"}}},
				{RepoURL: "https://github.com/cd", TargetRevision: "main-1", Ref: "values"},
			},
			syncedRevisions: []string{},
			result: appv1.RefTargetRevisionMapping{
				"$values": &appv1.RefTarget{
					Repo:           appv1.Repository{Repo: "https://github.com/cd"},
					TargetRevision: "",
					Chart:          "chart",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			syncedRefSources := GetSyncedRefSources(tt.refSources, tt.sources, tt.syncedRevisions)
			assert.Equal(t, tt.result, syncedRefSources)
		})
	}
}

// Covers https://github.com/argoproj/argo-cd/issues/27759:
// spec.sources with one element still sets HasMultipleSources(); GetRefSources must populate refSources so
// GetSyncedRefSources does not dereference a missing map entry.
func Test_GetRefSourcesAndSyncedRefSources_SingleSourceWithRef(t *testing.T) {
	repoPath, err := filepath.Abs("./../..")
	require.NoError(t, err)
	repo := &appv1.Repository{Repo: "file://" + repoPath}
	sources := appv1.ApplicationSources{
		{RepoURL: "file://" + repoPath, TargetRevision: "HEAD", Ref: "gitops"},
	}

	refSources, err := GetRefSources(t.Context(), sources, "", func(_ context.Context, _ string, _ string) (*appv1.Repository, error) {
		return repo, nil
	}, []string{})
	require.NoError(t, err)

	syncedRevisions := []string{"rev"}
	synced := GetSyncedRefSources(refSources, sources, syncedRevisions)

	expected := appv1.RefTargetRevisionMapping{
		"$gitops": &appv1.RefTarget{
			Repo:           *repo,
			TargetRevision: "rev",
		},
	}
	assert.Equal(t, expected, synced)
}
