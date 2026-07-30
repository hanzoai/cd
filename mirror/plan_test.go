package mirror

import "testing"

// The allowlist must refuse an org we do not own, including one nobody thought
// about. A denylist can only refuse what it was told to; this refuses by default,
// which is the property that matters when a repo is transferred somewhere new.
func TestOwnedRefusesAnythingNotOurs(t *testing.T) {
	for _, org := range []string{"hanzoai", "hanzobot", "hanzo-ml"} {
		if !Owned[org] {
			t.Errorf("%s is ours and must be allowed", org)
		}
	}
	for _, full := range []string{
		"someone-else/thing",
		"an-org-nobody-added-yet/repo",
		"hanzoai-lookalike/repo", // near-miss must not pass
		"nope",                   // no slash at all
	} {
		if owned(full) {
			t.Errorf("owned(%q) = true; the allowlist admitted something not ours", full)
		}
	}
}
