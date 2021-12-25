package errx

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Wrap 和 WrapWithoutStack 极为相似，放在一起测。
func TestWrapAndWrapWithoutStack(t *testing.T) {
	type args struct {
		msg   string
		cause error
	}
	tests := []struct {
		name      string
		args      args
		wantError string
		wantStack []string // 调用栈不好组装，给定一组特征，通过一组正则检验。如果是空的表示调用栈也是空的。
	}{
		{
			name:      "empty",
			args:      args{"", nil},
			wantError: "",
			wantStack: []string{},
		},

		{
			name:      "only-stack",
			args:      args{"", nil},
			wantError: "",
			wantStack: []string{`errx_test\.go`, `testing\.go`},
		},

		{
			name:      "no-prefix",
			args:      args{"", errors.New("cause")},
			wantError: "cause",
			wantStack: []string{`errx_test\.go`, `testing\.go`},
		},

		{
			name:      "no-cause",
			args:      args{"pre", nil},
			wantError: "pre",
			wantStack: []string{`errx_test\.go`, `testing\.go`},
		},

		{
			name:      "nested-stack",
			args:      args{"pre1", Wrap("pre2", errors.New("e"))},
			wantError: "pre1: pre2: e",
			wantStack: []string{`errx_test\.go`, `testing\.go`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 两个函数极为接近，放一起测试。
			for _, withStack := range []bool{true, false} {
				var got *ErrorWrapper
				if withStack {
					got = Wrap(tt.args.msg, tt.args.cause)
				} else {
					got = WrapWithoutStack(tt.args.msg, tt.args.cause)
				}

				a := assert.New(t)
				a.NotNil(got)
				a.Equal(got.ErrorCause.Err, got.Unwrap(), "ErrorCause.Unwrap()")
				a.Equal(got.ErrorCause.Err, got.Cause(), "ErrorCause.Cause()")

				if tt.args.cause == nil {
					a.Nil(got.Cause(), "Cause()")
				} else {
					a.Equal(tt.args.cause.Error(), got.Cause().Error(), "Cause().Error()")
				}

				a.Equal(tt.wantError, got.Error(), "Error()")

				if withStack {
					for _, v := range tt.wantStack {
						matched, err := regexp.MatchString(v, got.Stack())
						a.NoError(err)
						a.True(matched, "stack must match '%v', got '%v'", v, got.Stack())
					}
				} else {
					a.Equal("", got.Stack())
				}
			}
		})
	}
}

func TestDescribe(t *testing.T) {
	tests := []struct {
		name     string
		e        error
		patterns []string
	}{
		{
			"nil",
			nil,
			[]string{"^$"},
		},

		{
			"nostack",
			errors.New("gg"),
			[]string{"gg"},
		},

		{
			"unwrapable",
			fmt.Errorf("pre1: %w", fmt.Errorf("pre2: %w", errors.New("inner"))),
			[]string{
				`^pre1: pre2: inner\n=== pre2: inner\n=== inner$`,
			},
		},

		{
			"stackful",
			Wrap("pre1", Wrap("pre2", errors.New("gg"))),
			[]string{
				`^pre1: pre2: gg\n--- \[.+errx_test\.go:\d+\]`,
				`=== pre2: gg\n--- \[.+errx_test\.go:\d+\]`,
				`=== gg`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Describe(tt.e)

			for _, p := range tt.patterns {
				assert.Regexp(t, p, res)
			}
		})
	}
}
