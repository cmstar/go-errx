# errx - 更精准的定位和跟踪错误

[![GoDoc](https://pkg.go.dev/badge/github.com/cmstar/go-errx)](https://pkg.go.dev/github.com/cmstar/go-errx)
[![Go](https://github.com/cmstar/go-errx/workflows/Go/badge.svg)](https://github.com/cmstar/go-errx/actions?query=workflow%3AGo)
[![codecov](https://codecov.io/gh/cmstar/go-errx/branch/master/graph/badge.svg)](https://codecov.io/gh/cmstar/go-errx)
[![License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat)](https://opensource.org/licenses/MIT)

功能：
- 封装引起错误的错误。
- 为错误追加方法调用栈信息。
- 业务预定义异常 `BizError` 。
- 用于处理 recover() 结果的 `PreserveRecover` 方法。

安装：
```
go get -u github.com/cmstar/go-errx@latest
```

## 调用栈和错误链

Go 程序通常小而精，更多的用于中间件和系统编程，但有时仍会被用在上层的复杂业务里，和 Java 、 .net 同台，此时 Go 的错误处理模式就变得捉襟见肘起来。也许这就不是 Go 的设计意图，但真实的编码场景里，我们免不了碰到这样的情况。

Go 的 error 只是“不太特殊”的值而已（[Errors are values](https://go.dev/blog/errors-are-values)），它太过于普通以至于不能像 Java/.net 的 `Exception` 一样携带足够多的信息。

在复杂的上层业务中，快速定位错误的位置显得极为重要，有时可以用运行性能换工作效率。

---

### Wrap 方法

`errx.Wrap` 为错误添加更多的细节，返回一个 `StackfulError` 接口的实现，它包含方法：
- Cause：记录引起错误的错误，类似 Java 的 `Throwable.getCause()` 或 .net 的 `Exception.InnerException` 。
- Stack：记录创建错误（即调用 Wrap 方法）时的方法调用栈，类似 Java 的 `Exception.printStackTrace()` 或 .net 的 `Exception.StackTrace` 。
- Error：即标准库的 `error.Error()` ，但它包含了 `Cause` 和 `Stack` 格式化后的信息。输出格式见下文《Describe 方法》。
- ErrorWithoutStack：只包含错误的描述，不包含 `Stack` 的信息。

> 调用栈信息使用标准库的 `runtime.CallersFrames` 方法获取，有一定的性能开销。

## BizError

在业务交互中，我们可能需要根据错误的类别进行不同的处理，原始的 `error` 等同于一个字符串，难以判断和分类。 errx 包定义了 `BizError` ，以便对错误进行分类。它是一个特殊的 `error` ，可通过 `errx.NewBizError` 方法创建。

`BizError` 提供：
- 为每个错误添加一个整数型的错误码 `Code` ，以便更精准的对错误进行分类和定位，特别是在日志搜索时。通常0表示没有错误，其余值表示有错误，值由具体业务指定。
- 包含调用栈信息 `Stack` 。
- 包含 `Cause` ，即引起此错误的错误。

`BizError.Error()` 返回值格式为： `(Code) Message` ，不包含 `Cause` 和 `Stack` 。

`BizError` 的使用样例可参考 [GoDoc 示例](https://pkg.go.dev/github.com/cmstar/go-errx#example-BizError) 。

> [go-webapi](https://github.com/cmstar/go-webapi#%E9%94%99%E8%AF%AF%E5%A4%84%E7%90%86) 框架使用 `BizError` 区分需要返回的业务错误和其他内部错误。

### Describe 方法

`errx.Describe` 将 `errx.Wrap` 添加的信息抽取出来，形成一段完整的错误描述，它包含各层级错误的信息及调用栈。

格式为：
```
最外层错误描述
--- 最外层错误的调用栈信息
=== 第1层内部错误的描述
--- 第1层内部错误的调用栈信息
=== 第2层内部错误的描述
--- 第2层内部错误的调用栈信息
...（逐层展示）
=== 最内层错误的描述
--- 最内层错误的描述的调用栈信息
```

实际示例可参考 [GoDoc 示例](https://pkg.go.dev/github.com/cmstar/go-errx#example-package-ErrorChain) 。

---

总体而言，就是让 Go 的 `error` 更像 Java/.net 的 `Exception` 。

当一个 `error` 在 `Wrap` 之后返回给其调用者，调用者再次使用 `Wrap` 并返回给更上层的调用者， error 就形成了一个链条。

### PreserveRecover 方法

我们可能需要利用应对 `panic` ，并将相关的错误信息保留下来，代码如下：

```go
func do() (err error) {
    defer func() {
        r := recover()
        if r == nil {
            return
        }

        // 留存错误，丢失了调用栈信息。
        err = fmt.Errorf("do: %v", r)
    }()

    somethingThatMayPanic()
    return nil
}
```

利用 `PreserveRecover` 方法，这段代码可以简化成这样，当 `recover()` 的结果不是 `nil` 时，它会被 `Wrap()` 在一个错误里：

```go
func do() (err error) {
    defer func() {
        // 留存错误，并保留调用栈信息。若 recover() 结果为 nil ，则 err 也是 nil 。
        err = errx.PreserveRecover("do", recover())
    }()

    somethingThatMayPanic()
    return nil
}
```
