package util

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
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
