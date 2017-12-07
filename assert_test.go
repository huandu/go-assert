// Copyright 2017 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package assert

import (
	"testing"
)

func TestParseFalseKind(t *testing.T) {
	if k := parseFalseKind(12); k != falseKindPositive {
		t.Fatalf("unexpected kind. [k:%v]", k)
	}

	if k := parseFalseKind(nil); k != falseKindNil {
		t.Fatalf("unexpected kind. [k:%v]", k)
	}

	if k := parseFalseKind(0); k != falseKindZero {
		t.Fatalf("unexpected kind. [k:%v]", k)
	}

	if k := parseFalseKind(false); k != falseKindFalse {
		t.Fatalf("unexpected kind. [k:%v]", k)
	}

	if k := parseFalseKind([]int{}); k != falseKindPositive {
		t.Fatalf("unexpected kind. [k:%v]", k)
	}

	var i1 interface{} = 123
	if k := parseFalseKind(i1); k != falseKindPositive {
		t.Fatalf("unexpected kind. [k:%v]", k)
	}

	s := ""
	var i2 interface{} = s
	if k := parseFalseKind(i2); k != falseKindEmptyString {
		t.Fatalf("unexpected kind. [k:%v]", k)
	}
}

func TestCallerArgExpr(t *testing.T) {
	if s, err := callerArgExpr(falseKindPositive, 2, "callerArgExpr", 0); err != nil || s != "falseKindPositive" {
		t.Fatalf("unexpected expr. [expr:%v]", s)
	}

	if s, err := callerArgExpr(falseKindPositive, 2, "caller"+"ArgExpr", 2); err != nil || s != `"caller"+"ArgExpr"` {
		t.Fatalf("unexpected expr. [expr:%v]", s)
	}

	if s, err := callerArgExpr(falseKindNil, 2, "caller"+"ArgExpr", 0); err != nil || s != `falseKindNil != nil` {
		t.Fatalf("unexpected expr. [expr:%v]", s)
	}

	if s, err := callerArgExpr(falseKindFalse, 2, "caller"+"ArgExpr", 0); err != nil || s != `falseKindFalse != true` {
		t.Fatalf("unexpected expr. [expr:%v]", s)
	}

	if s, err := callerArgExpr(falseKindZero, 2, "caller"+"ArgExpr", 0); err != nil || s != `falseKindZero != 0` {
		t.Fatalf("unexpected expr. [expr:%v]", s)
	}

	if s, err := callerArgExpr(falseKindEmptyString, 2, "caller"+"ArgExpr", 0); err != nil || s != `falseKindEmptyString != ""` {
		t.Fatalf("unexpected expr. [expr:%v]", s)
	}
}
