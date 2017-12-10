# Package `assert` - assert macro for Go #

[![Build Status](https://travis-ci.org/huandu/go-assert.svg?branch=master)](https://travis-ci.org/huandu/go-assert)
[![GoDoc](https://godoc.org/github.com/huandu/go-assert?status.svg)](https://godoc.org/github.com/huandu/go-assert)

Package `assert` provides developer a way to assert expression and print expression source code in test cases. It works like C macro `assert`.

Without this package, developers must use negative logic to test expressions and call `t.Fatalf` to print meaningful failure message for debugging. Just like following.

```go
func TestSomething(t *testing.T) {
    str := "actual"

    // We want `str` to be "expected", but we have to negative logic to check it.
    // Obviously, it's not straight forward.
    if str != "expected" {
        // We have to write some messages to let us know what's called and why it fails.
        t.Fatalf("invalid str. [str:%v] [expected:%v]", str, "expected")
    }
}
```

With this package, we can significantly simplify test code which works similar as above.

```go
import . "github.com/huandu/go-assert"

func TestSomething(t *testing.T) {
    str := "actual"
    Assert(str == "expected")

    // This case fails with following message.
    //
    //     Assertion failed: str == "expected"
}
```

## Install ##

Use `go get` to install this package.

    go get -u github.com/huandu/go-assert

## Usage ##

Use it in a test file.

```go
import . "github.com/huandu/go-assert"

func TestSomething(t *testing.T) {
    a, b := 1, 2
    Assert(t, a > b)
    
    // This case fails with message:
    //     Assertion failed: a > b
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
    //     Assertion failed: map[string]int{
    //         "foo": 1,
    //         "bar": -2,
    //     } == map[string]int{
    //         "bar": -2,
    //         "foo": 10000,
    //     }
    //         v1 = map[foo:1 bar:-2]
    //         v2 = map[bar:-2 foo:10000]
}
```

One can wrap `t` in an `Assertion` to validate results returned by a function.

```go
import "github.com/huandu/go-assert/assertion"

func TestCallAFunction(t *testing.T) {
    a := assertion.New(t)

    f := func(bool, int) (int, string, error) {
        return 0, "", errors.New("an error")
    }
    a.NilError(f(true, 42))
    
    // This case fails with message:
    //     Assertion failed: f(true, 42) should return nil error.
    //         err = an error
}
```
