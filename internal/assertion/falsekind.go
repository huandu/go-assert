package assertion

import (
	"reflect"
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
