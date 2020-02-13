# Package `assert` - assert macro for Go #

[![Build Status](https://travis-ci.org/huandu/go-assert.svg?branch=master)](https://travis-ci.org/huandu/go-assert)
[![GoDoc](https://godoc.org/github.com/huandu/go-assert?status.svg)](https://godoc.org/github.com/huandu/go-assert)

Package `assert` provides developer a way to assert expression and magically output lots of useful contextual information when a case fails.

For example, if we write `Assert(t, a > b)` when `a = 1` and `b = 2`, we can read `Assertion failed: a > b` in the failure message. The `a > b` is the expression evaluated in `Assert`.

Here is a quick sample.

```go
import . "github.com/huandu/go-assert"

func TestSomething(t *testing.T) {
    str := Foo(42)
    Assert(t, str == "expected")

    // This case fails with following message.
    //
    //     Assertion failed:
    //         str == "expected"
    //     Referenced variables are assigned in following statements:
    //         str := Foo(42)
}
```

## Import ##

Use `go get` to install this package.

    go get github.com/huandu/go-assert

Current stable version is `v1.*`. Old versions tagged by `v0.*` are obsoleted.

## Usage ##

### Assertions ###

If we just want to use functions like `Assert`, `AssertEqual` or `AssertNotEqual`, it's recommended to import this package as `.`.

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
    //     [1] -> (map[string]int)map[bar:-2 foo:1]
    //     [2] -> (map[string]int)map[bar:-2 foo:10000]
}
```

### Advanced assertion wrapper: type `A` ###

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

If we want to print variables referenced by assertion expression automatically, call `A#Use` to track variables.

```go
import "github.com/huandu/go-assert"

func TestSomething(t *testing.T) {
    a := assert.New(t)
    v1 := 123
    v2 := "wrong"
    v3 := 3.45
    a.Use(&v1, &v2, &v3)

    a.Assert(v1 == 123 && v2 == "right")

    // This case fails with following message.
    //
    //     Assertion failed:
    //         v1 == 123 && v2 == "right"
    //     Referenced variables are assigned in following statements:
    //         v1 := 123
    //         v2 := "wrong"
    //     Referenced variables:
    //         v1 -> (int)123
    //         v2 -> (string)wrong
}
```