package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IsNamespaceEnabled(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		namespace         string
		serverNamespace   string
		enabledNamespaces []string
		expectedResult    bool
	}{
		{
			"namespace is empty",
			"cd",
			"cd",
			[]string{},
			true,
		},
		{
			"namespace is explicitly server namespace",
			"cd",
			"cd",
			[]string{},
			true,
		},
		{
			"namespace is allowed namespace",
			"allowed",
			"cd",
			[]string{"allowed"},
			true,
		},
		{
			"namespace matches pattern",
			"test-ns",
			"cd",
			[]string{"test-*"},
			true,
		},
		{
			"namespace is not allowed namespace",
			"disallowed",
			"cd",
			[]string{"allowed"},
			false,
		},
		{
			"match everything but specified word: fail",
			"disallowed",
			"cd",
			[]string{"/^((?!disallowed).)*$/"},
			false,
		},
		{
			"match everything but specified word: pass",
			"allowed",
			"cd",
			[]string{"/^((?!disallowed).)*$/"},
			true,
		},
	}

	for _, tc := range testCases {
		tcc := tc
		t.Run(tcc.name, func(t *testing.T) {
			t.Parallel()
			result := IsNamespaceEnabled(tcc.namespace, tcc.serverNamespace, tcc.enabledNamespaces)
			assert.Equal(t, tcc.expectedResult, result)
		})
	}
}
