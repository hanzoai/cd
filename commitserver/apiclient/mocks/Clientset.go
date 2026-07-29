package mocks

import (
	"github.com/hanzoai/cd/commitserver/apiclient"
	utilio "github.com/hanzoai/cd/util/io"
)

type Clientset struct {
	CommitServiceClient apiclient.CommitServiceClient
}

func (c *Clientset) NewCommitServerClient() (utilio.Closer, apiclient.CommitServiceClient, error) {
	return utilio.NopCloser, c.CommitServiceClient, nil
}
