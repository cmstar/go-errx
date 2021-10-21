package errx

import "runtime"

// StackfulError 是一个包含调用栈信息的 error 。
type StackfulError interface {
	error

	// Stack 返回调用栈信息。若未记录调用栈，可返回空值。
	Stack() string
}

// ErrorCause 用于封装一个 error ，它支持 errors.Unwrap() 。
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

// ErrorStack 用于存放调用栈信息，以便实现 StackfulError 。
type ErrorStack struct {
	stack string
}

// Stack 实现 StackfulError.Stack() 。
func (e ErrorStack) Stack() string {
	return e.stack
}

// GetErrorStack 创建一个带有调用栈信息的 ErrorStack 。
// 调用栈信息使用 runtime.Stack(buf, false) 获取。
//
// 得到的 ErrorStack.Stack() 有一个固定的开头“--- ”，末尾会有一个空行。格式为：
//   --- stack text
//
func GetErrorStack() ErrorStack {
	const initialBufSize = 12  // 1K.
	const stackPrefix = "--- " // 拼接调用栈信息时，每组信息的开头部分，作为分隔符。

	// 这段代码算法参考 debug.Stack() ，其使用 n < len(buf) 判断，而不是 <= 。
	// 似乎 <= 才是对的，可以避免一次不必要的扩容。在没搞透彻前，还是抄之。
	// 按照格式约定，在 buf 前面额外追加一个分隔符。
	// runtime.Stack() 会重置和填充给定的切片，这里利用切片特性，将扣除了头部分隔符的切片传入，
	// 使其填充在分隔符后面，同时保留分隔符。这样避免了有一次数据冗余，提高内存利用率。
	buf := make([]byte, 0, initialBufSize)
	buf = append(buf, stackPrefix...)
	bufWithoutPrefix := buf[len(stackPrefix):cap(buf)]
	for {
		n := runtime.Stack(bufWithoutPrefix, false)
		if n < len(bufWithoutPrefix) {
			buf = buf[:n+len(stackPrefix)] // -> stackPrefix + bufWithoutPrefix
			break
		}
		buf = make([]byte, 0, 2*cap(buf))
		buf = append(buf, stackPrefix...)
		bufWithoutPrefix = buf[len(stackPrefix):cap(buf)]
	}

	e := ErrorStack{string(buf)}
	return e
}
