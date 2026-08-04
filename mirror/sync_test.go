package mirror

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// exitErr models a failed process the way os/exec does: the status IS the
// message. The fake has to carry it, because the code under test now reads exit 1
// as an answer ("not an ancestor") and anything else as a failure to answer — a
// fake that returns a bare error can only test one of those two paths.
type exitErr struct{ code int }

func (e exitErr) Error() string { return fmt.Sprintf("exit status %d", e.code) }
func (e exitErr) ExitCode() int { return e.code }

// fakeGit records what it was asked to run and answers ancestry from a set.
type fakeGit struct {
	ancestors map[string]bool // "child<-parent" -> is-ancestor
	pushErr   error
	fetchErr  error
	ancErr    error // when set, merge-base fails with this instead of answering
	calls     []string
	dirs      []string // the working directory each call ran in
}

func (f *fakeGit) run(_ context.Context, dir string, args ...string) (string, error) {
	f.dirs = append(f.dirs, dir)
	f.calls = append(f.calls, strings.Join(args, " "))
	switch args[0] {
	case "fetch":
		return "", f.fetchErr
	case "merge-base":
		// merge-base --is-ancestor <maybe-ancestor> <descendant>
		if f.ancErr != nil {
			return "fatal: Not a valid object name", f.ancErr
		}
		if f.ancestors[args[2]+"<-"+args[3]] {
			return "", nil
		}
		return "", exitErr{1}
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
	res := FastForward(context.Background(), g.run, t.TempDir(), planned(),
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
	res := FastForward(context.Background(), g.run, t.TempDir(), planned(),
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
	res := FastForward(context.Background(), g.run, t.TempDir(), planned(),
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
		res := FastForward(context.Background(), g.run, t.TempDir(), planned(),
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
	res := FastForward(context.Background(), g.run, t.TempDir(), planned(),
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

// git failing to ANSWER is not the same as git answering "diverged". Before the
// fetch existed, ghTip named an object git did not have, merge-base exited 128,
// and this code called it divergence — so every repo demanded a human look at a
// history that was never in conflict. Exit 1 is an answer; 128 is a broken job.
func TestUnreadableAncestryIsNotDivergence(t *testing.T) {
	g := &fakeGit{ancErr: exitErr{128}}
	res := FastForward(context.Background(), g.run, t.TempDir(), planned(),
		"http://forge/hanzoai/cloud.git", "https://github.com/hanzoai/cloud.git",
		"main", "aaaaaaaa", "bbbbbbbb")

	if res.Outcome != Failed {
		t.Fatalf("outcome = %q, want %q — an unreadable repo must not be reported as a question for a person (detail: %s)",
			res.Outcome, Failed, res.Detail)
	}
}

// BOTH tips have to be local before anything can be asked about them. ghURL was
// accepted and never used, so nothing was ever fetched; the forge side was never
// fetched either, which left a genuinely diverged tip unresolvable and reported as
// "ancestry undetermined" instead of by name.
func TestBothSidesAreFetchedBeforeComparing(t *testing.T) {
	g := &fakeGit{ancestors: map[string]bool{"aaaaaaaa<-bbbbbbbb": true}}
	FastForward(context.Background(), g.run, t.TempDir(), planned(),
		"http://forge/hanzoai/cloud.git", "https://github.com/hanzoai/cloud.git",
		"main", "aaaaaaaa", "bbbbbbbb")

	var fetched []string
	for _, c := range g.calls {
		if strings.HasPrefix(c, "merge-base") || strings.HasPrefix(c, "push") {
			break
		}
		if strings.HasPrefix(c, "fetch ") {
			fetched = append(fetched, c)
		}
	}
	if len(fetched) != 2 {
		t.Fatalf("fetches before the first comparison = %d, want 2 (forge then github); calls: %v", len(fetched), g.calls)
	}
	if !strings.Contains(fetched[0], "http://forge/hanzoai/cloud.git") {
		t.Errorf("the forge is fetched first, so its objects are offered in negotiation; got %q", fetched[0])
	}
	if !strings.Contains(fetched[1], "https://github.com/hanzoai/cloud.git") {
		t.Errorf("github was not fetched: %q", fetched[1])
	}
}

// EVERY git command must run inside a repository. Nothing here declared one, so
// each command ran in whatever directory the process started in — for a scheduled
// job, none — and git answered "fatal: not a git repository" to the fetch. That
// error arrives on the path that reports GitHub as unreadable, so a reconciler with
// perfect credentials and perfect config would have called every repo in the fleet
// unreachable and asked a person to look.
func TestEveryGitCallRunsInARepository(t *testing.T) {
	g := &fakeGit{ancestors: map[string]bool{"aaaaaaaa<-bbbbbbbb": true}}
	root := t.TempDir()
	FastForward(context.Background(), g.run, root, planned(),
		"http://forge/hanzoai/cloud.git", "https://github.com/hanzoai/cloud.git",
		"main", "aaaaaaaa", "bbbbbbbb")

	if len(g.dirs) == 0 {
		t.Fatal("no git ran at all")
	}
	for i, d := range g.dirs {
		if d == "" {
			t.Fatalf("%q ran with no working directory — git has no repository there", g.calls[i])
		}
		if !strings.HasPrefix(d, root) {
			t.Fatalf("%q ran in %q, outside the workspace root %q", g.calls[i], d, root)
		}
	}
}

// The store is per repo and reusable: a second run over the same root must find the
// first run's repository rather than start again. That is the whole reason a fetch
// from GitHub costs a delta instead of a repository, and it is why the job is
// affordable on a ten-minute schedule.
func TestWorkspaceCreatesThenReuses(t *testing.T) {
	root := t.TempDir()
	e := Entry{Org: "hanzoai", Name: "cloud", Direction: Native}

	first, err := Workspace(context.Background(), RealGit, root, e)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	// A marker no `git init` writes: if the second call reinitialised from scratch
	// the objects would be gone, which is exactly what a delta fetch depends on.
	marker := filepath.Join(first, "objects", "info", "mirror-test")
	if err := os.WriteFile(marker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	second, err := Workspace(context.Background(), RealGit, root, e)
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if second != first {
		t.Fatalf("a repo's store moved between runs: %q then %q", first, second)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("the existing store was discarded, so every fetch pays full price: %v", err)
	}
	if _, err := os.Stat(filepath.Join(first, "HEAD")); err != nil {
		t.Fatalf("not a git repository: %v", err)
	}
}

// A repo we cannot read is a failure of this job, not a verdict about history.
func TestUnreachableGitHubFailsWithoutJudging(t *testing.T) {
	g := &fakeGit{fetchErr: exitErr{128}}
	res := FastForward(context.Background(), g.run, t.TempDir(), planned(),
		"http://forge/hanzoai/cloud.git", "https://github.com/hanzoai/cloud.git",
		"main", "aaaaaaaa", "bbbbbbbb")

	if res.Outcome != Failed {
		t.Fatalf("outcome = %q, want %q", res.Outcome, Failed)
	}
	for _, c := range g.calls {
		if strings.HasPrefix(c, "push ") {
			t.Fatal("pushed after failing to read github")
		}
	}
}

// End to end against real git, because every fake in this file agrees with the
// code about what git is. Two bare repositories stand in for the forge and
// GitHub; the forge is left one commit behind and must end up on GitHub's tip.
//
// This is the test that fails outright without a workspace. Every git command
// here used to run in whatever directory the process started in, and a scheduled
// job starts in none, so git answered "fatal: not a git repository" to the fetch
// — on the branch that reports GITHUB as unreadable. Perfect credentials and a
// perfect table would still have produced a fleet-wide "cannot read github".
func TestFastForwardMovesARealRepository(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	ctx := context.Background()
	tmp := t.TempDir()
	run := func(dir string, args ...string) string {
		t.Helper()
		out, err := RealGit(ctx, dir, args...)
		if err != nil {
			t.Fatalf("git %v in %s: %v", args, dir, err)
		}
		return strings.TrimSpace(out)
	}

	forge := filepath.Join(tmp, "forge.git")
	github := filepath.Join(tmp, "github.git")
	for _, d := range []string{forge, github} {
		if err := os.MkdirAll(d, 0o750); err != nil {
			t.Fatal(err)
		}
		run(d, "init", "--bare", "--quiet", "--initial-branch=main")
	}

	src := filepath.Join(tmp, "src")
	if err := os.MkdirAll(src, 0o750); err != nil {
		t.Fatal(err)
	}
	run(src, "init", "--quiet", "--initial-branch=main")
	run(src, "config", "user.email", "test@hanzo.ai")
	run(src, "config", "user.name", "test")

	commit := func(name string) string {
		t.Helper()
		if err := os.WriteFile(filepath.Join(src, name), []byte(name), 0o600); err != nil {
			t.Fatal(err)
		}
		run(src, "add", name)
		run(src, "commit", "--quiet", "-m", name)
		return run(src, "rev-parse", "HEAD")
	}

	base := commit("one")
	run(src, "push", "--quiet", forge, "main")
	run(src, "push", "--quiet", github, "main")
	ahead := commit("two")
	run(src, "push", "--quiet", github, "main") // github moves, the forge does not

	res := FastForward(ctx, RealGit, filepath.Join(tmp, "work"), planned(),
		forge, github, "main", base, ahead)

	if res.Outcome != Moved {
		t.Fatalf("outcome = %q, want %q (detail: %s)", res.Outcome, Moved, res.Detail)
	}
	if got := run(forge, "rev-parse", "main"); got != ahead {
		t.Fatalf("the forge is at %s, want github's tip %s", got, ahead)
	}

	// And the safety property, proven by git rather than by argument inspection:
	// rewind GitHub so the two genuinely disagree, and the forge must not move.
	run(src, "reset", "--quiet", "--hard", base)
	sideways := commit("three")
	run(src, "push", "--quiet", "--force", github, "main")

	res = FastForward(ctx, RealGit, filepath.Join(tmp, "work"), planned(),
		forge, github, "main", ahead, sideways)

	if res.Outcome != Diverged {
		t.Fatalf("outcome = %q, want %q (detail: %s)", res.Outcome, Diverged, res.Detail)
	}
	if got := run(forge, "rev-parse", "main"); got != ahead {
		t.Fatalf("a diverged history was overwritten: forge is at %s, was %s", got, ahead)
	}
}
