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

	x := NewQDMyInterface[string, IIImpl](
		WithQDMyInterfaceDataQDII[string, IIImpl](&data),
		WithQDMyInterfaceGet[string, IIImpl](func(qdCtx *QDMyInterfaceContext, ctx context.Context, name string) (string, error) {
			d := qdCtx.Data.(*int)
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
