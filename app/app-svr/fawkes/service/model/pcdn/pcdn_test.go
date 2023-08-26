package pcdn

import (
	"testing"
	"time"
)

func TestVersionId(t *testing.T) {
	ti := time.Now()       // 2022-09-08 15:49:42
	println(VersionId(ti)) // 20220908154942
}
