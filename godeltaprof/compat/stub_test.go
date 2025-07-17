package compat

import (
	"go/types"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

func checkSignature(t *testing.T, pkg string, name string, expectedSignature string) {
	t.Helper()
	cfg := &packages.Config{
		Mode:  packages.NeedImports | packages.NeedExportFile | packages.NeedTypes | packages.NeedSyntax,
		Tests: true,
	}
	pkgs, err := packages.Load(cfg, pkg)
	require.NoError(t, err)
	found := false
	for _, p := range pkgs {
		if strings.Contains(p.ID, ".test") {
			continue
		}
		f := p.Types.Scope().Lookup(name)
		if f != nil {
			found = true
			ff, ok := f.(*types.Func)
			require.True(t, ok)
			assert.Equal(t, expectedSignature, ff.String())
		}
	}
	assert.Truef(t, found, "function %s %s not found", pkg, name)
}
