# Package `assert` - Magic assert macros for Go #

[![Build Status](https://travis-ci.org/huandu/go-assert.svg?branch=master)](https://travis-ci.org/huandu/go-assert)
[![GoDoc](https://godoc.org/github.com/huandu/go-assert?status.svg)](https://godoc.org/github.com/huandu/go-assert)

Package `assert` provides developer a way to assert expression and output useful contextual information automatically when a case fails.
With this package, we can focus on writing test code without worrying about how to print lots of verbose debug information for debug.

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

### Assertion methods ###

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

There are lots of useful assert methods implemented in `A`.

* [`Assert`](https://godoc.org/github.com/huandu/go-assert#A.Assert)/[`Eqaul`](https://godoc.org/github.com/huandu/go-assert#A.Equal)/[`NotEqual`](https://godoc.org/github.com/huandu/go-assert#A.NotEqual): Basic assertion methods.
* [`NilError`](https://godoc.org/github.com/huandu/go-assert#A.NilError)/[`NonNilError`](https://godoc.org/github.com/huandu/go-assert#A.NonNilError): Test if a func/method returns expected error.
* [`Use`](https://godoc.org/github.com/huandu/go-assert#A.Use): Track variables. If any assert method fails, all variables tracked by `A` and related in assert method will be printed out automatically in assertion message.

Here is a sample to demonstrate how to use `A#Use` to print related variables in assertion message.

```go
import "github.com/huandu/go-assert"

func TestSomething(t *testing.T) {
    a := assert.New(t)
    v1 := 123
    v2 := []string{"wrong", "right"}
    v3 := v2[0]
    v4 := "not related"
    a.Use(&v1, &v2, &v3, &v4)

    a.Assert(v1 == 123 && v3 == "right")

    // This case fails with following message.
    //
    //     Assertion failed:
    //         v1 == 123 && v3 == "right"
    //     Referenced variables are assigned in following statements:
    //         v1 := 123
    //         v3 := v2[0]
    //     Related variables:
    //         v1 -> (int)123
    //         v2 -> ([]string)[wrong right]
    //         v3 -> (string)wrong
}
```