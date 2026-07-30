package mirror

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// Owned is the set of orgs this may touch. An ALLOWLIST, not a denylist, and the
// difference is the whole point.
//
// The script this replaces carried a regex of forbidden names, which meant the
// forbidden names were written down in our source — the exact thing the rule
// against them forbids. It needed that regex because it swept entire orgs, so
// anything could turn up in a listing.
//
// This never enumerates an org: it acts only on the explicitly declared table. The
// one residual risk is a declared repo being TRANSFERRED into an org we do not own,
// which Plan checks after resolution. Naming what we own answers that without
// naming anything we do not, and fails closed on an org nobody thought about —
// including ones that do not exist yet.
var Owned = map[string]bool{
	"hanzoai":    true,
	"hanzobot":   true,
	"hanzo-js":   true,
	"hanzo-ml":   true,
	"hanzo-apps": true,
	"hanzodao":   true,
	"hanzoid":    true,
}

// owned reports whether a full name "org/repo" sits in an org we own.
func owned(full string) bool {
	org, _, ok := strings.Cut(full, "/")
	return ok && Owned[org]
}

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
		if !Owned[e.Org] {
			errs = append(errs, fmt.Errorf("%s: org %q is not one we own", e, e.Org))
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
		// A transfer can move a repo anywhere, including out of our estate entirely.
		// Refusing loudly beats syncing a repo that is no longer ours.
		if !owned(r.FullName) {
			errs = append(errs, fmt.Errorf("%s now resolves to %s, outside the orgs we own — refusing", e, r.FullName))
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
