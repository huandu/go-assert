// Copyright 2017 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package assertion

import (
	"testing"
)

func TestParseFalseKind(t *testing.T) {
	if k := ParseFalseKind(12); k != Positive {
		t.Fatalf("unexpected kind. [k:%v]", k)
	}

	if k := ParseFalseKind(nil); k != Nil {
		t.Fatalf("unexpected kind. [k:%v]", k)
	}

	if k := ParseFalseKind(0); k != Zero {
		t.Fatalf("unexpected kind. [k:%v]", k)
	}

	if k := ParseFalseKind(false); k != False {
		t.Fatalf("unexpected kind. [k:%v]", k)
	}

	if k := ParseFalseKind([]int{}); k != Positive {
		t.Fatalf("unexpected kind. [k:%v]", k)
	}

	var i1 interface{} = 123
	if k := ParseFalseKind(i1); k != Positive {
		t.Fatalf("unexpected kind. [k:%v]", k)
	}

	s := ""
	var i2 interface{} = s
	if k := ParseFalseKind(i2); k != EmptyString {
		t.Fatalf("unexpected kind. [k:%v]", k)
	}
}

func TestParseArgs(t *testing.T) {
	if args, filename, _, err := ParseArgs("ParseArgs", 0, 0); err != nil || len(args) != 1 || args[0] != `"ParseArgs"` || filename != "assertion_test.go" {
		t.Fatalf("unexpected expr. [expr:%v]", args[0])
	}

	if args, filename, _, err := ParseArgs("Parse"+"Args", 0, 0); err != nil || len(args) != 1 || args[0] != `"Parse"+"Args"` || filename != "assertion_test.go" {
		t.Fatalf("unexpected expr. [expr:%v]", args[0])
	}

	if args, filename, _, err := ParseArgs("Parse"+"Args", 0, -1, 0, -4); err != nil || len(args) != 3 ||
		args[0] != "-4" || args[1] != `"Parse"+"Args"` || args[2] != "0" || filename != "assertion_test.go" {
		t.Fatalf("unexpected expr. [expr:%v]", args)
	}
}
