package main

import (
	"fmt"
	"io"
)

// -type=Reader -type-package=io -overwrite=true -force-package-name=main

func main() {
	reader := NewReader(
		WithRead(func(qdCtx *ReaderContext, p []byte) (n int, err error) {
			n = copy(p, []byte("test"))
			return n, nil
		}),
	)

	readInterface(reader)
}

func readInterface(r io.Reader) {
	b := make([]byte, 10)

	n, err := r.Read(b)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%d: %v\n", n, b)
}
