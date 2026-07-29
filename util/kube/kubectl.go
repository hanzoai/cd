package kube

import (
	"os"

	"github.com/go-logr/logr"
	"k8s.io/client-go/rest"

	"github.com/hanzoai/cd/util/log"

	"github.com/hanzoai/cd/gitops-engine/pkg/diff"
	"github.com/hanzoai/cd/gitops-engine/pkg/utils/kube"
	"github.com/hanzoai/cd/gitops-engine/pkg/utils/tracing"
)

var (
	tracer tracing.Tracer = &tracing.NopTracer{}
	logger logr.Logger    = log.NewLogrusLogger(log.NewWithCurrentConfig())
)

func init() {
	if os.Getenv("CD_TRACING_ENABLED") == "1" {
		tracer = tracing.NewLoggingTracer(logger)
	}
}

func NewKubectl() kube.Kubectl {
	return &kube.KubectlCmd{Tracer: tracer, Log: logger}
}

func ManageServerSideDiffDryRuns(config *rest.Config, onKubectlRun kube.OnKubectlRunFunc) (diff.KubeApplier, func(), error) {
	k := &kube.KubectlCmd{
		Log:          logger,
		Tracer:       tracer,
		OnKubectlRun: onKubectlRun,
	}
	return k.ManageServerSideDiffDryRuns(config)
}
