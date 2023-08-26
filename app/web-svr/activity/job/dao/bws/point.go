package bws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/bws"

	"github.com/pkg/errors"
)

const (
	_rechargeAwardURI = "/x/internal/activity/bws/recharge/award"
)

func bwsLotteryKey(bid, awardID int64) string {
	return fmt.Sprintf("bws_lott_%d_%d", bid, awardID)
}

func bwsSpecLotteryKey(bid int64) string {
	return fmt.Sprintf("bws_spec_lott_%d", bid)
}

// RechargeAward .
func (d *Dao) RechargeAward(c context.Context, bid int64) (awards []*bws.Award, err error) {
	var res struct {
		Code int                `json:"code"`
		Data *bws.RechargeAward `json:"data"`
	}
	params := url.Values{}
	params.Set("bid", strconv.FormatInt(bid, 10))
	if err = d.httpClient.Get(c, d.rechargeURL, "", params, &res); err != nil {
		log.Error("AddAchieve:d.httpClient.Get bid(%d) error(%v)", bid, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.rechargeURL+"?"+params.Encode())
	}
	if res.Data != nil {
		for _, v := range res.Data.Recharge {
			if v != nil && len(v.Unlock) > 0 {
				awards = append(awards, v.Unlock...)
			}
		}
	}
	return
}

// SetLotteryCache .
func (d *Dao) SetLotteryCache(c context.Context, bid, awardID int64, users []*bws.LotteryUser) (err error) {
	var (
		conn = d.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	key := bwsLotteryKey(bid, awardID)
	if bs, err = json.Marshal(users); err != nil {
		log.Error("SetLotteryCache json.Marsha(%v) error(%v)", users, err)
		return
	}
	if _, err = conn.Do("SET", key, bs); err != nil {
		log.Error("conn.Send(SET, %s, %s) error(%v)", key, bs, err)
	}
	return
}

func (d *Dao) SetSpecLotteryCache(c context.Context, bid int64, user *bws.LotteryUser) (err error) {
	var (
		conn = d.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	key := bwsSpecLotteryKey(bid)
	if bs, err = json.Marshal(user); err != nil {
		log.Error("SetLotteryCache json.Marsha(%v) error(%v)", user, err)
		return
	}
	if _, err = conn.Do("SET", key, bs); err != nil {
		log.Error("conn.Send(SET, %s, %s) error(%v)", key, bs, err)
	}
	return
}
