// errx 包提供一组类型与方法，用于将错误信息封装起来形成错误链，以便在程序中更精准的定位和跟踪错误。
package errx

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ErrorWrapper 是一个 StackfulError ，封装另一个 error ，其表示引起当前错误的原因。
type ErrorWrapper struct {
	ErrorCause
	ErrorStack
	msg string
}

var _ StackfulError = (*ErrorWrapper)(nil)
var _ fmt.Formatter = (*ErrorWrapper)(nil)

// Error 返回以 Describe() 的格式输出错误信息。
func (w *ErrorWrapper) Error() string {
	// Describe() 不会调用 BizError 之外的 StackfulError.Error() ，所以不会死循环。
	return Describe(w)
}

// Message 返回错误的描述信息，但不包含 Stack 。
//
// 返回格式如下，当前实例的 message 为空时，前置的“message:”部分被省略。
//  - 若 cause 为 nil，则仅返回当前实例的错误信息；
//  - 若 cause 为 StackfulError， 则返回： message: cause.ErrorWithoutStack() ；
//  - 若 cause 不是 StackfulError， 则返回： message: cause.Error() 。
func (w *ErrorWrapper) ErrorWithoutStack() string {
	c := w.Cause()
	if c == nil {
		return w.msg
	}

	prefix := w.msg
	if w.msg != "" {
		prefix += ": "
	}

	se, ok := w.Cause().(StackfulError)
	if ok {
		return prefix + se.ErrorWithoutStack()
	}
	return prefix + c.Error()
}

// Format 实现 fmt.Formatter.Formats() 。
// 支持：
//   %s  输出 ErrorWithoutStack()
//   %q  输出 strconv.Quote(ErrorWithoutStack())
//   %v  输出 ErrorWithoutStack()
//   %+v 输出 Error()
//
func (w *ErrorWrapper) Format(f fmt.State, verb rune) {
	var out string

	switch verb {
	case 's':
		out = w.ErrorWithoutStack()
	case 'q':
		out = strconv.Quote(w.ErrorWithoutStack())
	case 'v':
		if f.Flag('+') {
			out = w.Error()
		} else {
			out = w.ErrorWithoutStack()
		}
	default:
		// 其他不支持的格式，输出： BADFORMAT:Message()
		out = "BADFORMAT:" + w.ErrorWithoutStack()
	}

	io.WriteString(f, out)
}

// Wrap 封装给定的 error ，返回 StackfulError 。
// 错误信息的格式为： message: cause.Error() 。若 cause 为 nil，则仅返回 message  。
//
// 得到的 StackfulError.Stack() 有一个固定的开头“--- ”，末尾会有一个空行。格式为：
//   --- stack text
//
func Wrap(message string, cause error) StackfulError {
	return &ErrorWrapper{
		ErrorCause: ErrorCause{cause},
		ErrorStack: GetErrorStack(3), // 调用栈不包括当前函数。
		msg:        message,
	}
}

// WrapWithoutStack 封装给定的 error 。和 Wrap() 类似，但不带有调用栈信息。
// 错误信息的格式为： message: cause.Error() 。若 cause 为 nil，则仅返回 message  。
func WrapWithoutStack(message string, cause error) StackfulError {
	return &ErrorWrapper{
		ErrorCause: ErrorCause{cause},
		msg:        message,
	}
}

// PreserveRecover 用于封装从 panic 中 recover 的数据，返回 StackfulError 。
// 此方法的调用应放在 defer 过程里。
func PreserveRecover(message string, recovered interface{}) StackfulError {
	if recovered == nil {
		return nil
	}

	var cause error
	switch e := recovered.(type) {
	case error:
		cause = e
	case string:
		cause = fmt.Errorf(e)
	default:
		// panic 的不是 error 和字符串也应该是个能转成字符串的东西。
		cause = fmt.Errorf("%v", e)
	}

	return &ErrorWrapper{
		ErrorCause: ErrorCause{cause},
		ErrorStack: GetErrorStack(4), // 忽略当前函数、 panic 调用和 defer 的函数。
		msg:        message,
	}
}

// Describe 返回一个字符串描述给定的错误。如果给定 nil ，返回空字符串。
//
// 递归使用 errors.Unwrap() 获取内部错误，并追加在描述信息上。如果错误是 StackfulError ，则描述携带调用栈信息。
// 若不能获取到对应的信息，则该部分省略。
//
// 可通过此方法获取完整的错误链信息。
//
// 输出格式为：
//   最外层错误描述
//   --- 最外层错误的调用栈信息
//   === 第1层内部错误的描述
//   --- 第1层内部错误的调用栈信息
//   === 第2层内部错误的描述
//   --- 第2层内部错误的调用栈信息
//   ...（逐层展示）
//   === 最内层错误的描述
//   --- 最内层错误的描述的调用栈信息
//
func Describe(err error) string {
	if err == nil {
		return ""
	}

	var msg strings.Builder
	isStackful := false
	for {
		if msg.Len() > 0 {
			// stack 末尾自带换行。如果之前不是 stack ，就要单独添加一个。
			if !isStackful {
				msg.WriteByte('\n')
			}
			msg.WriteString("=== ")
		}

		switch e := err.(type) {
		case StackfulError:
			msg.WriteString(e.ErrorWithoutStack())
			msg.WriteString("\n--- ")
			msg.WriteString(e.Stack())
			isStackful = true
		default:
			msg.WriteString(e.Error())
		}

		err = errors.Unwrap(err)
		if err == nil {
			break
		}
	}

	return msg.String()
}
