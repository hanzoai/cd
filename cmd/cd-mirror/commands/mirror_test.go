package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hanzoai/cd/mirror"
)

func write(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "repos.json")
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

// Direction decides which side may overwrite the other, so an entry that omits it
// must be refused rather than assigned a default. Defaulting to native would turn
// every typo into a push at a repository we do not own; defaulting to mirror would
// silently stop syncing one we do. Neither is a guess a sync job gets to make.
func TestDirectionIsNeverAssumed(t *testing.T) {
	_, err := load(write(t, `[{"org":"hanzoai","name":"cloud"}]`))
	if err == nil {
		t.Fatal("an entry with no direction was accepted")
	}
	if !strings.Contains(err.Error(), "direction is required") {
		t.Fatalf("the error should name the missing field, got %q", err)
	}
}

func TestIncompleteEntryIsRefused(t *testing.T) {
	for _, body := range []string{
		`[{"org":"hanzoai","direction":"native"}]`,
		`[{"name":"cloud","direction":"native"}]`,
	} {
		if _, err := load(write(t, body)); err == nil {
			t.Errorf("accepted an entry with no org or no name: %s", body)
		}
	}
}

// An unreadable table must stop the run. Treating it as empty would report a clean
// reconcile of nothing — the exact shape of silence this program exists to remove.
func TestAbsentTableIsAnError(t *testing.T) {
	if _, err := load(filepath.Join(t.TempDir(), "nope.json")); err == nil {
		t.Fatal("a missing table was accepted")
	}
}

// The shipped table is the one the scheduled run reads, so it has to survive the
// same validation as any other. It caught nothing for weeks while living as a Go
// literal; parsing it here is what makes it config.
func TestShippedTableIsValid(t *testing.T) {
	entries, err := load("../../../mirror/repos.json")
	if err != nil {
		t.Fatalf("the shipped table does not load: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("the shipped table is empty")
	}
	for _, e := range entries {
		if e.Direction != mirror.Native && e.Direction != mirror.Mirror {
			t.Errorf("%s: direction %q is neither %q nor %q", e, e.Direction, mirror.Native, mirror.Mirror)
		}
	}
}

// Every declared repo must sit in an org the allowlist owns, or the entry is
// dead on arrival: `owned` refuses it, the plan drops it, and the report says
// only that fewer entries resolved than were declared. Adding a repo in a new
// org is therefore TWO edits, and this is what says so — otherwise the second
// one is discovered as a silence, which is the failure mode this whole package
// exists to remove.
//
// One direction only. The allowlist is deliberately broader than the table: a
// repo can be transferred into an org that has no entry yet, and `owned` guards
// the name GitHub resolves to, not the name declared here.
func TestEveryDeclaredOrgIsOwned(t *testing.T) {
	entries, err := load("../../../mirror/repos.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if !mirror.Owned[e.Org] {
			t.Errorf("%s: org %q is not in mirror.Owned, so this entry can never resolve", e, e.Org)
		}
	}
}

// The table replaced a Python dict, and the port dropped 16 of its 40 entries --
// among them hanzoai/ci, the reusable pipeline every repo in three orgs imports
// as `hanzoai/ci/...@v1`. It sat 18 commits behind with v1 pinned to a pipeline
// from months earlier, and a stale pipeline runs GREEN, so every caller believed
// it was current. A count is a blunt guard, but the thing it guards against is
// blunt too: a table that quietly loses rows.
func TestTableCoversEveryRepoTheScriptCarried(t *testing.T) {
	entries, err := load("../../../mirror/repos.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) < 40 {
		t.Errorf("table has %d entries, want at least the 40 the script carried", len(entries))
	}
	for _, want := range []string{"hanzoai/ci", "hanzoai/console", "hanzo-go/sdk", "hanzo-rs/sdk", "luxfi/aml"} {
		found := false
		for _, e := range entries {
			if e.Org+"/"+e.Name == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s is missing; it was in the table this replaced", want)
		}
	}
}
