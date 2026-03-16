package util

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"

	"github.com/dave/jennifer/jen"
)

// pkgName extracts the package name from src, which must contain a package header.
func pkgName(src string) string {
	const kw = "package "
	if i := strings.Index(src, kw); i >= 0 {
		after := src[i+len(kw):]
		n := len(after)
		if i := strings.IndexAny(after, "\n\t ;/"); i >= 0 {
			n = i
		}
		return after[:n]
	}
	panic("missing package header: " + src)
}

func mustParse(fset *token.FileSet, src string) *ast.File {
	f, err := parser.ParseFile(fset, pkgName(src), src, parser.ParseComments)
	if err != nil {
		panic(err) // so we don't need to pass *testing.T
	}
	return f
}

func typecheck(src string, conf *types.Config, info *types.Info) (*types.Package, error) {
	// TODO(adonovan): plumb this from caller.
	fset := token.NewFileSet()
	f := mustParse(fset, src)
	if conf == nil {
		conf = &types.Config{
			Error:    func(err error) {}, // collect all errors
			Importer: importer.Default(),
		}
	}
	return conf.Check(f.Name.Name, fset, []*ast.File{f}, info)
}

func mustTypecheck(src string, conf *types.Config, info *types.Info) *types.Package {
	pkg, err := typecheck(src, conf, info)
	if err != nil {
		panic(err) // so we don't need to pass *testing.T
	}
	return pkg
}

type tc struct {
	// path
	path string
	// description for locating the test case
	desc string
	// code to generate
	code   jen.Code
	codeFn func() jen.Code
	// expected generated source
	expect string
	// expected imports
	expectImports map[string]string
}

func runTestCases(t *testing.T, cases []tc) {
	for i, c := range cases {
		onlyTest := ""
		if onlyTest != "" && c.desc != onlyTest {
			continue
		}
		var code jen.Code
		if c.code != nil && c.codeFn != nil {
			t.Fatal("code OR codeFn must be set")
		}
		if c.code != nil {
			code = c.code
		} else if c.codeFn != nil {
			code = c.codeFn()
		} else {
			t.Fatal("code or codeFn must be set")
		}

		rendered := fmt.Sprintf("%#v", code)

		expected, err := format.Source([]byte(c.expect))
		if err != nil {
			panic(fmt.Sprintf("Error formatting expected source in test case %d. Description: %s\nError:\n%s", i, c.desc, err))
		}

		if strings.TrimSpace(string(rendered)) != strings.TrimSpace(string(expected)) {
			t.Errorf("Test case %d failed. Description: %s\nExpected:\n%s\nOutput:\n%s", i, c.desc, expected, rendered)
		}

		// if c.expectImports != nil {
		//	f := FromContext(ctx)
		//	if !reflect.DeepEqual(f.Imports, c.expectImports) {
		//		t.Errorf("Test case %d failed. Description: %s\nImports expected:\n%s\nOutput:\n%s", i, c.desc, c.expectImports, f.Imports)
		//	}
		// }
	}
}
