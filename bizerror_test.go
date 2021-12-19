package errx

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBizError(t *testing.T) {
	got := NewBizError(123, "msg", errors.New("cause"))
	a := assert.New(t)
	a.NotNil(got)
	a.Equal(123, got.Code())
	a.Equal("cause", got.Cause().Error())
	a.Equal("msg", got.Message())
	a.Equal("(123) msg", got.Error())

	// 调用栈不好判断，用正则匹配。堆栈有换行，需给定 s flag 。
	a.Regexp(`(?s)goroutine.+testing\.go`, got.Stack())
}

func TestNewBizErrorWithoutStack(t *testing.T) {
	got := NewBizErrorWithoutStack(123, "msg", errors.New("cause"))
	a := assert.New(t)
	a.NotNil(got)
	a.Equal(123, got.Code())
	a.Equal("cause", got.Cause().Error())
	a.Equal("msg", got.Message())
	a.Equal("(123) msg", got.Error())
	a.Equal("", got.Stack())
}
