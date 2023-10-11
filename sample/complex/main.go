package main

import (
	"context"
	"fmt"
)

type IIImpl struct {
}

func (I IIImpl) Setme() {
}

// -type=MyInterface -data-type=any -overwrite=true

func main() {
	data := 12

	x := NewDebugMyInterface[string, IIImpl](
		WithDebugMyInterfaceDataQDII[string, IIImpl](&data),
		WithDebugMyInterfaceGet[string, IIImpl](func(debugCtx *DebugMyInterfaceContext, ctx context.Context, name string) (string, error) {
			d := debugCtx.Data.(*int)
			*d++
			return fmt.Sprintf("a%v", *d), nil
		}),
	)

	v, err := x.Get(context.Background(), "a")
	if err != nil {
		panic(err)
	}
	fmt.Println(v)

	v, err = x.Get(context.Background(), "b")
	if err != nil {
		panic(err)
	}
	fmt.Println(v)
}
