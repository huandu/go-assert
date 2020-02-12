# Package `assert` - assert macro for Go #

[![Build Status](https://travis-ci.org/huandu/go-assert.svg?branch=master)](https://travis-ci.org/huandu/go-assert)
[![GoDoc](https://godoc.org/github.com/huandu/go-assert?status.svg)](https://godoc.org/github.com/huandu/go-assert)

Package `assert` provides developer a way to assert expression and print expression source code in test cases. It works like C macro `assert`. When assertion fails, source code of the expression in assert function is printed.

For example, if we write `Assert(t, a > b)` when `a = 1` and `b = 2`, we can read `Assertion failed: a > b` in the failure message. The `a > b` is the expression evaluated in `Assert`.

Without this package, developers must use negate logic to test expressions and call `t.Fatalf` with lots of redundant information to print meaningful failure message for debugging. Just like following.

```go
func TestSomething(t *testing.T) {
    str := Foo(42)

    // We expect `str` to be "expected". To verify it, we need to use negate logic.
    // It's not straight forward.
    if str != "expected" {
        // We have to write some messages to let us know what's called and why it fails.
        t.Fatalf("invalid str when calling Foo(42). [str:%v] [expected:%v]", str, "expected")
    }
}
```

With this package, we can significantly simplify test code which works similar as above.

```go
import . "github.com/huandu/go-assert"

func TestSomething(t *testing.T) {
    str := Foo(42)
    Assert(t, str == "expected")

    // This case fails with following message.
    //
    //     Assertion failed:
    //         str == "expected"
    
    // If we're aware of the value of str, use AssertEqual.
    AssertEqual(t, str, "expected")
    
    // This case fails with following message with lots of useful information.
    //
    //     Assertion failed:
    //     The value of following expression should equal.
    //     [1] str
    //         str := Foo(42)
    //     [2] "expected"
    //     Values:
    //     [1] = (string)actual
    //     [2] = (string)expected
}
```

## Import ##

Use `go get` to install this package.

    go get github.com/huandu/go-assert

Current stable version is `v1.*`. Old versions tagged by `v0.*` are obsoleted.

## Usage ##

If we just want to use functions like `Assert` or `AssertEqual`, it's recommended to import this package as `.`.

```go
import . "github.com/huandu/go-assert"

func TestSomething(t *testing.T) {
    a, b := 1, 2
    Assert(t, a > b)
    
    // This case fails with message:
    //     Assertion failed:
    //         a > b
}

func TestAssertEquality(t *testing.T) {
    AssertEqual(t, map[string]int{
        "foo": 1,
        "bar": -2,
    }, map[string]int{
        "bar": -2,
        "foo": 10000,
    })
    
    // This case fails with message:
    //     Assertion failed:
    //     The value of following expression should equal.
    //     [1] map[string]int{
    //             "foo": 1,
    //             "bar": -2,
    //         }
    //     [2] map[string]int{
    //             "bar": -2,
    //             "foo": 10000,
    //         }
    //     Values:
    //     [1] = (map[string]int)map[bar:-2 foo:1]
    //     [2] = (map[string]int)map[bar:-2 foo:10000]
}
```

If we want more controls on assertion, it's recommended to wrap `t` in an `A`.
One huge benifit of using `A` is that it can assert a function call directly like following.

```go
import "github.com/huandu/go-assert"

func TestCallAFunction(t *testing.T) {
    a := assert.New(t)

    f := func(bool, int) (int, string, error) {
        return 0, "", errors.New("an error")
    }
    a.NilError(f(true, 42)) // Assert calls to f to make test code more readable.
    
    // This case fails with message:
    //     Assertion failed:
    //     Following expression should return a nil error.
    //         f(true, 42)
    //     The error is:
    //         an error
}
```
