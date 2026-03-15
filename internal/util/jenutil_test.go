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

type Item[T any, S comparable] interface {
	AddItem(item T)
	Compare(other S) int
}

var (
	stringValue string
	arrayValue  [4]int
	sliceValue  []int
)

var (
	itemValue Item[int, string]
	itemArrayValue [6]Item[int, string]
	itemArrayItemValue [6]Item[Item[int, float64], string]
	itemSliceValue []Item[int, string]
	itemSliceItemValue []Item[Item[int, float64], string]
)
	`
	pkg := mustTypecheck(src, nil, nil)
	objList := pkg.Scope().Lookup("List")
	assert.Assert(t, objList != nil)

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
			desc:   `named`,
			code:   jen.Var().Id("list").Add(GetQualCode(objList.Type())),
			expect: `var list p.List`,
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
