package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/hanzoai/cd/common"
	appsv1 "github.com/hanzoai/cd/pkg/apis/application/v1alpha1"
	"github.com/hanzoai/cd/util/settings"
)

var repoCD = &corev1.Secret{
	ObjectMeta: metav1.ObjectMeta{
		Namespace: testNamespace,
		Name:      "some-repo-secret",
		Annotations: map[string]string{
			common.AnnotationKeyManagedBy: common.AnnotationValueManagedByCD,
		},
		Labels: map[string]string{
			common.LabelKeySecretType: common.LabelValueSecretTypeRepository,
		},
	},
	Data: map[string][]byte{
		"name":     []byte("SomeRepo"),
		"url":      []byte("git@github.com:hanzoai/cd.git"),
		"username": []byte("someUsername"),
		"password": []byte("somePassword"),
		"type":     []byte("git"),
	},
}

var repoExampleApps = &corev1.Secret{
	ObjectMeta: metav1.ObjectMeta{
		Namespace: testNamespace,
		Name:      "some-other-repo-secret",
		Annotations: map[string]string{
			common.AnnotationKeyManagedBy: common.AnnotationValueManagedByCD,
		},
		Labels: map[string]string{
			common.LabelKeySecretType: common.LabelValueSecretTypeRepository,
		},
	},
	Data: map[string][]byte{
		"name":     []byte("OtherRepo"),
		"url":      []byte("git@github.com:hanzocd/example-apps.git"),
		"username": []byte("someUsername"),
		"password": []byte("somePassword"),
		"type":     []byte("git"),
	},
}

var repoCDWrite = &corev1.Secret{
	ObjectMeta: metav1.ObjectMeta{
		Namespace: testNamespace,
		Name:      "some-repo-secret",
		Annotations: map[string]string{
			common.AnnotationKeyManagedBy: common.AnnotationValueManagedByCD,
		},
		Labels: map[string]string{
			common.LabelKeySecretType: common.LabelValueSecretTypeRepositoryWrite,
		},
	},
	Data: map[string][]byte{
		"name":     []byte("SomeRepo"),
		"url":      []byte("git@github.com:hanzoai/cd.git"),
		"username": []byte("someUsername"),
		"password": []byte("somePassword"),
		"type":     []byte("git"),
	},
}

var repoExampleAppsWrite = &corev1.Secret{
	ObjectMeta: metav1.ObjectMeta{
		Namespace: testNamespace,
		Name:      "some-other-repo-secret",
		Annotations: map[string]string{
			common.AnnotationKeyManagedBy: common.AnnotationValueManagedByCD,
		},
		Labels: map[string]string{
			common.LabelKeySecretType: common.LabelValueSecretTypeRepositoryWrite,
		},
	},
	Data: map[string][]byte{
		"name":     []byte("OtherRepo"),
		"url":      []byte("git@github.com:hanzocd/example-apps.git"),
		"username": []byte("someUsername"),
		"password": []byte("somePassword"),
		"type":     []byte("git"),
	},
}

func TestDb_CreateRepository(t *testing.T) {
	t.Parallel()
	clientset := getClientset()
	settingsManager := settings.NewSettingsManager(t.Context(), clientset, testNamespace)
	testee := &db{
		ns:            testNamespace,
		kubeclientset: clientset,
		settingsMgr:   settingsManager,
	}

	input := &appsv1.Repository{
		Name:     "TestRepo",
		Repo:     "git@github.com:hanzoai/cd.git",
		Username: "someUsername",
		Password: "somePassword",
	}

	// The repository was indeed created successfully
	output, err := testee.CreateRepository(t.Context(), input)
	require.NoError(t, err)
	assert.Same(t, input, output)

	secret, err := clientset.CoreV1().Secrets(testNamespace).Get(
		t.Context(),
		RepoURLToSecretName(repoSecretPrefix, input.Repo, ""),
		metav1.GetOptions{},
	)
	assert.NotNil(t, secret)
	require.NoError(t, err)
}

func TestDb_GetRepository(t *testing.T) {
	t.Parallel()
	clientset := getClientset(repoCD, repoExampleApps)
	settingsManager := settings.NewSettingsManager(t.Context(), clientset, testNamespace)
	testee := &db{
		ns:            testNamespace,
		kubeclientset: clientset,
		settingsMgr:   settingsManager,
	}

	repository, err := testee.GetRepository(t.Context(), "git@github.com:hanzocd/example-apps.git", "")
	require.NoError(t, err)
	require.NotNil(t, repository)
	assert.Equal(t, "OtherRepo", repository.Name)

	repository, err = testee.GetRepository(t.Context(), "git@github.com:hanzoai/cd.git", "")
	require.NoError(t, err)
	require.NotNil(t, repository)
	assert.Equal(t, "SomeRepo", repository.Name)

	repository, err = testee.GetRepository(t.Context(), "git@github.com:hanzoai/not-existing.git", "")
	require.NoError(t, err)
	assert.NotNil(t, repository)
	assert.Equal(t, "git@github.com:hanzoai/not-existing.git", repository.Repo)
}

func TestDb_GetWriteRepository(t *testing.T) {
	t.Parallel()
	clientset := getClientset(repoCDWrite, repoExampleAppsWrite)
	settingsManager := settings.NewSettingsManager(t.Context(), clientset, testNamespace)
	testee := &db{
		ns:            testNamespace,
		kubeclientset: clientset,
		settingsMgr:   settingsManager,
	}

	repository, err := testee.GetWriteRepository(t.Context(), "git@github.com:hanzocd/example-apps.git", "")
	require.NoError(t, err)
	require.NotNil(t, repository)
	assert.Equal(t, "OtherRepo", repository.Name)

	repository, err = testee.GetWriteRepository(t.Context(), "git@github.com:hanzoai/cd.git", "")
	require.NoError(t, err)
	require.NotNil(t, repository)
	assert.Equal(t, "SomeRepo", repository.Name)
}

func TestDb_GetWriteRepository_SecretNotFound_DefaultRepo(t *testing.T) {
	t.Parallel()
	clientset := getClientset(repoCD)
	settingsManager := settings.NewSettingsManager(t.Context(), clientset, testNamespace)
	testee := &db{
		ns:            testNamespace,
		kubeclientset: clientset,
		settingsMgr:   settingsManager,
	}

	repository, err := testee.GetWriteRepository(t.Context(), "git@github.com:hanzoai/cd.git", "")
	require.NoError(t, err)
	require.NotNil(t, repository)
	assert.Empty(t, repository.Name)
}

func TestDb_ListRepositories(t *testing.T) {
	t.Parallel()
	clientset := getClientset(repoCD, repoExampleApps)
	settingsManager := settings.NewSettingsManager(t.Context(), clientset, testNamespace)
	testee := &db{
		ns:            testNamespace,
		kubeclientset: clientset,
		settingsMgr:   settingsManager,
	}

	repositories, err := testee.ListRepositories(t.Context())
	require.NoError(t, err)
	assert.Len(t, repositories, 2)
}

func TestDb_UpdateRepository(t *testing.T) {
	t.Parallel()
	secretRepository := &appsv1.Repository{
		Name:     "SomeRepo",
		Repo:     "git@github.com:hanzoai/cd.git",
		Username: "someUsername",
		Password: "somePassword",
		Type:     "git",
	}

	clientset := getClientset(repoCD)
	settingsManager := settings.NewSettingsManager(t.Context(), clientset, testNamespace)
	testee := &db{
		ns:            testNamespace,
		kubeclientset: clientset,
		settingsMgr:   settingsManager,
	}

	secretRepository.Username = "UpdatedUsername"
	repository, err := testee.UpdateRepository(t.Context(), secretRepository)
	require.NoError(t, err)
	assert.NotNil(t, repository)
	assert.Same(t, secretRepository, repository)

	secret, err := clientset.CoreV1().Secrets(testNamespace).Get(
		t.Context(),
		"some-repo-secret",
		metav1.GetOptions{},
	)
	require.NoError(t, err)
	assert.NotNil(t, secret)
	assert.Equal(t, "UpdatedUsername", string(secret.Data["username"]))
}

func TestDb_DeleteRepository(t *testing.T) {
	t.Parallel()
	clientset := getClientset(repoCD, repoExampleApps)
	settingsManager := settings.NewSettingsManager(t.Context(), clientset, testNamespace)
	testee := &db{
		ns:            testNamespace,
		kubeclientset: clientset,
		settingsMgr:   settingsManager,
	}

	err := testee.DeleteRepository(t.Context(), "git@github.com:hanzocd/example-apps.git", "")
	require.NoError(t, err)

	err = testee.DeleteRepository(t.Context(), "git@github.com:hanzoai/cd.git", "")
	require.NoError(t, err)

	_, err = clientset.CoreV1().Secrets(testNamespace).Get(t.Context(), "some-repo-secret", metav1.GetOptions{})
	require.Error(t, err)
}

func TestDb_GetRepositoryCredentials(t *testing.T) {
	t.Parallel()
	gitHubRepoCredsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      "some-repocreds-secret",
			Labels: map[string]string{
				common.LabelKeySecretType: common.LabelValueSecretTypeRepoCreds,
			},
		},
		Data: map[string][]byte{
			"type":     []byte("git"),
			"url":      []byte("git@github.com:hanzoai"),
			"username": []byte("someUsername"),
			"password": []byte("somePassword"),
		},
	}
	gitLabRepoCredsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      "some-other-repocreds-secret",
			Labels: map[string]string{
				common.LabelKeySecretType: common.LabelValueSecretTypeRepoCreds,
			},
		},
		Data: map[string][]byte{
			"type":     []byte("git"),
			"url":      []byte("git@gitlab.com"),
			"username": []byte("someUsername"),
			"password": []byte("somePassword"),
		},
	}

	clientset := getClientset(gitHubRepoCredsSecret, gitLabRepoCredsSecret)
	testee := NewDB(testNamespace, settings.NewSettingsManager(t.Context(), clientset, testNamespace), clientset)

	repoCreds, err := testee.GetRepositoryCredentials(t.Context(), "git@github.com:hanzoai/cd.git")
	require.NoError(t, err)
	require.NotNil(t, repoCreds)
	assert.Equal(t, "git@github.com:hanzoai", repoCreds.URL)

	repoCreds, err = testee.GetRepositoryCredentials(t.Context(), "git@gitlab.com:someorg/foobar.git")
	require.NoError(t, err)
	require.NotNil(t, repoCreds)
	assert.Equal(t, "git@gitlab.com", repoCreds.URL)

	repoCreds, err = testee.GetRepositoryCredentials(t.Context(), "git@github.com:example/not-existing.git")
	require.NoError(t, err)
	assert.Nil(t, repoCreds)
}

func TestRepoURLToSecretName(t *testing.T) {
	t.Parallel()
	tables := []struct {
		repoURL    string
		secretName string
		project    string
	}{{
		repoURL:    "git://git@github.com:hanzoai/CD.git",
		secretName: "repo-2548978051",
		project:    "",
	}, {
		repoURL:    "git://git@github.com:hanzoai/CD.git",
		secretName: "repo-3943644754",
		project:    "foobar",
	}, {
		repoURL:    "https://github.com/hanzoai/CD",
		secretName: "repo-1502159595",
		project:    "",
	}, {
		repoURL:    "https://github.com/hanzoai/CD",
		secretName: "repo-3522625594",
		project:    "foobar",
	}, {
		repoURL:    "https://github.com/hanzoai/cd",
		secretName: "repo-2044163115",
		project:    "",
	}, {
		repoURL:    "https://github.com/hanzoai/cd",
		secretName: "repo-1336240250",
		project:    "foobar",
	}, {
		repoURL:    "https://github.com/hanzoai/cd.git",
		secretName: "repo-2463124965",
		project:    "",
	}, {
		repoURL:    "https://github.com/hanzoai/cd.git",
		secretName: "repo-247532680",
		project:    "foobar",
	}, {
		repoURL:    "https://github.com/hanzoai/hanzo_cd.git",
		secretName: "repo-3196856040",
		project:    "",
	}, {
		repoURL:    "https://github.com/hanzoai/hanzo_cd.git",
		secretName: "repo-310744421",
		project:    "foobar",
	}, {
		repoURL:    "ssh://git@github.com/hanzoai/cd.git",
		secretName: "repo-1441016006",
		project:    "",
	}, {
		repoURL:    "ssh://git@github.com/hanzoai/cd.git",
		secretName: "repo-3989133391",
		project:    "foobar",
	}}

	for _, v := range tables {
		sn := RepoURLToSecretName(repoSecretPrefix, v.repoURL, v.project)
		assert.Equal(t, sn, v.secretName, "Expected secret name %q for repo %q; instead, got %q", v.secretName, v.repoURL, sn)
	}
}

func Test_CredsURLToSecretName(t *testing.T) {
	tables := map[string]string{
		"git://git@github.com:hanzoai":  "creds-2689407981",
		"git://git@github.com:hanzoai/": "creds-459453030",
		"git@github.com:hanzoai":        "creds-2292954401",
		"git@github.com:hanzoai/":       "creds-877528330",
	}

	for k, v := range tables {
		sn := RepoURLToSecretName(credSecretPrefix, k, "")
		assert.Equal(t, sn, v, "Expected secret name %q for repo %q; instead, got %q", v, k, sn)
	}
}

func Test_GetProjectRepositories(t *testing.T) {
	repoSecretWithProject := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      "some-repo-secret",
			Labels: map[string]string{
				common.LabelKeySecretType: common.LabelValueSecretTypeRepository,
			},
		},
		Data: map[string][]byte{
			"type":    []byte("git"),
			"url":     []byte("git@github.com:hanzoai/cd"),
			"project": []byte("some-project"),
		},
	}

	repoSecretWithoutProject := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      "some-other-repo-secret",
			Labels: map[string]string{
				common.LabelKeySecretType: common.LabelValueSecretTypeRepository,
			},
		},
		Data: map[string][]byte{
			"type": []byte("git"),
			"url":  []byte("git@github.com:hanzoai/cd"),
		},
	}

	clientset := getClientset(repoSecretWithProject, repoSecretWithoutProject)
	repoDB := NewDB(testNamespace, settings.NewSettingsManager(t.Context(), clientset, testNamespace), clientset)

	repos, err := repoDB.GetProjectRepositories("some-project")
	require.NoError(t, err)
	assert.Len(t, repos, 1)
	assert.Equal(t, "git@github.com:hanzoai/cd", repos[0].Repo)
}
