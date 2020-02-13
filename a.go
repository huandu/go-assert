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
// Usage:
//
//     import "github.com/huandu/go-assert"
//
//     func TestSomething(t *testing.T) {
//         a := assert.New(t)
//         x, y := 1, 2
//         a.Assert(t, x > y) // This case fails with message "Assertion failed: x > y".
//     }
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
// Usage:
//
//     import "github.com/huandu/go-assert"
//
//     func TestSomething(t *testing.T) {
//         a := assert.New(t)
//         a.NilError(os.Open("path/to/a/file")) // This case fails if os.Open returns error.
//     }
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
// Usage:
//
//     import "github.com/huandu/go-assert"
//
//     func TestSomething(t *testing.T) {
//         a := assert.New(t)
//         f := func() (int, error) { return 0, errors.New("expected") }
//         a.NilError(f()) // This case fails.
//     }
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
// Usage:
//
//     import "github.com/huandu/go-assert"
//
//     func TestSomething(t *testing.T) {
//         a := assert.New(t)
//         a.Equal([]int{1,2}, []int{1})
//
//         // This case fails with message:
//         //     Assertion failed: []int{1,2} == []int{1}
//         //         v1 = [1 2]
//         //         v2 = [1]
//     }
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
// Usage:
//
//     import "github.com/huandu/go-assert"
//
//     func TestSomething(t *testing.T) {
//         a := assert.New(t)
//         a.NotEqual(t, []int{1}, []int{1})
//
//         // This case fails with message:
//         //     Assertion failed: []int{1} != []int{1}
//     }
func (a *A) NotEqual(v1, v2 interface{}) {
	assertion.AssertNotEqual(a.T, v1, v2, &assertion.Trigger{
		FuncName: "NotEqual",
		Skip:     1,
		Args:     []int{0, 1},
		Vars:     a.vars,
	})
}

// Use saves args in context and prints related args automatically in assertion method.
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
