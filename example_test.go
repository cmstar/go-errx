package errx_test

import (
	"bufio"
	"errors"
	"fmt"
	"strings"

	"github.com/cmstar/go-errx"
)

func Example_errorChain() {
	// 当前示例演示如何利用 Wrap 方法封装错误，并利用 Describe 方法获取完整的错误描述。

	// 调用顺序 A() -> B() -> Source() 。
	err := A()

	// 获取带有调用栈的错误描述。
	// 也可以通过 fmt.Sprintf("%+v", err) 得到一样的结果。
	msg := errx.Describe(err)

	// 简化一下输出。
	msg = simplify(msg)
	fmt.Println(msg)

	// Output:
	// from A: from B: the original error
	// --- go-errx_test.A
	// go-errx_test.Example_errorChain
	// === from B: the original error
	// --- go-errx_test.B
	// go-errx_test.A
	// go-errx_test.Example_errorChain
	// === the original error
}

//go:noinline
func A() error {
	e := B(0)
	return errx.Wrap("from A", e)
}

//go:noinline
func B(int) error {
	e := Source()
	return errx.Wrap("from B", e)
}

//go:noinline
func Source() error {
	return errors.New("the original error")
}

func simplify(stack string) string {
	/*
		完整的信息类似：
		  from A: from B: the original error
		  --- [/home/my/go-errx/example_test.go:42] go-errx_test.A
		  [/home/my/go-errx/example_test.go:16] go-errx_test.Example_errorChain
		  [/path/testing/run_example.go:64] testing.runExample
		  [/path/testing/example.go:44] testing.runExamples
		  [/path/testing/testing.go:1505] testing.(*M).Run
		  [_testmain.go:57] main.main

		过于冗长，耦合物理路径难以断言输出，将其简化，仅留下方法名，并去掉非本地代码的部分。
	*/
	s := bufio.NewScanner(strings.NewReader(stack))
	b := new(strings.Builder)
	for s.Scan() {
		line := s.Text()
		if strings.Contains(line, "testing.") || strings.Contains(line, "main") {
			continue
		}

		idx := strings.Index(line, "[")
		if idx >= 0 {
			idxSlash := strings.LastIndex(line, " ")
			if idxSlash >= 0 {
				line = line[:idx] + line[idxSlash+1:]
			}
		}

		b.WriteString(line)
		b.WriteRune('\n')
	}
	return b.String()
}
