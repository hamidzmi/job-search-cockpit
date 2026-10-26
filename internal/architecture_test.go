package internal_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCleanArchitectureBoundaries(t *testing.T) {
	tests := []struct {
		name  string
		dir   string
		check func(t *testing.T, relPath string, f *ast.File)
	}{
		{
			name: "domain imports only stdlib and has no serialization struct tags",
			dir:  "domain",
			check: func(t *testing.T, relPath string, f *ast.File) {
				for _, imp := range f.Imports {
					pkgPath := strings.Trim(imp.Path.Value, `"`)
					if strings.Contains(pkgPath, ".") {
						t.Errorf("%s: domain package must only import stdlib, found %q", relPath, pkgPath)
					}
				}
				ast.Inspect(f, func(n ast.Node) bool {
					field, ok := n.(*ast.Field)
					if !ok || field.Tag == nil {
						return true
					}
					if strings.Contains(field.Tag.Value, "json:") || strings.Contains(field.Tag.Value, "db:") {
						t.Errorf("%s: domain struct field must not carry json/db tags: %s", relPath, field.Tag.Value)
					}
					return true
				})
			},
		},
		{
			name: "usecase never imports adapters or infrastructure packages",
			dir:  "usecase",
			check: func(t *testing.T, relPath string, f *ast.File) {
				banned := []string{
					"database/sql",
					"net/http",
					"os",
					"job-search-cockpit/internal/adapter",
					"job-search-cockpit/internal/app",
				}
				for _, imp := range f.Imports {
					pkgPath := strings.Trim(imp.Path.Value, `"`)
					for _, prefix := range banned {
						if pkgPath == prefix || strings.HasPrefix(pkgPath, prefix+"/") {
							t.Errorf("%s: usecase package must not import %q", relPath, pkgPath)
						}
					}
				}
			},
		},
		{
			name: "inbound and outbound port packages declare only interfaces",
			dir:  filepath.Join("usecase", "port"),
			check: func(t *testing.T, relPath string, f *ast.File) {
				for _, decl := range f.Decls {
					gen, ok := decl.(*ast.GenDecl)
					if !ok || gen.Tok != token.TYPE {
						continue
					}
					for _, spec := range gen.Specs {
						ts, ok := spec.(*ast.TypeSpec)
						if !ok {
							continue
						}
						if _, isIface := ts.Type.(*ast.InterfaceType); !isIface {
							t.Errorf("%s: port package type %q must be an interface, found concrete type", relPath, ts.Name.Name)
						}
					}
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			err := filepath.WalkDir(tc.dir, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
					return nil
				}
				fileAst, parseErr := parser.ParseFile(fset, path, nil, parser.ParseComments)
				if parseErr != nil {
					t.Fatalf("failed to parse %s: %v", path, parseErr)
				}
				tc.check(t, path, fileAst)
				return nil
			})
			if err != nil {
				t.Fatalf("failed walking %s: %v", tc.dir, err)
			}
		})
	}
}
