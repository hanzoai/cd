package generators

import (
	"errors"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	appv1alpha1 "github.com/hanzoai/cd/pkg/apis/application/v1alpha1"
	"github.com/hanzoai/cd/util/env"
)

// Generator defines the interface implemented by all ApplicationSet generators.
type Generator interface {
	// GenerateParams interprets the ApplicationSet and generates all relevant parameters for the application template.
	// The expected / desired list of parameters is returned, it then will be render and reconciled
	// against the current state of the Applications in the cluster.
	GenerateParams(appSetGenerator *appv1alpha1.ApplicationSetGenerator, applicationSetInfo *appv1alpha1.ApplicationSet, client client.Client) ([]map[string]any, error)

	// GetRequeueAfter is the generator can controller the next reconciled loop
	// In case there is more then one generator the time will be the minimum of the times.
	// In case NoRequeueAfter is empty, it will be ignored
	GetRequeueAfter(appSetGenerator *appv1alpha1.ApplicationSetGenerator) time.Duration

	// GetTemplate returns the inline template from the spec if there is any, or an empty object otherwise
	GetTemplate(appSetGenerator *appv1alpha1.ApplicationSetGenerator) *appv1alpha1.ApplicationSetTemplate
}

var (
	ErrEmptyAppSetGenerator = errors.New("ApplicationSet is empty")
	NoRequeueAfter          time.Duration
)

const (
	DefaultRequeueAfter = 3 * time.Minute
)

func getDefaultRequeueAfter() time.Duration {
	// Default is 3 minutes, min is 1 second, max is 1 year
	return env.ParseDurationFromEnv("CD_APPLICATIONSET_CONTROLLER_REQUEUE_AFTER", DefaultRequeueAfter, 1*time.Second, 8760*time.Hour)
}
