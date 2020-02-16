package assert

import (
	"errors"
	"os"
	"testing"
)

func TestSample_Assert(t *testing.T) {
	a, b := 1, 2
	Assert(t, a > b)
}

func TestSample_AssertEqual(t *testing.T) {
	Equal(t, []int{1, 2}, []int{1})
}

func TestSample_AssertNotEqual(t *testing.T) {
	NotEqual(t, []int{1}, []int{1})
}

func TestSample_A_Assert(t *testing.T) {
	a := New(t)
	x, y := 1, 2
	a.Assert(x > y)
}

func TestSample_A_NilError(t *testing.T) {
	a := New(t)
	a.NilError(os.Open("path/to/a/file"))
}

func TestSample_A_NonNilError(t *testing.T) {
	a := New(t)
	f := func() (int, error) { return 0, errors.New("expected") }
	a.NilError(f())
}

func TestSample_A_Equal(t *testing.T) {
	a := New(t)
	a.Equal([]int{1, 2}, []int{1})
}

func TestSample_A_Use(t *testing.T) {
	a := New(t)
	v1 := 123
	v2 := []string{"wrong", "right"}
	v3 := v2[0]
	v4 := "not related"
	a.Use(&v1, &v2, &v3, &v4)

	a.Assert(v1 == 123 && v3 == "right")
}
