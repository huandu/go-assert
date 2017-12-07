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

// Assertion implements useful methods to assert expressions or function call.
type Assertion testing.T

// New returns a T wrapping testing.T.
func New(t *testing.T) *Assertion {
	return (*Assertion)(t)
}

// Assert tests expr and call `t.Fatalf` to terminate test case if expr is false-equivalent value.
// `false`, 0, nil and empty string are false-equivalent values.
// Usage:
//
//     import "github.com/huandu/go-assert/assertion"
//
//     func TestSomething(t *testing.T) {
//         fa := assertion.New(t)
//         a, b := 1, 2
//         fa.Assert(t, a > b) // This case fails with message "Assertion failed: a > b".
//     }
func (t *Assertion) Assert(expr interface{}) {
	k := ParseFalseKind(expr)

	if k == Positive {
		return
	}

	TriggerAssertion((*testing.T)(t), k, "Assert", 0)
}

// NilError tests result of returned by a function and terminate test case
// using `t.Fatalf` if the last parameter is `error` and not nil.
//
// Usage:
//
//     import "github.com/huandu/go-assert/assertion"
//
//     func TestSomething(t *testing.T) {
//         a := assertion.New(t)
//         a.NilError(os.Open("path/to/a/file")) // This case fails if os.Open returns error.
//     }
func (t *Assertion) NilError(result ...interface{}) {
	if len(result) == 0 {
		return
	}

	pos := len(result) - 1
	e := result[pos]

	if ee, ok := e.(error); !ok || ee == nil {
		return
	}

	s, err := callerArgExpr(Positive, 3, "NilError", -1)

	if err != nil {
		t.Fatalf("Assertion failed. [err:%v]", err)
		return
	}

	t.Fatalf(`Assertion failed: %v returns error "%v".`, s, e)
}

// TriggerAssertion calls t.Fatalf to terminate a test case.
// It must be called by an assert function which will be directly used in test cases.
// See code in `Assertion#Assert` as a sample.
func TriggerAssertion(t *testing.T, k FalseKind, name string, argIndex int) {
	s, err := callerArgExpr(k, 4, name, argIndex)

	if err != nil {
		t.Fatalf("Assertion failed. [err:%v]", err)
		return
	}

	t.Fatalf("Assertion failed: %v", s)
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

// callerArgExpr finds the source code calling assert function and returns the text.
func callerArgExpr(k FalseKind, skip int, name string, argIndex int) (string, error) {
	pc := make([]uintptr, 1)
	n := runtime.Callers(skip, pc)

	if n == 0 {
		return "", fmt.Errorf("fail to read call stack")
	}

	pc = pc[:n]
	frames := runtime.CallersFrames(pc)

	var frame runtime.Frame
	var err error
	result := &bytes.Buffer{}
	done := false

	frame, _ = frames.Next()
	filename := frame.File
	line := frame.Line

	if filename == "" || line == 0 {
		return "", fmt.Errorf("fail to read source code information")
	}

	dotIdx := strings.LastIndex(name, ".")

	if dotIdx >= 0 {
		name = name[dotIdx+1:]
	}

	// Load AST and find target function at target line.
	fset := token.NewFileSet()
	src, parsedAst, err := parseFile(fset, filename)

	if err != nil {
		return "", err
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

		if argIndex < 0 {
			argIndex += len(call.Args)
		}

		if len(call.Args) <= argIndex {
			return true
		}

		arg := call.Args[argIndex]

		pos := fset.Position(arg.Pos())
		posEnd := fset.Position(arg.End())

		if line < pos.Line || line > posEnd.Line {
			return true
		}

		result.Write(src[arg.Pos()-1 : arg.End()-1])

		if _, ok := arg.(*ast.Ident); ok {
			switch k {
			case Nil:
				result.WriteString(" != nil")
			case False:
				result.WriteString(" != true")
			case Zero:
				result.WriteString(" != 0")
			case EmptyString:
				result.WriteString(` != ""`)
			}
		}

		done = true
		return false
	})

	s := result.String()

	if s == "" {
		s = "<EMPTY>"
	}

	return s, err
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
