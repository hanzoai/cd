package scm_provider

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hanzoai/cd/pkg/apis/application/v1alpha1"
)

func githubMockHandler(t *testing.T) func(http.ResponseWriter, *http.Request) {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.RequestURI {
		case "/api/v3/orgs/hanzoai/repos?per_page=100":
			_, err := io.WriteString(w, `[
				{
				  "id": 1296269,
				  "node_id": "MDEwOlJlcG9zaXRvcnkxMjk2MjY5",
				  "name": "argo-cd",
				  "full_name": "hanzoai/cd",
				  "owner": {
					"login": "hanzoai",
					"id": 1,
					"node_id": "MDQ6VXNlcjE=",
					"avatar_url": "https://github.com/images/error/hanzoai_happy.gif",
					"gravatar_id": "",
					"url": "https://api.github.com/users/hanzoai",
					"html_url": "https://github.com/hanzoai",
					"followers_url": "https://api.github.com/users/hanzoai/followers",
					"following_url": "https://api.github.com/users/hanzoai/following{/other_user}",
					"gists_url": "https://api.github.com/users/hanzoai/gists{/gist_id}",
					"starred_url": "https://api.github.com/users/hanzoai/starred{/owner}{/repo}",
					"subscriptions_url": "https://api.github.com/users/hanzoai/subscriptions",
					"organizations_url": "https://api.github.com/users/hanzoai/orgs",
					"repos_url": "https://api.github.com/users/hanzoai/repos",
					"events_url": "https://api.github.com/users/hanzoai/events{/privacy}",
					"received_events_url": "https://api.github.com/users/hanzoai/received_events",
					"type": "User",
					"site_admin": false
				  },
				  "private": false,
				  "html_url": "https://github.com/hanzoai/cd",
				  "description": "This your first repo!",
				  "fork": false,
				  "url": "https://api.github.com/repos/hanzoai/cd",
				  "archive_url": "https://api.github.com/repos/hanzoai/cd/{archive_format}{/ref}",
				  "assignees_url": "https://api.github.com/repos/hanzoai/cd/assignees{/user}",
				  "blobs_url": "https://api.github.com/repos/hanzoai/cd/git/blobs{/sha}",
				  "branches_url": "https://api.github.com/repos/hanzoai/cd/branches{/branch}",
				  "collaborators_url": "https://api.github.com/repos/hanzoai/cd/collaborators{/collaborator}",
				  "comments_url": "https://api.github.com/repos/hanzoai/cd/comments{/number}",
				  "commits_url": "https://api.github.com/repos/hanzoai/cd/commits{/sha}",
				  "compare_url": "https://api.github.com/repos/hanzoai/cd/compare/{base}...{head}",
				  "contents_url": "https://api.github.com/repos/hanzoai/cd/contents/{path}",
				  "contributors_url": "https://api.github.com/repos/hanzoai/cd/contributors",
				  "deployments_url": "https://api.github.com/repos/hanzoai/cd/deployments",
				  "downloads_url": "https://api.github.com/repos/hanzoai/cd/downloads",
				  "events_url": "https://api.github.com/repos/hanzoai/cd/events",
				  "forks_url": "https://api.github.com/repos/hanzoai/cd/forks",
				  "git_commits_url": "https://api.github.com/repos/hanzoai/cd/git/commits{/sha}",
				  "git_refs_url": "https://api.github.com/repos/hanzoai/cd/git/refs{/sha}",
				  "git_tags_url": "https://api.github.com/repos/hanzoai/cd/git/tags{/sha}",
				  "git_url": "git:github.com/hanzoai/cd.git",
				  "issue_comment_url": "https://api.github.com/repos/hanzoai/cd/issues/comments{/number}",
				  "issue_events_url": "https://api.github.com/repos/hanzoai/cd/issues/events{/number}",
				  "issues_url": "https://api.github.com/repos/hanzoai/cd/issues{/number}",
				  "keys_url": "https://api.github.com/repos/hanzoai/cd/keys{/key_id}",
				  "labels_url": "https://api.github.com/repos/hanzoai/cd/labels{/name}",
				  "languages_url": "https://api.github.com/repos/hanzoai/cd/languages",
				  "merges_url": "https://api.github.com/repos/hanzoai/cd/merges",
				  "milestones_url": "https://api.github.com/repos/hanzoai/cd/milestones{/number}",
				  "notifications_url": "https://api.github.com/repos/hanzoai/cd/notifications{?since,all,participating}",
				  "pulls_url": "https://api.github.com/repos/hanzoai/cd/pulls{/number}",
				  "releases_url": "https://api.github.com/repos/hanzoai/cd/releases{/id}",
				  "ssh_url": "git@github.com:hanzoai/cd.git",
				  "stargazers_url": "https://api.github.com/repos/hanzoai/cd/stargazers",
				  "statuses_url": "https://api.github.com/repos/hanzoai/cd/statuses/{sha}",
				  "subscribers_url": "https://api.github.com/repos/hanzoai/cd/subscribers",
				  "subscription_url": "https://api.github.com/repos/hanzoai/cd/subscription",
				  "tags_url": "https://api.github.com/repos/hanzoai/cd/tags",
				  "teams_url": "https://api.github.com/repos/hanzoai/cd/teams",
				  "trees_url": "https://api.github.com/repos/hanzoai/cd/git/trees{/sha}",
				  "clone_url": "https://github.com/hanzoai/cd.git",
				  "mirror_url": "git:git.example.com/hanzoai/cd",
				  "hooks_url": "https://api.github.com/repos/hanzoai/cd/hooks",
				  "svn_url": "https://svn.github.com/hanzoai/cd",
				  "homepage": "https://github.com",
				  "language": null,
				  "forks_count": 9,
				  "stargazers_count": 80,
				  "watchers_count": 80,
				  "size": 108,
				  "default_branch": "master",
				  "open_issues_count": 0,
				  "is_template": false,
				  "topics": [
					"hanzoai",
					"atom",
					"electron",
					"api"
				  ],
				  "has_issues": true,
				  "has_projects": true,
				  "has_wiki": true,
				  "has_pages": false,
				  "has_downloads": true,
				  "archived": false,
				  "disabled": false,
				  "visibility": "public",
				  "pushed_at": "2011-01-26T19:06:43Z",
				  "created_at": "2011-01-26T19:01:12Z",
				  "updated_at": "2011-01-26T19:14:43Z",
				  "permissions": {
					"admin": false,
					"push": false,
					"pull": true
				  },
				  "template_repository": null
				},
				{
				  "id": 1296270,
				  "node_id": "MDEwOlJlcGsddRvcnkxMjk2MjY5",
				  "name": "another-repo",
				  "full_name": "hanzoai/another-repo",
				  "owner": {
					"login": "hanzoai",
					"id": 1,
					"node_id": "MDQ6VXNlcjE=",
					"avatar_url": "https://github.com/images/error/hanzoai_happy.gif",
					"gravatar_id": "",
					"url": "https://api.github.com/users/hanzoai",
					"html_url": "https://github.com/hanzoai",
					"followers_url": "https://api.github.com/users/hanzoai/followers",
					"following_url": "https://api.github.com/users/hanzoai/following{/other_user}",
					"gists_url": "https://api.github.com/users/hanzoai/gists{/gist_id}",
					"starred_url": "https://api.github.com/users/hanzoai/starred{/owner}{/repo}",
					"subscriptions_url": "https://api.github.com/users/hanzoai/subscriptions",
					"organizations_url": "https://api.github.com/users/hanzoai/orgs",
					"repos_url": "https://api.github.com/users/hanzoai/repos",
					"events_url": "https://api.github.com/users/hanzoai/events{/privacy}",
					"received_events_url": "https://api.github.com/users/hanzoai/received_events",
					"type": "User",
					"site_admin": false
				  },
				  "private": false,
				  "html_url": "https://github.com/hanzoai/another-repo",
				  "description": "This your first repo!",
				  "fork": false,
				  "url": "https://api.github.com/repos/hanzoai/another-repo",
				  "archive_url": "https://api.github.com/repos/hanzoai/another-repo/{archive_format}{/ref}",
				  "assignees_url": "https://api.github.com/repos/hanzoai/another-repo/assignees{/user}",
				  "blobs_url": "https://api.github.com/repos/hanzoai/another-repo/git/blobs{/sha}",
				  "branches_url": "https://api.github.com/repos/hanzoai/another-repo/branches{/branch}",
				  "collaborators_url": "https://api.github.com/repos/hanzoai/another-repo/collaborators{/collaborator}",
				  "comments_url": "https://api.github.com/repos/hanzoai/another-repo/comments{/number}",
				  "commits_url": "https://api.github.com/repos/hanzoai/another-repo/commits{/sha}",
				  "compare_url": "https://api.github.com/repos/hanzoai/another-repo/compare/{base}...{head}",
				  "contents_url": "https://api.github.com/repos/hanzoai/another-repo/contents/{path}",
				  "contributors_url": "https://api.github.com/repos/hanzoai/another-repo/contributors",
				  "deployments_url": "https://api.github.com/repos/hanzoai/another-repo/deployments",
				  "downloads_url": "https://api.github.com/repos/hanzoai/another-repo/downloads",
				  "events_url": "https://api.github.com/repos/hanzoai/another-repo/events",
				  "forks_url": "https://api.github.com/repos/hanzoai/another-repo/forks",
				  "git_commits_url": "https://api.github.com/repos/hanzoai/another-repo/git/commits{/sha}",
				  "git_refs_url": "https://api.github.com/repos/hanzoai/another-repo/git/refs{/sha}",
				  "git_tags_url": "https://api.github.com/repos/hanzoai/another-repo/git/tags{/sha}",
				  "git_url": "git:github.com/hanzoai/another-repo.git",
				  "issue_comment_url": "https://api.github.com/repos/hanzoai/another-repo/issues/comments{/number}",
				  "issue_events_url": "https://api.github.com/repos/hanzoai/another-repo/issues/events{/number}",
				  "issues_url": "https://api.github.com/repos/hanzoai/another-repo/issues{/number}",
				  "keys_url": "https://api.github.com/repos/hanzoai/another-repo/keys{/key_id}",
				  "labels_url": "https://api.github.com/repos/hanzoai/another-repo/labels{/name}",
				  "languages_url": "https://api.github.com/repos/hanzoai/another-repo/languages",
				  "merges_url": "https://api.github.com/repos/hanzoai/another-repo/merges",
				  "milestones_url": "https://api.github.com/repos/hanzoai/another-repo/milestones{/number}",
				  "notifications_url": "https://api.github.com/repos/hanzoai/another-repo/notifications{?since,all,participating}",
				  "pulls_url": "https://api.github.com/repos/hanzoai/another-repo/pulls{/number}",
				  "releases_url": "https://api.github.com/repos/hanzoai/another-repo/releases{/id}",
				  "ssh_url": "git@github.com:hanzoai/another-repo.git",
				  "stargazers_url": "https://api.github.com/repos/hanzoai/another-repo/stargazers",
				  "statuses_url": "https://api.github.com/repos/hanzoai/another-repo/statuses/{sha}",
				  "subscribers_url": "https://api.github.com/repos/hanzoai/another-repo/subscribers",
				  "subscription_url": "https://api.github.com/repos/hanzoai/another-repo/subscription",
				  "tags_url": "https://api.github.com/repos/hanzoai/another-repo/tags",
				  "teams_url": "https://api.github.com/repos/hanzoai/another-repo/teams",
				  "trees_url": "https://api.github.com/repos/hanzoai/another-repo/git/trees{/sha}",
				  "clone_url": "https://github.com/hanzoai/another-repo.git",
				  "mirror_url": "git:git.example.com/hanzoai/another-repo",
				  "hooks_url": "https://api.github.com/repos/hanzoai/another-repo/hooks",
				  "svn_url": "https://svn.github.com/hanzoai/another-repo",
				  "homepage": "https://github.com",
				  "language": null,
				  "forks_count": 9,
				  "stargazers_count": 80,
				  "watchers_count": 80,
				  "size": 108,
				  "default_branch": "master",
				  "open_issues_count": 0,
				  "is_template": false,
				  "topics": [
					"hanzoai",
					"atom",
					"electron",
					"api"
				  ],
				  "has_issues": true,
				  "has_projects": true,
				  "has_wiki": true,
				  "has_pages": false,
				  "has_downloads": true,
				  "archived": true,
				  "disabled": false,
				  "visibility": "public",
				  "pushed_at": "2011-01-26T19:06:43Z",
				  "created_at": "2011-01-26T19:01:12Z",
				  "updated_at": "2011-01-26T19:14:43Z",
				  "permissions": {
					"admin": false,
					"push": false,
					"pull": true
				  },
				  "template_repository": null
				}
			  ]`)
			if err != nil {
				t.Fail()
			}
		case "/api/v3/repos/hanzoai/cd/branches?per_page=100":
			_, err := io.WriteString(w, `[
				{
				  "name": "master",
				  "commit": {
					"sha": "c5b97d5ae6c19d5c5df71a34c7fbeeda2479ccbc",
					"url": "https://api.github.com/repos/hanzoai/cd/commits/c5b97d5ae6c19d5c5df71a34c7fbeeda2479ccbc"
				  },
				  "protected": true,
				  "protection": {
					"required_status_checks": {
					  "enforcement_level": "non_admins",
					  "contexts": [
						"ci-test",
						"linter"
					  ]
					}
				  },
				  "protection_url": "https://api.github.com/repos/hanzoai/hello-world/branches/master/protection"
				},
				{
				  "name": "test",
				  "commit": {
					"sha": "80a6e93f16e8093e24091b03c614362df3fb9b92",
					"url": "https://api.github.com/repos/hanzoai/cd/commits/80a6e93f16e8093e24091b03c614362df3fb9b92"
				  },
				  "protected": true,
				  "protection": {
					"required_status_checks": {
					  "enforcement_level": "non_admins",
					  "contexts": [
						"ci-test",
						"linter"
					  ]
					}
				  },
				  "protection_url": "https://api.github.com/repos/hanzoai/hello-world/branches/master/protection"
				}
			  ]
			`)
			if err != nil {
				t.Fail()
			}
		case "/api/v3/repos/hanzoai/another-repo/branches?per_page=100":
			_, err := io.WriteString(w, `[
				{
				  "name": "main",
				  "commit": {
					"sha": "19b016818bc0e0a44ddeaab345838a2a6c97fa67",
					"url": "https://api.github.com/repos/hanzoai/another-repo/commits/19b016818bc0e0a44ddeaab345838a2a6c97fa67"
				  },
				  "protected": true,
				  "protection": {
					"required_status_checks": {
					  "enforcement_level": "non_admins",
					  "contexts": [
						"ci-test",
						"linter"
					  ]
					}
				  },
				  "protection_url": "https://api.github.com/repos/hanzoai/hello-world/branches/master/protection"
					}
			  ]
			`)
			if err != nil {
				t.Fail()
			}
		case "/api/v3/repos/hanzoai/cd/contents/pkg?ref=master":
			_, err := io.WriteString(w, `{
				"type": "file",
				"encoding": "base64",
				"size": 5362,
				"name": "pkg/",
				"path": "pkg/",
				"content": "encoded content ...",
				"sha": "3d21ec53a331a6f037a91c368710b99387d012c1",
				"url": "https://api.github.com/repos/octokit/octokit.rb/contents/README.md",
				"git_url": "https://api.github.com/repos/octokit/octokit.rb/git/blobs/3d21ec53a331a6f037a91c368710b99387d012c1",
				"html_url": "https://github.com/octokit/octokit.rb/blob/master/README.md",
				"download_url": "https://raw.githubusercontent.com/octokit/octokit.rb/master/README.md",
				"_links": {
				  "git": "https://api.github.com/repos/octokit/octokit.rb/git/blobs/3d21ec53a331a6f037a91c368710b99387d012c1",
				  "self": "https://api.github.com/repos/octokit/octokit.rb/contents/README.md",
				  "html": "https://github.com/octokit/octokit.rb/blob/master/README.md"
				}
			  }`)
			if err != nil {
				t.Fail()
			}
		case "/api/v3/repos/hanzoai/cd/branches/master":
			_, err := io.WriteString(w, `{
				"name": "master",
				"commit": {
				  "sha": "c5b97d5ae6c19d5c5df71a34c7fbeeda2479ccbc",
				  "url": "https://api.github.com/repos/octocat/Hello-World/commits/c5b97d5ae6c19d5c5df71a34c7fbeeda2479ccbc"
				},
				"protected": true,
				"protection": {
				  "required_status_checks": {
					"enforcement_level": "non_admins",
					"contexts": [
					  "ci-test",
					  "linter"
					]
				  }
				},
				"protection_url": "https://api.github.com/repos/octocat/hello-world/branches/master/protection"
			  }`)
			if err != nil {
				t.Fail()
			}
		case "/api/v3/repos/hanzoai/cd/branches/test":
			_, err := io.WriteString(w, `{
				"name": "test",
				"commit": {
				  "sha": "80a6e93f16e8093e24091b03c614362df3fb9b92",
				  "url": "https://api.github.com/repos/octocat/Hello-World/commits/80a6e93f16e8093e24091b03c614362df3fb9b92"
				},
				"protected": true,
				"protection": {
				  "required_status_checks": {
					"enforcement_level": "non_admins",
					"contexts": [
					  "ci-test",
					  "linter"
					]
				  }
				},
				"protection_url": "https://api.github.com/repos/octocat/hello-world/branches/test/protection"
			  }`)
			if err != nil {
				t.Fail()
			}
		case "/api/v3/repos/hanzoai/another-repo/branches/main":
			_, err := io.WriteString(w, `{
					"name": "main",
					"commit": {
						"sha": "19b016818bc0e0a44ddeaab345838a2a6c97fa67",
						"url": "https://api.github.com/repos/octocat/Hello-World/commits/c5b97d5ae6c19d5c5df71a34c7fbeeda2479ccbc"
					},
					"protected": true,
					"protection": {
						"required_status_checks": {
						"enforcement_level": "non_admins",
						"contexts": [
							"ci-test",
							"linter"
						]
						}
					},
					"protection_url": "https://api.github.com/repos/octocat/hello-world/branches/master/protection"
					}`)
			if err != nil {
				t.Fail()
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func TestGithubListRepos(t *testing.T) {
	t.Parallel()
	idptr := func(i int64) *int64 {
		return &i
	}
	// Test cases for ListRepos
	cases := []struct {
		name, proto           string
		hasError, allBranches bool
		excludeArchivedRepos  bool
		expectedRepos         []*Repository
		filters               []v1alpha1.SCMProviderGeneratorFilter
	}{
		{
			name:                 "blank protocol",
			allBranches:          true,
			excludeArchivedRepos: false,
			expectedRepos: []*Repository{
				{
					Organization: "hanzoai",
					Repository:   "argo-cd",
					Branch:       "master",
					URL:          "git@github.com:hanzoai/cd.git",
					SHA:          "c5b97d5ae6c19d5c5df71a34c7fbeeda2479ccbc",
					Labels: []string{
						"hanzoai",
						"atom",
						"electron",
						"api",
					},
					RepositoryId: idptr(1296269),
				},
				{
					Organization: "hanzoai",
					Repository:   "argo-cd",
					Branch:       "test",
					URL:          "git@github.com:hanzoai/cd.git",
					SHA:          "80a6e93f16e8093e24091b03c614362df3fb9b92",
					Labels: []string{
						"hanzoai",
						"atom",
						"electron",
						"api",
					},
					RepositoryId: idptr(1296269),
				},
				{
					Organization: "hanzoai",
					Repository:   "another-repo",
					Branch:       "main",
					URL:          "git@github.com:hanzoai/another-repo.git",
					SHA:          "19b016818bc0e0a44ddeaab345838a2a6c97fa67",
					Labels: []string{
						"hanzoai",
						"atom",
						"electron",
						"api",
					},
					RepositoryId: idptr(1296270),
				},
			},
			filters: []v1alpha1.SCMProviderGeneratorFilter{
				{},
			},
		},
		{
			name:                 "ssh protocol",
			proto:                "ssh",
			allBranches:          true,
			excludeArchivedRepos: false,
			expectedRepos: []*Repository{
				{
					Organization: "hanzoai",
					Repository:   "argo-cd",
					Branch:       "master",
					URL:          "git@github.com:hanzoai/cd.git",
					SHA:          "c5b97d5ae6c19d5c5df71a34c7fbeeda2479ccbc",
					Labels: []string{
						"hanzoai",
						"atom",
						"electron",
						"api",
					},
					RepositoryId: idptr(1296269),
				},
				{
					Organization: "hanzoai",
					Repository:   "argo-cd",
					Branch:       "test",
					URL:          "git@github.com:hanzoai/cd.git",
					SHA:          "80a6e93f16e8093e24091b03c614362df3fb9b92",
					Labels: []string{
						"hanzoai",
						"atom",
						"electron",
						"api",
					},
					RepositoryId: idptr(1296269),
				},
				{
					Organization: "hanzoai",
					Repository:   "another-repo",
					Branch:       "main",
					URL:          "git@github.com:hanzoai/another-repo.git",
					SHA:          "19b016818bc0e0a44ddeaab345838a2a6c97fa67",
					Labels: []string{
						"hanzoai",
						"atom",
						"electron",
						"api",
					},
					RepositoryId: idptr(1296270),
				},
			},
			filters: []v1alpha1.SCMProviderGeneratorFilter{
				{},
			},
		},
		{
			name:                 "https protocol",
			proto:                "https",
			allBranches:          true,
			excludeArchivedRepos: false,
			expectedRepos: []*Repository{
				{
					Organization: "hanzoai",
					Repository:   "argo-cd",
					Branch:       "master",
					URL:          "https://github.com/hanzoai/cd.git",
					SHA:          "c5b97d5ae6c19d5c5df71a34c7fbeeda2479ccbc",
					Labels: []string{
						"hanzoai",
						"atom",
						"electron",
						"api",
					},
					RepositoryId: idptr(1296269),
				},
				{
					Organization: "hanzoai",
					Repository:   "argo-cd",
					Branch:       "test",
					URL:          "https://github.com/hanzoai/cd.git",
					SHA:          "80a6e93f16e8093e24091b03c614362df3fb9b92",
					Labels: []string{
						"hanzoai",
						"atom",
						"electron",
						"api",
					},
					RepositoryId: idptr(1296269),
				},
				{
					Organization: "hanzoai",
					Repository:   "another-repo",
					Branch:       "main",
					URL:          "https://github.com/hanzoai/another-repo.git",
					SHA:          "19b016818bc0e0a44ddeaab345838a2a6c97fa67",
					Labels: []string{
						"hanzoai",
						"atom",
						"electron",
						"api",
					},
					RepositoryId: idptr(1296270),
				},
			},
			filters: []v1alpha1.SCMProviderGeneratorFilter{
				{},
			},
		},
		{
			name:                 "other protocol",
			proto:                "other",
			hasError:             true,
			excludeArchivedRepos: false,
			expectedRepos:        []*Repository{},
			filters: []v1alpha1.SCMProviderGeneratorFilter{
				{},
			},
		},
		{
			name:                 "all branches with archived repos",
			allBranches:          true,
			proto:                "ssh",
			excludeArchivedRepos: false,
			expectedRepos: []*Repository{
				{
					Organization: "hanzoai",
					Repository:   "argo-cd",
					Branch:       "master",
					URL:          "git@github.com:hanzoai/cd.git",
					SHA:          "c5b97d5ae6c19d5c5df71a34c7fbeeda2479ccbc",
					Labels: []string{
						"hanzoai",
						"atom",
						"electron",
						"api",
					},
					RepositoryId: idptr(1296269),
				},
				{
					Organization: "hanzoai",
					Repository:   "argo-cd",
					Branch:       "test",
					URL:          "git@github.com:hanzoai/cd.git",
					SHA:          "80a6e93f16e8093e24091b03c614362df3fb9b92",
					Labels: []string{
						"hanzoai",
						"atom",
						"electron",
						"api",
					},
					RepositoryId: idptr(1296269),
				},
				{
					Organization: "hanzoai",
					Repository:   "another-repo",
					Branch:       "main",
					URL:          "git@github.com:hanzoai/another-repo.git",
					SHA:          "19b016818bc0e0a44ddeaab345838a2a6c97fa67",
					Labels: []string{
						"hanzoai",
						"atom",
						"electron",
						"api",
					},
					RepositoryId: idptr(1296270),
				},
			},
			filters: []v1alpha1.SCMProviderGeneratorFilter{
				{},
			},
		},
		{
			name:                 "test repo all branches without archived repos",
			allBranches:          true,
			excludeArchivedRepos: true,
			proto:                "https",
			expectedRepos: []*Repository{
				{
					Organization: "hanzoai",
					Repository:   "argo-cd",
					Branch:       "master",
					URL:          "https://github.com/hanzoai/cd.git",
					SHA:          "c5b97d5ae6c19d5c5df71a34c7fbeeda2479ccbc",
					Labels: []string{
						"hanzoai",
						"atom",
						"electron",
						"api",
					},
					RepositoryId: idptr(1296269),
				},
				{
					Organization: "hanzoai",
					Repository:   "argo-cd",
					Branch:       "test",
					URL:          "https://github.com/hanzoai/cd.git",
					SHA:          "80a6e93f16e8093e24091b03c614362df3fb9b92",
					Labels: []string{
						"hanzoai",
						"atom",
						"electron",
						"api",
					},
					RepositoryId: idptr(1296269),
				},
			},
			filters: []v1alpha1.SCMProviderGeneratorFilter{
				{},
			},
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		githubMockHandler(t)(w, r)
	}))
	t.Cleanup(ts.Close)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			provider, _ := NewGithubProvider("hanzoai", "", ts.URL, c.allBranches, c.excludeArchivedRepos)
			rawRepos, err := ListRepos(t.Context(), provider, c.filters, c.proto)
			if c.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				repos := []*Repository{}
				repos = append(rawRepos, repos...)

				assert.NotEmpty(t, repos)
				assert.Len(t, repos, len(c.expectedRepos))
				assert.ElementsMatch(t, c.expectedRepos, repos)
			}
		})
	}
}

/*
	metricsCtx := &services.MetricsContext{
		AppSetNamespace: "test-ns",
		AppSetName:      "test-appset",
	}

httpClient := services.NewGitHubMetricsClient(metricsCtx)
*/
func TestGithubHasPath(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		githubMockHandler(t)(w, r)
	}))
	defer ts.Close()
	host, _ := NewGithubProvider("hanzoai", "", ts.URL, false, false)
	repo := &Repository{
		Organization: "hanzoai",
		Repository:   "argo-cd",
		Branch:       "master",
	}
	ok, err := host.RepoHasPath(t.Context(), repo, "pkg/")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = host.RepoHasPath(t.Context(), repo, "notathing/")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestGithubGetBranches(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		githubMockHandler(t)(w, r)
	}))
	defer ts.Close()
	host, _ := NewGithubProvider("hanzoai", "", ts.URL, false, false)
	repo := &Repository{
		Organization: "hanzoai",
		Repository:   "argo-cd",
		Branch:       "master",
	}
	repos, err := host.GetBranches(t.Context(), repo)
	if err != nil {
		require.NoError(t, err)
	} else {
		assert.Equal(t, "master", repos[0].Branch)
	}
	// Branch Doesn't exists instead of error will return no error
	repo2 := &Repository{
		Organization: "hanzoai",
		Repository:   "applicationset",
		Branch:       "main",
	}
	_, err = host.GetBranches(t.Context(), repo2)
	require.NoError(t, err)

	// Get all branches
	host.allBranches = true
	repos, err = host.GetBranches(t.Context(), repo)
	if err != nil {
		require.NoError(t, err)
	} else {
		// considering master  branch to  exist.
		assert.Len(t, repos, 2)
	}
}
