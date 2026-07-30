package mirror

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// fakeGit records what it was asked to run and answers ancestry from a set.
type fakeGit struct {
	ancestors map[string]bool // "child<-parent" -> is-ancestor
	pushErr   error
	calls     []string
}

func (f *fakeGit) run(_ context.Context, args ...string) (string, error) {
	f.calls = append(f.calls, strings.Join(args, " "))
	switch args[0] {
	case "merge-base":
		// merge-base --is-ancestor <maybe-ancestor> <descendant>
		if f.ancestors[args[2]+"<-"+args[3]] {
			return "", nil
		}
		return "", errors.New("exit status 1")
	case "push":
		return "", f.pushErr
	}
	return "", nil
}

func planned() Planned {
	p := Planned{Entry: Entry{Org: "hanzoai", Name: "cloud", Direction: Native}}
	p.GitHub = &Repo{FullName: "hanzoai/cloud", Default: "main"}
	return p
}

// THE safety property: the push refspec must never carry a leading '+'. With one,
// git force-updates and a diverged branch is silently rewritten — which is how a
// push mirror once pruned a contiguous tag range off a production repo. Asserting
// on the argument, not on the outcome, because the outcome depends on git's mood
// and this depends only on us.
func TestFastForwardNeverForces(t *testing.T) {
	g := &fakeGit{ancestors: map[string]bool{"aaaaaaaa<-bbbbbbbb": true}}
	res := FastForward(context.Background(), g.run, planned(),
		"http://forge/hanzoai/cloud.git", "https://github.com/hanzoai/cloud.git",
		"main", "aaaaaaaa", "bbbbbbbb")

	if res.Outcome != Moved {
		t.Fatalf("outcome = %q, want %q (detail: %s)", res.Outcome, Moved, res.Detail)
	}
	for _, c := range g.calls {
		if strings.HasPrefix(c, "push") && strings.Contains(c, "+") {
			t.Fatalf("the push refspec carries a '+', which force-updates:\n  %s", c)
		}
		if strings.HasPrefix(c, "push") && strings.Contains(c, "--force") {
			t.Fatalf("the push uses --force:\n  %s", c)
		}
	}
}

// Divergence must be REPORTED and nothing touched. A sync job that settles a
// disagreement destroys whichever side lost the race.
func TestFastForwardRefusesToSettleDivergence(t *testing.T) {
	g := &fakeGit{ancestors: map[string]bool{}} // nothing is an ancestor of anything
	res := FastForward(context.Background(), g.run, planned(),
		"http://forge/x.git", "https://github.com/x.git", "main", "deadbeef", "cafebabe")

	if res.Outcome != Diverged {
		t.Fatalf("outcome = %q, want %q", res.Outcome, Diverged)
	}
	for _, c := range g.calls {
		if strings.HasPrefix(c, "push") {
			t.Fatalf("a diverged branch was pushed anyway: %s", c)
		}
	}
	if !strings.Contains(res.Detail, "not an ancestor") {
		t.Errorf("detail should say why a human is needed, got %q", res.Detail)
	}
}

func TestFastForwardEqualTipsDoesNothing(t *testing.T) {
	g := &fakeGit{}
	res := FastForward(context.Background(), g.run, planned(),
		"http://forge/x.git", "https://github.com/x.git", "main", "same", "same")
	if res.Outcome != InSync {
		t.Fatalf("outcome = %q, want %q", res.Outcome, InSync)
	}
	if len(g.calls) != 0 {
		t.Fatalf("steady state should cost no git calls, ran: %v", g.calls)
	}
}

// An unknown tip is not "equal". Treating it as equal would silently skip a repo
// forever; treating it as diverged would page a human over a missing branch.
func TestFastForwardUnknownTipIsSkippedNotEqual(t *testing.T) {
	for _, tc := range []struct{ forge, gh string }{{"", "abc"}, {"abc", ""}} {
		g := &fakeGit{}
		res := FastForward(context.Background(), g.run, planned(),
			"http://forge/x.git", "https://github.com/x.git", "main", tc.forge, tc.gh)
		if res.Outcome != Skipped {
			t.Errorf("forge=%q gh=%q: outcome = %q, want %q", tc.forge, tc.gh, res.Outcome, Skipped)
		}
		if len(g.calls) != 0 {
			t.Errorf("forge=%q gh=%q: should not touch git, ran %v", tc.forge, tc.gh, g.calls)
		}
	}
}

// A push that fails must report the reason WITHOUT the credential in the URL.
func TestFastForwardRedactsCredentialsOnFailure(t *testing.T) {
	g := &fakeGit{
		ancestors: map[string]bool{"aaaa<-bbbb": true},
		pushErr:   errors.New("fatal: unable to access 'http://oauth2:s3cr3t-token@forge/x.git': 403"),
	}
	res := FastForward(context.Background(), g.run, planned(),
		"http://oauth2:s3cr3t-token@forge/x.git", "https://github.com/x.git", "main", "aaaa", "bbbb")

	if res.Outcome != Failed {
		t.Fatalf("outcome = %q, want %q", res.Outcome, Failed)
	}
	if strings.Contains(res.Detail, "s3cr3t-token") {
		t.Fatalf("the token reached the reported detail: %q", res.Detail)
	}
	if !strings.Contains(res.Detail, "***") {
		t.Errorf("detail should show the redaction, got %q", res.Detail)
	}
}

// pushURL must preserve the configured scheme. The in-cluster forge is plain http,
// and assuming https makes every push fail against the only address a controller
// uses.
func TestPushURLKeepsTheConfiguredScheme(t *testing.T) {
	c := NewForge("http://hanzo-git.hanzo.svc", "tok")
	got := c.pushURL("http://hanzo-git.hanzo.svc", "hanzoai", "cloud")
	if !strings.HasPrefix(got, "http://oauth2:tok@hanzo-git.hanzo.svc/") {
		t.Fatalf("scheme or credential wrong: %s", redact(got))
	}
	if strings.HasPrefix(got, "https://") {
		t.Fatal("https was assumed; the in-cluster address is http")
	}
}

func TestSummaryFlagsOnlyWhatNeedsAHuman(t *testing.T) {
	counts, human := Summary([]Result{
		{Outcome: InSync}, {Outcome: InSync}, {Outcome: Moved},
		{Outcome: Diverged, Entry: Entry{Org: "o", Name: "d"}},
		{Outcome: Failed, Entry: Entry{Org: "o", Name: "f"}},
		{Outcome: Skipped},
	})
	if counts[InSync] != 2 || counts[Moved] != 1 || counts[Skipped] != 1 {
		t.Errorf("counts wrong: %v", counts)
	}
	if len(human) != 2 {
		t.Fatalf("needsHuman = %d, want 2 (diverged + failed)", len(human))
	}
	// Skipped must NOT page anyone: a repo with no such branch is a normal state.
	for _, r := range human {
		if r.Outcome == Skipped {
			t.Error("a skipped repo was reported as needing a human")
		}
	}
}
