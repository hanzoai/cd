package pull_request

import (
	"context"
	"net/http"

	"github.com/hanzoai/cd/applicationset/services/github_app_auth"
	"github.com/hanzoai/cd/applicationset/services/internal/github_app"
	appsetutils "github.com/hanzoai/cd/applicationset/utils"
)

func NewGithubAppService(ctx context.Context, g github_app_auth.Authentication, url, owner, repo string, labels []string, optionalHTTPClient ...*http.Client) (PullRequestService, error) {
	httpClient := appsetutils.GetOptionalHTTPClient(optionalHTTPClient...)
	client, err := github_app.Client(ctx, g, url, owner, httpClient)
	if err != nil {
		return nil, err
	}
	return &GithubService{
		client: client,
		owner:  owner,
		repo:   repo,
		labels: labels,
	}, nil
}
