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

func TestAssertIdent(t *testing.T) {
	a := 0
	Assert(t, a)
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

func TestAssertEqualityTypeMismatch(t *testing.T) {
	v1 := struct {
		Foo string
		Bar int
	}{"should pass", 1}
	v2 := struct {
		Foo string
		Bar int
	}{"should pass", 1}
	AssertEqual(t, v1, v2)

	v3 := []int{1, 2, 3}
	v4 := []int64{1, 2, 3}
	AssertEqual(t, v3, v4)
}

func TestAssertEqualityWithAssertion(t *testing.T) {
	a := assertion.New(t)
	a.Equal(map[string]int{
		"foo": 1,
		"bar": -2,
	}, map[string]int{
		"bar": -2,
		"foo": 1,
	})

	a.Equal(map[string]int{
		"foo": 1,
		"bar": -2,
	}, map[string]int{
		"bar": -2,
		"foo": 10000,
	})
}

func TestAssertEqualityTypeMismatchWithAssertion(t *testing.T) {
	a := assertion.New(t)
	v1 := struct {
		Foo string
		Bar int
	}{"should pass", 1}
	v2 := struct {
		Foo string
		Bar int
	}{"should pass", 1}
	a.Equal(v1, v2)

	v3 := []int{1, 2, 3}
	v4 := []int64{1, 2, 3}
	a.Equal(v3, v4)
}

func TestAssertNotEqual(t *testing.T) {
	v1 := struct {
		Foo string
		Bar int
	}{"should pass", 1}
	v2 := struct {
		Bar int
		Foo string
	}{1, "should pass"}
	AssertNotEqual(t, v1, v2)

	v3 := []int{1, 2, 3}
	v4 := []int{1, 2, 3}
	AssertNotEqual(t, v3, v4)
}

func TestAssertNotEqualWithAssertion(t *testing.T) {
	a := assertion.New(t)
	v1 := struct {
		Foo string
		Bar int
	}{"should pass", 1}
	v2 := struct {
		Bar int
		Foo string
	}{1, "should pass"}
	a.NotEqual(v1, v2)

	v3 := []int{1, 2, 3}
	v4 := []int{1, 2, 3}
	a.NotEqual(v3, v4)
}
