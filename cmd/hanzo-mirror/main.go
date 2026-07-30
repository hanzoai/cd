// Command hanzo-mirror reconciles which repositories live on git.hanzo.ai and
// which direction each one syncs.
//
// It replaces hanzoai/mirrors' reconcile.py + sync.py. The behaviour it preserves
// is deliberate and each part was learned the hard way; the behaviour it drops is
// the part that made those lessons necessary:
//
//   - The direction table is CONFIG, read from a file, not a list literal in
//     source. Adding a repo was a code edit, which is not what declared state
//     means. hanzoai/cloud was added that way hours ago, and until then it was in
//     neither sync leg: is_mirror=0 so nothing pulled in, absent from the table so
//     nothing pushed in, and its CI therefore ran on a commit four behind for a
//     day while looking merely "red".
//   - Names are RESOLVED, never matched against a listing. A listing reports a
//     repo only under its current name in its current org, so six entries whose
//     repos had been renamed or transferred were skipped silently — an absent name
//     raises nothing.
//   - Fast-forward is git's guarantee, not this program's: no leading '+' in the
//     refspec. Divergence is reported and left alone.
//
// Exit status is 0 when every declared repo is in sync or was moved forward, and 1
// when any diverged or failed — so a scheduler can tell "nothing to do" from
// "someone needs to look", which a script that only printed could not.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/hanzoai/cd/mirror"
)

func main() {
	var (
		cfgPath   = flag.String("config", "", "path to the repository table (JSON)")
		forgeURL  = flag.String("forge", envOr("FORGE_URL", "http://hanzo-git.hanzo.svc"), "forge base URL")
		dryRun    = flag.Bool("dry-run", false, "resolve and report the plan; change nothing")
		timeout   = flag.Duration("timeout", 100*time.Minute, "overall budget")
		forgeTok  = os.Getenv("FORGE_TOKEN")
		githubTok = os.Getenv("GITHUB_TOKEN")
	)
	flag.Parse()

	if *cfgPath == "" {
		fail("-config is required: the repository table is declared state, not a default")
	}
	// Fail on an ABSENT credential rather than proceeding unauthenticated. An empty
	// token produces a 401 that reads like a permissions problem, which is the
	// failure mode that hid a broken npm publish for weeks — every request
	// unauthenticated because a name resolved to "" and nothing said so.
	if forgeTok == "" {
		fail("FORGE_TOKEN is empty — refusing to reconcile unauthenticated")
	}
	if githubTok == "" {
		fail("GITHUB_TOKEN is empty — refusing to resolve names unauthenticated")
	}

	entries, err := load(*cfgPath)
	if err != nil {
		fail(err.Error())
	}
	if len(entries) == 0 {
		fail("the table is empty: that is a config mistake, not an instruction to do nothing")
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	forge := mirror.NewForge(*forgeURL, forgeTok)
	gh := mirror.NewGitHub(githubTok)

	plan, planErrs := mirror.Plan(ctx, gh, entries)
	for _, e := range planErrs {
		fmt.Fprintf(os.Stderr, "!! %v\n", e)
	}
	fmt.Printf("== %d of %d entries resolved\n", len(plan), len(entries))
	for _, p := range plan {
		if p.Moved {
			// Printed loudly: a moved repo is exactly what the predecessor could not see.
			fmt.Printf("   %s  <- %s (renamed or transferred)\n", p.Entry, p.GitHub.FullName)
		}
	}

	if *dryRun {
		// A dry run must exit non-zero on a plan error, or "it printed fine" reads as
		// "the config is good".
		if len(planErrs) > 0 {
			os.Exit(1)
		}
		return
	}

	results := mirror.Reconcile(ctx, forge, gh, mirror.RealGit, *forgeURL, plan)
	for _, r := range results {
		switch r.Outcome {
		case mirror.Moved:
			fmt.Printf("++ %s: %s -> %s\n", r.Entry, r.From[:min(8, len(r.From))], r.To[:min(8, len(r.To))])
		case mirror.Diverged, mirror.Failed:
			fmt.Fprintf(os.Stderr, "!! %s: %s — %s\n", r.Entry, r.Outcome, r.Detail)
		case mirror.Skipped:
			fmt.Printf("-- %s: %s\n", r.Entry, r.Detail)
		}
	}

	counts, needsHuman := mirror.Summary(results)
	fmt.Printf("== %s\n", format(counts))
	if len(planErrs) > 0 || len(needsHuman) > 0 {
		fmt.Fprintf(os.Stderr, "%d need a person\n", len(needsHuman)+len(planErrs))
		os.Exit(1)
	}
}

// load reads the table. Direction defaults to native only when the file says so —
// never implicitly, because the default direction decides which side may overwrite
// the other.
func load(path string) ([]mirror.Entry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read table: %w", err)
	}
	var entries []mirror.Entry
	if err := json.Unmarshal(raw, &entries); err != nil {
		return nil, fmt.Errorf("parse table: %w", err)
	}
	for i, e := range entries {
		if e.Org == "" || e.Name == "" {
			return nil, fmt.Errorf("entry %d: org and name are both required", i)
		}
		if e.Direction == "" {
			return nil, fmt.Errorf("%s: direction is required (%q or %q) — it decides which side may overwrite the other",
				e, mirror.Native, mirror.Mirror)
		}
	}
	return entries, nil
}

func format(counts map[mirror.Outcome]int) string {
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, string(k))
	}
	sort.Strings(keys)
	out := ""
	for _, k := range keys {
		out += fmt.Sprintf("%s=%d  ", k, counts[mirror.Outcome(k)])
	}
	return out
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func fail(msg string) {
	fmt.Fprintf(os.Stderr, "hanzo-mirror: %s\n", msg)
	os.Exit(2)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
