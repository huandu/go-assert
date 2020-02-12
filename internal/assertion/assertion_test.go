// Copyright 2017 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package assertion

import (
	"testing"
)

func assertEqual(t *testing.T, v1, v2 interface{}) {
	AssertEqual(t, v1, v2, &Trigger{
		FuncName: "assertEqual",
		Skip:     1,
		Args:     []int{1, 2},
	})
}

func TestParseFalseKind(t *testing.T) {
	cases := []struct {
		Value interface{}
		Kind  FalseKind
	}{
		{
			12, Positive,
		},
		{
			nil, Nil,
		},
		{
			0, Zero,
		},
		{
			false, False,
		},
		{
			[]int{}, Positive,
		},
		{
			([]int)(nil), Nil,
		},
		{
			"", EmptyString,
		},
	}

	for i, c := range cases {
		t.Logf("case %v: %v", i, c)
		k := ParseFalseKind(c.Value)
		assertEqual(t, c.Kind, k)
	}
}

func TestParseArgs(t *testing.T) {
	cases := []struct {
		ArgIndex    []int
		Args        []string
		Assignments []string
	}{
		{
			[]int{0},
			[]string{`"ParseArgs"`},
			[]string{""},
		},
		{
			[]int{1},
			[]string{`skip`},
			[]string{`skip = 0`},
		},
		{
			[]int{-1, 0, -2},
			[]string{`c.ArgIndex`, `"ParseArgs"`, `skip`},
			[]string{`i, c := range cases`, "", `skip = 0`},
		},
	}

	for i, c := range cases {
		t.Logf("case %v: %v", i, c)
		skip := 1
		skip = 0
		args, assignments, _, _, err := ParseArgs("ParseArgs", skip, c.ArgIndex)
		skip = 2

		assertEqual(t, err, nil)
		assertEqual(t, args, c.Args)
		assertEqual(t, assignments, c.Assignments)
	}
}
