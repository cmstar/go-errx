package errx

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
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

		{
			name:      "nested-biz",
			args:      args{"pre1", NewBizError(100, "biz", errors.New("e"))},
			wantError: "pre1: (100) biz",
			wantStack: []string{`errx_test\.go`, `testing\.go`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 两个函数极为接近，放一起测试。
			for _, withStack := range []bool{true, false} {
				var got *ErrorWrapper
				if withStack {
					got = Wrap(tt.args.msg, tt.args.cause).(*ErrorWrapper)
				} else {
					got = WrapWithoutStack(tt.args.msg, tt.args.cause).(*ErrorWrapper)
				}

				a := require.New(t)
				a.NotNil(got)
				a.Equal(got.ErrorCause.Err, got.Unwrap(), "ErrorCause.Unwrap()")
				a.Equal(got.ErrorCause.Err, got.Cause(), "ErrorCause.Cause()")

				if tt.args.cause == nil {
					a.Nil(got.Cause(), "Cause()")
				} else {
					a.Equal(tt.args.cause.Error(), got.Cause().Error(), "Cause().Error()")
				}

				a.Equal(tt.wantError, got.ErrorWithoutStack(), "ErrorWithoutStack()")

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
	check := func(t *testing.T, e error, patterns []string) {
		res := Describe(e)

		for _, p := range patterns {
			require.Regexp(t, p, res)
		}
	}

	t.Run("nil", func(t *testing.T) {
		check(t, nil, []string{"^$"})
	})

	t.Run("nostack", func(t *testing.T) {
		check(t, errors.New("gg"), []string{"gg"})
	})

	t.Run("unwrapable", func(t *testing.T) {
		check(t,
			fmt.Errorf("pre1: %w", fmt.Errorf("pre2: %w", errors.New("inner"))),
			[]string{
				`^pre1: pre2: inner\n=== pre2: inner\n=== inner\n`,
			})
	})

	t.Run("stackful", func(t *testing.T) {
		check(t,
			Wrap("pre1", Wrap("pre2", errors.New("gg"))),
			[]string{
				`^pre1: pre2: gg\n--- \[.+errx_test\.go:\d+\]`,
				`=== pre2: gg\n--- \[.+errx_test\.go:\d+\]`,
				`=== gg`,
			})
	})

	t.Run("wrap-biz", func(t *testing.T) {
		check(t,
			Wrap("pre1", NewBizError(100, "biz", errors.New("gg"))),
			[]string{
				`^pre1: \(100\) biz\n--- \[.+errx_test\.go:\d+\]`,
				`=== \(100\) biz\n--- \[.+errx_test\.go:\d+\]`,
				`=== gg`,
			})
	})

	t.Run("biz-wrap", func(t *testing.T) {
		check(t,
			NewBizError(100, "biz", Wrap("inner", errors.New("gg"))),
			[]string{
				`^\(100\) biz\n--- \[.+errx_test\.go:\d+\]`,
				`=== inner: gg\n--- \[.+errx_test\.go:\d+\]`,
				`=== gg`,
			})
	})
}

func TestErrorWrapper_ErrorWithoutStack(t *testing.T) {
	t.Run("p1", func(t *testing.T) {
		w := Wrap("p1", errors.New(``))
		require.Equal(t, `p1: `, w.ErrorWithoutStack())
	})

	t.Run("p1-e", func(t *testing.T) {
		w := Wrap("p1", nil)
		require.Equal(t, `p1`, w.ErrorWithoutStack())
	})

	t.Run("p1-e", func(t *testing.T) {
		w := Wrap("p1", errors.New(`e`))
		require.Equal(t, `p1: e`, w.ErrorWithoutStack())
	})

	t.Run("p1-p2-e", func(t *testing.T) {
		w := Wrap("p1", Wrap("p2", errors.New(`e`)))
		require.Equal(t, `p1: p2: e`, w.ErrorWithoutStack())
	})

	t.Run("e", func(t *testing.T) {
		w := Wrap("", errors.New(`e`))
		require.Equal(t, `e`, w.ErrorWithoutStack())
	})

	t.Run("p2-e", func(t *testing.T) {
		w := Wrap("", Wrap("p2", errors.New(`e`)))
		require.Equal(t, `p2: e`, w.ErrorWithoutStack())
	})

	t.Run("p1--e", func(t *testing.T) {
		w := Wrap("p1", Wrap("", errors.New(`e`)))
		require.Equal(t, `p1: e`, w.ErrorWithoutStack())
	})
}

func TestErrorWrapper_Format(t *testing.T) {
	w := Wrap("pre1", Wrap("pre2", errors.New(`"gg"`)))

	got := fmt.Sprintf("%s", w)
	require.Equal(t, `pre1: pre2: "gg"`, got)

	got = fmt.Sprintf("%q", w)
	require.Equal(t, `"pre1: pre2: \"gg\""`, got)

	got = fmt.Sprintf("%v", w)
	require.Regexp(t, `^pre1: pre2: "gg"\n--- \[.+errx_test\.go:\d+\]`, got)

	got = fmt.Sprintf("%+v", w)
	require.Regexp(t, `^pre1: pre2: "gg"\n--- \[.+errx_test\.go:\d+\]`, got)

	got = fmt.Sprintf("%d", w)
	require.Equal(t, `BADFORMAT:pre1: pre2: "gg"`, got)
}

func TestPreserveRecover(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		err := func() (err error) {
			defer func() {
				err = PreserveRecover("prefix", recover())
			}()
			return
		}()

		require.Equal(t, "", Describe(err))
	})

	t.Run("stackful", func(t *testing.T) {
		err := func() (err error) {
			defer func() {
				err = PreserveRecover("prefix", recover())
			}()

			panic(fmt.Errorf("msg"))
		}()

		require.Regexp(t, `^prefix: msg\n--- \[`, Describe(err))
	})

	t.Run("string", func(t *testing.T) {
		err := func() (err error) {
			defer func() {
				err = PreserveRecover("prefix", recover())
			}()

			panic("gg")
		}()

		require.Regexp(t, `^prefix: gg\n--- \[`, Describe(err))
	})

	t.Run("other", func(t *testing.T) {
		err := func() (err error) {
			defer func() {
				err = PreserveRecover("prefix", recover())
			}()

			panic(99)
		}()

		require.Regexp(t, `^prefix: 99\n--- \[`, Describe(err))
	})
}

func TestRun(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		err := Run(func() {})
		require.Nil(t, err)
	})

	t.Run("panic", func(t *testing.T) {
		err := Run(func() { panic("some error") })
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "some error")
	})
}

func TestRunE(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		err := RunE(func() error { return nil })
		require.Nil(t, err)
	})

	t.Run("panic", func(t *testing.T) {
		err := RunE(func() error { panic("some error") })
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "some error")
	})

	t.Run("err", func(t *testing.T) {
		err := RunE(func() error { return errors.New("some error") })
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "some error")
	})
}
