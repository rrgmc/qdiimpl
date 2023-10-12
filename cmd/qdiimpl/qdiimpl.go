package main

import (
	"flag"
	"fmt"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/RangelReale/qdiimpl/internal/util"
	. "github.com/dave/jennifer/jen"
	"golang.org/x/tools/go/packages"
)

var (
	typeName         = flag.String("type", "", "type name; must be set")
	typePackageName  = flag.String("type-package", "", "type package path if not the current directory")
	forcePackageName = flag.String("force-package-name", "", "force generated package name")
	samePackage      = flag.Bool("same-package", true, "if false will import source package and qualify the types")
	namePrefix       = flag.String("name-prefix", "QD", "interface name prefix")
	nameSuffix       = flag.String("name-suffix", "", "interface name suffix (default blank)")
	dataType         = flag.String("data-type", "", "add a data member of this type (e.g.: `any`, `package.com/data.XData`)")
	output           = flag.String("output", "", "output file name; default srcdir/<type>_qdii.go")
	buildTags        = flag.String("tags", "", "comma-separated list of build tags to apply")
	doSync           = flag.Bool("sync", true, "use mutex to prevent concurrent accesses")
	overwrite        = flag.Bool("overwrite", false, "overwrite file if exists")
)

// Usage is a replacement usage function for the flags package.
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of qdiimpl:\n")
	fmt.Fprintf(os.Stderr, "\tqdiimpl [flags] -type T [directory]\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("qdiimpl: ")
	flag.Usage = Usage
	flag.Parse()
	if len(*typeName) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	var tags []string
	if len(*buildTags) > 0 {
		tags = strings.Split(*buildTags, ",")
	}

	args := flag.Args()
	if len(args) == 0 {
		// Default: process whole package in current directory.
		args = []string{"."}
	} else if len(args) > 1 {
		log.Println("only one directory must be set")
	}

	err := run(args[0], *typeName, tags)
	if err != nil {
		log.Fatalf("error: %s", err)
	}
}

func run(source, typ string, tags []string) error {
	srcPkg, err := util.PkgInfoFromPath(
		source, *typePackageName, packages.NeedName|packages.NeedSyntax|packages.NeedTypes, tags,
	)
	if err != nil {
		return fmt.Errorf("couldn't load source package: %s", err)
	}

	obj := srcPkg.Types.Scope().Lookup(typ)
	if obj == nil {
		return fmt.Errorf("interface not found: %s", typ)
	}

	if !types.IsInterface(obj.Type()) {
		return fmt.Errorf("%s (%s) is not an interface", typ, obj.Type())
	}

	outputName := *output
	if outputName == "" {
		baseName := fmt.Sprintf("%s_qdii.go", obj.Name())
		outputName = filepath.Join(source, strings.ToLower(baseName))
	}
	if _, err := os.Stat(outputName); err == nil {
		if *overwrite {
			_ = os.Truncate(outputName, 0)
		} else {
			return fmt.Errorf("file '%s' already exists", outputName)
		}
	}

	err = gen(outputName, obj, obj.Type().Underlying().(*types.Interface).Complete())
	if err != nil {
		return err
	}

	return nil
}

func gen(outputName string, obj types.Object, iface *types.Interface) error {
	var f *File
	filePackageName := obj.Pkg().Name()
	if *forcePackageName != "" {
		filePackageName = *forcePackageName
	}

	if *typePackageName != "" || !*samePackage {
		f = NewFile(filePackageName)
	} else {
		f = NewFilePathName(obj.Pkg().Path(), filePackageName)
	}

	f.PackageComment("// Code generated by \"qdiimpl\"; DO NOT EDIT.")

	objName := *namePrefix + obj.Name() + *nameSuffix
	objContext := objName + "Context"
	objOption := objName + "Option"

	objNamedType := obj.Type().(*types.Named) // interfaces are always named types

	var err error
	var codeDataType *Statement
	if *dataType != "" {
		codeDataType, err = util.TypeNameCode(*dataType)
		if err != nil {
			return err
		}
	}

	dataParamName := util.GetUniqueName("Data", func(nameExists string) bool {
		for j := 0; j < iface.NumMethods(); j++ {
			if iface.Method(j).Name() == nameExists {
				return true
			}
		}
		return false
	})

	// default interface generic types
	codeObjectTypes := util.AddTypeParamsList(objNamedType.TypeParams(), false)
	codeObjectTypesWithType := util.AddTypeParamsList(objNamedType.TypeParams(), true)

	// Debug Context
	// # type QDTYPEContext struct {}
	f.Type().Id(objContext).
		StructFunc(func(sgroup *Group) {
			sgroup.Id("ExecCount").Int()
			sgroup.Id("CallerFunc").String()
			sgroup.Id("CallerFile").String()
			sgroup.Id("CallerLine").Int()
			if codeDataType != nil {
				sgroup.Id("Data").Add(codeDataType)
			}
		})
	f.Line()

	// Struct implementation
	// # type debugTYPE struct {}
	f.Type().Id(objName).
		TypesFunc(func(tgroup *Group) {
			for t := 0; t < objNamedType.TypeParams().Len(); t++ {
				tparam := objNamedType.TypeParams().At(t)
				tgroup.Id(tparam.Obj().Name()).Add(util.GetQualCode(tparam.Constraint()))
			}
		}).
		StructFunc(func(group *Group) {
			if codeDataType != nil {
				group.Id(dataParamName).Add(codeDataType)
				group.Line()
			}
			if *doSync {
				group.Id("lock").Qual("sync", "Mutex")
			}
			group.Id("execCount").Map(String()).Int()

			// interface method impls
			for j := 0; j < iface.NumMethods(); j++ {
				mtd := iface.Method(j)
				sig := mtd.Type().(*types.Signature)

				// # implMETHOD  func(qdCtx *QDTYPEContext, METHODPARAMS...) (METHODRESULTS...)
				group.Id("impl" + mtd.Name()).Func().ParamsFunc(func(pgroup *Group) {
					// add debug context parameter
					qdCtxName := util.GetUniqueName("qdCtx", func(nameExists string) bool {
						for k := 0; k < sig.Params().Len(); k++ {
							if sig.Params().At(k).Name() == nameExists {
								return true
							}
						}
						return false
					})
					pgroup.Id(qdCtxName).Op("*").Id(objContext)
					for k := 0; k < sig.Params().Len(); k++ {
						sigParam := sig.Params().At(k)
						pgroup.Id(util.ParamName(k, sigParam)).Add(util.GetQualCode(sigParam.Type()))
					}
				}).ParamsFunc(func(rgroup *Group) {
					for k := 0; k < sig.Results().Len(); k++ {
						sigParam := sig.Results().At(k)
						rgroup.Id(sigParam.Name()).Add(util.GetQualCode(sigParam.Type()))
					}
				})
			}
		})

	// ensure struct implements interface
	if objNamedType.TypeParams().Len() == 0 { // with generics, it is harder to find suitable types
		f.Line()
		// # var _ TYPE = (*debugTYPE)(nil)
		f.Var().Id("_").Add(util.GetQualCode(obj.Type()).TypesFunc(codeObjectTypes)).Op("=").
			Parens(Op("*").Id(objName).TypesFunc(codeObjectTypes)).Parens(Nil())
	}

	f.Line()

	// option type
	// # type QDTYPEOption func(*debugTYPE)
	f.Type().Id(objOption).TypesFunc(codeObjectTypesWithType).Func().Params(Op("*").Id(objName).TypesFunc(codeObjectTypes))

	f.Line()

	// constructor
	// # func NewQDTYPE(options ...QDTYPEOption) *QDTYPE {}
	f.Func().Id("New"+objName).
		TypesFunc(codeObjectTypesWithType).
		Params(
			Id("options").Op("...").Id(objOption).TypesFunc(codeObjectTypes),
		).
		Op("*").Id(objName).TypesFunc(codeObjectTypes).
		Block(
			Id("ret").Op(":=").Op("&").Id(objName).TypesFunc(codeObjectTypes).Values(
				Id("execCount").Op(":").Map(String()).Int().Values(),
			),
			// parse options
			For(List(Id("_"), Id("opt")).Op(":=").Op("range").Id("options")).Block(
				Id("opt").Call(Id("ret")),
			),
			Return(Id("ret")),
		)

	f.Line()

	// interface methods
	for j := 0; j < iface.NumMethods(); j++ {
		f.Line()

		mtd := iface.Method(j)
		sig := mtd.Type().(*types.Signature)

		// # func (d *debugTYPE) METHOD(METHODPARAMS...) (METHODRESULTS...) {}
		f.Commentf("%s implements [%s.%s].", mtd.Name(), util.FormatObjectName(obj), mtd.Name())
		f.Func().Params(Id("d").Op("*").Id(objName).TypesFunc(codeObjectTypes)).Id(mtd.Name()).ParamsFunc(func(pgroup *Group) {
			for k := 0; k < sig.Params().Len(); k++ {
				sigParam := sig.Params().At(k)
				pgroup.Id(util.ParamName(k, sigParam)).Add(util.GetQualCode(sigParam.Type()))
			}
		}).ParamsFunc(func(rgroup *Group) {
			for k := 0; k < sig.Results().Len(); k++ {
				sigParam := sig.Results().At(k)
				rgroup.Id(sigParam.Name()).Add(util.GetQualCode(sigParam.Type()))
			}
		}).Block(
			Do(func(s *Statement) {
				call := Id("d").Dot("impl" + mtd.Name()).CallFunc(func(cgroup *Group) {
					cgroup.Id("d").Dot("createContext").Call(
						Lit(mtd.Name()), Id("d").Dot("impl"+mtd.Name()).Op("==").Nil(),
					)
					for k := 0; k < sig.Params().Len(); k++ {
						sigParam := sig.Params().At(k)
						cgroup.Id(util.ParamName(k, sigParam))
					}
				})
				if sig.Results().Len() == 0 {
					s.Add(call)
				} else {
					s.Add(Return(call))
				}
			}),
		)
	}

	// helper methods
	f.Line()

	// getCallerFuncName
	// # func (d *debugTYPE) getCallerFuncName(skip int) (funcName string, file string, line int) {}
	f.Func().Params(Id("d").Op("*").Id(objName).TypesFunc(codeObjectTypes)).
		Id("getCallerFuncName").
		Params(
			Id("skip").Int(),
		).
		Params(
			Id("funcName").String(),
			Id("file").String(),
			Id("line").Int()).
		Block(
			List(Id("counter"), Id("file"), Id("line"), Id("success")).
				Op(":=").Qual("runtime", "Caller").Call(Id("skip")),
			If(Op("!").Id("success")).Block(
				Panic(Lit("runtime.Caller failed")),
			),
			Return(
				Qual("runtime", "FuncForPC").Call(Id("counter")).Dot("Name").Call(),
				Id("file"),
				Id("line"),
			),
		)

	f.Line()

	// checkCallMethod
	// # func (d *debugTYPE) checkCallMethod(methodName string, implIsNil bool) (count int) {}
	f.Func().Params(Id("d").Op("*").Id(objName).TypesFunc(codeObjectTypes)).
		Id("checkCallMethod").
		Params(
			Id("methodName").String(),
			Id("implIsNil").Bool(),
		).
		Params(
			Id("count").Int(),
		).
		BlockFunc(func(bgroup *Group) {
			bgroup.If(Id("implIsNil")).Block(
				Panic(Qual("fmt", "Errorf").
					Call(Lit(fmt.Sprintf("[%s] method '%%s' not implemented", objName)), Id("methodName"))),
			)
			bgroup.Id("d").Dot("lock").Dot("Lock").Call()
			bgroup.Defer().Id("d").Dot("lock").Dot("Unlock").Call()

			bgroup.Id("d").Dot("execCount").Index(Id("methodName")).Op("++")
			bgroup.Return(Id("d").Dot("execCount").Index(Id("methodName")))
		})

	f.Line()

	// createContext
	// # func (d *debugTYPE[T, X]) createContext(methodName string, implIsNil bool) *QDTYPEContext {}
	f.Func().Params(Id("d").Op("*").Id(objName).TypesFunc(codeObjectTypes)).
		Id("createContext").
		Params(
			Id("methodName").String(),
			Id("implIsNil").Bool(),
		).
		Params(
			Op("*").Id(objContext),
		).
		Block(
			List(Id("callerFunc"), Id("callerFile"), Id("callerLine")).Op(":=").
				Id("d").Dot("getCallerFuncName").Call(Lit(3)),
			Return(
				Op("&").Id(objContext).ValuesFunc(func(vgroup *Group) {
					vgroup.Id("ExecCount").Op(":").Id("d").
						Dot("checkCallMethod").Call(Id("methodName"), Id("implIsNil"))
					vgroup.Id("CallerFunc").Op(":").Id("callerFunc")
					vgroup.Id("CallerFile").Op(":").Id("callerFile")
					vgroup.Id("CallerLine").Op(":").Id("callerLine")
					if codeDataType != nil {
						vgroup.Id("Data").Op(":").Id("d").Dot(dataParamName)
					}
				})),
		)

	f.Line()
	f.Comment("Options")
	f.Line()

	if codeDataType != nil {
		// WithData option
		// # func WithQDTYPEData(data any) QDTYPEOption {}
		f.Func().Id("With" + objName + dataParamName).TypesFunc(codeObjectTypesWithType).Params(
			Id("data").Add(codeDataType),
		).Params(Id(objOption).TypesFunc(codeObjectTypes)).Block(
			Return(Func().Params(Id("d").Op("*").Id(objName).TypesFunc(codeObjectTypes)).Block(
				Id("d").Dot(dataParamName).Op("=").Id("data"),
			)),
		)
	}

	// method options
	for j := 0; j < iface.NumMethods(); j++ {
		mtd := iface.Method(j)
		sig := mtd.Type().(*types.Signature)

		f.Line()

		// # func WithQDTYPEMETHOD(implMETHOD func(qdCtx *QDTYPEContext, METHODPARAMS...) (METHODRESULTS...)) QDTYPEOption {}
		f.Commentf("With%s%s implements [%s.%s].", objName, mtd.Name(), util.FormatObjectName(obj), mtd.Name())
		f.Func().Id("With" + objName + mtd.Name()).TypesFunc(codeObjectTypesWithType).Params(
			Id("impl" + mtd.Name()).Func().ParamsFunc(func(pgroup *Group) {
				// add debug context parameter
				qdCtxName := util.GetUniqueName("qdCtx", func(nameExists string) bool {
					for k := 0; k < sig.Params().Len(); k++ {
						if sig.Params().At(k).Name() == nameExists {
							return true
						}
					}
					return false
				})
				pgroup.Id(qdCtxName).Op("*").Id(objContext)
				for k := 0; k < sig.Params().Len(); k++ {
					sigParam := sig.Params().At(k)
					pgroup.Id(util.ParamName(k, sigParam)).Add(util.GetQualCode(sigParam.Type()))
				}
			}).ParamsFunc(func(rgroup *Group) {
				for k := 0; k < sig.Results().Len(); k++ {
					sigParam := sig.Results().At(k)
					rgroup.Id(sigParam.Name()).Add(util.GetQualCode(sigParam.Type()))
				}
			}),
		).Params(Id(objOption).TypesFunc(codeObjectTypes)).Block(
			Return(Func().Params(Id("d").Op("*").Id(objName).TypesFunc(codeObjectTypes)).Block(
				Id("d").Dot("impl" + mtd.Name()).Op("=").Id("impl" + mtd.Name()),
			)),
		)
	}

	// Write to file.
	fmt.Printf("Writing file %s...", outputName)

	outFile, err := os.Create(outputName)
	if err != nil {
		return err
	}
	defer outFile.Close()

	err = f.Render(outFile)
	if err != nil {
		return err
	}

	// output
	// fmt.Printf("%#v", f)

	return nil
}
