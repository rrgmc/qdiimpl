package util

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestAddTypeParamsList(t *testing.T) {
	const src = `
package p

type T struct {
	P int
}

func (T) M(int) {}
func (T) N() (i int) { return }

type G[P any] struct {
	F P
}

func (G[P]) M(P) {}
func (G[P]) N() (p P) { return }

type Inst = G[int]
	`
	pkg := mustTypecheck(src, nil, nil)

	T := pkg.Scope().Lookup("T").Type()

	assert.Assert(t, T != nil)
}
