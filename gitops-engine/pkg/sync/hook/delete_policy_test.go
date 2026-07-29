package hook

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hanzoai/cd/gitops-engine/pkg/sync/common"
	testingutils "github.com/hanzoai/cd/gitops-engine/pkg/utils/testing"
)

func TestDeletePolicies(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []common.HookDeletePolicy{common.HookDeletePolicyBeforeHookCreation}, DeletePolicies(testingutils.NewPod()))
	assert.Equal(t, []common.HookDeletePolicy{common.HookDeletePolicyBeforeHookCreation}, DeletePolicies(testingutils.Annotate(testingutils.NewPod(), "cd.hanzo.ai/hook-delete-policy", "garbage")))
	assert.Equal(t, []common.HookDeletePolicy{common.HookDeletePolicyBeforeHookCreation}, DeletePolicies(testingutils.Annotate(testingutils.NewPod(), "cd.hanzo.ai/hook-delete-policy", "BeforeHookCreation")))
	assert.Equal(t, []common.HookDeletePolicy{common.HookDeletePolicyHookSucceeded}, DeletePolicies(testingutils.Annotate(testingutils.NewPod(), "cd.hanzo.ai/hook-delete-policy", "HookSucceeded")))
	assert.Equal(t, []common.HookDeletePolicy{common.HookDeletePolicyHookFailed}, DeletePolicies(testingutils.Annotate(testingutils.NewPod(), "cd.hanzo.ai/hook-delete-policy", "HookFailed")))
	// Helm test
	assert.Equal(t, []common.HookDeletePolicy{common.HookDeletePolicyHookSucceeded}, DeletePolicies(testingutils.Annotate(testingutils.NewPod(), "helm.sh/hook-delete-policy", "hook-succeeded")))
}
