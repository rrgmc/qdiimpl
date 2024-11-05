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
		WithOnMethodNotImplemented[string, IIImpl](func(qdCtx *QDMyInterfaceContext) error {
			return fmt.Errorf("the method '%s' was not implemented", qdCtx.MethodName)
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

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Recovered from panic. Error:\n%v", r)
			}
		}()
		_ = x.Set(context.Background(), "c", "d")
	}()
}
