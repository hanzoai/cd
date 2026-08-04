package mirror

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Outcome is what a reconcile did to one repo. Diverged is not an error state to be
// retried — it is a question for a person, and saying so is the whole point.
type Outcome string

const (
	InSync   Outcome = "in-sync"
	Moved    Outcome = "moved"
	Diverged Outcome = "diverged"
	Skipped  Outcome = "skipped"
	Failed   Outcome = "failed"
)

// Result is one repo's reconcile.
type Result struct {
	Entry   Entry
	Outcome Outcome
	From    string
	To      string
	Detail  string
}

// credential strips userinfo from anything before it is printed. A URL carrying a
// token must never reach a log, and the only reliable way to guarantee that is to
// pass every message through here rather than remembering to at each call site.
var credential = regexp.MustCompile(`//[^/@\s]*@`)

func redact(s string) string { return credential.ReplaceAllString(s, "//***@") }

// Git runs git in a working directory. Injectable so the fast-forward decision is
// testable without a network or a real repository — the FF guarantee is the safety
// property here, and one that can only be tested against live remotes is one nobody
// tests.
//
// The directory is a parameter rather than the process's own, because every command
// here needs a repository to hold objects in and a scheduled job starts in none.
// Run from a bare working directory git answers "fatal: not a git repository", and
// this package would have reported that as GitHub being unreadable.
type Git func(ctx context.Context, dir string, args ...string) (string, error)

// RealGit runs the git binary.
func RealGit(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, redact(string(out)))
	}
	return string(out), nil
}

// Workspace prepares the object store for one repo and returns its path.
//
// One store per repo, under a root that MAY survive between runs: git's fetch
// negotiation offers whatever is already here, so a kept root makes every later run
// a delta. It is only ever an optimisation — an empty root is equally correct, just
// slower — so this needs no volume to be right, and losing the volume costs time
// rather than correctness.
func Workspace(ctx context.Context, git Git, root string, e Entry) (string, error) {
	dir := filepath.Join(root, e.Org, e.Name+".git")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", fmt.Errorf("prepare %s: %w", dir, err)
	}
	// init is idempotent — on an existing store it rereads the config and leaves
	// every object alone — so "create or reuse" is one command instead of a stat and
	// a branch on its answer.
	if out, err := git(ctx, dir, "init", "--bare", "--quiet"); err != nil {
		return "", fmt.Errorf("prepare %s: %s", dir, detailOf(out, err))
	}
	return dir, nil
}

// exitCode reports a failed command's process exit status, or -1 when the error
// is not one. git uses status to say WHICH answer it gave — merge-base
// --is-ancestor exits 1 for "no" and reserves other codes for "I could not tell
// you" — so a caller that only checks err != nil cannot tell an answer from a
// failure. Read through wrapping, since RealGit annotates with %w.
func exitCode(err error) int {
	var ec interface{ ExitCode() int }
	if errors.As(err, &ec) {
		return ec.ExitCode()
	}
	return -1
}

// detailOf prefers git's own first line over the wrapped error: the command says
// what went wrong far better than the exit status does. Redacted, because these
// URLs carry an admin token.
func detailOf(out string, err error) string {
	if d := redact(firstLine(out)); d != "" {
		return d
	}
	return redact(err.Error())
}

// FastForward moves the forge's branch up to GitHub's, or refuses and says why.
//
// The guarantee is GIT's, not this function's: the refspec carries no leading '+',
// so a non-fast-forward is refused by git itself. That ordering matters — a check
// written here could be reasoned around by a later edit, whereas a missing '+'
// cannot be. A push mirror once pruned the v1.0-v1.31 tag range off an IAM repo and
// the objects were then collected; every safety property here exists because of
// something like that.
//
// Divergence returns Diverged with both tips and touches nothing. Two people
// disagreeing about history is not a race for a sync job to settle: settling it
// destroys whichever side lost.
func FastForward(ctx context.Context, git Git, work string, p Planned, forgeURL, ghURL, branch, forgeTip, ghTip string) Result {
	res := Result{Entry: p.Entry, From: forgeTip, To: ghTip}

	if forgeTip == ghTip {
		res.Outcome = InSync
		return res
	}
	if forgeTip == "" || ghTip == "" {
		res.Outcome = Skipped
		res.Detail = "a tip is unknown; nothing to compare"
		return res
	}

	// Nothing above this line touches disk, so a fleet in steady state costs no
	// repository at all — the store is prepared only for a repo that has actually
	// moved.
	dir, err := Workspace(ctx, git, work, p.Entry)
	if err != nil {
		res.Outcome = Failed
		res.Detail = redact(err.Error())
		return res
	}

	// Both tips must be present locally before anything can be asked about them.
	// ghURL used to be accepted and never used: nothing fetched, so ghTip named an
	// object git did not have, merge-base could not resolve it, and the error below
	// was read as divergence. Every repo then reported "diverged — needs a person",
	// which is the worst failure a reconciler can have: it invents work for a human
	// and is unfalsifiable from its own output. The push would have failed for the
	// same missing-object reason had anything ever reached it.
	//
	// The FORGE is fetched first, and not only so a diverged tip can be named. With
	// its history present git offers those objects in negotiation, so the GitHub
	// fetch carries the delta instead of the repository — which is what keeps this
	// affordable every ten minutes over two dozen repos. Fetching GitHub alone
	// leaves a diverged forge tip absent, and the honest report for that is
	// "ancestry undetermined" — true, and the one case a person most needs named.
	//
	// The refspec's leading '+' is NOT the property this package guards. These are
	// scratch refs in a store we own, where a forced update destroys nothing; they
	// exist so the next fetch has something to negotiate with. The refspec that must
	// never carry one is the push, below.
	for _, side := range []struct{ name, url string }{{"forge", forgeURL}, {"github", ghURL}} {
		refspec := "+refs/heads/" + branch + ":refs/remotes/" + side.name + "/" + branch
		if out, err := git(ctx, dir, "fetch", "--no-tags", side.url, refspec); err != nil {
			res.Outcome = Failed
			res.Detail = "cannot read " + side.name + " " + p.GitHub.FullName + ": " + detailOf(out, err)
			return res
		}
	}

	// Ancestry decides, and it is asked of git rather than inferred from commit
	// counts or dates. "Behind by N" is a summary; "is an ancestor" is the fact.
	//
	// Exit 1 is git's ANSWER — "no, not an ancestor". Any other status is git
	// failing to answer at all, and the two must not collapse into one outcome:
	// divergence is a real state a person should look at, while an unreadable
	// repository is a broken job. Reporting the second as the first is what sent
	// people to inspect histories that were never in conflict.
	if out, err := git(ctx, dir, "merge-base", "--is-ancestor", forgeTip, ghTip); err != nil {
		if exitCode(err) != 1 {
			res.Outcome = Failed
			res.Detail = "ancestry undetermined for " + p.GitHub.FullName + ": " + detailOf(out, err)
			return res
		}
		res.Outcome = Diverged
		res.Detail = fmt.Sprintf("forge %s is not an ancestor of github %s (%s)",
			short(forgeTip), short(ghTip), p.GitHub.FullName)
		return res
	}

	// No leading '+'. git refuses anything but a fast-forward.
	if out, err := git(ctx, dir, "push", forgeURL, ghTip+":refs/heads/"+branch); err != nil {
		res.Outcome = Failed
		res.Detail = redact(firstLine(out))
		if res.Detail == "" {
			res.Detail = redact(err.Error())
		}
		return res
	}
	res.Outcome = Moved
	return res
}

func short(s string) string {
	if len(s) > 8 {
		return s[:8]
	}
	return s
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// Reconcile brings every declared NATIVE repo's forge branch up to GitHub's.
//
// Only native entries: a mirror tracks an upstream we do not own, and pushing INTO
// one would fight the thing it exists to follow.
//
// One repo's failure never ends the run. A reconcile that stops at the first
// problem leaves every repo after it unexamined, and the ordering is by size, so
// the unexamined ones would be arbitrary.
func Reconcile(ctx context.Context, forge, gh *Client, git Git, forgeBase, work string, plan []Planned) []Result {
	var out []Result
	for _, p := range plan {
		if p.Direction != Native {
			continue
		}
		branch := p.GitHub.Default
		if branch == "" {
			branch = "main"
		}

		forgeTip, err := forge.Tip(ctx, p.Org, p.Name, branch)
		if err != nil {
			// No such branch here is not a failure to fix by inventing one: creating a
			// branch is not a fast-forward of anything, and guessing that the forge
			// wants it is how a sync job invents state.
			o := Failed
			if errors.Is(err, ErrNotFound) {
				o = Skipped
			}
			out = append(out, Result{Entry: p.Entry, Outcome: o, Detail: redact(err.Error())})
			continue
		}
		ghOwner, ghName := split(p.GitHub.FullName)
		ghTip, err := gh.Tip(ctx, ghOwner, ghName, branch)
		if err != nil {
			out = append(out, Result{Entry: p.Entry, Outcome: Skipped, Detail: redact(err.Error())})
			continue
		}

		res := FastForward(ctx, git, work, p,
			forge.pushURL(forgeBase, p.Org, p.Name), p.GitHub.CloneURL, branch, forgeTip, ghTip)
		out = append(out, res)
	}
	return out
}

// pushURL builds a pushable forge URL carrying the admin token.
//
// The scheme comes from the configured base rather than being assumed https: the
// in-cluster address is plain http, and hardcoding https here would make every
// push fail against the only address a controller actually uses.
func (c *Client) pushURL(base, org, name string) string {
	base = strings.TrimRight(base, "/")
	scheme, rest, ok := strings.Cut(base, "://")
	if !ok {
		return base + "/" + org + "/" + name + ".git"
	}
	return fmt.Sprintf("%s://oauth2:%s@%s/%s/%s.git", scheme, c.token, rest, org, name)
}

func split(full string) (string, string) {
	org, name, _ := strings.Cut(full, "/")
	return org, name
}

// Summary counts outcomes, and reports whether a human is needed.
func Summary(rs []Result) (counts map[Outcome]int, needsHuman []Result) {
	counts = map[Outcome]int{}
	for _, r := range rs {
		counts[r.Outcome]++
		if r.Outcome == Diverged || r.Outcome == Failed {
			needsHuman = append(needsHuman, r)
		}
	}
	return counts, needsHuman
}
