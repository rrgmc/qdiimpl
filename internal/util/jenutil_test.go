package util

import (
	"go/types"
	"testing"

	"github.com/dave/jennifer/jen"
	"gotest.tools/v3/assert"
)

func TestGetQualCode(t *testing.T) {
	const src = `
package p

type List interface {
	AddItem(value int)
}

type List2 = List

type Item[T any, S comparable] interface {
	AddItem(item T)
	Compare(other S) int
}

var (
	stringValue string
	arrayValue  [4]int
	sliceValue  []int
	pointerValue *float32
	mapValue map[float32]int8
	chanValue chan int
	chanSendValue chan <-int
	chanRecvValue <-chan int
)

var (
	itemValue Item[int, string]
	itemArrayValue [6]Item[int, string]
	itemArrayItemValue [6]Item[Item[int, float64], string]
	itemSliceValue []Item[int, string]
	itemSliceItemValue []Item[Item[int, float64], string]
	itemPointerValue *Item[int, string]
	itemPointerItemValue *Item[Item[int, float64], string]
	itemMapValue map[string]Item[int, string]
	itemMapItemValue map[string]Item[Item[int, float64], string]
	itemChanValue chan Item[int, string]
	itemChanItemValue chan Item[Item[int, float64], string]
)

func NewValue(v1 string, v2 int) *string {
	return nil
}

func NewItem(v1 string, v2 Item[string, int]) *Item[string, int] {
	return nil
}
	`
	pkg := mustTypecheck(src, nil, nil)
	objList := pkg.Scope().Lookup("List")
	assert.Assert(t, objList != nil)
	assert.Assert(t, types.IsInterface(objList.Type()))

	// objNamedType := obj.Type().(*types.Named) // interfaces are always named types

	var cases = []tc{
		{
			desc:   `basic`,
			code:   jen.Var().Id("aValue").Add(GetQualCode(pkg.Scope().Lookup("stringValue").Type())),
			expect: `var aValue string`,
		},
		{
			desc:   `array`,
			code:   jen.Var().Id("aValue").Add(GetQualCode(pkg.Scope().Lookup("arrayValue").Type())),
			expect: `var aValue [4]int`,
		},
		{
			desc:   `array item`,
			code:   jen.Var().Id("aValue").Add(GetQualCode(pkg.Scope().Lookup("itemArrayValue").Type())),
			expect: `var aValue [6]p.Item[int, string]`,
		},
		{
			desc:   `array item item`,
			code:   jen.Var().Id("aValue").Add(GetQualCode(pkg.Scope().Lookup("itemArrayItemValue").Type())),
			expect: `var aValue [6]p.Item[p.Item[int, float64], string]`,
		},
		{
			desc:   `slice`,
			code:   jen.Var().Id("aValue").Add(GetQualCode(pkg.Scope().Lookup("sliceValue").Type())),
			expect: `var aValue []int`,
		},
		{
			desc:   `slice item`,
			code:   jen.Var().Id("aValue").Add(GetQualCode(pkg.Scope().Lookup("itemSliceValue").Type())),
			expect: `var aValue []p.Item[int, string]`,
		},
		{
			desc:   `slice item item`,
			code:   jen.Var().Id("aValue").Add(GetQualCode(pkg.Scope().Lookup("itemSliceItemValue").Type())),
			expect: `var aValue []p.Item[p.Item[int, float64], string]`,
		},
		{
			desc:   `pointer`,
			code:   jen.Var().Id("aValue").Add(GetQualCode(pkg.Scope().Lookup("pointerValue").Type())),
			expect: `var aValue *float32`,
		},
		{
			desc:   `pointer item`,
			code:   jen.Var().Id("aValue").Add(GetQualCode(pkg.Scope().Lookup("itemPointerValue").Type())),
			expect: `var aValue *p.Item[int, string]`,
		},
		{
			desc:   `pointer item item`,
			code:   jen.Var().Id("aValue").Add(GetQualCode(pkg.Scope().Lookup("itemPointerItemValue").Type())),
			expect: `var aValue *p.Item[p.Item[int, float64], string]`,
		},
		{
			desc: `tuple`,
			codeFn: func() jen.Code {
				objFn := pkg.Scope().Lookup("NewValue")
				assert.Assert(t, objFn != nil)
				objFnType := objFn.Type().(*types.Signature)
				return jen.Func().Id("OtherNewValue").ParamsFunc(func(rgroup *jen.Group) {
					for sigParam := range objFnType.Params().Variables() {
						rgroup.Id(sigParam.Name()).Add(GetQualCode(sigParam.Type()))
					}
				})
			},
			expect: `func OtherNewValue(v1 string, v2 int)`,
		},
		{
			desc: `tuple item`,
			codeFn: func() jen.Code {
				objFn := pkg.Scope().Lookup("NewItem")
				assert.Assert(t, objFn != nil)
				objFnType := objFn.Type().(*types.Signature)
				return jen.Func().Id("OtherNewItem").ParamsFunc(func(rgroup *jen.Group) {
					for sigParam := range objFnType.Params().Variables() {
						rgroup.Id(sigParam.Name()).Add(GetQualCode(sigParam.Type()))
					}
				})
			},
			expect: `func OtherNewItem(v1 string, v2 p.Item[string, int])`,
		},
		{
			desc:   `interface`,
			code:   jen.Var().Id("list").Add(GetQualCode(objList.Type().Underlying().(*types.Interface))),
			expect: `var list interface{ AddItem(value int) }`,
		},
		{
			desc:   `map`,
			code:   jen.Var().Id("aValue").Add(GetQualCode(pkg.Scope().Lookup("mapValue").Type())),
			expect: `var aValue map[float32]int8`,
		},
		{
			desc:   `map item`,
			code:   jen.Var().Id("aValue").Add(GetQualCode(pkg.Scope().Lookup("itemMapValue").Type())),
			expect: `var aValue map[string]p.Item[int, string]`,
		},
		{
			desc:   `map item item`,
			code:   jen.Var().Id("aValue").Add(GetQualCode(pkg.Scope().Lookup("itemMapItemValue").Type())),
			expect: `var aValue map[string]p.Item[p.Item[int, float64], string]`,
		},
		{
			desc:   `chan`,
			code:   jen.Var().Id("aValue").Add(GetQualCode(pkg.Scope().Lookup("chanValue").Type())),
			expect: `var aValue chan int`,
		},
		{
			desc:   `chan send`,
			code:   jen.Var().Id("aValue").Add(GetQualCode(pkg.Scope().Lookup("chanSendValue").Type())),
			expect: `var aValue chan <-int`,
		},
		{
			desc:   `chan recv`,
			code:   jen.Var().Id("aValue").Add(GetQualCode(pkg.Scope().Lookup("chanRecvValue").Type())),
			expect: `var aValue <-chan int`,
		},
		{
			desc:   `chan item`,
			code:   jen.Var().Id("aValue").Add(GetQualCode(pkg.Scope().Lookup("itemChanValue").Type())),
			expect: `var aValue chan p.Item[int, string]`,
		},
		{
			desc:   `chan item item`,
			code:   jen.Var().Id("aValue").Add(GetQualCode(pkg.Scope().Lookup("itemChanItemValue").Type())),
			expect: `var aValue chan p.Item[p.Item[int, float64], string]`,
		},
		{
			desc:   `named`,
			code:   jen.Var().Id("list").Add(GetQualCode(objList.Type())),
			expect: `var list p.List`,
		},
		{
			desc:   `named item`,
			code:   jen.Var().Id("item").Add(GetQualCode(pkg.Scope().Lookup("Item").Type())),
			expect: `var item p.Item`,
		},
		{
			desc:   `alias`,
			code:   jen.Type().Id("NewList2").Op("=").Add(GetQualCode(pkg.Scope().Lookup("List2").Type())),
			expect: `type NewList2 = p.List2`,
		},
		{
			desc:   `signature`,
			code:   jen.Type().Id("NewValue2").Add(GetQualCode(pkg.Scope().Lookup("NewValue").Type())),
			expect: `type NewValue2 func(v1 string, v2 int) *string`,
		},
	}

	runTestCases(t, cases)
}

func TestAddTypeParamsList(t *testing.T) {
	const src = `
package p

type Item[T any, S comparable] interface {
	AddItem(T)
}
	`
	pkg := mustTypecheck(src, nil, nil)
	obj := pkg.Scope().Lookup("Item")
	assert.Assert(t, obj != nil)

	objNamedType := obj.Type().(*types.Named) // interfaces are always named types

	codeObjectTypes := AddTypeParamsList(objNamedType.TypeParams(), false)
	codeObjectTypesWithType := AddTypeParamsList(objNamedType.TypeParams(), true)

	var cases = []tc{
		{
			desc:   `type_params_decl`,
			code:   jen.Type().Id("ItemImpl").TypesFunc(codeObjectTypesWithType).Interface(),
			expect: `type ItemImpl[T any, S comparable] interface{}`,
		},
		{
			desc: `type_params_use`,
			code: jen.Func().Params(jen.Id("d").Op("*").Id("ItemImpl").TypesFunc(codeObjectTypes)).
				Id("AddItem").
				Params(
					jen.Id("s").Id("T"),
				),
			expect: `func (d *ItemImpl[T, S]) AddItem(s T)`,
		},
	}

	runTestCases(t, cases)
}
