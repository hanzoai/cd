package status

import (
	"github.com/hanzoai/cd/pkg/apis/application/v1alpha1"
)

func BuildResourceStatus(statusMap map[string]v1alpha1.ResourceStatus, apps []v1alpha1.Application) map[string]v1alpha1.ResourceStatus {
	appMap := map[string]v1alpha1.Application{}
	for _, app := range apps {
		appMap[app.Name] = app

		gvk := app.GroupVersionKind()
		var status v1alpha1.ResourceStatus
		status.Group = gvk.Group
		status.Version = gvk.Version
		status.Kind = gvk.Kind
		status.Name = app.Name
		status.Namespace = app.Namespace
		status.Status = app.Status.Sync.Status
		status.Health = &v1alpha1.HealthStatus{Status: app.Status.Health.Status}

		statusMap[app.Name] = status
	}
	cleanupDeletedApplicationStatuses(statusMap, appMap)

	return statusMap
}

func GetResourceStatusMap(appset *v1alpha1.ApplicationSet) map[string]v1alpha1.ResourceStatus {
	statusMap := map[string]v1alpha1.ResourceStatus{}
	for _, status := range appset.Status.Resources {
		statusMap[status.Name] = status
	}
	return statusMap
}

func cleanupDeletedApplicationStatuses(statusMap map[string]v1alpha1.ResourceStatus, apps map[string]v1alpha1.Application) {
	for name := range statusMap {
		if _, ok := apps[name]; !ok {
			delete(statusMap, name)
		}
	}
}
