// Package mirror reconciles which repositories live on git.hanzo.ai and which
// direction each one syncs.
//
// It replaces a Python script (hanzoai/mirrors reconcile.py + sync.py) that drove
// the forge over untyped HTTP. That script worked, and its two worst bugs were
// both consequences of being untyped against a typed API:
//
//   - The direction table was matched against a GitHub org LISTING rather than
//     resolved per name. A listing reports a repo only under its CURRENT name in
//     its CURRENT org, so six entries whose repos had been renamed or transferred
//     were never visited even once — and nothing failed, because a name absent
//     from a listing raises nothing. Here Resolve is a distinct method from List,
//     so the two cannot be confused at a call site.
//   - `mirror` meant two different things — a vestigial git config flag and the
//     forge's actual push refspecs — and reading the wrong one produced a
//     confident wrong conclusion about whether a push could delete history.
//     Direction is one typed value here, with one meaning.
package mirror

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Direction is which way a repository syncs. There are exactly two, and the
// difference is which side is allowed to be authoritative.
type Direction string

const (
	// Native: the forge is canonical and github.com is a push-mirror. Our own code.
	Native Direction = "native"
	// Mirror: an upstream we do not own. Read-only here, so it keeps following.
	Mirror Direction = "mirror"
)

// ErrNotFound is returned when a name resolves to nothing on the remote.
var ErrNotFound = errors.New("mirror: not found")

// Repo is a repository as the remote reports it. The fields are the ones a
// reconcile decision actually reads; the wire carries far more, and decoding only
// what is used keeps a schema change from silently altering behaviour.
type Repo struct {
	FullName string `json:"full_name"`
	Owner    struct {
		Login string `json:"login"`
	} `json:"owner"`
	Name     string `json:"name"`
	Private  bool   `json:"private"`
	Empty    bool   `json:"empty"`
	Mirror   bool   `json:"mirror"`
	CloneURL string `json:"clone_url"`
	Desc     string `json:"description"`
	Size     int    `json:"size"`
	Default  string `json:"default_branch"`
}

// Direction reports the direction this repo is currently configured for.
//
// Derived from the forge's own `mirror` field and nothing else. The git config
// flag of the same name is vestigial — the push path passes explicit refspecs
// that override it — so reading the config would answer a different question than
// the one being asked.
func (r Repo) Direction() Direction {
	if r.Mirror {
		return Mirror
	}
	return Native
}

// Client talks to one forge or to github.com. Two instances, one type: the APIs
// are compatible for everything here, and giving them separate clients would mean
// two places to fix a bug in either.
type Client struct {
	base  string
	token string
	auth  string // "token" for Gitea, "Bearer" for GitHub
	http  *http.Client
}

// NewForge returns a client for a Gitea-compatible forge.
func NewForge(base, token string) *Client {
	return &Client{
		base:  strings.TrimRight(base, "/") + "/v1",
		token: token,
		auth:  "token",
		http:  &http.Client{Timeout: 10 * time.Minute},
	}
}

// NewGitHub returns a client for github.com.
func NewGitHub(token string) *Client {
	return &Client{
		base:  "https://api.github.com",
		token: token,
		auth:  "Bearer",
		http:  &http.Client{Timeout: 2 * time.Minute},
	}
}

func (c *Client) do(ctx context.Context, method, path string, body any) (int, []byte, error) {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, rdr)
	if err != nil {
		return 0, nil, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", c.auth+" "+c.token)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	// Bounded: a forge answering with an unexpected body must not become a memory
	// problem in a controller that runs forever.
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	return resp.StatusCode, raw, err
}

// Resolve answers "what is org/name TODAY", following renames and transfers.
//
// This is deliberately NOT List-and-search. GET /repos/{org}/{name} follows a
// rename and a transfer permanently; a listing does neither, which is how six
// repos went unreconciled and silent. Keeping these as separate methods means a
// caller has to choose, rather than reach for whichever is at hand.
func (c *Client) Resolve(ctx context.Context, org, name string) (*Repo, error) {
	st, raw, err := c.do(ctx, http.MethodGet, "/repos/"+url.PathEscape(org)+"/"+url.PathEscape(name), nil)
	if err != nil {
		return nil, err
	}
	if st == http.StatusNotFound {
		return nil, fmt.Errorf("%w: %s/%s", ErrNotFound, org, name)
	}
	if st != http.StatusOK {
		return nil, fmt.Errorf("mirror: resolve %s/%s: http %d", org, name, st)
	}
	var r Repo
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// Branch is the tip of one branch.
type Branch struct {
	Name   string `json:"name"`
	Commit struct {
		ID  string `json:"id"`
		SHA string `json:"sha"`
	} `json:"commit"`
}

// Tip returns the commit a branch points at, accepting either field name — Gitea
// answers `commit.id` and GitHub `commit.sha`, and a reconcile that only read one
// would silently see an empty tip on the other and treat "unknown" as "equal".
func (c *Client) Tip(ctx context.Context, org, name, branch string) (string, error) {
	st, raw, err := c.do(ctx, http.MethodGet,
		"/repos/"+url.PathEscape(org)+"/"+url.PathEscape(name)+"/branches/"+url.PathEscape(branch), nil)
	if err != nil {
		return "", err
	}
	if st == http.StatusNotFound {
		return "", fmt.Errorf("%w: %s/%s@%s", ErrNotFound, org, name, branch)
	}
	if st != http.StatusOK {
		return "", fmt.Errorf("mirror: tip %s/%s@%s: http %d", org, name, branch, st)
	}
	var b Branch
	if err := json.Unmarshal(raw, &b); err != nil {
		return "", err
	}
	if b.Commit.ID != "" {
		return b.Commit.ID, nil
	}
	return b.Commit.SHA, nil
}
