package mocks

import (
	"github.com/hanzoai/cd/reposerver/apiclient"
	utilio "github.com/hanzoai/cd/util/io"
)

type Clientset struct {
	RepoServerServiceClient apiclient.RepoServerServiceClient
}

func (c *Clientset) NewRepoServerClient() (utilio.Closer, apiclient.RepoServerServiceClient, error) {
	return utilio.NopCloser, c.RepoServerServiceClient, nil
}
