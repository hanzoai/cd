package mocks

import (
	"github.com/hanzoai/deploy/commitserver/apiclient"
	utilio "github.com/hanzoai/deploy/util/io"
)

type Clientset struct {
	CommitServiceClient apiclient.CommitServiceClient
}

func (c *Clientset) NewCommitServerClient() (utilio.Closer, apiclient.CommitServiceClient, error) {
	return utilio.NopCloser, c.CommitServiceClient, nil
}
