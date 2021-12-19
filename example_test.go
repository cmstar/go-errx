package errx

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
)

func Example_errorChain() {
	// 当前示例演示如何利用 Wrap 方法封装错误，并利用 Describe 方法获取完整的错误描述。

	// 调用顺序 A() -> B() -> Source() 。
	err := A()
	msg := Describe(err)

	// 简化一下输出。
	msg = simplify(msg)
	fmt.Println(msg)

	// Output:
	// from A: from B: the original error
	// --- goroutine 1 [running]:
	// go-errx.GetErrorStack()
	// go-errx.Wrap(...)
	// go-errx.A()
	// go-errx.Example_errorChain()
	// main.main()
	// ===
	// from B: the original error
	// --- goroutine 1 [running]:
	// go-errx.GetErrorStack()
	// go-errx.Wrap(...)
	// go-errx.B()
	// go-errx.A()
	// go-errx.Example_errorChain()
	// main.main()
	// ===
	// the original error
}

func A() error {
	e := B()
	return Wrap("from A", e)
}

func B() error {
	e := Source()
	return Wrap("from B", e)
}

func Source() error {
	return errors.New("the original error")
}

func simplify(stack string) string {
	/*
		完整的信息类似：
		from A: from B: the original error
		--- goroutine 1 [running]:
		github.com/cmstar/go-errx.GetErrorStack()
			/path/go-errx/stackful.go:57 +0x7a
		github.com/cmstar/go-errx.Wrap(...)
			/path/go-errx/errx.go:43
		github.com/cmstar/go-errx.A()
			/path/go-errx/example_test.go:66 +0x2c
		github.com/cmstar/go-errx.Example()
			/path/go-errx/example_test.go:11 +0x19
		testing.runExample({{0x5b0039, 0x7}, 0x5c3f38, {0x5be5ed, 0x143}, 0x0})
			/gopath/src/testing/run_example.go:64 +0x28d
		testing.runExamples(0xc0000c3e70, {0x7092e0, 0x1, 0x5})
			/gopath/src/testing/example.go:44 +0x18e
		testing.(*M).Run(0xc0000ce100)
			/gopath/src/testing/testing.go:1505 +0x587
		main.main()
			_testmain.go:53 +0x14b

		过于冗长且不稳定，耦合物理路径难以断言输出，将其简化，仅保留下方法名称。
	*/

	isLineNum := func(v string) bool {
		return strings.Contains(v, ".go")
	}
	isMethodCall := func(v string) bool {
		return strings.Contains(v, ")")
	}
	isTestingCall := func(v string) bool {
		// testing.(*M).Run(0xc0000ce100)
		return strings.HasPrefix(v, "testing")
	}

	s := bufio.NewScanner(strings.NewReader(stack))
	b := new(strings.Builder)
	for s.Scan() {
		line := s.Text()
		if isLineNum(line) || isTestingCall(line) {
			continue
		}

		if isMethodCall(line) {
			idx := strings.LastIndex(line, "/")
			line = line[idx+1:]
		}

		b.WriteString(line)
		b.WriteRune('\n')
	}
	return b.String()
}