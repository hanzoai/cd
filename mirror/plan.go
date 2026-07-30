package mirror

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Poison is absolute: nothing matching it is created, listed, or named in output.
// Carried over verbatim from the script it replaces — the constraint is a hard one
// and does not weaken because the language changed.
var Poison = regexp.MustCompile(`(?i)liquid|satschel|simplici|equitytable|onyx-plus|vccross|securegate|morningline`)

// Entry is one declared repository: where it lives here, and which way it syncs.
//
// This is data, not code. Adding a repo used to be a source edit to a list literal
// inside the script; it is now a line of config, which is what "declared state"
// has to mean if the word is doing any work.
type Entry struct {
	Org       string    `json:"org" yaml:"org"`
	Name      string    `json:"name" yaml:"name"`
	Direction Direction `json:"direction" yaml:"direction"`
}

func (e Entry) String() string { return e.Org + "/" + e.Name }

// Planned is one entry resolved against the remote: what we declared, and what the
// remote says that name is today.
type Planned struct {
	Entry
	// GitHub is the repo the entry resolves to NOW. Its full name differs from the
	// entry whenever GitHub has renamed or moved the repo, which is the case the
	// listing-based predecessor could not see at all.
	GitHub *Repo
	// Moved is true when the entry's name is no longer the remote's name.
	Moved bool
}

// Plan resolves every declared entry against GitHub, in the order it will be acted
// on: smallest first, so a budget spent on one enormous repo cannot starve every
// small one behind it.
//
// Two entries resolving to ONE GitHub repo is a config bug, not a merge: both would
// push-mirror to the same remote and overwrite each other. The first wins and the
// collision is reported, because silently dropping one is how a repo stops syncing
// without anyone being told.
func Plan(ctx context.Context, gh *Client, entries []Entry) ([]Planned, []error) {
	var out []Planned
	var errs []error
	claimed := map[string]Entry{}

	for _, e := range entries {
		if Poison.MatchString(e.Org) || Poison.MatchString(e.Name) {
			continue
		}
		if e.Direction != Native && e.Direction != Mirror {
			errs = append(errs, fmt.Errorf("%s: direction %q is neither %q nor %q", e, e.Direction, Native, Mirror))
			continue
		}
		r, err := gh.Resolve(ctx, e.Org, e.Name)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", e, err))
			continue
		}
		// A transfer can move a repo anywhere, including somewhere the guard forbids.
		if Poison.MatchString(r.FullName) {
			continue
		}
		if prev, dup := claimed[r.FullName]; dup {
			errs = append(errs, fmt.Errorf("%s and %s both resolve to %s — one repo, one entry", e, prev, r.FullName))
			continue
		}
		claimed[r.FullName] = e
		out = append(out, Planned{Entry: e, GitHub: r, Moved: !strings.EqualFold(r.FullName, e.String())})
	}

	sort.SliceStable(out, func(i, j int) bool { return out[i].GitHub.Size < out[j].GitHub.Size })
	return out, errs
}
