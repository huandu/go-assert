// Copyright 2017 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package assert

import (
	"errors"
	"os"
	"strings"
	"testing"
)

// TestMain hacks the testing process and runs cases only if flag -test.run is specified.
// Due to the nature of this package, all "successful" cases will always fail.
// With this hack, we can run selected case manually without breaking travis-ci system.
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
	a := New(t)
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
	a := New(t)
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
	Equal(t, map[string]int{
		"foo": 1,
		"bar": -2,
	}, map[string]int{
		"bar": -2,
		"foo": 1,
	})

	Equal(t, map[string]int{
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
	Equal(t, v1, v2)

	v3 := []int{1, 2, 3}
	v4 := []int64{1, 2, 3}
	Equal(t, v3, v4)
}

func TestAssertEqualityWithAssertion(t *testing.T) {
	a := New(t)
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
	a := New(t)
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
	NotEqual(t, v1, v2)

	v3 := []int{1, 2, 3}
	v4 := []int{1, 2, 3}
	NotEqual(t, v3, v4)
}

func TestAssertNotEqualWithAssertion(t *testing.T) {
	a := New(t)
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

func TestUse(t *testing.T) {
	a := New(t)
	v1 := 123
	v2 := []string{"foo", "bar"}
	v3 := v2[0]
	a.Use(v1, &v2) // v2 is used but v1 is not used due to missing `&`.

	// Should pass.
	a.Assert(v1 == 123 && v3 == "foo")

	// Should fail.
	v1 = 345
	v3 = v2[1]
	a.Assert(v1 > 123 && v3 != "bar")
}
