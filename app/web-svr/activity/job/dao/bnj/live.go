package bnj

import (
	"context"
	"strconv"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/bnj"
)

func (d *Dao) SendLiveItem(c context.Context, mid int64) (err error) {
	nowT := strconv.FormatInt(time.Now().Unix(), 10)
	midStr := strconv.FormatInt(mid, 10)
	msgID := nowT + midStr
	headRewards := &bnj.LiveMsg{
		Uids:   []int64{mid},
		MsgID:  msgID,
		Source: d.c.Bnj2020.Live.Source,
	}
	for _, v := range d.c.Bnj2020.Live.Items {
		item := &bnj.LiveReward{
			RewardID:   v.RewardID,
			ExpireTime: d.c.Bnj2020.Live.ExpireTime,
			Type:       v.Type,
		}
		switch v.Type {
		case 6: //头衔续期卡
			item.StartTime = v.StartTime
			item.Num = v.Num
			item.ExtraData = &struct {
				Source int64 `json:"source"`
			}{Source: d.c.Bnj2020.Live.Source}
		case 8: //辣条
			item.Num = v.Num
		case 11: //弹幕颜色
			item.ExtraData = &struct {
				Type   string `json:"type"`
				Value  int64  `json:"value"`
				Roomid int64  `json:"roomid"`
			}{Type: v.ExtraData.Type, Value: v.ExtraData.Value, Roomid: v.ExtraData.Roomid}
		}
		headRewards.Rewards = append(headRewards.Rewards, item)
	}
	if err = d.liveItemPub.Send(c, midStr, headRewards); err != nil {
		log.Error("SendHeadMsg: d.liveItemPub.Send(%+v) error(%v)", headRewards, err)
	}
	return
}
