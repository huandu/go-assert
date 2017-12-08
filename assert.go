// Copyright 2017 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

// Package assert provides API to implement C-like assert macro.
package assert

import (
	"reflect"
	"testing"

	"github.com/huandu/go-assert/assertion"
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
	k := assertion.ParseFalseKind(expr)

	if k == assertion.Positive {
		return
	}

	assertion.TriggerAssertion(t, k, "Assert", 1)
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
//         //     Assertion failed: []int{1,2} != []int{1}
//     }
func AssertEqual(t *testing.T, v1, v2 interface{}) {
	typeMismatch := false
	t1 := reflect.TypeOf(v1)
	t2 := reflect.TypeOf(v2)

	if !t1.AssignableTo(t2) && !t2.AssignableTo(t1) {
		typeMismatch = true
	}

	if reflect.DeepEqual(v1, v2) {
		return
	}

	args, err := assertion.ParseArgs("AssertEqual", 1, 1, 2)

	if err != nil {
		t.Fatalf("Assertion failed with an internal error: %v", err)
		return
	}

	if typeMismatch {
		t.Fatalf("Assertion failed: %v and %v type mismatch.", args[0], args[1])
	} else {
		t.Fatalf("Assertion failed: %v != %v", args[0], args[1])
	}
}
