package e2e

import (
	"testing"

	"github.com/hanzoai/deploy/common"
	. "github.com/hanzoai/deploy/pkg/apis/application/v1alpha1"
	. "github.com/hanzoai/deploy/test/e2e/fixture/app"
)

func TestAppSkipReconcileTrue(t *testing.T) {
	Given(t).
		Path(guestbookPath).
		When().
		// app should have no status
		CreateFromFile(func(app *Application) {
			app.Annotations = map[string]string{common.AnnotationKeyAppSkipReconcile: "true"}
		}).
		Then().
		Expect(NoStatus())
}

func TestAppSkipReconcileFalse(t *testing.T) {
	Given(t).
		Path(guestbookPath).
		When().
		// app should have status
		CreateFromFile(func(app *Application) {
			app.Annotations = map[string]string{common.AnnotationKeyAppSkipReconcile: "false"}
		}).
		Then().
		Expect(StatusExists())
}

func TestAppSkipReconcileNonBooleanValue(t *testing.T) {
	Given(t).
		Path(guestbookPath).
		When().
		// app should have status
		CreateFromFile(func(app *Application) {
			app.Annotations = map[string]string{common.AnnotationKeyAppSkipReconcile: "not a boolean value"}
		}).
		Then().
		Expect(StatusExists())
}
