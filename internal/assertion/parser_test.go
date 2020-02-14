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
			uint(0), Zero,
		},
		{
			0.0, Zero,
		},
		{
			complex(0, 0), Zero,
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
		Assignments [][]string
		RelatedVars []string
	}{
		{
			[]int{0},
			[]string{`prefix + args`},
			[][]string{
				{`f(&args)`, `prefix := s.(type)`},
			},
			[]string{`args`, `prefix`, `s`},
		},
		{
			[]int{1},
			[]string{`skip`},
			[][]string{
				{`skip = 0`},
			},
			[]string{},
		},
		{
			[]int{-1, 0, -2, 4},
			[]string{`c.ArgIndex`, `prefix + args`, `skip`, ""},
			[][]string{
				{`i, c := range cases`},
				{`f(&args)`, `prefix := s.(type)`},
				{`skip = 0`},
				nil,
			},
			[]string{`args`, `c`, `i`, `prefix`, `s`},
		},
	}
	p := new(Parser)

	for i, c := range cases {
		skip := i
		skip = 0 // The last assignment to `skip` should be chosen.
		args := "foo"
		f := func(s *string) { *s = "Args" }
		f(&args)

		var s interface{} = "Parse"
		switch prefix := s.(type) { // Test init stmt in SwitchStmt.
		case string:
			f, err := p.ParseArgs(prefix+args, skip, c.ArgIndex)
			info := p.ParseInfo(f)
			skip = 2

			assertEqual(t, err, nil)
			assertEqual(t, info.Args, c.Args)
			assertEqual(t, info.Assignments, c.Assignments)
			assertEqual(t, info.RelatedVars, c.RelatedVars)
		}
	}
}
