package s10

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-gateway/app/web-svr/activity/interface/model/s10"

	"go-common/library/log"
)

func (d *Dao) LiveUserTitlePub(ctx context.Context, mid int64, uniqueID, extra string) error {
	var err error
	strs := strings.Split(extra, ":")
	if len(strs) < 2 {
		log.Errorc(ctx, "s10 goods info error! uniqueID:%s", uniqueID)
		return fmt.Errorf("商品信息错误")
	}
	ints := make([]int64, len(strs))
	for i := 0; i < 2; i++ {
		if ints[i], err = strconv.ParseInt(strs[i], 10, 64); err != nil {
			log.Errorc(ctx, "s10 d.dao.LiveUserTitlePub extra  to Num fail mid:%d, extra:%s error:%v", mid, extra, err)
			return err
		}
	}
	res := &s10.UserTitle{
		MsgID:  uniqueID,
		Source: 1508,
		Uids:   []int64{mid},
		Rewards: []*s10.UserTitleReward{{
			RewardID:   int32(ints[1]),
			ExpireTime: time.Now().Unix() + ints[0]*24*60*60,
			Type:       5,
			ExtraData:  &s10.UserTitleExtraData{},
		}},
	}
	if err = d.liveDataBusPub.Send(ctx, fmt.Sprintf("%d", mid), res); err != nil {
		log.Errorc(ctx, "s10 d.dao.LiveUserTitlePubDataBus(mid:%s,extra:%s) error(%v)", uniqueID, extra, err)
	}
	return err
}

func (d *Dao) LiverBulletPub(ctx context.Context, mid int64, uniqueID, extra string) error {
	var (
		err     error
		roomeID []int64
	)
	strs := strings.Split(extra, ":")
	if len(strs) < 4 {
		log.Errorc(ctx, "s10 goods info error! uniqueID:%s", uniqueID)
		return fmt.Errorf("商品信息错误")
	}
	ints := make([]int64, len(strs))
	for i := 0; i < 4; i++ {
		switch i {
		case 0:
			if ints[i], err = strconv.ParseInt(strs[i], 10, 64); err != nil {
				log.Errorc(ctx, "s10 d.dao.LiverBulletPub extra  to Num fail mid:%d, extra:%s error:%v", mid, extra, err)
				return err
			}
		case 2:
			if ints[i], err = strconv.ParseInt(strs[i], 16, 64); err != nil {
				log.Errorc(ctx, "s10 d.dao.LiverBulletPub extra  to Num fail mid:%d, extra:%s error:%v", mid, extra, err)
				return err
			}
		case 3:
			intstr := strings.Split(strs[i], ",")
			roomeID = make([]int64, len(intstr))
			for j, str := range intstr {
				if roomeID[j], err = strconv.ParseInt(str, 10, 64); err != nil {
					log.Errorc(ctx, "s10 d.dao.LiverBulletPub extra  to Num fail mid:%d, extra:%s error:%v", mid, extra, err)
					return err
				}
			}
		}
	}
	rewards := make([]*s10.BulletReward, 0, len(roomeID))
	for _, v := range roomeID {
		rewards = append(rewards, &s10.BulletReward{
			RewardID:   6,
			ExpireTime: time.Now().Unix() + ints[0]*24*60*60,
			Type:       11,
			ExtraData: &s10.BulletExtraData{
				Type:   "color",
				Value:  ints[2],
				RoomID: int32(v),
			}})
	}
	res := &s10.Bullet{
		Uids:    []int64{mid},
		MsgID:   uniqueID,
		Source:  1508,
		Rewards: rewards,
	}
	if err = d.liveDataBusPub.Send(ctx, fmt.Sprintf("%d", mid), res); err != nil {
		log.Errorc(ctx, "s10 d.dao.LiveUserTitlePubDataBus(uniqueID:%s,extra:%s) error(%v)", uniqueID, extra, err)
	}
	return err
}
