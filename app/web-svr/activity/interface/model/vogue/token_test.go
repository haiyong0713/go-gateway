package model

import (
	"testing"
)

func TestToken(t *testing.T) {
	for i := int64(0); i < 10000000; i++ {
		if a, e := TokenEncode(TokenDecode(i)); e != nil {
			panic(i)
		} else if a != i {
			panic(i)
		}
	}
}

func TestTokenUnique(t *testing.T) {
	m := make(map[string]struct{})
	for i := int64(0); i < 10000000; i++ {
		t := TokenDecode(i)
		if _, ok := m[t]; ok {
			panic(t)
		}
		m[t] = struct{}{}
	}
}
