package rewards

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/ecode"
	"go-common/library/log"
	xhttp "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/api"
	model "go-gateway/app/web-svr/activity/interface/model/rewards"
	"net/http"
	"time"
)

const (
	//漫画折扣卷
	rewardTypeComicsCoupon = "Comics"

	_cartoonBnj2021URI = "/twirp/activity.v0.Activity/SendRewards"
)

func init() {
	awardsSendFuncMap[rewardTypeComicsCoupon] = Client.comicsBnj2021CouponSender
	awardsConfigMap[rewardTypeComicsCoupon] = &model.ComicsCouponConfig{}
}

//comicsBnj2021CouponSender: 漫画折扣券发放
/*
http://comic.bilibili.co/api-doc/bainianji-2021/activity/v0/activity.html#activityv0activitysendrewards
   // 1-限免卡，只能新人发，先调用查询新人接口
   // 2-发放打折卡，每次发一张
   // 3-畅读卡月卡
   // 4-畅读卡季卡
   // 5-畅读卡年卡
*/
func (s *service) comicsBnj2021CouponSender(ctx context.Context, c *api.RewardsAwardInfo, mid int64, uniqueID, business string) (extraInfo map[string]string, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards.comicsBnj2021CouponSender mids(%v) uniqueId(%v) config (%v) error:%v", mid, uniqueID, c.JsonStr, err)
			return
		}
		s.sendAwardNotifyCard(mid, c, c.NotifyJumpUri1, c.NotifyJumpUri2)
	}()
	config := &model.ComicsCouponConfig{}
	if err = json.Unmarshal([]byte(c.JsonStr), &config); err != nil {
		return
	}
	param := &struct {
		Uid  int64 `json:"uid"`
		Type int64 `json:"type"`
	}{
		Uid:  mid,
		Type: config.Type,
	}

	paramJSON, err := json.Marshal(param)
	if err != nil {
		return
	}
	var req *http.Request
	var res struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	for i := 0; i < 3; i++ {
		time.Sleep(time.Duration(i*10) * time.Millisecond)
		if req, err = http.NewRequest("POST", s.comicBnj2021CouponURL, bytes.NewReader(paramJSON)); err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		err = xhttp.NewClient(s.c.HTTPClientComic).Do(ctx, req, &res)
		if err != nil {
			log.Errorc(ctx, "rewards.comicsBnj2021CouponSender fail, wait next retry. error: %v", err)
			continue
		}
		if res.Code != 0 {
			if res.Code == 4 {
				log.Errorc(ctx, "日志告警: 漫画优惠券(type=%v)库存不足", config.Type)
				err = fmt.Errorf("漫画优惠券(type=%v)库存不足", config.Type)
				break
			}
			if res.Code == 3 { //当前用户已领取
				err = nil
			}
			err = fmt.Errorf("code: %v, message: %v", ecode.Int(res.Code).Error(), res.Msg)
		}
		if err == nil {
			break
		}
		log.Errorc(ctx, "rewards.comicsBnj2021CouponSender fail, wait next retry. error: %v", err)
	}

	return
}
