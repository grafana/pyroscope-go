package compat

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errNoExportFile = errors.New("no export file")

var exportFileCache = make(map[string]string)

func goListExport(pkg string) (string, error) {
	if v, ok := exportFileCache[pkg]; ok {
		return v, nil
	}
	out, err := exec.Command("go", "list", "-export", "-f", "{{.Export}}", pkg).Output()
	if err != nil {
		return "", fmt.Errorf("go list -export %s: %w", pkg, err)
	}
	path := strings.TrimSpace(string(out))
	if path == "" {
		return "", fmt.Errorf("%s: %w", pkg, errNoExportFile)
	}
	exportFileCache[pkg] = path

	return path, nil
}

type goListJSON struct {
	Dir     string   `json:"Dir"`
	GoFiles []string `json:"GoFiles"`
}

func checkSignature(t *testing.T, pkg string, name string, expectedSignature string) {
	t.Helper()

	// Use go list -json to find source files for the package.
	// We need source (not just export data) because the functions
	// we're checking are unexported.
	out, err := exec.Command("go", "list", "-json", pkg).Output()
	require.NoError(t, err, "go list -json failed for %s", pkg)

	var pkgInfo goListJSON
	require.NoError(t, json.Unmarshal(out, &pkgInfo))

	fset := token.NewFileSet()
	files := make([]*ast.File, 0, len(pkgInfo.GoFiles))
	for _, fname := range pkgInfo.GoFiles {
		f, err := parser.ParseFile(fset, filepath.Join(pkgInfo.Dir, fname), nil, 0)
		require.NoError(t, err)
		files = append(files, f)
	}

	// Type-check the package from source.
	// Dependencies are loaded from export data via go list -export.
	conf := types.Config{
		Importer: importer.ForCompiler(fset, "gc", func(path string) (io.ReadCloser, error) {
			exportFile, err := goListExport(path)
			if err != nil {
				return nil, err
			}

			return os.Open(exportFile) //nolint:gosec // export file path from go list is trusted
		}),
	}
	p, err := conf.Check(pkg, fset, files, nil)
	require.NoError(t, err, "type-check failed for %s", pkg)

	f := p.Scope().Lookup(name)
	require.NotNilf(t, f, "function %s not found in %s", name, pkg)
	ff, ok := f.(*types.Func)
	require.True(t, ok)
	assert.Equal(t, expectedSignature, ff.String())
}
