package show

import (
	"fmt"
	"testing"
)

func TestBatchEditQuerySQL(t *testing.T) {
	quers := []*SearchWebQuery{
		{
			ID:    1,
			SID:   2,
			Value: "test",
		},
		{
			ID:    2,
			SID:   2,
			Value: "test2",
		},
	}
	sql, param := BatchEditQuerySQL(quers)
	fmt.Println(sql)
	fmt.Println(param)
}
