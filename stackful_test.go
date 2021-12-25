package errx

import (
	"bufio"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetErrorStack(t *testing.T) {
	/*
	   利用一个递归来获得指定层数的调用，格式为：
	   [/path/stackful_test.go:62] go-errx.(*caller).call
	   [/path/stackful_test.go:62] go-errx.(*caller).call
	   ... （递归指定次数）
	   [d:/Workspace/my/go/go-errx/stackful_test.go:62] go-errx.TestGetErrorStack.func1
	   ... （更下层的方法）
	*/

	check := func(s string, calls int) {
		scanner := bufio.NewScanner(strings.NewReader(s))

		// 计算并校验前置的 call 递归部分的数量。
		count := 0
		var line, rx string
		for scanner.Scan() {
			line = scanner.Text()
			if count == calls {
				break
			}

			rx = `\[.+stackful_test\.go:\d+\] ` + regexp.QuoteMeta("go-errx.(*caller).call")
			if !assert.Regexp(t, rx, line) {
				assert.Fail(t, "", "expected calls %d, got %d", calls, count)
				return
			}

			count++
		}

		// 紧跟着入口函数。
		rx = `\[.+stackful_test\.go:\d+\] go-errx\.TestGetErrorStack.*`
		assert.Regexp(t, rx, line)
	}

	run := func(recursiveLevel int) {
		t.Run(strconv.Itoa(recursiveLevel), func(t *testing.T) {
			e := new(caller).call(recursiveLevel)
			s := e.Stack()
			check(s, recursiveLevel+1) // 递归次数+最后一次非递归的调用。
		})
	}

	run(0)
	run(1)
	run(5)
	run(14)
	run(16)
	run(43)
}

type caller struct {
	count int
}

// 递归指定的次数。
func (c *caller) call(recursiveLevel int) ErrorStack {
	if c.count == recursiveLevel {
		return GetErrorStack(2)
	}
	c.count++
	return c.call(recursiveLevel)
}
