// Copyright 2017 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

// Package assertion is the implementation detail of package assert.
// One can use API to create a customized assert function with this package
package assertion

import (
	"reflect"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

// Trigger represents the method which triggers assertion.
type Trigger struct {
	Parser   *Parser
	FuncName string
	Skip     int
	Args     []int
	Vars     map[string]interface{}
}

// P returns a valid parser.
func (t *Trigger) P() *Parser {
	if t.Parser != nil {
		return t.Parser
	}

	return &Parser{}
}

// Assert tests expr and call `t.Fatalf` to terminate test case if expr is false-equivalent value.
func Assert(t *testing.T, expr interface{}, trigger *Trigger) {
	k := ParseFalseKind(expr)

	if k == Positive {
		return
	}

	f, err := trigger.P().ParseArgs(trigger.FuncName, trigger.Skip+1, trigger.Args)

	if err != nil {
		t.Fatalf("Assertion failed with an internal error: %v", err)
		return
	}

	info := trigger.P().ParseInfo(f)
	suffix := ""
	arg := info.Args[0]

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

	assignment := indentAssignments(info.Assignments[0], 4)

	if assignment != "" {
		assignment = "\nReferenced variables are assigned in following statements:" + assignment
	}

	t.Fatalf("\n%v:%v: Assertion failed:\n    %v%v%v%v",
		f.Filename, f.Line, indentCode(arg, 4), suffix,
		assignment, formatRelatedVars(info.RelatedVars, trigger.Vars),
	)
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

	f, err := trigger.P().ParseArgs(trigger.FuncName, trigger.Skip+1, trigger.Args)

	if err != nil {
		t.Fatalf("Assertion failed with an internal error: %v", err)
		return
	}

	info := trigger.P().ParseInfo(f)
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
	msg := "The value of following expression should equal."

	if typeMismatch {
		msg = "The type of following expressions should be the same."
	}

	t.Fatalf("\n%v:%v: Assertion failed:\n    %v\n%v\n[1] %v%v\n[2] %v%v\nValues:\n[1] -> %v\n[2] -> %v%v",
		f.Filename, f.Line, indentCode(info.Source, 4), msg,
		indentCode(info.Args[0], 4), indentAssignments(info.Assignments[0], 4),
		indentCode(info.Args[1], 4), indentAssignments(info.Assignments[1], 4),
		v1Dump, v2Dump, formatRelatedVars(info.RelatedVars, trigger.Vars),
	)
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

	f, err := trigger.P().ParseArgs(trigger.FuncName, trigger.Skip+1, trigger.Args)

	if err != nil {
		t.Fatalf("Assertion failed with an internal error: %v", err)
		return
	}

	info := trigger.P().ParseInfo(f)
	t.Fatalf("\n%v:%v: Assertion failed:\n    %v\nThe value of following expression should not equal.\n[1] %v%v\n[2] %v%v%v",
		f.Filename, f.Line, indentCode(info.Source, 4),
		indentCode(info.Args[0], 4), indentAssignments(info.Assignments[0], 4),
		indentCode(info.Args[1], 4), indentAssignments(info.Assignments[1], 4),
		formatRelatedVars(info.RelatedVars, trigger.Vars),
	)
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

	f, err := trigger.P().ParseArgs(trigger.FuncName, trigger.Skip+1, trigger.Args)

	if err != nil {
		t.Fatalf("Assertion failed with an internal error: %v", err)
		return
	}

	info := trigger.P().ParseInfo(f)
	t.Fatalf("\n%v:%v: Assertion failed:\nFollowing expression should return a nil error.\n    %v%v\nThe error is:\n    %v%v",
		f.Filename, f.Line,
		indentCode(info.Args[0], 4), indentAssignments(info.Assignments[0], 4),
		e, formatRelatedVars(info.RelatedVars, trigger.Vars),
	)
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

	f, err := trigger.P().ParseArgs(trigger.FuncName, trigger.Skip+1, trigger.Args)

	if err != nil {
		t.Fatalf("Assertion failed with an internal error: %v", err)
		return
	}

	info := trigger.P().ParseInfo(f)
	t.Fatalf("\n%v:%v: Assertion failed:\nFollowing expression should return an error.\n    %v%v%v",
		f.Filename, f.Line,
		indentCode(info.Args[0], 4), indentAssignments(info.Assignments[0], 4),
		formatRelatedVars(info.RelatedVars, trigger.Vars),
	)
}

func indentCode(code string, spaces int) string {
	if code == "" {
		return ""
	}

	lines := strings.Split(code, "\n")
	indented := make([]string, 0, len(lines))
	space := strings.Repeat(" ", spaces)

	indented = append(indented, lines[0])
	lines = lines[1:]

	for _, line := range lines {
		indented = append(indented, space+line)
	}

	return strings.Join(indented, "\n")
}

func indentAssignments(assignments []string, spaces int) string {
	if len(assignments) == 0 {
		return ""
	}

	space := strings.Repeat(" ", spaces)
	output := make([]string, 0)
	output = append(output, "") // Add a newline at the front.

	for _, code := range assignments {
		lines := strings.Split(code, "\n")
		indented := make([]string, 0, len(lines))

		indented = append(indented, space+lines[0])
		lines = lines[1:]

		for _, line := range lines {
			indented = append(indented, space+line)
		}

		output = append(output, indented...)
	}

	return strings.Join(output, "\n")
}

func formatRelatedVars(related []string, vars map[string]interface{}) string {
	if len(related) == 0 || len(vars) == 0 {
		return ""
	}

	values := make([]interface{}, 0, len(related))
	names := make([]string, 0, len(related))
	fields := make([]string, 0, len(related))

	for _, name := range related {
		if v, ok := vars[name]; ok {
			values = append(values, v)
			names = append(names, name)
			fields = append(fields, "")
			continue
		}

		parts := strings.Split(name, ".")
		parts = parts[:len(parts)-1]

		for len(parts) > 0 {
			n := strings.Join(parts, ".")

			if v, ok := vars[n]; ok {
				values = append(values, v)
				names = append(names, n)
				fields = append(fields, name[len(n)+1:])
				break
			}

			parts = parts[:len(parts)-1]
		}
	}

	if len(values) == 0 {
		return ""
	}

	config := &spew.ConfigState{
		DisableMethods:          true,
		DisablePointerMethods:   true,
		DisablePointerAddresses: true,
		DisableCapacities:       true,
		SortKeys:                true,
		SpewKeys:                true,
	}
	lines := make([]string, 0, len(values)+1)
	lines = append(lines, "\nRelated variables:")
	visitedNames := map[string]struct{}{}

	for i, v := range values {
		val := reflect.ValueOf(v)

		if !val.IsValid() || val.Kind() != reflect.Ptr {
			continue
		}

		val = val.Elem()
		field, v, ok := getValue(fields[i], val)

		if !ok {
			continue
		}

		name := names[i]

		if field != "" {
			name += "." + field
		}

		if _, ok := visitedNames[name]; ok {
			continue
		}

		lines = append(lines, config.Sprintf("    "+name+" = %#v", v))
		visitedNames[name] = struct{}{}
	}

	// No valid related variables.
	if len(lines) == 1 {
		return ""
	}

	return strings.Join(lines, "\n")
}

func getValue(field string, v reflect.Value) (actualField string, value interface{}, ok bool) {
	if field == "" {
		value = v.Interface()
		ok = true
		return
	}

	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	if !v.IsValid() {
		return
	}

	if v.Kind() != reflect.Struct {
		value = v.Interface()
		ok = true
		return
	}

	parts := strings.Split(field, ".")
	f := v.FieldByName(parts[0])

	if !f.IsValid() {
		value = v.Interface()
		ok = true
		return
	}

	// If f is not printable, use f's parent.
	switch f.Kind() {
	case reflect.Func, reflect.Chan:
		actualField = parts[0]
		value = v.Interface()
		ok = true
		return
	}

	actual, value, ok := getValue(strings.Join(parts[1:], "."), f)

	if !ok {
		return
	}

	actualField = parts[0]

	if actual != "" {
		actualField += "." + actual
	}
	return
}
