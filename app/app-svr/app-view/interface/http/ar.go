package http

import bm "go-common/library/net/http/blademaster"

func doAr(c *bm.Context) {
	v := new(struct {
		Sid            int64   `form:"sid" validate:"min=1"`
		TotalTime      int64   `form:"total_time"`
		MatchedPercent float32 `form:"matched_percent"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	// 19年ar打卡活动仅开放过蓝版白名单，产品已确认可下线（https://www.tapd.bilibili.co/20095661/prong/stories/view/1120095661002124671）
	c.JSON(map[string]interface{}{"days": 0}, nil)
}
