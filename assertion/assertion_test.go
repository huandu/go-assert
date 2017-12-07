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

func TestCallerArgExpr(t *testing.T) {
	if s, err := callerArgExpr(Positive, 2, "callerArgExpr", 0); err != nil || s != "Positive" {
		t.Fatalf("unexpected expr. [expr:%v]", s)
	}

	if s, err := callerArgExpr(Positive, 2, "caller"+"ArgExpr", 2); err != nil || s != `"caller"+"ArgExpr"` {
		t.Fatalf("unexpected expr. [expr:%v]", s)
	}

	if s, err := callerArgExpr(Nil, 2, "caller"+"ArgExpr", 0); err != nil || s != `Nil != nil` {
		t.Fatalf("unexpected expr. [expr:%v]", s)
	}

	if s, err := callerArgExpr(False, 2, "caller"+"ArgExpr", 0); err != nil || s != `False != true` {
		t.Fatalf("unexpected expr. [expr:%v]", s)
	}

	if s, err := callerArgExpr(Zero, 2, "caller"+"ArgExpr", 0); err != nil || s != `Zero != 0` {
		t.Fatalf("unexpected expr. [expr:%v]", s)
	}

	if s, err := callerArgExpr(EmptyString, 2, "caller"+"ArgExpr", 0); err != nil || s != `EmptyString != ""` {
		t.Fatalf("unexpected expr. [expr:%v]", s)
	}
}
