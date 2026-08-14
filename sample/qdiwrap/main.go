package main

import (
	"fmt"
	"io"
	"strings"
)

// -type=ReadWriter -type-package=io -overwrite=true -force-package-name=main

func main() {
	// 1. Read is overridden, so the callback wins.
	rw := NewReadWriter(
		WithRead(func(p []byte) (n int, err error) {
			return copy(p, []byte("test")), nil
		}),
	)
	readInterface(rw)

	// 2. Read is not overridden, but a fallback is set, so it delegates to the fallback.
	rw = NewReadWriter(
		WithFallback(struct {
			io.Reader
			io.Writer
		}{Reader: strings.NewReader("fallback"), Writer: io.Discard}),
	)
	readInterface(rw)

	// 3. Write has neither an override nor a fallback, so it panics.
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Recovered from panic. Error:\n%v\n", r)
			}
		}()
		_, _ = NewReadWriter().Write([]byte("nope"))
	}()
}

func readInterface(r io.Reader) {
	b := make([]byte, 10)

	n, err := r.Read(b)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%d: %s\n", n, b[:n])
}
