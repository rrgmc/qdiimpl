package main

import (
	"fmt"
	"go/types"

	"github.com/dave/jennifer/jen"
)

func getQualCode(typ types.Type) *jen.Statement {
	var st jen.Statement
	for {
		switch tt := typ.(type) {
		case *types.Named:
			if tt.Obj().Pkg() != nil {
				return st.Add(jen.Qual(tt.Obj().Pkg().Path(), tt.Obj().Name()).TypesFunc(addTypeList(tt.TypeArgs())))
			}
			return st.Add(jen.Id(tt.Obj().Name()).TypesFunc(addTypeList(tt.TypeArgs())))
		case *types.Interface:
			return st.Add(jen.Id(tt.String()))
		case *types.Basic:
			return st.Add(jen.Id(tt.Name()))
		case *types.TypeParam:
			return st.Add(jen.Id(tt.Obj().Name()))
		case *types.Pointer:
			st.Add(jen.Op("*"))
			typ = tt.Elem()
		case *types.Slice:
			return st.Add(jen.Index().Add(getQualCode(tt.Elem())))
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
		default:
			panic(fmt.Errorf("unknown type %T", typ))
		}
	}
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
