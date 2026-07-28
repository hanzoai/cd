package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hanzoai/deploy/reposerver/apiclient"
)

func TestGetCommonRootPath(t *testing.T) {
	t.Parallel()

	repoRoot := "/tmp/_cd-repo/7a58c52a-0030-4fd9-8cc5-35b2d8b4e731"

	tests := []struct {
		name             string
		annotation       string
		appPath          string
		expectedRootPath string
	}{
		{"app path", ".", "/tmp/_cd-repo/7a58c52a-0030-4fd9-8cc5-35b2d8b4e731/services/helloworld", "/tmp/_cd-repo/7a58c52a-0030-4fd9-8cc5-35b2d8b4e731/services/helloworld"},
		{"app path and relative", "../../overlays;.", "/tmp/_cd-repo/7a58c52a-0030-4fd9-8cc5-35b2d8b4e731/services/helloworld", repoRoot},
		{"app path and absolute path", "/services;.", "/tmp/_cd-repo/7a58c52a-0030-4fd9-8cc5-35b2d8b4e731/services/helloworld", "/tmp/_cd-repo/7a58c52a-0030-4fd9-8cc5-35b2d8b4e731/services"},
		{"several relative paths", "../../;..;.", "/tmp/_cd-repo/7a58c52a-0030-4fd9-8cc5-35b2d8b4e731/services/team/helloworld", "/tmp/_cd-repo/7a58c52a-0030-4fd9-8cc5-35b2d8b4e731/services"},
		// backward compatibility test
		{"no annotation", "", "/tmp/_cd-repo/7a58c52a-0030-4fd9-8cc5-35b2d8b4e731/services/helloworld", repoRoot},
		// appPath should be the lower calculated root path
		{"relative subdir", "./manifests", "/tmp/_cd-repo/7a58c52a-0030-4fd9-8cc5-35b2d8b4e731/services/team/helloworld", "/tmp/_cd-repo/7a58c52a-0030-4fd9-8cc5-35b2d8b4e731/services/team/helloworld"},
		// glob pattern
		{"glob", "/services/shared/*-secret.yaml", "/tmp/_cd-repo/7a58c52a-0030-4fd9-8cc5-35b2d8b4e731/services/helloworld", "/tmp/_cd-repo/7a58c52a-0030-4fd9-8cc5-35b2d8b4e731/services"},
		{"relative glob", "../*", "/tmp/_cd-repo/7a58c52a-0030-4fd9-8cc5-35b2d8b4e731/services/helloworld", "/tmp/_cd-repo/7a58c52a-0030-4fd9-8cc5-35b2d8b4e731/services"},
		{"duplicate slashes", "//services/shared/*-secret.yaml", "/tmp/_cd-repo/7a58c52a-0030-4fd9-8cc5-35b2d8b4e731/services/helloworld", "/tmp/_cd-repo/7a58c52a-0030-4fd9-8cc5-35b2d8b4e731/services"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := &apiclient.ManifestRequest{AnnotationManifestGeneratePaths: tt.annotation}
			rootPath := getApplicationRootPath(req, tt.appPath, repoRoot)
			assert.Equal(t, tt.expectedRootPath, rootPath, "input and output should match")
		})
	}
}
