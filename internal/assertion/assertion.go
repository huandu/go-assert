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
	"go/printer"
	"go/token"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
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

func indentCode(code string, spaces int, indentFirstLine bool, newLine bool) string {
	if code == "" {
		return ""
	}

	lines := strings.Split(code, "\n")
	indented := make([]string, 0, len(lines))
	space := strings.Repeat(" ", spaces)
	firstLine := lines[0]

	if newLine {
		firstLine = ""
	} else {
		lines = lines[1:]
	}

	if indentFirstLine {
		indented = append(indented, space+firstLine)
	} else {
		indented = append(indented, firstLine)
	}

	for _, line := range lines {
		indented = append(indented, space+line)
	}

	return strings.Join(indented, "\n")
}

// Assert tests expr and call `t.Fatalf` to terminate test case if expr is false-equivalent value.
func Assert(t *testing.T, expr interface{}, trigger *Trigger) {
	k := ParseFalseKind(expr)

	if k == Positive {
		return
	}

	args, assignments, filename, line, err := ParseArgs(trigger.FuncName, trigger.Skip+1, trigger.Args)

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

	assignment := indentCode(assignments[0], 4, false, true)

	if assignment != "" {
		assignment = "\nWhich is assigned in this statement:" + assignment
	}

	t.Fatalf("\n%v:%v: Assertion failed:\n%v%v%v",
		filename, line, indentCode(arg, 4, true, false), suffix, assignment)
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

	args, assignments, filename, line, err := ParseArgs(trigger.FuncName, trigger.Skip+1, trigger.Args)

	if err != nil {
		t.Fatalf("Assertion failed with an internal error: %v", err)
		return
	}

	config := &spew.ConfigState{
		DisableMethods:          true,
		DisablePointerMethods:   true,
		DisablePointerAddresses: true,
		DisableCapacities:       true,
		SortKeys:                true,
		SpewKeys:                true,
	}
	v1Dump := config.Sprintf("%#v", v1)
	v2Dump := config.Sprintf("%#v", v2)

	if typeMismatch {
		t.Fatalf("\n%v:%v: Assertion failed:\nThe type of following expressions should be the same.\n[1] %v%v\n[2] %v%v\nValues:\n[1] = %v\n[2] = %v",
			filename, line, indentCode(args[0], 4, false, false), indentCode(assignments[0], 4, false, true), indentCode(args[1], 4, false, false), indentCode(assignments[1], 4, false, true), v1Dump, v2Dump)
	} else {
		t.Fatalf("\n%v:%v: Assertion failed:\nThe value of following expression should equal.\n[1] %v%v\n[2] %v%v\nValues:\n[1] = %v\n[2] = %v",
			filename, line, indentCode(args[0], 4, false, false), indentCode(assignments[0], 4, false, true), indentCode(args[1], 4, false, false), indentCode(assignments[1], 4, false, true), v1Dump, v2Dump)
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

	args, assignments, filename, line, err := ParseArgs(trigger.FuncName, trigger.Skip+1, trigger.Args)

	if err != nil {
		t.Fatalf("Assertion failed with an internal error: %v", err)
		return
	}

	t.Fatalf("\n%v:%v: Assertion failed:\nThe value of following expression should not equal.\n[1] %v%v\n[2] %v%v",
		filename, line, indentCode(args[0], 4, false, false), indentCode(assignments[0], 4, false, true), indentCode(args[1], 4, false, false), indentCode(assignments[1], 4, false, true))
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

	args, assignments, filename, line, err := ParseArgs(trigger.FuncName, trigger.Skip+1, trigger.Args)

	if err != nil {
		t.Fatalf("Assertion failed with an internal error: %v", err)
		return
	}

	t.Fatalf("\n%v:%v: Assertion failed:\nFollowing expression should return a nil error.\n%v%v\nThe error is:\n    %v",
		filename, line, indentCode(args[0], 4, true, false), indentCode(assignments[0], 4, false, true), e)
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

	args, assignments, filename, line, err := ParseArgs(trigger.FuncName, trigger.Skip+1, trigger.Args)

	if err != nil {
		t.Fatalf("Assertion failed with an internal error: %v", err)
		return
	}

	t.Fatalf("\n%v:%v: Assertion failed:\nFollowing expression should return an error.\n%v%v",
		filename, line, indentCode(args[0], 4, true, false), indentCode(assignments[0], 4, false, true))
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
		case reflect.Invalid:
			return Nil
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
		case reflect.Complex64, reflect.Complex128:
			if n := v.Complex(); n == 0 {
				return Zero
			}
		case reflect.String:
			if s := v.String(); s == "" {
				return EmptyString
			}
		case reflect.Interface:
			if v.IsNil() {
				return Nil
			}

			v = v.Elem()
			continue
		case reflect.Ptr, reflect.Chan, reflect.Func, reflect.Slice:
			if v.IsNil() {
				return Nil
			}
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
func ParseArgs(name string, skip int, argIndex []int) (args []string, assignments []string, filename string, line int, err error) {
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
	parsedAst, err := parseFile(fset, filename)
	filename = path.Base(filename)

	if err != nil {
		return
	}

	argNodes := make([]ast.Node, 0, len(argIndex))
	maxArgIdx := 0
	var funcDecl *ast.FuncDecl

	for _, idx := range argIndex {
		if idx > maxArgIdx {
			maxArgIdx = idx
		}
	}

	ast.Inspect(parsedAst, func(node ast.Node) bool {
		if node == nil || done {
			return false
		}

		if decl, ok := node.(*ast.FuncDecl); ok {
			funcDecl = decl
			return true
		}

		call, ok := node.(*ast.CallExpr)

		if !ok {
			return true
		}

		var fn string
		switch expr := call.Fun.(type) {
		case *ast.Ident:
			fn = expr.Name
		case *ast.SelectorExpr:
			fn = expr.Sel.Name
		}

		if fn != name {
			return true
		}

		pos := fset.Position(call.Pos())
		posEnd := fset.Position(call.End())

		if line < pos.Line || line > posEnd.Line {
			return true
		}

		for _, idx := range argIndex {
			if idx < 0 {
				idx += len(call.Args)
			}

			if idx < 0 || idx >= len(call.Args) {
				// Ignore invalid idx.
				argNodes = append(argNodes, nil)
				continue
			}

			arg := call.Args[idx]
			argNodes = append(argNodes, arg)
		}

		done = true
		return false
	})

	args = make([]string, 0, len(argIndex))
	assignments = make([]string, 0, len(argIndex))

	// If args contains any arg which is an ident, find out where it's assigned.
	for _, arg := range argNodes {
		args = append(args, formatNode(fset, arg))
		assignments = append(assignments, findAssignment(fset, funcDecl, line, arg))
	}

	return
}

func parseFile(fset *token.FileSet, filename string) (*ast.File, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}

	defer file.Close()
	f, err := parser.ParseFile(fset, filename, file, 0)
	return f, err
}

func formatNode(fset *token.FileSet, node ast.Node) string {
	if node == nil {
		return ""
	}

	buf := &bytes.Buffer{}
	config := &printer.Config{
		Mode:     printer.UseSpaces,
		Tabwidth: 4,
	}
	config.Fprint(buf, fset, node)
	return buf.String()
}

func findAssignment(fset *token.FileSet, decl *ast.FuncDecl, line int, arg ast.Node) string {
	if decl == nil || arg == nil {
		return ""
	}

	ident := findIdent(arg)

	if ident == "" {
		return ""
	}

	// Find the last assignment for ident.
	done := false
	var assignment ast.Stmt
	ast.Inspect(decl, func(n ast.Node) bool {
		if n == nil || done {
			return false
		}

		if pos := fset.Position(n.Pos()); pos.Line >= line {
			done = true
			return false
		}

		switch stmt := n.(type) {
		case *ast.AssignStmt:
			for _, left := range stmt.Lhs {
				switch expr := left.(type) {
				case *ast.Ident:
					if ident == expr.Name {
						assignment = stmt
						return true
					}
				}
			}

			for _, right := range stmt.Rhs {
				switch expr := right.(type) {
				case *ast.UnaryExpr:
					// Treat `&a` as a kind of assignment to `a`.
					if id, ok := expr.X.(*ast.Ident); ok && expr.Op == token.AND && id.Name == ident {
						assignment = stmt
						return true
					}
				}
			}
		case *ast.RangeStmt:
			if stmt.Key == nil {
				return true
			}

			switch expr := stmt.Key.(type) {
			case *ast.Ident:
				if ident == expr.Name {
					assignment = stmt
					return true
				}
			}

			if stmt.Value == nil {
				return true
			}

			switch expr := stmt.Value.(type) {
			case *ast.Ident:
				if ident == expr.Name {
					assignment = stmt
					return true
				}
			}
		}

		return true
	})

	// For RangeStmt, only use the code like `k, v := range arr`.
	if rng, ok := assignment.(*ast.RangeStmt); ok {
		stmt := *rng
		body := *rng.Body
		body.List = nil
		stmt.Body = &body
		code := formatNode(fset, &stmt)
		return code[stmt.Key.Pos()-stmt.Pos() : stmt.X.End()-stmt.Pos()]
	}

	return formatNode(fset, assignment)
}

// findIdent parses arg to find ident in arg.
// Only if arg is an ident, selector, or an unary expr against ident/selector like `*a` or `&a.b`,
// the ident will be returned. Otherwise, return empty string.
func findIdent(arg ast.Node) string {
	switch expr := arg.(type) {
	case *ast.Ident:
		return expr.Name
	case *ast.SelectorExpr:
		return findIdent(expr.X)
	case *ast.UnaryExpr:
		return findIdent(expr.X)
	}

	return ""
}
