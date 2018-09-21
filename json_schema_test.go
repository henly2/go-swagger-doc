package swagger

import (
	"testing"
	"reflect"
)

func testJsonSchema(t *testing.T)  {
	type H2 struct {
		AA string
		BB int
	}

	type H3 struct {
		AA string
		BB int
	}

	ss := &H2{}
	tt := reflect.TypeOf(ss)

	beginParse()
	// 1
	_, a1 := existAndAddType(tt)
	if a1 != false {
		t.Fatal()
	}

	// 2
	_, c1 := existAndAddType(tt)
	if c1 != true {
		t.Fatal()
	}

	return
}
