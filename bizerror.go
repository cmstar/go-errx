package errx

import "fmt"

// ErrorCode 表示错误信息的状态码。0 表示无错误；其余均表示错误。
type ErrorCode int

// 预定义的状态码。1000以下基本抄 HTTP 状态码。
const (
	// 未定义。
	ErrorCodeUndefined = -1

	// 无错误。
	ErrorCodeSuccess = 0 // 这个不抄 HTTP 的 200 。

	// 不合规的请求数据。
	ErrorCodeBadRequest = 400

	// 发生内部错误。
	ErrorCodeInternalError = 500
)

// 表示一个业务预定义错误。当遇到业务预定义错误时，方法可返回或 panic 此错误。
type BizError struct {
	// 对应 ApiResponse.Code 。
	Code ErrorCode

	// 对应 ApiResponse.Message 。
	Message string

	// Inner 记录引起此错误的内部错误。
	// 当认为一个 error 可以被作为 BizError 对待时，可创建 BizError 并设置此字段。
	// 若没有内部错误，为 nil 。
	Inner error
}

// Error 实现 error 接口，返回 BizError 的数据，格式为： (Code) Message 。
func (e BizError) Error() string {
	return fmt.Sprintf("(%v) %v", e.Code, e.Message)
}
