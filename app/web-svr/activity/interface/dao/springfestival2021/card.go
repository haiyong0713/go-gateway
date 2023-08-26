package springfestival2021

import (
	"context"
	"go-common/library/log"

	springfestival2021 "go-gateway/app/web-svr/activity/interface/model/springfestival2021"
)

// GetMidCards 用户已经获得的卡
func (d *Dao) GetMidCards(c context.Context, mid int64) (res *springfestival2021.MidNums, err error) {
	cards, err := d.MidCardDetail(c, mid)
	if err != nil {
		log.Errorc(c, "d.MidCardDetail err(%v)", err)
	}
	if cards != nil && err == nil {
		return cards, nil
	}
	giftMid, err := d.MidNums(c, mid)
	if err != nil {
		log.Errorc(c, "d.MidNums(c, %d) err(%v)", mid, err)
		return nil, err
	}

	err = d.AddMidCardDetail(c, mid, giftMid)
	if err != nil {
		log.Errorc(c, " d.AddMidCardDetail err(%v)", err)
	}
	return giftMid, nil
}
