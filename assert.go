// Copyright 2017 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

// Package assert provides API to implement C-like assert macro.
package assert

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

type falseKind int

const (
	falseKindPositive falseKind = iota
	falseKindNil
	falseKindFalse
	falseKindZero
	falseKindEmptyString
)

// Assert tests expr and call `t.Fatalf` to terminate test case if expr is false-equivalent value.
// `false`, 0, nil and empty string are false-equivalent values.
func Assert(t *testing.T, expr interface{}) {
	k := parseFalseKind(expr)

	if k == falseKindPositive {
		return
	}

	assertion(t, k, "Assert", 1)
}

func assertion(t *testing.T, k falseKind, name string, argIndex int) {
	s, err := callerArgExpr(k, 4, name, argIndex)

	if err != nil {
		t.Fatalf("Assertion failed. [err:%v]", err)
		return
	}

	t.Fatalf("Assertion failed: %v", s)
}

// parseFalseKind checks expr value and return false when expr is `false`, 0, `nil` and empty string.
// Otherwise, return true.
func parseFalseKind(expr interface{}) falseKind {
	if expr == nil {
		return falseKindNil
	}

	if v, ok := expr.(bool); ok && !v {
		return falseKindFalse
	}

	v := reflect.ValueOf(expr)

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if n := v.Int(); n == 0 {
			return falseKindZero
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if n := v.Uint(); n == 0 {
			return falseKindZero
		}
	case reflect.Float32, reflect.Float64:
		if n := v.Float(); n == 0 {
			return falseKindZero
		}
	case reflect.String:
		if s := v.String(); s == "" {
			return falseKindEmptyString
		}
	}

	return falseKindPositive
}

// callerArgExpr finds the source code calling assert function and returns the text.
func callerArgExpr(k falseKind, skip int, name string, argIndex int) (string, error) {
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
			case falseKindNil:
				result.WriteString(" != nil")
			case falseKindFalse:
				result.WriteString(" != true")
			case falseKindZero:
				result.WriteString(" != 0")
			case falseKindEmptyString:
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
