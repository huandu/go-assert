// Copyright 2017 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

// Package assertion is the implementation detail of package assert.
// One can use API to create a customized assert function with this package
package assertion

import (
	"reflect"
	"strings"
	"testing"
	"unsafe"

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
		value = getValueInterface(v)
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
		value = getValueInterface(v)
		ok = true
		return
	}

	parts := strings.Split(field, ".")
	f := v.FieldByName(parts[0])

	if !f.IsValid() {
		value = getValueInterface(v)
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

func getValueInterface(v reflect.Value) interface{} {
	if v.CanInterface() {
		return v.Interface()
	}

	// src is an unexported field value. Copy its value.
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool()

	case reflect.Int:
		return int(v.Int())
	case reflect.Int8:
		return int8(v.Int())
	case reflect.Int16:
		return int16(v.Int())
	case reflect.Int32:
		return int32(v.Int())
	case reflect.Int64:
		return v.Int()

	case reflect.Uint:
		return uint(v.Uint())
	case reflect.Uint8:
		return uint8(v.Uint())
	case reflect.Uint16:
		return uint16(v.Uint())
	case reflect.Uint32:
		return uint32(v.Uint())
	case reflect.Uint64:
		return v.Uint()
	case reflect.Uintptr:
		return uintptr(v.Uint())

	case reflect.Float32:
		return float32(v.Float())
	case reflect.Float64:
		return v.Float()

	case reflect.Complex64:
		return complex64(v.Complex())
	case reflect.Complex128:
		return v.Complex()

	case reflect.Array:
		arr := reflect.New(v.Type()).Elem()
		num := v.Len()

		for i := 0; i < num; i++ {
			arr.Index(i).Set(reflect.ValueOf(getValueInterface(v.Index(i))))
		}

		return arr.Interface()

	case reflect.Chan:
		ch := reflect.MakeChan(v.Type(), v.Cap())
		return ch.Interface()

	case reflect.Func:
		// src.Pointer is the PC address of a func.
		pc := reflect.New(reflect.TypeOf(uintptr(0)))
		pc.Elem().SetUint(uint64(v.Pointer()))

		fn := reflect.New(v.Type())
		*(*uintptr)(unsafe.Pointer(fn.Pointer())) = pc.Pointer()
		return fn.Elem().Interface()

	case reflect.Interface:
		iface := reflect.New(v.Type())
		*(*[2]uintptr)(unsafe.Pointer(iface.Pointer())) = v.InterfaceData()
		return iface.Elem().Interface()

	case reflect.Map:
		m := reflect.MakeMap(v.Type())
		keys := v.MapKeys()

		for _, key := range keys {
			m.SetMapIndex(key, reflect.ValueOf(getValueInterface(v.MapIndex(key))))
		}

		return m.Interface()

	case reflect.Ptr:
		ptr := reflect.New(v.Type())
		*(*uintptr)(unsafe.Pointer(ptr.Pointer())) = v.Pointer()
		return ptr.Elem().Interface()

	case reflect.Slice:
		slice := reflect.MakeSlice(v.Type(), v.Len(), v.Cap())
		num := v.Len()

		for i := 0; i < num; i++ {
			slice.Index(i).Set(reflect.ValueOf(getValueInterface(v.Index(i))))
		}

		return slice.Interface()

	case reflect.String:
		return v.String()

	case reflect.Struct:
		st := reflect.New(v.Type()).Elem()
		num := v.NumField()

		for i := 0; i < num; i++ {
			st.Field(i).Set(reflect.ValueOf(getValueInterface(v.Field(i))))
		}

		return st.Interface()

	case reflect.UnsafePointer:
		ptr := reflect.New(v.Type())
		*(*uintptr)(unsafe.Pointer(ptr.Pointer())) = v.Pointer()
		return ptr.Elem().Interface()
	}

	panic("go-assert: never be here")
}
