// Copyright 2017 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

// Package assert provides API to implement C-like assert macro.
package assert

import (
	"testing"

	"github.com/huandu/go-assert/internal/assertion"
)

// Assert tests expr and call `t.Fatalf` to terminate test case if expr is false-equivalent value.
// `false`, 0, nil and empty string are false-equivalent values.
//
// Usage:
//
//     import . "github.com/huandu/go-assert"
//
//     func TestSomething(t *testing.T) {
//         a, b := 1, 2
//         Assert(t, a > b) // This case fails with message "Assertion failed: a > b".
//     }
func Assert(t *testing.T, expr interface{}) {
	assertion.Assert(t, expr, &assertion.Trigger{
		FuncName: "Assert",
		Skip:     1,
		Args:     []int{1},
	})
}

// AssertEqual uses `reflect.DeepEqual` to test v1 and v2 equality.
//
// Usage:
//
//     import . "github.com/huandu/go-assert"
//
//     func TestSomething(t *testing.T) {
//         AssertEqual(t, []int{1,2}, []int{1})
//
//         // This case fails with message:
//         //     Assertion failed: []int{1,2} == []int{1}
//         //         v1 = [1,2]
//         //         v2 = [1]
//     }
func AssertEqual(t *testing.T, v1, v2 interface{}) {
	assertion.AssertEqual(t, v1, v2, &assertion.Trigger{
		FuncName: "AssertEqual",
		Skip:     1,
		Args:     []int{1, 2},
	})
}

// AssertNotEqual uses `reflect.DeepEqual` to test v1 and v2 equality.
//
// Usage:
//
//     import . "github.com/huandu/go-assert"
//
//     func TestSomething(t *testing.T) {
//         AssertNotEqual(t, []int{1}, []int{1})
//
//         // This case fails with message:
//         //     Assertion failed: []int{1} != []int{1}
//     }
func AssertNotEqual(t *testing.T, v1, v2 interface{}) {
	assertion.AssertNotEqual(t, v1, v2, &assertion.Trigger{
		FuncName: "AssertNotEqual",
		Skip:     1,
		Args:     []int{1, 2},
	})
}
