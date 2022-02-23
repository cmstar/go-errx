package errx_test

import (
	"fmt"

	"github.com/cmstar/go-errx"
)

func ExamplePreserveRecover() {
	demo := func() (e error) {
		defer func() {
			e = errx.PreserveRecover("oops, it panics", recover())
		}()

		panic(123)
	}

	err := demo()
	fmt.Println(fmt.Sprintf("%+v", err)[:25] + "the callstack ...")

	// Output:
	// oops, it panics: 123
	// --- the callstack ...
}
