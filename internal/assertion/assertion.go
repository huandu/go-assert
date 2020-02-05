// Copyright 2017 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

// Package assertion is the implementation detail of package assert.
// One can use API to create a customized assert function with this package
package assertion

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// FalseKind is the kind of a false-equivalent value.
type FalseKind int

// Valid kinds for all false-equivalent values.
const (
	Positive FalseKind = iota
	Nil
	False
	Zero
	EmptyString
)

// Trigger represents the method which triggers assertion.
type Trigger struct {
	FuncName string
	Skip     int
	Args     []int
}

// Assert tests expr and call `t.Fatalf` to terminate test case if expr is false-equivalent value.
func Assert(t *testing.T, expr interface{}, trigger *Trigger) {
	k := ParseFalseKind(expr)

	if k == Positive {
		return
	}

	TriggerAssert(t, trigger.FuncName, trigger.Skip+1, trigger.Args, k)
}

// AssertEqual uses `reflect.DeepEqual` to test v1 and v2 equality.
func AssertEqual(t *testing.T, v1, v2 interface{}, trigger *Trigger) {
	if reflect.DeepEqual(v1, v2) {
		return
	}

	typeMismatch := false

	if v1 != nil && v2 != nil {
		t1 := reflect.TypeOf(v1)
		t2 := reflect.TypeOf(v2)

		if !t1.AssignableTo(t2) && !t2.AssignableTo(t1) {
			typeMismatch = true
		}
	} else {
		v1Val := reflect.ValueOf(v1)
		v2Val := reflect.ValueOf(v2)

		// Treat (*T)(nil) as nil.
		if isNil(v1Val) && isNil(v2Val) {
			return
		}
	}

	args, filename, line, err := ParseArgs(trigger.FuncName, trigger.Skip+1, trigger.Args)

	if err != nil {
		t.Fatalf("Assertion failed with an internal error: %v", err)
		return
	}

	if typeMismatch {
		t.Fatalf("\n%v:%v: Assertion failed: type of %v and %v should be the same.\n\tv1 = %v (type %[5]T)\n\tv2 = %v (type %[6]T)",
			filename, line, args[0], args[1], v1, v2)
	} else {
		t.Fatalf("\n%v:%v: Assertion failed: %v == %v\n\tv1 = %v (type %[5]T)\n\tv2 = %v (type %[6]T)",
			filename, line, args[0], args[1], v1, v2)
	}
}

func isNil(val reflect.Value) bool {
	switch val.Kind() {
	case reflect.Invalid:
		return true
	case reflect.Interface, reflect.Chan, reflect.Func, reflect.Slice, reflect.Map, reflect.Ptr:
		return val.IsNil()
	}

	return false
}

// AssertNotEqual uses `reflect.DeepEqual` to test v1 and v2 equality.
func AssertNotEqual(t *testing.T, v1, v2 interface{}, trigger *Trigger) {
	if !reflect.DeepEqual(v1, v2) {
		return
	}

	args, filename, line, err := ParseArgs(trigger.FuncName, trigger.Skip+1, trigger.Args)

	if err != nil {
		t.Fatalf("Assertion failed with an internal error: %v", err)
		return
	}

	t.Fatalf("\n%v:%v: Assertion failed: %v != %v", filename, line, args[0], args[1])
}

// AssertNilError expects a function return a nil error.
// Otherwise, it will terminate the test case using `t.Fatalf`.
func AssertNilError(t *testing.T, result []interface{}, trigger *Trigger) {
	if len(result) == 0 {
		return
	}

	pos := len(result) - 1
	e := result[pos]

	if ee, ok := e.(error); !ok || ee == nil {
		return
	}

	args, filename, line, err := ParseArgs(trigger.FuncName, trigger.Skip+1, trigger.Args)

	if err != nil {
		t.Fatalf("Assertion failed with an internal error: %v", err)
		return
	}

	t.Fatalf("\n%v:%v: Assertion failed: %v should return nil error\n\t err = %v",
		filename, line, args[0], e)
}

// AssertNonNilError expects a function return a non-nil error.
// Otherwise, it will terminate the test case using `t.Fatalf`.
func AssertNonNilError(t *testing.T, result []interface{}, trigger *Trigger) {
	if len(result) == 0 {
		return
	}

	pos := len(result) - 1
	e := result[pos]

	if e != nil {
		if _, ok := e.(error); !ok {
			return
		}

		if v := reflect.ValueOf(e); !v.IsNil() {
			return
		}
	}

	args, filename, line, err := ParseArgs(trigger.FuncName, trigger.Skip+1, trigger.Args)

	if err != nil {
		t.Fatalf("Assertion failed with an internal error: %v", err)
		return
	}

	t.Fatalf("\n%v:%v: Assertion failed: expect %v returns an error.", filename, line, args[0])
}

// TriggerAssert calls t.Fatalf to terminate a test case.
// It must be called by an assert function which will be directly used in test cases.
// See code in `Assertion#Assert` as a sample.
func TriggerAssert(t *testing.T, name string, skip int, argIndex []int, k FalseKind) {
	args, filename, line, err := ParseArgs(name, skip, argIndex)

	if err != nil {
		t.Fatalf("Assertion failed with an internal error: %v", err)
		return
	}

	suffix := ""
	arg := args[0]

	if !strings.ContainsRune(arg, ' ') {
		switch k {
		case Nil:
			suffix = " != nil"
		case False:
			suffix = " != true"
		case Zero:
			suffix = " != 0"
		case EmptyString:
			suffix = ` != ""`
		}
	}

	t.Fatalf("\n%v:%v: Assertion failed: %v%v", filename, line, arg, suffix)
}

// ParseFalseKind checks expr value and return false when expr is `false`, 0, `nil` and empty string.
// Otherwise, return true.
func ParseFalseKind(expr interface{}) FalseKind {
	if expr == nil {
		return Nil
	}

	if v, ok := expr.(bool); ok && !v {
		return False
	}

	v := reflect.ValueOf(expr)

	for {
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if n := v.Int(); n == 0 {
				return Zero
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if n := v.Uint(); n == 0 {
				return Zero
			}
		case reflect.Float32, reflect.Float64:
			if n := v.Float(); n == 0 {
				return Zero
			}
		case reflect.String:
			if s := v.String(); s == "" {
				return EmptyString
			}
		case reflect.Interface:
			v = v.Elem()
			continue
		}

		return Positive
	}
}

// ParseArgs parses caller's source code, finds out the right call expression by name
// and returns the argument source code.
//
// Skip is the stack frame calling an assert function. If skip is 0, the stack frame for
// ParseArgs is selected.
// In most cases, caller should set skip to 1 to skip ParseArgs itself.
func ParseArgs(name string, skip int, argIndex []int) (args []string, filename string, line int, err error) {
	if len(argIndex) == 0 {
		err = fmt.Errorf("missing argIndex")
		return
	}

	const minimumSkip = 2 // Skip 2 frames running runtime functions.

	pc := make([]uintptr, 1)
	n := runtime.Callers(skip+minimumSkip, pc)

	if n == 0 {
		err = fmt.Errorf("fail to read call stack")
		return
	}

	pc = pc[:n]
	frames := runtime.CallersFrames(pc)

	var frame runtime.Frame
	done := false

	frame, _ = frames.Next()
	filename = frame.File
	line = frame.Line

	if filename == "" || line == 0 {
		err = fmt.Errorf("fail to read source code information")
		return
	}

	dotIdx := strings.LastIndex(name, ".")

	if dotIdx >= 0 {
		name = name[dotIdx+1:]
	}

	// Load AST and find target function at target line.
	fset := token.NewFileSet()
	src, parsedAst, err := parseFile(fset, filename)
	filename = path.Base(filename)

	if err != nil {
		return
	}

	args = make([]string, len(argIndex))
	maxArgIdx := 0

	for _, idx := range argIndex {
		if idx > maxArgIdx {
			maxArgIdx = idx
		}
	}

	ast.Inspect(parsedAst, func(node ast.Node) bool {
		if node == nil || done || err != nil {
			return false
		}

		call, ok := node.(*ast.CallExpr)

		if !ok {
			return true
		}

		fn := string(src[call.Fun.Pos()-1 : call.Fun.End()-1])
		dotIdx := strings.LastIndex(fn, ".")

		if dotIdx >= 0 {
			fn = fn[dotIdx+1:]
		}

		if fn != name {
			return true
		}

		pos := fset.Position(call.Pos())
		posEnd := fset.Position(call.End())

		if line < pos.Line || line > posEnd.Line {
			return true
		}

		for i, idx := range argIndex {
			if idx < 0 {
				idx += len(call.Args)
			}

			if idx < 0 || idx >= len(call.Args) {
				// Ignore invalid idx.
				continue
			}

			arg := call.Args[idx]
			args[i] = string(src[arg.Pos()-1 : arg.End()-1])
		}

		done = true
		return false
	})

	return
}

func parseFile(fset *token.FileSet, filename string) ([]byte, *ast.File, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, nil, err
	}

	defer file.Close()

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, file)
	data := buf.Bytes()

	if err != nil {
		return nil, nil, err
	}

	f, err := parser.ParseFile(fset, filename, buf, 0)
	return data, f, err
}
