package tool

import (
	"testing"
)

// go test -v berserker_test.go berserker.go
func TestSignBiz(t *testing.T) {
	_, signStr := genSignStr(
		"ac59b99d3f6d0ec1dc82a9a2aefb01a5",
		"e69f9822c4a8d8820e1d6a818de393e3",
		"1.0",
		BerserkerSignMethodOfMD5,
		"2018-02-07 17:24:17",
		"")
	if signStr != "8B9E843BA64AEE747D68EFC8FCB192DA" {
		t.Errorf("sign should as %v, but now %v", "8B9E843BA64AEE747D68EFC8FCB192DA", signStr)
	}
}
