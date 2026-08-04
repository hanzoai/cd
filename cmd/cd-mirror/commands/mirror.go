// Package commands reconciles which repositories live on git.hanzo.ai and which
// direction each one syncs.
//
// It replaces hanzoai/mirrors' reconcile.py + sync.py. The behaviour it preserves
// is deliberate and each part was learned the hard way; the behaviour it drops is
// the part that made those lessons necessary:
//
//   - The direction table is CONFIG, read from a file, not a list literal in
//     source. Adding a repo was a code edit, which is not what declared state
//     means. hanzoai/cloud was added that way, and until then it was in neither
//     sync leg: is_mirror=0 so nothing pulled in, absent from the table so nothing
//     pushed in, and its CI therefore ran on a commit four behind for a day while
//     looking merely "red".
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
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/hanzoai/cd/common"
	"github.com/hanzoai/cd/mirror"
)

// NewCommand returns the mirror reconciler.
//
// A subcommand of the one CD binary rather than a program of its own: the fleet
// already builds, pins and ships exactly one image, and a second executable would
// be a second thing to version and a second chance for the two to disagree about
// which repos exist.
func NewCommand() *cobra.Command {
	var (
		table   string
		work    string
		forge   string
		dryRun  bool
		timeout time.Duration
	)

	cmd := &cobra.Command{
		Use:   common.CommandMirror,
		Short: "Reconcile git.hanzo.ai against github.com",
		// The one binary reports its own errors (cmd/main.go), so cobra must not
		// also print them — otherwise every failure is logged twice and a reader
		// counting them sees two problems where there is one.
		SilenceErrors: true,
		SilenceUsage:  true, // past argument parsing, a failure is not a usage error
		RunE: func(_ *cobra.Command, _ []string) error {

			// Fail on an ABSENT credential rather than proceeding unauthenticated. An
			// empty token produces a 401 that reads like a permissions problem, which
			// is the failure mode that hid a broken npm publish for weeks — every
			// request unauthenticated because a name resolved to "" and nothing said
			// so.
			forgeTok, githubTok := os.Getenv("FORGE_TOKEN"), os.Getenv("GITHUB_TOKEN")
			if forgeTok == "" {
				return fmt.Errorf("FORGE_TOKEN is empty — refusing to reconcile unauthenticated")
			}
			if githubTok == "" {
				return fmt.Errorf("GITHUB_TOKEN is empty — refusing to resolve names unauthenticated")
			}

			entries, err := load(table)
			if err != nil {
				return err
			}
			if len(entries) == 0 {
				return fmt.Errorf("%s is empty: that is a config mistake, not an instruction to do nothing", table)
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

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

			if dryRun {
				// A dry run must exit non-zero on a plan error, or "it printed fine"
				// reads as "the config is good".
				if len(planErrs) > 0 {
					return failed(len(planErrs))
				}
				return nil
			}

			results := mirror.Reconcile(ctx, mirror.NewForge(forge, forgeTok), gh, mirror.RealGit, forge, work, plan)
			for _, r := range results {
				switch r.Outcome {
				case mirror.Moved:
					fmt.Printf("++ %s: %s -> %s\n", r.Entry, short(r.From), short(r.To))
				case mirror.Diverged, mirror.Failed:
					fmt.Fprintf(os.Stderr, "!! %s: %s — %s\n", r.Entry, r.Outcome, r.Detail)
				case mirror.Skipped:
					fmt.Printf("-- %s: %s\n", r.Entry, r.Detail)
				case mirror.InSync:
				}
			}

			counts, needsHuman := mirror.Summary(results)
			fmt.Printf("== %s\n", format(counts))
			if n := len(planErrs) + len(needsHuman); n > 0 {
				return failed(n)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&table, "table", common.DefaultMirrorTable, "repository table (JSON)")
	cmd.Flags().StringVar(&work, "work", filepath.Join(os.TempDir(), "cd-mirror"), "where git object stores are kept between runs")
	cmd.Flags().StringVar(&forge, "forge", envOr("FORGE_URL", "http://hanzo-git.hanzo.svc"), "forge base URL")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "resolve and report the plan; change nothing")
	cmd.Flags().DurationVar(&timeout, "timeout", 100*time.Minute, "overall budget")
	return cmd
}

// failed reports how many repos need a person, as an error so cobra sets the exit
// status. The count is the message: a scheduler reads the status, a person reads
// the line above it.
func failed(n int) error { return fmt.Errorf("%d need a person", n) }

// load reads the table. Direction is never defaulted, because the default decides
// which side may overwrite the other — and a table that silently means "native"
// makes every unlisted-direction entry a push into a repo we do not own.
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

func short(s string) string {
	if len(s) > 8 {
		return s[:8]
	}
	return s
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
