package bnj2021

import (
	"encoding/json"
	"testing"
)

type TestStruct struct {
	A int64 `json:"a,omitempty"`
}

// go test -v -count=1 service_test.go service.go reserve_lottery.go live_lottery.go biz_limit_tool.go exam_stats.go
func TestServiceBiz(t *testing.T) {
	t.Run("watch configuration biz", watchConfigurationTesting)
}

func watchConfigurationTesting(t *testing.T) {
	err := UpdateBnjReserveLiveAwardCfg()
	if err != nil {
		t.Error(err)

		return
	}

	bs, err := json.Marshal(BnjAward4ReserveLive)
	if err != nil {
		t.Error(err)

		return
	}

	t.Log(string(bs))

	//RegisterFileWatcher()
	//time.Sleep(time.Hour)
}
