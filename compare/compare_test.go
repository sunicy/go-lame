package compare

import (
	"testing"
	"fmt"
)

type TS struct {
	A string
	B struct {
		C int
		D interface{}
	}
}

type B struct {
	M string
}

func Test_Compare(t *testing.T) {

	diffs, err := Compare(
		&TS{
			A: "aaa",
			B: struct {
				C int
				D interface{}
			}{
				C: 3,
				D: &B{
					M: "mmm",
				},
			},
		},
		&TS{
			A: "aaa",
			B: struct {
				C int
				D interface{}
			}{
				C: 3,
				D: &B{
					M: "m2mm",
				},
			},
		},
	)
	fmt.Println(diffs)
	fmt.Println(err)
}

