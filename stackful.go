package errx

import (
	"runtime"
	"strconv"
	"strings"
)

// StackfulError 是一个包含调用栈信息的 error 。
//
// 通常， Error() 方法返回带有调用栈信息的错误描述。
// 可通过 ErrorWithoutStack() 获取没有调用栈的错误描述。
//
type StackfulError interface {
	error

	// ErrorWithoutStack 返回错误的描述信息，但不包含 Stack 。
	ErrorWithoutStack() string

	// Cause 返回引起当前错误的错误。
	Cause() error

	// Stack 返回调用栈信息。若未记录调用栈，可返回空值。
	Stack() string
}

// ErrorCause 用于封装一个 error ，表示引起另一个错误的错误，它支持 errors.Unwrap() 。
type ErrorCause struct {
	Err error
}

// Unwrap 返回引起当前错误的错误。
func (e ErrorCause) Unwrap() error {
	return e.Err
}

// Cause 同 Unwrap()。返回引起当前错误的错误。
func (e ErrorCause) Cause() error {
	return e.Err
}

// frame 存放了 runtime.Frame 的部分字段，用于自定义输出格式。
type frame struct {
	name string // 方法名称。
	line string // 行号。
	file string // 文件名。
}

// shortFuncName 从一个完整的函数描述中获取短名称，去掉路径部分： github.com/user/pkg.Name -> pkg.Name 。
func (f frame) shortName() string {
	idx := strings.LastIndex(f.name, "/")
	if idx < 0 {
		return f.name
	}
	return f.name[idx+1:]
}

// ErrorStack 用于存放调用栈信息，以便实现 StackfulError 。
// 输出调用栈格式为（末尾有一个空行）：
//   [file0:line] func0
//   [file1:line] func1
//   [file2:line] func2
//
type ErrorStack struct {
	frames []frame
}

// Stack 实现 StackfulError.Stack() 。
func (e ErrorStack) Stack() string {
	b := new(strings.Builder)
	for i := 0; i < len(e.frames); i++ {
		f := e.frames[i]
		b.WriteRune('[')
		b.WriteString(f.file)
		b.WriteRune(':')
		b.WriteString(f.line)
		b.WriteString("] ")
		b.WriteString(f.shortName())
		b.WriteRune('\n')
	}
	return b.String()
}

// GetErrorStack 创建一个带有调用栈信息的 ErrorStack 。
// 调用栈信息使用 runtime.CallersFrames() 获取，skip 参数传递给 runtime.Callers() 。
// 要跳过当前函数，至少为 2 ：分别跳过 runtime.Callers() 和当前函数。
func GetErrorStack(skip int) ErrorStack {
	const batchSize = 24 // Go 的调用层级通常不会很多，此大小足够应付多数场景。

	var localFrames []frame
	ps := make([]uintptr, batchSize)
	for {
		num := runtime.Callers(skip, ps)
		runtimeFrames := runtime.CallersFrames(ps[:num])

		if localFrames == nil {
			localFrames = make([]frame, 0, num)
		}

		for {
			f, more := runtimeFrames.Next()

			name := f.Func.Name()
			localFrames = append(localFrames, frame{
				name: name,
				file: f.File,
				line: strconv.Itoa(f.Line),
			})

			if !more {
				break
			}
		}

		if num < batchSize {
			break
		}

		// 当 num == batchSize ，说明可能还没获取完整，推进到下一批次。
		skip += batchSize
	}

	// 将末尾的系统调用去掉，让信息“干净”点。
	localFrames = excludeRuntimeFrame(localFrames)
	return ErrorStack{localFrames}
}

// excludeRuntimeFrame 将 fs 末尾的标准库 runtime 包的调用去掉。
func excludeRuntimeFrame(fs []frame) []frame {
	var i int
	var f frame
	for i = len(fs) - 1; i >= 0; i-- {
		f = fs[i]
		if f.file == "" || f.line == "0" {
			// 最底下可能有个什么信息都没有的调用，应该来自非 GO 代码。
			continue
		}
		if !strings.HasPrefix(f.name, "runtime.") {
			break
		}
	}
	return fs[:i+1]
}
