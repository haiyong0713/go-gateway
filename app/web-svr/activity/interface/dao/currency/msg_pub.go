package currency

import (
	"context"
	"strconv"
	"time"

	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/interface/model/currency"
)

func (d *Dao) SendHeadMsg(c context.Context, mid int64) (err error) {
	var uids = []int64{mid}
	expireTime := xtime.Time(time.Now().Unix() + 30*86400)
	nowT := strconv.FormatInt(time.Now().Unix(), 10)
	midStr := strconv.FormatInt(mid, 10)
	msgID := nowT + midStr
	var headRewards = []*currency.HeadRewards{{RewardID: d.c.Live.HeadRewardID, ExpireTime: expireTime, Type: d.c.Live.HeadType}}
	msg := &currency.LiveHeadMsg{Uids: uids, MsgID: msgID, Source: d.c.Live.Source, Rewards: headRewards}
	if err = d.liveItemPub.Send(c, midStr, msg); err != nil {
		log.Error("SendHeadMsg: d.liveItemPub.Send(%v) error(%v)", msg, err)
		return
	}
	log.Info("SendHeadMsg: d.liveItemPub.Send(%v)", msg)
	return
}

func (d *Dao) SendPropsMsg(c context.Context, mid int64) (err error) {
	var uids = []int64{mid}
	nowT := strconv.FormatInt(time.Now().Unix(), 10)
	midStr := strconv.FormatInt(mid, 10)
	msgID := nowT + midStr
	var propsRewards = []*currency.PropsRewards{{RewardID: d.c.Live.PropsRewardID, ExpireTime: d.c.Live.ExpireTime, Type: d.c.Live.PropsType, Num: d.c.Live.Num, ExtraData: &currency.PropsExtraData{MsgID: msgID, Source: d.c.Live.Source}}}
	msg := &currency.LivePropsMsg{Uids: uids, MsgID: msgID, Source: d.c.Live.Source, Rewards: propsRewards}
	if err = d.liveItemPub.Send(c, midStr, msg); err != nil {
		log.Error("SendPropsMsg: d.liveItemPub.Send(%v) error(%v)", msg, err)
		return
	}
	log.Info("SendPropsMsg: d.liveItemPub.Send(%v)", msg)
	return
}
