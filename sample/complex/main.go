package main

import (
	"context"
	"fmt"
)

type IIImpl struct {
}

func (I IIImpl) Setme() {
}

// -type=MyInterface -data-type=any -overwrite=true -name-prefix=QD

func main() {
	data := 12

	x := NewQDMyInterface[string, IIImpl](
		WithDataQDII[string, IIImpl](&data),
		WithGet[string, IIImpl](func(qdCtx *QDMyInterfaceContext, ctx context.Context, name string) (string, error) {
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
