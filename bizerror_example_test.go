package errx_test

import (
	"errors"
	"fmt"

	"github.com/cmstar/go-errx"
)

func ExampleBizError() {
	// 当前示例演示如何通过 BizError 对错误进行分类。

	printErr := func(e error) {
		biz, ok := e.(errx.BizError)
		if ok {
			switch biz.Code() {
			case 1:
				fmt.Println("BizError1: " + biz.Error())
			case 2:
				fmt.Println("BizError2: " + biz.Error())
			}
		} else {
			fmt.Println("non-BizError: " + e.Error())
		}
	}
	printErr(GetError(true, 1))
	printErr(GetError(true, 2))
	printErr(GetError(false, 0))

	// Output:
	// BizError1: (1) hello
	// BizError2: (2) hello
	// non-BizError: other error
}

func GetError(biz bool, code int) error {
	if biz {
		return errx.NewBizError(code, "hello", nil)
	}
	return errors.New("other error")
}
