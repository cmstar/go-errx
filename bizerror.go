package errx

import (
	"strconv"
	"strings"
)

// ErrorCode 表示错误信息的状态码。0 表示无错误；其余均表示错误。
type ErrorCode int

const (
	// 未定义。
	ErrorCodeUndefined = -1

	// 无错误。
	ErrorCodeSuccess = 0
)

// BizError 是一个 error 。增加了错误码等属性，可更精确的表示一个预定义的错误。
// BizError 实现 StackfulError ，可以携带调用栈信息。
type BizError interface {
	StackfulError

	// Code 表示错误码。
	// 在 webapi 中，对应 ApiResponse.Code 。
	Code() ErrorCode

	// Message 返回错误的描述信息。
	Message() string

	// Cause 记录引起此错误的内部错误。
	// 当一个 error 可以被作为 BizError 对待时，可创建 BizError 并设置此字段。
	// 若没有内部错误，为 nil 。
	Cause() error
}

// bizErr 实现 BizError 。
type bizErr struct {
	ErrorCause
	ErrorStack
	code    ErrorCode
	message string
}

// Ensure implementation.
var _ BizError = (*bizErr)(nil)

// Code 返回错误码。通常 0 是一个保留值，表示没有错误，不用于任何错误。
func (e *bizErr) Code() ErrorCode {
	return e.code
}

// Message 返回错误的描述信息，不含错误码。
func (e *bizErr) Message() string {
	return e.message
}

// Error 实现 error 接口，返回 BizError 的数据，格式为： (Code) Message 。
func (e *bizErr) Error() string {
	var b strings.Builder
	b.WriteRune('(')
	b.WriteString(strconv.Itoa(int(e.code)))
	b.WriteString(") ")
	b.WriteString(e.message)
	res := b.String()
	return res
}

// NewBizError 创建一个 BizError 。
// 可通过 cause 指定引发此错误的错误，可以为 nil 。
// 此方法创建的 BizError 会包含方法调用栈信息，通过 debug.Stack() 获取。
func NewBizError(code ErrorCode, message string, cause error) BizError {
	bizErr := &bizErr{
		code:       code,
		message:    message,
		ErrorCause: ErrorCause{cause},
		ErrorStack: GetErrorStack(),
	}
	return bizErr
}

// NewBizErrorWithoutStack 创建一个 BizError 。
// 可通过 cause 指定引发此错误的错误，可以为 nil 。
// 和 NewBizError() 类似，但不带调用栈信息， BizError.Stack() 返回空字符串。
func NewBizErrorWithoutStack(code ErrorCode, message string, cause error) BizError {
	bizErr := &bizErr{
		code:       code,
		message:    message,
		ErrorCause: ErrorCause{cause},
	}
	return bizErr
}
