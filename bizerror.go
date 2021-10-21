package errx

import "fmt"

// ErrorCode 表示错误信息的状态码。0 表示无错误；其余均表示错误。
type ErrorCode int

const (
	// 未定义。
	ErrorCodeUndefined = -1

	// 无错误。
	ErrorCodeSuccess = 0
)

// BizError 实现 error 。表示一个业务预定义错误。当遇到业务预定义错误时，方法可返回或 panic 此错误。
type BizError struct {
	// Code 表示错误码。
	Code ErrorCode

	// Message 是描述错误的信息。
	Message string

	// Inner 记录引起此错误的内部错误。
	// 当一个 error 可以被作为 BizError 对待时，可创建 BizError 并设置此字段。
	// 若没有内部错误，为 nil 。
	Inner error
}

// Ensure implementation.
var _ error = BizError{}

// Error 实现 error 接口，返回 BizError 的数据，格式为： (Code) Message 。
func (e BizError) Error() string {
	return fmt.Sprintf("(%v) %v", e.Code, e.Message)
}
