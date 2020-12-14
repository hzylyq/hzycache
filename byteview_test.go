package hzycache

import (
	"fmt"
	"testing"
)

func TestByteView(t *testing.T) {
	for _, s := range []string{"", "x", "yy"} {
		for _, v := range []ByteView{of([]byte(s)), of(s)} {
			name := fmt.Sprintf("string %q, view %+v", s, v)
			if v.Len() != len(s) {
				t.Errorf("%s: Len = %d; want %d", name, v.Len(), len(s))
			}
			if v.String() != s {
				t.Errorf("%s: String = %q; want %q", name, v.String(), s)
			}
		}
	}
}

// of returns a byte view of the []byte or string in x.
func of(x interface{}) ByteView {
	if bytes, ok := x.([]byte); ok {
		return ByteView{b: bytes}
	}

	return ByteView{b: []byte(x.(string))}
}
