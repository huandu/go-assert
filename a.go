// Copyright 2017 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package assert

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"reflect"
	"testing"

	"github.com/huandu/go-assert/internal/assertion"
)

// The A is a wrapper of testing.T with some extra help methods.
type A struct {
	*testing.T

	vars map[string]interface{}
}

// New creates an assertion object wraps t.
func New(t *testing.T) *A {
	return &A{
		T:    t,
		vars: make(map[string]interface{}),
	}
}

// Assert tests expr and call `t.Fatalf` to terminate test case if expr is false-equivalent value.
// `false`, 0, nil and empty string are false-equivalent values.
//
// Sample code.
//
//     func TestSomething(t *testing.T) {
//         a := assert.New(t)
//         x, y := 1, 2
//         a.Assert(x > y)
//     }
//
// Output:
//
//     Assertion failed:
//         x > y
//     Referenced variables are assigned in following statements:
//         x, y := 1, 2
func (a *A) Assert(expr interface{}) {
	assertion.Assert(a.T, expr, &assertion.Trigger{
		FuncName: "Assert",
		Skip:     1,
		Args:     []int{0},
		Vars:     a.vars,
	})
}

// NilError expects a function return a nil error.
// Otherwise, it will terminate the test case using `t.Fatalf`.
//
// Sample code.
//
//     func TestSomething(t *testing.T) {
//         a := assert.New(t)
//         a.NilError(os.Open("path/to/a/file"))
//     }
//
// Output:
//
//     Assertion failed:
//     Following expression should return a nil error.
//         os.Open("path/to/a/file")
//     The error is:
//         open path/to/a/file: no such file or directory
func (a *A) NilError(result ...interface{}) {
	assertion.AssertNilError(a.T, result, &assertion.Trigger{
		FuncName: "NilError",
		Skip:     1,
		Args:     []int{-1},
		Vars:     a.vars,
	})
}

// NonNilError expects a function return a non-nil error.
// Otherwise, it will terminate the test case using `t.Fatalf`.
//
// Sample code.
//
//     func TestSomething(t *testing.T) {
//         a := assert.New(t)
//         f := func() (int, error) { return 0, errors.New("expected") }
//         a.NilError(f())
//     }
//
// Output:
//
//     Assertion failed:
//     Following expression should return a nil error.
//         f()
//         f := func() (int, error) { return 0, errors.New("expected") }
//     The error is:
//         expected
func (a *A) NonNilError(result ...interface{}) {
	assertion.AssertNonNilError(a.T, result, &assertion.Trigger{
		FuncName: "NonNilError",
		Skip:     1,
		Args:     []int{-1},
		Vars:     a.vars,
	})
}

// Equal uses `reflect.DeepEqual` to test v1 and v2 equality.
//
// Sample code.
//
//     func TestSomething(t *testing.T) {
//         a := assert.New(t)
//         a.Equal([]int{1,2}, []int{1})
//     }
//
// Output:
//
//     Assertion failed:
//         a.Equal([]int{1, 2}, []int{1})
//     The value of following expression should equal.
//     [1] []int{1, 2}
//     [2] []int{1}
//     Values:
//     [1] -> ([]int)[1 2]
//     [2] -> ([]int)[1]
func (a *A) Equal(v1, v2 interface{}) {
	assertion.AssertEqual(a.T, v1, v2, &assertion.Trigger{
		FuncName: "Equal",
		Skip:     1,
		Args:     []int{0, 1},
		Vars:     a.vars,
	})
}

// NotEqual uses `reflect.DeepEqual` to test v1 and v2 equality.
//
// Sample code.
//
//     func TestSomething(t *testing.T) {
//         a := assert.New(t)
//         a.NotEqual(t, []int{1}, []int{1})
//     }
//
// Output:
//
//     Assertion failed:
//         a.NotEqual(t, []int{1}, []int{1})
//     The value of following expression should not equal.
//     [1] []int{1}
//     [2] []int{1}
func (a *A) NotEqual(v1, v2 interface{}) {
	assertion.AssertNotEqual(a.T, v1, v2, &assertion.Trigger{
		FuncName: "NotEqual",
		Skip:     1,
		Args:     []int{0, 1},
		Vars:     a.vars,
	})
}

// Use saves args in context and prints related args automatically in assertion method when referenced.
//
// Sample code.
//
//     func TestSomething(t *testing.T) {
//         a := assert.New(t)
//         v1 := 123
//         v2 := []string{"wrong", "right"}
//         v3 := v2[0]
//         v4 := "not related"
//         a.Use(&v1, &v2, &v3, &v4)
//     }
//
// Output:
//
//     Assertion failed:
//         v1 == 123 && v3 == "right"
//     Referenced variables are assigned in following statements:
//         a.Use(&v1, &v2, &v3, &v4)
//     Related variables:
//         v1 = (int)123
//         v3 = (string)wrong
func (a *A) Use(args ...interface{}) {
	if len(args) == 0 {
		return
	}

	argIndex := make([]int, 0, len(args))
	values := make([]interface{}, 0, len(args))

	for i := range args {
		if args[i] == nil {
			continue
		}

		val := reflect.ValueOf(args[i])

		if val.Kind() != reflect.Ptr {
			continue
		}

		val = val.Elem()

		if !val.IsValid() {
			continue
		}

		argIndex = append(argIndex, i)
		values = append(values, args[i])
	}

	if len(argIndex) == 0 {
		return
	}

	f, err := assertion.ParseArgs("Use", 1, argIndex)

	if err != nil {
		return
	}

	for i, arg := range f.Args {
		// Arg must be something like `&a` or `&a.b`.
		// Otherwise, ignore the arg.
		expr, ok := arg.(*ast.UnaryExpr)
		if !ok || expr.Op != token.AND {
			continue
		}

		if !assertion.IsVar(expr.X) {
			continue
		}

		buf := &bytes.Buffer{}
		printer.Fprint(buf, f.FileSet, expr.X)
		a.vars[buf.String()] = values[i]
	}
}
