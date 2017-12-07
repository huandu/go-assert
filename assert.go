// Copyright 2017 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

// Package assert provides API to implement C-like assert macro.
package assert

import (
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
