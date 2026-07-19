package e2e

import (
	"testing"

	"github.com/hanzoai/deploy/gitops-engine/pkg/health"
	. "github.com/hanzoai/deploy/gitops-engine/pkg/sync/common"

	. "github.com/hanzoai/deploy/pkg/apis/application/v1alpha1"
	. "github.com/hanzoai/deploy/test/e2e/fixture/app"
)

// ensure that cluster scoped objects, like a cluster role, as a hook, can be successfully deployed
func TestClusterRoleBindingHook(t *testing.T) {
	Given(t).
		Path("cluster-role-hook").
		When().
		CreateApp().
		Sync().
		Then().
		Expect(OperationPhaseIs(OperationSucceeded)).
		Expect(HealthIs(health.HealthStatusHealthy)).
		Expect(SyncStatusIs(SyncStatusCodeSynced))
}
