// Package qdiwrap is a POC alternative to the qdiimpl generator that builds its output from
// text/template templates driven by github.com/hexdigest/gowrap instead of the dave/jennifer
// AST DSL.
//
// It supports only the two core features of qdiimpl: overriding each interface method
// individually with a WithMETHOD functional option, and setting a fallback implementation to
// delegate to when a method was not overridden. There is no qdii context, no exec counts, no
// caller introspection and no NotSupported() retry chain.
package qdiwrap

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hexdigest/gowrap/generator"
	"github.com/rrgmc/qdiimpl/internal/util"
	"golang.org/x/tools/go/packages"
)

var (
	typeName         = flag.String("type", "", "type name; must be set")
	typePackageName  = flag.String("type-package", "", "type package path if not the current directory")
	forcePackageName = flag.String("force-package-name", "", "force generated package name")
	namePrefix       = flag.String("name-prefix", "", "interface name prefix")
	nameSuffix       = flag.String("name-suffix", "", "interface name suffix")
	optionNamePrefix = flag.String("option-name-prefix", "", "option name prefix (WithXXXMethod)")
	output           = flag.String("output", "", "output file name; default srcdir/<type>_qdiw.go")
	overwrite        = flag.Bool("overwrite", false, "overwrite file if exists")
)

// Usage is a replacement usage function for the flags package.
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of qdiwrap:\n")
	fmt.Fprintf(os.Stderr, "\tqdiwrap [flags] -type T [directory]\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func Main() {
	log.SetFlags(0)
	log.SetPrefix("qdiwrap: ")
	flag.Usage = Usage
	flag.Parse()
	if len(*typeName) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	args := flag.Args()
	if len(args) == 0 {
		// Default: process the package in the current directory.
		args = []string{"."}
	} else if len(args) > 1 {
		log.Println("only one directory must be set")
	}

	err := run(args[0], *typeName)
	if err != nil {
		log.Fatalf("error: %s", err)
	}
}

func run(source, typ string) error {
	// gowrap loads and parses the interface itself, so unlike qdiimpl there is no
	// packages.Load / scope lookup / IsInterface step here.
	sourcePackage := *typePackageName
	if sourcePackage == "" {
		sourcePackage = source
		if !filepath.IsAbs(sourcePackage) && !strings.HasPrefix(sourcePackage, ".") {
			sourcePackage = "./" + sourcePackage
		}
	}

	outputName := *output
	if outputName == "" {
		outputName = filepath.Join(source, strings.ToLower(typ)+"_qdiw.go")
	}

	return gen(sourcePackage, outputName, typ)
}

func gen(sourcePackage, outputName, typ string) error {
	// gowrap parses the whole source directory with parser.DeclarationErrors and loads the
	// destination directory as a package, both of which choke on a stale generated file.
	// Remove it up front rather than truncating.
	if _, err := os.Stat(outputName); err == nil {
		if !*overwrite {
			return fmt.Errorf("file '%s' already exists", outputName)
		}
		if err := os.Remove(outputName); err != nil {
			return fmt.Errorf("could not remove existing output file: %w", err)
		}
	}

	// The struct is exported, so an initial-lowercase prefix is upper-cased; a prefix that
	// already starts upper-case is left alone.
	prefix := *namePrefix
	if prefix != "" {
		prefix = util.InitialToUpper(prefix)
	}
	structName := prefix + typ + *nameSuffix

	// text/template renders a missing map key as "<no value>", so every key the templates
	// reference must be present.
	vars := map[string]any{
		"StructName":   structName,
		"OptionName":   structName + "Option",
		"CtorName":     "New" + structName,
		"OptionPrefix": *optionNamePrefix,
		"PackageName":  *forcePackageName,
	}

	// gowrap would otherwise alias the source package as "_sourceIo", which renders as an
	// unresolvable doc link in the generated comments. Use the package's own name instead.
	// The alias is ignored entirely when generating into the source package itself.
	srcPkg, err := util.PkgInfoFromPath(".", sourcePackage, packages.NeedName, nil)
	if err != nil {
		return fmt.Errorf("couldn't load source package: %w", err)
	}

	g, err := generator.NewGenerator(generator.Options{
		InterfaceName:      typ,
		SourcePackage:      sourcePackage,
		SourcePackageAlias: srcPkg.Name,
		OutputFile:         outputName,
		HeaderTemplate:     headerTemplate,
		BodyTemplate:       bodyTemplate,
		Vars:               vars,
	})
	if err != nil {
		return err
	}

	// Generate already runs goimports over the result, so the output is formatted and its
	// imports are resolved. On failure it wraps the whole unformatted buffer in the error.
	var buf bytes.Buffer
	if err := g.Generate(&buf); err != nil {
		return err
	}

	fmt.Printf("Writing file %s...\n", outputName)
	return os.WriteFile(outputName, buf.Bytes(), 0o644)
}
