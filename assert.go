// Copyright 2017 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

// Package assert provides developer a way to assert expression and output useful contextual information automatically when a case fails.
// With this package, we can focus on writing test code without worrying about how to print lots of verbose debug information for debug.
//
// See project page for more samples.
// https://github.com/huandu/go-assert
package assert

import (
	"testing"

	"github.com/huandu/go-assert/internal/assertion"
)

// Assert tests expr and call `t.Fatalf` to terminate test case if expr is false-equivalent value.
// `false`, 0, nil and empty string are false-equivalent values.
//
// Sample code.
//
//     import "github.com/huandu/go-assert"
//
//     func TestSomething(t *testing.T) {
//         a, b := 1, 2
//         assert.Assert(t, a > b)
//     }
//
// Output:
//
//     Assertion failed:
//         a > b
//     Referenced variables are assigned in following statements:
//         a, b := 1, 2
func Assert(t *testing.T, expr interface{}) {
	assertion.Assert(t, expr, &assertion.Trigger{
		FuncName: "Assert",
		Skip:     1,
		Args:     []int{1},
	})
}

// Equal uses `reflect.DeepEqual` to test v1 and v2 equality.
//
// Sample code.
//
//     import "github.com/huandu/go-assert"
//
//     func TestSomething(t *testing.T) {
//         assert.Equal(t, []int{1,2}, []int{1})
//     }
//
// Output:
//
//     Assertion failed:
//         assert.Equal(t, []int{1, 2}, []int{1})
//     The value of following expression should equal.
//     [1] []int{1, 2}
//     [2] []int{1}
//     Values:
//     [1] -> ([]int)[1 2]
//     [2] -> ([]int)[1]
func Equal(t *testing.T, v1, v2 interface{}) {
	assertion.AssertEqual(t, v1, v2, &assertion.Trigger{
		FuncName: "Equal",
		Skip:     1,
		Args:     []int{1, 2},
	})
}

// NotEqual uses `reflect.DeepEqual` to test v1 and v2 equality.
//
// Sample code.
//
//     import "github.com/huandu/go-assert"
//
//     func TestSomething(t *testing.T) {
//         assert.NotEqual(t, []int{1}, []int{1})
//     }
//
// Output:
//
//     Assertion failed:
//         assert.NotEqual(t, []int{1}, []int{1})
//     The value of following expression should not equal.
//     [1] []int{1}
//     [2] []int{1}
func NotEqual(t *testing.T, v1, v2 interface{}) {
	assertion.AssertNotEqual(t, v1, v2, &assertion.Trigger{
		FuncName: "NotEqual",
		Skip:     1,
		Args:     []int{1, 2},
	})
}

// AssertEqual uses `reflect.DeepEqual` to test v1 and v2 equality.
//
// Note: as golint dislike the name of this function,
// it will be removed in the future. Use Equal instead.
//
// Sample code.
//
//     import "github.com/huandu/go-assert"
//
//     func TestSomething(t *testing.T) {
//         assert.AssertEqual(t, []int{1,2}, []int{1})
//     }
//
// Output:
//
//     Assertion failed:
//         assert.AssertEqual(t, []int{1, 2}, []int{1})
//     The value of following expression should equal.
//     [1] []int{1, 2}
//     [2] []int{1}
//     Values:
//     [1] -> ([]int)[1 2]
//     [2] -> ([]int)[1]
func AssertEqual(t *testing.T, v1, v2 interface{}) {
	assertion.AssertEqual(t, v1, v2, &assertion.Trigger{
		FuncName: "AssertEqual",
		Skip:     1,
		Args:     []int{1, 2},
	})
}

// AssertNotEqual uses `reflect.DeepEqual` to test v1 and v2 equality.
//
// Note: as golint dislike the name of this function,
// it will be removed in the future. Use NotEqual instead.
//
// Sample code.
//
//     import "github.com/huandu/go-assert"
//
//     func TestSomething(t *testing.T) {
//         assert.AssertNotEqual(t, []int{1}, []int{1})
//     }
//
// Output:
//
//     Assertion failed:
//         assert.AssertNotEqual(t, []int{1}, []int{1})
//     The value of following expression should not equal.
//     [1] []int{1}
//     [2] []int{1}
func AssertNotEqual(t *testing.T, v1, v2 interface{}) {
	assertion.AssertNotEqual(t, v1, v2, &assertion.Trigger{
		FuncName: "AssertNotEqual",
		Skip:     1,
		Args:     []int{1, 2},
	})
}
