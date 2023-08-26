package v1

import (
	"fmt"
	"testing"
	"time"
)

func TestApplyV1PlayItems(t *testing.T) {
	o := &SortOption{
		Order: ListOrder_ORDER_REVERSE,
	}
	in := []*PlayItem{
		{Oid: 1},
		{Oid: 2},
		{Oid: 3},
		{Oid: 4},
		{Oid: 5},
	}
	t.Logf("%+v", o.ApplyOrderToV1PlayItems(in, nil))
}

func TestDetailItem_ApplyHistoryTag(t *testing.T) {
	_, offset := time.Now().Zone()
	now := time.Now()
	timeElapsed := (now.Unix() + int64(offset)) % (3600 * 24)
	fmt.Println(now.Add(-(time.Second * time.Duration(timeElapsed))).String())
}
