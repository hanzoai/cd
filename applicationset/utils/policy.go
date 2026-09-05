package utils

import (
	"github.com/hanzoai/cd/pkg/apis/application/v1alpha1"
)

// Policies is a registry of available policies.
var Policies = map[string]v1alpha1.ApplicationsSyncPolicy{
	"create-only":   v1alpha1.ApplicationsSyncPolicyCreateOnly,
	"create-update": v1alpha1.ApplicationsSyncPolicyCreateUpdate,
	"create-delete": v1alpha1.ApplicationsSyncPolicyCreateDelete,
	"sync":          v1alpha1.ApplicationsSyncPolicySync,
	// Default is "sync"
	"": v1alpha1.ApplicationsSyncPolicySync,
}

func DefaultPolicy(appSetSyncPolicy *v1alpha1.ApplicationSetSyncPolicy, controllerPolicy v1alpha1.ApplicationsSyncPolicy, enablePolicyOverride bool) v1alpha1.ApplicationsSyncPolicy {
	if appSetSyncPolicy == nil || appSetSyncPolicy.ApplicationsSync == nil || !enablePolicyOverride {
		return controllerPolicy
	}
	return *appSetSyncPolicy.ApplicationsSync
}
