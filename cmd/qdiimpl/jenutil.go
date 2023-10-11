package main

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/dave/jennifer/jen"
)

func getQualCode(typ types.Type) *jen.Statement {
	var st jen.Statement
	for {
		switch tt := typ.(type) {
		case *types.Basic:
			return st.Add(jen.Id(tt.Name()))
		case *types.Array:
			return st.Add(jen.Index(jen.Lit(tt.Len())).Add(getQualCode(tt.Elem())))
		case *types.Slice:
			return st.Add(jen.Index().Add(getQualCode(tt.Elem())))
		case *types.Pointer:
			st.Add(jen.Op("*"))
			typ = tt.Elem()
		case *types.Tuple:
			var items jen.Statement
			for i := 0; i < tt.Len(); i++ {
				items.Add(jen.Id(tt.At(i).Name()).Add(getQualCode(tt.At(i).Type())))
			}
			return st.Add(jen.Params(items...))
		case *types.Interface:
			return st.Add(jen.Id(tt.String()))
		case *types.Map:
			return st.Add(jen.Map(getQualCode(tt.Key())).Add(getQualCode(tt.Elem())))
		case *types.Chan:
			var chanDesc *jen.Statement
			switch tt.Dir() {
			case types.SendRecv:
				chanDesc = jen.Chan()
			case types.SendOnly:
				chanDesc = jen.Chan().Op("<-")
			case types.RecvOnly:
				chanDesc = jen.Op("<-").Chan()
			default:
				panic("unknown channel direction")
			}
			return st.Add(chanDesc.Add(getQualCode(tt.Elem())))
		case *types.Named:
			if tt.Obj().Pkg() != nil {
				return st.Add(jen.Qual(tt.Obj().Pkg().Path(), tt.Obj().Name()).TypesFunc(addTypeList(tt.TypeArgs())))
			}
			return st.Add(jen.Id(tt.Obj().Name()).TypesFunc(addTypeList(tt.TypeArgs())))
		case *types.TypeParam:
			return st.Add(jen.Id(tt.Obj().Name()))
		default:
			panic(fmt.Errorf("unknown type %T", typ))
		}
	}
}

func typeNameCode(typeName string) (*jen.Statement, error) {
	typeName, isPtr := strings.CutPrefix(typeName, "*")

	var st jen.Statement
	if isPtr {
		st.Add(jen.Op("*"))
	}

	lastIndex := strings.LastIndexAny(typeName, "/.")
	if lastIndex == -1 {
		return st.Add(jen.Id(typeName)), nil
	}
	if typeName[lastIndex:lastIndex+1] == "." {
		return st.Add(jen.Qual(typeName[:lastIndex], typeName[lastIndex+1:])), nil
	}
	return nil, fmt.Errorf("invalid type name format (must have a dot to determine the type): %s", typeName)
}

func paramName(idx int, param *types.Var) string {
	if param.Name() != "" {
		return param.Name()
	}
	return fmt.Sprintf("p%d", idx)
}

func addTypeParamsList(typeList *types.TypeParamList, withType bool) func(*jen.Group) {
	return func(tgroup *jen.Group) {
		for t := 0; t < typeList.Len(); t++ {
			tparam := typeList.At(t)
			if withType {
				tgroup.Id(tparam.Obj().Name()).Add(getQualCode(tparam.Constraint()))
			} else {
				tgroup.Id(tparam.Obj().Name())
			}
		}
	}
}

func addTypeList(typeList *types.TypeList) func(*jen.Group) {
	return func(tgroup *jen.Group) {
		for t := 0; t < typeList.Len(); t++ {
			tparam := typeList.At(t)
			tgroup.Add(getQualCode(tparam))
		}
	}
}
