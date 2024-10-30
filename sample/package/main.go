package main

import (
	"fmt"
	"io"
)

// -type=Reader -type-package=io -overwrite=true -force-package-name=main -name-prefix=QD

func main() {
	reader := NewQDReader(
		WithRead(func(qdCtx *QDReaderContext, p []byte) (n int, err error) {
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
