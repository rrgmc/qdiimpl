package cmd

import (
	"flag"
	"fmt"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	. "github.com/dave/jennifer/jen"
	"github.com/rrgmc/qdiimpl/internal/util"
	"golang.org/x/tools/go/packages"
)

var (
	typeName             = flag.String("type", "", "type name; must be set")
	allInterfaces        = flag.Bool("all-interfaces", false, "if true will ignore -type and output all interfaces")
	skipInterfaces       = flag.String("skip-interfaces", "", "list of interface names to skip (if -all-interfaces)")
	typePackageName      = flag.String("type-package", "", "type package path if not the current directory")
	forcePackageName     = flag.String("force-package-name", "", "force generated package name")
	samePackage          = flag.Bool("same-package", true, "if false will import source package and qualify the types")
	namePrefix           = flag.String("name-prefix", "", "interface name prefix")
	nameSuffix           = flag.String("name-suffix", "", "interface name suffix")
	optionNamePrefix     = flag.String("option-name-prefix", "", "option name prefix (WithXXXMethod)")
	optionNamePrefixSelf = flag.Bool("option-name-prefix-self", false, "use self name as option name prefix (WithXXXMethod)")
	dataType             = flag.String("data-type", "", "add a data member of this type (e.g.: `any`, `package.com/data.XData`)")
	dataTypeSelf         = flag.Bool("data-type-self", false, "add a data member of the self type with `Data` suffix")
	output               = flag.String("output", "", "output file name; default srcdir/<type>_qdii.go")
	buildTags            = flag.String("tags", "", "comma-separated list of build tags to apply")
	doSync               = flag.Bool("sync", true, "use mutex to prevent concurrent accesses")
	callLogger           = flag.Bool("call-logger", false, "add call logger")
	exportType           = flag.Bool("export-type", false, "whether to export the generated type (default false)")
	overwrite            = flag.Bool("overwrite", false, "overwrite file if exists")
)

// Usage is a replacement usage function for the flags package.
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of qdiimpl:\n")
	fmt.Fprintf(os.Stderr, "\tqdiimpl [flags] -type T [directory]\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func Main() {
	log.SetFlags(0)
	log.SetPrefix("qdiimpl: ")
	flag.Usage = Usage
	flag.Parse()
	if !*allInterfaces && len(*typeName) == 0 {
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

	if *allInterfaces {
		return runAllInterfaces(source, srcPkg, tags)
	}

	obj := srcPkg.Types.Scope().Lookup(typ)
	if obj == nil {
		return fmt.Errorf("interface not found: %s", typ)
	}

	return runType(source, typ, obj, tags)
}

func runAllInterfaces(source string, pkg *packages.Package, tags []string) error {
	skipList := strings.Split(*skipInterfaces, ",")

	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if obj == nil {
			continue
		}

		if !obj.Exported() {
			continue
		}

		var typ *types.Named

		ttyp := obj.Type()

		typ, ok := ttyp.(*types.Named)
		if !ok {
			continue
		}

		if !typ.Obj().Exported() || typ.Obj().Pkg() == nil {
			continue
		}

		_, ok = typ.Underlying().(*types.Interface)
		if !ok {
			continue
		}

		if slices.Contains(skipList, typ.Obj().Name()) {
			continue
		}

		err := runType(source, typ.Obj().Name(), obj, tags)
		if err != nil {
			return fmt.Errorf("error processing '%s': %w", obj.Name(), err)
		}
	}
	return nil
}

func runType(source string, typ string, obj types.Object, tags []string) error {
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

	return gen(outputName, obj, obj.Type().Underlying().(*types.Interface).Complete())
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

	var prefix string
	// if first already uppercase use the passed string, otherwise uppercase all
	if util.InitialIsUpper(*namePrefix) {
		prefix = *namePrefix
	} else {
		prefix = strings.ToUpper(*namePrefix)
	}
	uprefix := prefix
	if !*exportType {
		// if first already lowercase use the passed string, otherwise lowercase all
		if util.InitialIsLower(*namePrefix) {
			prefix = *namePrefix
		} else {
			prefix = strings.ToLower(*namePrefix)
		}
	}

	objName := prefix + obj.Name() + *nameSuffix
	objNameExported := uprefix + obj.Name() + *nameSuffix
	objContext := objNameExported + "Context"
	objOption := objNameExported + "Option"
	objOptionPrefix := *optionNamePrefix
	if *optionNamePrefixSelf {
		objOptionPrefix = obj.Name()
	}

	objNamedType := obj.Type().(*types.Named) // interfaces are always named types

	genDataType := *dataType
	if *dataTypeSelf {
		genDataType = fmt.Sprintf("%sData", obj.Name())
	}
	var err error
	var codeDataType *Statement
	if genDataType != "" {
		codeDataType, err = util.TypeNameCode(genDataType)
		if err != nil {
			return err
		}
	}

	dataParamName := util.GetUniqueName("Data", func(nameExists string) bool {
		for mtd := range iface.Methods() {
			if mtd.Name() == nameExists {
				return true
			}
		}
		return false
	})
	fallbackParamName := util.GetUniqueName("fallback", func(nameExists string) bool {
		for mtd := range iface.Methods() {
			if mtd.Name() == nameExists {
				return true
			}
		}
		return false
	})
	onMethodNotImplementedParamName := util.GetUniqueName("onMethodNotImplemented", func(nameExists string) bool {
		for mtd := range iface.Methods() {
			if mtd.Name() == nameExists {
				return true
			}
		}
		return false
	})
	onCallLoggerParamName := util.GetUniqueName("onCallLogger", func(nameExists string) bool {
		for mtd := range iface.Methods() {
			if mtd.Name() == nameExists {
				return true
			}
		}
		return false
	})

	// default interface generic types
	codeObjectTypes := util.AddTypeParamsList(objNamedType.TypeParams(), false)
	codeObjectTypesWithType := util.AddTypeParamsList(objNamedType.TypeParams(), true)

	// QD Context
	// # type TYPEContext struct {}
	f.Type().Id(objContext).
		StructFunc(func(sgroup *Group) {
			if codeDataType != nil {
				sgroup.Id("Data").Add(codeDataType)
				sgroup.Line()
			}
			sgroup.Id("methodName").String()
			sgroup.Id("execCount").Int()
			sgroup.Id("callerFunc").String()
			sgroup.Id("callerFile").String()
			sgroup.Id("callerLine").Int()
			sgroup.Id("isNotSupported").Bool()
		})
	f.Line()

	// # func (c *TYPEContext) NotSupported()
	f.Comment("NotSupported should be called if the current callback don't support the passed arguments.")
	f.Comment("The function return values will be ignored.")
	f.Func().Params(Id("c").Op("*").Id(objContext)).
		Id("NotSupported").
		Params().
		Block(
			Id("c").Dot("isNotSupported").Op("=").True(),
		)
	f.Line()

	f.Func().Params(Id("c").Op("*").Id(objContext)).
		Id("MethodName").
		Params().
		Params(String()).
		Block(
			Return(Id("c").Dot("methodName")),
		)
	f.Line()

	f.Func().Params(Id("c").Op("*").Id(objContext)).
		Id("ExecCount").
		Params().
		Params(Int()).
		Block(
			Return(Id("c").Dot("execCount")),
		)
	f.Line()

	f.Func().Params(Id("c").Op("*").Id(objContext)).
		Id("CallerFunc").
		Params().
		Params(String()).
		Block(
			Return(Id("c").Dot("callerFunc")),
		)
	f.Line()

	f.Func().Params(Id("c").Op("*").Id(objContext)).
		Id("CallerFile").
		Params().
		Params(String()).
		Block(
			Return(Id("c").Dot("callerFile")),
		)
	f.Line()

	f.Func().Params(Id("c").Op("*").Id(objContext)).
		Id("CallerLine").
		Params().
		Params(Int()).
		Block(
			Return(Id("c").Dot("callerLine")),
		)
	f.Line()

	// Struct implementation
	// # type TYPE struct {}
	f.Type().Id(objName).
		TypesFunc(func(tgroup *Group) {
			for tparam := range objNamedType.TypeParams().TypeParams() {
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
			group.Id(fallbackParamName).Add(util.GetQualCode(obj.Type()).TypesFunc(codeObjectTypes))
			if *callLogger {
				group.Id(onCallLoggerParamName).Add(Func().
					Params(
						Id("qdCtx").Op("*").Id(objContext),
						Id("params").Index().Any(),
					))
			}
			group.Id(onMethodNotImplementedParamName).Add(Func().
				Params(
					Id("qdCtx").Op("*").Id(objContext),
					Id("hasCallbacks").Bool(),
				).
				Params(
					Error(),
				))

			// interface method impls
			for mtd := range iface.Methods() {
				sig := mtd.Type().(*types.Signature)

				// # implMETHOD  func(qdCtx *TYPEContext, METHODPARAMS...) (METHODRESULTS...)
				group.Id("impl" + mtd.Name()).Index().Func().ParamsFunc(func(pgroup *Group) {
					// add qd context parameter
					qdCtxName := util.GetUniqueName("qdCtx", func(nameExists string) bool {
						for sigParam := range sig.Params().Variables() {
							if sigParam.Name() == nameExists {
								return true
							}
						}
						return false
					})
					pgroup.Id(qdCtxName).Op("*").Id(objContext)
					for k, sigParam := range util.IterWithIndex(sig.Params().Variables()) {
						pgroup.Id(util.ParamName(k, sigParam)).Add(util.GetSignatureParamQualCode(sig, k))
					}
				}).ParamsFunc(func(rgroup *Group) {
					for sigParam := range sig.Results().Variables() {
						rgroup.Id(sigParam.Name()).Add(util.GetQualCode(sigParam.Type()))
					}
				})
			}
		})

	// ensure struct implements interface
	if objNamedType.TypeParams().Len() == 0 { // with generics, it is harder to find suitable types
		f.Line()
		// # var _ TYPE = (*TYPE)(nil)
		f.Var().Id("_").Add(util.GetQualCode(obj.Type()).TypesFunc(codeObjectTypes)).Op("=").
			Parens(Op("*").Id(objName).TypesFunc(codeObjectTypes)).Parens(Nil())
	}

	f.Line()

	// option type
	// # type TYPEOption func(*TYPE)
	f.Type().Id(objOption).TypesFunc(codeObjectTypesWithType).Func().Params(Op("*").Id(objName).TypesFunc(codeObjectTypes))

	f.Line()

	// constructor
	// # func NewTYPE(options ...TYPEOption) *TYPE {}
	f.Func().Id("New"+objNameExported).
		TypesFunc(codeObjectTypesWithType).
		Params(
			Id("options").Op("...").Id(objOption).TypesFunc(codeObjectTypes),
		).
		ParamsFunc(func(pgroup *Group) {
			if *exportType {
				pgroup.Op("*").Id(objName).TypesFunc(codeObjectTypes)
			} else {
				pgroup.Add(util.GetQualCode(obj.Type()).TypesFunc(codeObjectTypes))
			}
		}).
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
	for mtd := range iface.Methods() {
		f.Line()

		sig := mtd.Type().(*types.Signature)

		// # func (d *TYPE) METHOD(METHODPARAMS...) (METHODRESULTS...) {}
		f.Commentf("%s implements [%s.%s].", mtd.Name(), util.FormatObjectName(obj), mtd.Name())
		f.Func().Params(Id("d").Op("*").Id(objName).TypesFunc(codeObjectTypes)).Id(mtd.Name()).ParamsFunc(func(pgroup *Group) {
			for k, sigParam := range util.IterWithIndex(sig.Params().Variables()) {
				pgroup.Id(util.MethodParamName(util.ParamName(k, sigParam))).Add(util.GetSignatureParamQualCode(sig, k))
			}
		}).ParamsFunc(func(rgroup *Group) {
			for sigParam := range sig.Results().Variables() {
				rgroup.Id(util.MethodParamName(sigParam.Name())).Add(util.GetQualCode(sigParam.Type()))
			}
		}).BlockFunc(func(s *Group) {
			s.Const().Id("methodName").Op("=").Lit(mtd.Name())

			if *callLogger {
				s.Id("d").Dot("callLogger").CallFunc(func(clgroup *Group) {
					clgroup.Id("d").Dot("createContext").Call(Id("methodName"))
					clgroup.Index().Any().ValuesFunc(func(vgroup *Group) {
						for k := range sig.Params().Len() {
							vgroup.Id(util.MethodParamName(sig.Params().At(k).Name()))
						}
					})
				})
			}

			s.For(List(Id("_"), Id("impl")).Op(":=").Range().Id("d").Dot("impl" + mtd.Name())).
				BlockFunc(func(fgroup *Group) {
					fgroup.Id("qctx").Op(":=").Id("d").Dot("createContext").Call(Id("methodName"))

					call := Id("impl").CallFunc(func(cgroup *Group) {
						cgroup.Id("qctx")
						for k := range sig.Params().Len() {
							cgroup.Id(util.GetSignatureParamCallCode(sig, k))
						}
					})

					var retVars []Code
					for k := range sig.Results().Len() {
						retVars = append(retVars, Id(fmt.Sprintf("r%d", k)))
					}

					if sig.Results().Len() == 0 {
						fgroup.Add(call)
					} else {
						fgroup.List(retVars...).Op(":=").Add(call)
					}

					fgroup.If(Op("!").Id("qctx").Dot("isNotSupported")).
						BlockFunc(func(rgroup *Group) {
							rgroup.Id("d").Dot("addCallMethod").Call(Id("methodName"))
							if sig.Results().Len() == 0 {
								rgroup.Return()
							} else {
								rgroup.Return(List(retVars...))
							}
						})
				})

			s.If(Id("d").Dot(fallbackParamName).Op("!=").Nil()).BlockFunc(func(bgroup *Group) {
				bgroup.Id("d").Dot("addCallMethod").Call(Id("methodName"))
				icall := Id("d").Dot(fallbackParamName).Dot(mtd.Name()).CallFunc(func(igroup *Group) {
					for k := range sig.Params().Len() {
						igroup.Id(util.GetSignatureParamCallCode(sig, k))
					}
				})
				if sig.Results().Len() == 0 {
					bgroup.Add(icall)
					bgroup.Return()
				} else {
					bgroup.Add(Return(icall))
				}
			})

			s.Panic(Id("d").Dot("methodNotImplemented").Call(
				Id("d").Dot("createContext").Call(Id("methodName")),
				Len(Id("d").Dot("impl"+mtd.Name())).Op(">").Lit(0),
			))
		})
	}

	// helper methods
	f.Line()

	// getCallerFuncName
	// # func (d *TYPE) getCallerFuncName(skip int) (funcName string, file string, line int) {}
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

	// addCallMethod
	// # func (d *TYPE) addCallMethod(methodName string) {}
	f.Func().Params(Id("d").Op("*").Id(objName).TypesFunc(codeObjectTypes)).
		Id("addCallMethod").
		Params(
			Id("methodName").String(),
		).
		BlockFunc(func(bgroup *Group) {
			bgroup.Id("d").Dot("lock").Dot("Lock").Call()
			bgroup.Defer().Id("d").Dot("lock").Dot("Unlock").Call()
			bgroup.Id("d").Dot("execCount").Index(Id("methodName")).Op("++")
		})

	f.Line()

	// createContext
	// # func (d *TYPE) createContext(methodName string) *TYPEContext {}
	f.Func().Params(Id("d").Op("*").Id(objName).TypesFunc(codeObjectTypes)).
		Id("createContext").
		Params(
			Id("methodName").String(),
		).
		Params(
			Op("*").Id(objContext),
		).
		Block(
			List(Id("callerFunc"), Id("callerFile"), Id("callerLine")).Op(":=").
				Id("d").Dot("getCallerFuncName").Call(Lit(3)),
			Id("d").Dot("lock").Dot("Lock").Call(),
			Defer().Id("d").Dot("lock").Dot("Unlock").Call(),
			Return(
				Op("&").Id(objContext).CustomFunc(Options{
					Open:      "{",
					Close:     "}",
					Separator: ",",
					Multi:     true,
				}, func(vgroup *Group) {
					vgroup.Id("methodName").Op(":").Id("methodName")
					vgroup.Id("execCount").Op(":").Id("d").
						Dot("execCount").Index(Id("methodName"))
					vgroup.Id("callerFunc").Op(":").Id("callerFunc")
					vgroup.Id("callerFile").Op(":").Id("callerFile")
					vgroup.Id("callerLine").Op(":").Id("callerLine")
					if codeDataType != nil {
						vgroup.Id("Data").Op(":").Id("d").Dot(dataParamName)
					}
				})),
		)

	f.Line()

	// methodNotImplemented
	// # func (d *TYPE) methodNotImplemented(qdCtx *TYPEContext, hasCallbacks bool) error {}
	f.Func().Params(Id("d").Op("*").Id(objName).TypesFunc(codeObjectTypes)).
		Id("methodNotImplemented").
		Params(
			Id("qdCtx").Op("*").Id(objContext),
			Id("hasCallbacks").Bool(),
		).
		Params(
			Error(),
		).
		BlockFunc(func(bgroup *Group) {
			bgroup.If(Id("d").Dot(onMethodNotImplementedParamName).Op("!=").Nil()).
				Block(
					If(Id("err").Op(":=").Id("d").Dot(onMethodNotImplementedParamName).Call(Id("qdCtx"), Id("hasCallbacks")),
						Id("err").Op("!=").Nil()).
						Block(
							Return(Id("err")),
						),
				)
			bgroup.Id("msg").Op(":=").Lit("not implemented")
			bgroup.If(Id("hasCallbacks")).
				Block(
					Id("msg").Op("=").Lit("not supported by any callbacks"),
				)
			bgroup.Return(Qual("fmt", "Errorf").
				Call(Lit(fmt.Sprintf("[%s] method '%%s' %%s", objName)),
					Id("qdCtx").Dot("MethodName"), Id("msg")))
		})

	f.Line()

	if *callLogger {
		// callLogger
		// # func (d *TYPE) callLogger(qdCtx *TYPEContext, params []any) {}
		f.Func().Params(Id("d").Op("*").Id(objName).TypesFunc(codeObjectTypes)).
			Id("callLogger").
			Params(
				Id("qdCtx").Op("*").Id(objContext),
				Id("params").Index().Any(),
			).
			BlockFunc(func(bgroup *Group) {
				bgroup.If(Id("d").Dot(onCallLoggerParamName).Op("!=").Nil()).
					Block(
						Id("d").Dot(onCallLoggerParamName).Call(Id("qdCtx"), Id("params")),
					)
			})

		f.Line()
	}

	f.Comment("Options")
	f.Line()

	if codeDataType != nil {
		// WithData option
		// # func WithData(data any) TYPEOption {}
		f.Func().Id("With" + objOptionPrefix + dataParamName).TypesFunc(codeObjectTypesWithType).Params(
			Id("data").Add(codeDataType),
		).Params(Id(objOption).TypesFunc(codeObjectTypes)).Block(
			Return(Func().Params(Id("d").Op("*").Id(objName).TypesFunc(codeObjectTypes)).Block(
				Id("d").Dot(dataParamName).Op("=").Id("data"),
			)),
		)
		f.Line()
	}

	// WithFallback option
	// # func WithFallback(fallback SOURCETYPE) TYPEOption {}
	f.Func().Id("With" + objOptionPrefix + util.InitialToUpper(fallbackParamName)).TypesFunc(codeObjectTypesWithType).Params(
		Id("fallback").Add(util.GetQualCode(obj.Type()).TypesFunc(codeObjectTypes)),
	).Params(Id(objOption).TypesFunc(codeObjectTypes)).Block(
		Return(Func().Params(Id("d").Op("*").Id(objName).TypesFunc(codeObjectTypes)).Block(
			Id("d").Dot(fallbackParamName).Op("=").Id("fallback"),
		)),
	)
	f.Line()

	// WithOnMethodNotImplemented option
	// # func WithOnMethodNotImplemented(m func (qdCtx *TYPEContext, hasCallbacks bool) error) TYPEOption {}
	f.Func().Id("With" + objOptionPrefix + util.InitialToUpper(onMethodNotImplementedParamName)).TypesFunc(codeObjectTypesWithType).
		Params(
			Id("m").Func().
				Params(
					Id("qdCtx").Op("*").Id(objContext),
					Id("hasCallbacks").Bool(),
				).
				Params(Error()),
		).
		Params(Id(objOption).TypesFunc(codeObjectTypes)).
		Block(
			Return(Func().Params(Id("d").Op("*").Id(objName).TypesFunc(codeObjectTypes)).Block(
				Id("d").Dot(onMethodNotImplementedParamName).Op("=").Id("m"),
			)),
		)
	f.Line()

	if *callLogger {
		// WithOnCallLogger option
		// # func WithOnCallLogger(m func (qdCtx *TYPEContext, params []any)) {}
		f.Func().Id("With" + objOptionPrefix + util.InitialToUpper(onCallLoggerParamName)).TypesFunc(codeObjectTypesWithType).
			Params(
				Id("m").Func().
					Params(
						Id("qdCtx").Op("*").Id(objContext),
						Id("params").Index().Any(),
					),
			).
			Params(Id(objOption).TypesFunc(codeObjectTypes)).
			Block(
				Return(Func().Params(Id("d").Op("*").Id(objName).TypesFunc(codeObjectTypes)).Block(
					Id("d").Dot(onCallLoggerParamName).Op("=").Id("m"),
				)),
			)
		f.Line()
	}

	// method options
	for mtd := range iface.Methods() {
		sig := mtd.Type().(*types.Signature)

		f.Line()

		// # func WithMETHOD(implMETHOD func(qdCtx *TYPEContext, METHODPARAMS...) (METHODRESULTS...)) TYPEOption {}
		f.Commentf("With%s%s implements [%s.%s].", objOptionPrefix, mtd.Name(), util.FormatObjectName(obj), mtd.Name())
		f.Func().Id("With" + objOptionPrefix + mtd.Name()).TypesFunc(codeObjectTypesWithType).Params(
			Id("impl" + mtd.Name()).Func().ParamsFunc(func(pgroup *Group) {
				// add qd context parameter
				qdCtxName := util.GetUniqueName("qdCtx", func(nameExists string) bool {
					for sigParam := range sig.Params().Variables() {
						if sigParam.Name() == nameExists {
							return true
						}
					}
					return false
				})
				pgroup.Id(qdCtxName).Op("*").Id(objContext)
				for k, sigParam := range util.IterWithIndex(sig.Params().Variables()) {
					pgroup.Id(util.ParamName(k, sigParam)).Add(util.GetSignatureParamQualCode(sig, k))
				}
			}).ParamsFunc(func(rgroup *Group) {
				for sigParam := range sig.Results().Variables() {
					rgroup.Id(sigParam.Name()).Add(util.GetQualCode(sigParam.Type()))
				}
			}),
		).Params(Id(objOption).TypesFunc(codeObjectTypes)).Block(
			Return(Func().Params(Id("d").Op("*").Id(objName).TypesFunc(codeObjectTypes)).Block(
				Id("d").Dot("impl"+mtd.Name()).Op("=").Append(Id("d").Dot("impl"+mtd.Name()), Id("impl"+mtd.Name())),
			)),
		)
	}

	// Write to file.
	fmt.Printf("Writing file %s...", outputName)
	defer fmt.Printf("\n")

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
