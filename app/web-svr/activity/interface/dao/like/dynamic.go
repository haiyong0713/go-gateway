package like

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	pb "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/model/like"
	"net/url"
	"strconv"
)

const _dynamicAuth = "/lottery_svr/v0/lottery_svr/user_check"
const _dynamicBind = "/lottery_svr/v0/lottery_svr/bind"
const _dynamicPrizeInfo = "/lottery_svr/v1/lottery_svr/detail_by_lid"

func GetDynamicLotteryBizID(typ int64) (bizID int64, err error) {
	if typ == int64(pb.UpActReserveRelationType_Archive) {
		bizID = like.DynamicLotteryArcBizID
		return
	}
	if typ == int64(pb.UpActReserveRelationType_Live) {
		bizID = like.DynamicLotteryLiveBizID
		return
	}
	err = fmt.Errorf("illegal type")
	return
}

func (d *Dao) GetDynamicLotteryAuth(ctx context.Context, mid int64, typ int64) (auth bool, err error) {
	businessID, err := GetDynamicLotteryBizID(typ)
	if err != nil {
		err = errors.Wrap(err, "GetDynamicLotteryBizID err")
		return
	}

	params := url.Values{}
	params.Set("sender_uid", strconv.FormatInt(mid, 10))
	params.Set("business_type", strconv.FormatInt(businessID, 10))

	rsp := new(struct {
		Code int `json:"code"`
		Data struct {
			Result int64 `json:"result"`
		} `json:"data"`
	})
	err = d.client.Get(ctx, d.dynamicLotteryAuth, metadata.String(ctx, metadata.RemoteIP), params, &rsp)
	log.Infoc(ctx, "d.client.Get dynamicLotteryAuth req(%+v) reply(%+v)", params, rsp)
	if err != nil {
		err = errors.Wrapf(err, "d.client.Get err req(%+v) reply(%+v)", params, rsp)
		return
	}
	if rsp.Code != ecode.OK.Code() {
		err = fmt.Errorf("d.client.Get response err req(%+v) reply(%+v)", params, rsp)
		return
	}
	if rsp.Data.Result == 1 {
		auth = true
		return
	}
	return
}

func (d *Dao) BindReserveAndDynamicLottery(ctx context.Context, mid int64, relation *like.UpActReserveRelationInfo) (err error) {
	businessID, err := GetDynamicLotteryBizID(relation.Type)
	if err != nil {
		err = errors.Wrap(err, "GetDynamicLotteryBizID err")
		return
	}

	params := url.Values{}
	params.Set("lottery_id", relation.LotteryID)
	params.Set("business_type", strconv.FormatInt(businessID, 10))
	params.Set("business_id", strconv.FormatInt(relation.Sid, 10))
	params.Set("uid", strconv.FormatInt(relation.Mid, 10))
	params.Set("lottery_time", strconv.FormatInt(relation.LivePlanStartTime.Time().Unix(), 10))

	rsp := new(struct {
		Code int `json:"code"`
	})
	err = d.client.Post(ctx, d.dynamicLotteryBind, metadata.String(ctx, metadata.RemoteIP), params, &rsp)
	log.Infoc(ctx, "d.client.Get dynamicLotteryBind params(%+v)", params)
	if err != nil {
		err = errors.Wrapf(err, "d.client.Get err params(%+v)", params)
		return
	}
	if rsp.Code != ecode.OK.Code() {
		err = fmt.Errorf("d.client.Get response err params(%+v) result(%+v)", params, rsp)
		return
	}
	return
}

func (d *Dao) GetDynamicLotteryPrizeInfo(ctx context.Context, relation *like.UpActReserveRelationInfo) (res *pb.UpActReserveRelationPrizeInfo, err error) {
	res = new(pb.UpActReserveRelationPrizeInfo)

	params := url.Values{}
	params.Set("lottery_id", relation.LotteryID)

	rsp := new(struct {
		Code int `json:"code"`
		Data struct {
			FirstPrizeCmt    string `json:"first_prize_cmt"`
			SecondPrizeCmt   string `json:"second_prize_cmt"`
			ThirdPrizeCmt    string `json:"third_prize_cmt"`
			LotteryDetailUrl string `json:"lottery_detail_url"`
		} `json:"data"`
	})
	err = d.client.Post(ctx, d.dynamicLotteryPrizeInfo, metadata.String(ctx, metadata.RemoteIP), params, &rsp)
	log.Infoc(ctx, "d.client.Get dynamicLotteryPrizeInfo params(%+v)", params)
	if err != nil {
		err = errors.Wrapf(err, "d.client.Get err params(%+v)", params)
		return
	}
	if rsp.Code != ecode.OK.Code() {
		err = fmt.Errorf("d.client.Get response err params(%+v) result(%+v)", params, rsp)
		return
	}

	if rsp.Data.FirstPrizeCmt == "" {
		err = fmt.Errorf("rsp.Data.FirstPrizeCmt empty rsp(%+v)", rsp)
		return
	}

	res.Text = "预约有奖：" + rsp.Data.FirstPrizeCmt
	if rsp.Data.SecondPrizeCmt != "" {
		res.Text += "、" + rsp.Data.SecondPrizeCmt
	}
	if rsp.Data.ThirdPrizeCmt != "" {
		res.Text += "、" + rsp.Data.ThirdPrizeCmt
	}
	res.JumpUrl = rsp.Data.LotteryDetailUrl

	return
}
