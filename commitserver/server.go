package commitserver

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/hanzoai/deploy/commitserver/apiclient"
	"github.com/hanzoai/deploy/commitserver/commit"
	"github.com/hanzoai/deploy/commitserver/metrics"
	versionpkg "github.com/hanzoai/deploy/pkg/apiclient/version"
	"github.com/hanzoai/deploy/server/version"
	"github.com/hanzoai/deploy/util/git"
)

// ArgoCDCommitServer is the server that handles commit requests.
type ArgoCDCommitServer struct {
	commitService *commit.Service
}

// NewServer returns a new instance of the commit server.
func NewServer(gitCredsStore git.CredsStore, metricsServer *metrics.Server) *ArgoCDCommitServer {
	return &ArgoCDCommitServer{commitService: commit.NewService(gitCredsStore, metricsServer)}
}

// CreateGRPC creates a new gRPC server.
func (a *ArgoCDCommitServer) CreateGRPC() *grpc.Server {
	server := grpc.NewServer(grpc.MaxRecvMsgSize(apiclient.MaxGRPCMessageSize))
	versionpkg.RegisterVersionServiceServer(server, version.NewServer(nil, func() (bool, error) {
		return true, nil
	}))
	apiclient.RegisterCommitServiceServer(server, a.commitService)

	healthService := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthService)

	return server
}
