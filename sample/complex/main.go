package main

import (
	"context"
	"fmt"
)

type IIImpl struct {
}

func (I IIImpl) Setme() {
}

func main() {
	x := NewDebugMyInterface[string, IIImpl](
		WithDebugMyInterfaceDataQDII[string, IIImpl](12),
		WithDebugMyInterfaceGet[string, IIImpl](func(debugCtx *DebugMyInterfaceContext, ctx context.Context, name string) (string, error) {
			return fmt.Sprintf("a%v", debugCtx.Data), nil
		}),
	)

	v, err := x.Get(context.Background(), "a")
	if err != nil {
		panic(err)
	}
	fmt.Println(v)
}
