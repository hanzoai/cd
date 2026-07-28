package generator

import "github.com/hanzoai/deploy/hack/gen-resources/util"

var labels = map[string]string{
	"app.kubernetes.io/generated-by": "cd-generator",
}

type Generator interface {
	Generate(opts *util.GenerateOpts) error
	Clean(opts *util.GenerateOpts) error
}
