package errx_test

import (
	"errors"
	"fmt"
	"strings"

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

func ExampleWrap() {
	origin := errors.New("the cause")
	w := errx.Wrap("something wrong", origin)

	fmt.Println("Cause():")
	fmt.Println(w.Cause())

	fmt.Println("\nErrorWithoutStack() has no stack:")
	fmt.Println(w.ErrorWithoutStack())

	fmt.Println("\nError() has stack:")
	e := w.Error()
	index := strings.Index(e, "--- ")
	fmt.Println(e[:index] + "--- the callstack ...")

	// Output:
	// Cause():
	// the cause
	//
	// ErrorWithoutStack() has no stack:
	// something wrong: the cause
	//
	// Error() has stack:
	// something wrong: the cause
	// --- the callstack ...
}
