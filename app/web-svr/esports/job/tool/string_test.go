package tool

import (
	"fmt"
	"testing"
)

// go test -v string_test.go string.go
func TestInt64JoinStr(t *testing.T) {
	t.Run("test 0 elems", testZeroElems)
	t.Run("test 1 elems", testOneElems)
	t.Run("test 1000 elems", testOneThousandElems)
}

func testZeroElems(t *testing.T) {
	arr := make([]int64, 0)
	expected := ""
	caret := ""

	for i := 0; i < 0; i++ {
		arr = append(arr, int64(i))
		if i == 0 {
			caret = ""
		} else {
			caret = ","
		}
		expected = fmt.Sprintf("%v%v%v", expected, caret, i)
	}

	newStr := Int64JoinStr(arr, ",")
	if newStr != expected {
		t.Errorf("int64JoibStr(%v) is not equal expected(%v)", newStr, expected)
	}
}

func testOneElems(t *testing.T) {
	arr := make([]int64, 0)
	expected := ""
	caret := ""

	for i := 0; i < 1; i++ {
		arr = append(arr, int64(i))
		if i == 0 {
			caret = ""
		} else {
			caret = ","
		}
		expected = fmt.Sprintf("%v%v%v", expected, caret, i)
	}

	newStr := Int64JoinStr(arr, ",")
	if newStr != expected {
		t.Errorf("int64JoibStr(%v) is not equal expected(%v)", newStr, expected)
	}
}

func testOneThousandElems(t *testing.T) {
	arr := make([]int64, 0)
	expected := ""
	caret := ""

	for i := 0; i < 1000; i++ {
		arr = append(arr, int64(i))
		if i == 0 {
			caret = ""
		} else {
			caret = ","
		}
		expected = fmt.Sprintf("%v%v%v", expected, caret, i)
	}

	newStr := Int64JoinStr(arr, ",")
	if newStr != expected {
		t.Errorf("int64JoibStr(%v) is not equal expected(%v)", newStr, expected)
	}
}
