package mirror

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
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

// Git runs git. Injectable so the fast-forward decision is testable without a
// network or a real repository — the FF guarantee is the safety property here, and
// one that can only be tested against live remotes is one nobody tests.
type Git func(ctx context.Context, args ...string) (string, error)

// RealGit runs the git binary.
func RealGit(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, redact(string(out)))
	}
	return string(out), nil
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
func FastForward(ctx context.Context, git Git, p Planned, forgeURL, ghURL, branch, forgeTip, ghTip string) Result {
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

	// Ancestry decides, and it is asked of git rather than inferred from commit
	// counts or dates. "Behind by N" is a summary; "is an ancestor" is the fact.
	if _, err := git(ctx, "merge-base", "--is-ancestor", forgeTip, ghTip); err != nil {
		res.Outcome = Diverged
		res.Detail = fmt.Sprintf("forge %s is not an ancestor of github %s (%s)",
			short(forgeTip), short(ghTip), p.GitHub.FullName)
		return res
	}

	// No leading '+'. git refuses anything but a fast-forward.
	if out, err := git(ctx, "push", forgeURL, ghTip+":refs/heads/"+branch); err != nil {
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
func Reconcile(ctx context.Context, forge, gh *Client, git Git, forgeBase string, plan []Planned) []Result {
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

		res := FastForward(ctx, git, p,
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
