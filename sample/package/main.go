package main

import (
	"fmt"
	"io"
)

// -type=Reader -package=io -overwrite=true -force-package=main

func main() {
	reader := NewDebugReader(
		WithDebugReaderData(15),
		WithDebugReaderRead(func(debugCtx *DebugReaderContext, p []byte) (n int, err error) {
			n = copy(p, []byte("test"))
			return n, nil
		}),
	)

	readInterface(reader)

	fmt.Println(reader.Data)
}

func readInterface(r io.Reader) {
	b := make([]byte, 10)

	n, err := r.Read(b)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%d: %v\n", n, b)
}
