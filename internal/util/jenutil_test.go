package util

import (
	"go/types"
	"testing"

	"github.com/dave/jennifer/jen"
	"gotest.tools/v3/assert"
)

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
