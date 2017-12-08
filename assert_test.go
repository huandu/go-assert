// Copyright 2017 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package assert

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/huandu/go-assert/assertion"
)

// TestMain hacks the testing process and runs cases only if flag -test.run is specified.
// With this hack, one can run selected case, which will always fail due to the
// nature of this package, without breaking travis-ci system, which expects all cases passing.
func TestMain(m *testing.M) {
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-test.run") {
			os.Exit(m.Run())
			return
		}
	}
}

func TestAssertCompareExpr(t *testing.T) {
	a, b := 1, 2
	Assert(t, a > b)
}

func TestAssertNilErrorFunctionCall(t *testing.T) {
	a := assertion.New(t)
	f := func(string, int) (float32, bool, error) {
		return 12, true, nil
	}
	a.NilError(f("should pass", 0))

	f = func(string, int) (float32, bool, error) {
		return 0, false, errors.New("expected")
	}
	a.NilError(f("should fail", 42))
}

func TestAssertNonNilErrorFunctionCall(t *testing.T) {
	a := assertion.New(t)
	f := func(string, int) (float32, bool, error) {
		return 12, true, errors.New("should pass")
	}
	a.NonNilError(f("should pass", 0))

	f = func(string, int) (float32, bool, error) {
		return 0, false, nil
	}
	a.NonNilError(f("should fail", 42))
}

func TestAssertEquality(t *testing.T) {
	AssertEqual(t, map[string]int{
		"foo": 1,
		"bar": -2,
	}, map[string]int{
		"bar": -2,
		"foo": 1,
	})

	AssertEqual(t, map[string]int{
		"foo": 1,
		"bar": -2,
	}, map[string]int{
		"bar": -2,
		"foo": 10000,
	})
}
