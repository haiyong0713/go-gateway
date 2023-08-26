package bnj2021

import (
	"fmt"
	"testing"
	"time"
)

// go test -v --count=1 biz_limit_tool_test.go biz_limit_tool.go service.go live_lottery.go reserve_lottery.go exam_stats.go
func TestBizLimitTool(t *testing.T) {
	if err := UpdateBizLimitRule(); err != nil {
		t.Error(err)

		return
	}

	RegisterFileWatcher()

	for {
		fmt.Println(isBizLimitReachedByBiz(limitKeyOfDraw, limitKeyOfDraw), time.Now())
		time.Sleep(200 * time.Millisecond)
	}
}
