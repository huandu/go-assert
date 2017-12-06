# Package `assert` - assert macro for Go #

[![Build Status](https://travis-ci.org/huandu/go-assert.png?branch=master)](https://travis-ci.org/huandu/go-assert)
[![GoDoc](https://godoc.org/github.com/huandu/go-assert?status.svg)](https://godoc.org/github.com/huandu/go-assert)

Package `assert` provides developer a way to assert expression and print expression source code in test cases. It works like C macro `assert`.

## Install ##

Use `go get` to install this package.

    go get -u github.com/huandu/go-assert

## Use it ##

Simply use it in a test file.

```go
import . "github.com/huandu/go-assert"

func TestSomething(t *testing.T) {
    a, b := 1, 2
    Assert(t, a > b) // This case fails with "Assertion failed: a > b".
}
```
