package jsondata

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/log"
	jsonmdl "go-gateway/app/web-svr/activity/interface/model/jsondata"
)

const (
	SummerGiftLink = "http://activity.hdslb.com/blackboard/static/jsonlist/228/6eVf3OAf7Z.json"
)

// GetSummerGift 获取夏日奖品
func (d *Dao) GetSummerGift(c context.Context, timeStamp int64) (res []*jsonmdl.SummerGiftList, err error) {
	params := url.Values{}
	params.Set("t", strconv.FormatInt(timeStamp, 10))
	res = make([]*jsonmdl.SummerGiftList, 0)
	if err = d.singleClient.Get(c, SummerGiftLink, "", params, &res); err != nil {
		log.Errorc(c, "GetSummerGift d.client.Get(%s) error(%+v)", SummerGiftLink+"?"+params.Encode(), err)
		return
	}
	log.Infoc(c, "GetSummerGift d.client.Get url(%s) params (%v) res(%v)", SummerGiftLink, params.Encode(), res)
	return
}
