package service

import (
	"strings"
	"testing"
)

func TestSwitchFallThrough(t *testing.T) {
	tcs := []struct {
		Thumb, Coin, Fav bool
		Expected         string
	}{
		{
			true, true, true, "三连成功",
		},
		{
			false, false, false, "三连失败",
		},
		{
			false, true, true, "点赞失败",
		},
		{
			true, false, true, "投币失败",
		},
		{
			true, true, false, "收藏失败",
		},
		{
			false, false, true, "点赞投币失败",
		},
		{
			false, true, false, "点赞收藏失败",
		},
		{
			true, false, false, "投币收藏失败",
		},
	}
	for i, tc := range tcs {
		ret := ""
		if tc.Thumb && tc.Coin && tc.Fav {
			ret = "三连成功"
		} else if !tc.Thumb && !tc.Coin && !tc.Fav {
			ret = "三连失败"
		} else {
			failedActs := make([]string, 0, 2)
			for _, o := range []struct {
				Text   string
				Failed bool
			}{
				{"点赞", !tc.Thumb},
				{"投币", !tc.Coin},
				{"收藏", !tc.Fav},
			} {
				if o.Failed {
					failedActs = append(failedActs, o.Text)
				}
			}
			ret = strings.Join(failedActs, "") + "失败"
		}
		if ret != tc.Expected {
			t.Errorf("failed case %d Expected(%s) but got(%s)", i, tc.Expected, ret)
		}
	}

}
