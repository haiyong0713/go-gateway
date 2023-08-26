package http

import (
	"context"
	"encoding/json"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	activityEcode "go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/model/like"
	l "go-gateway/app/web-svr/activity/interface/model/lottery_v2"
	riskmdl "go-gateway/app/web-svr/activity/interface/model/risk"
	"go-gateway/app/web-svr/activity/interface/service"
	"strconv"
	"strings"
	"time"
)

func reserveIncr(c *bm.Context) {
	arg := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
		Mid int64 `form:"mid" validate:"min=1"`
		Num int32 `form:"num" validate:"min=1"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(service.LikeSvc.InterReserve(c, arg.Sid, arg.Mid, arg.Num))
}

func reserve(c *bm.Context) {
	arg := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
		like.HTTPReserveReport
	})
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	if err := c.Bind(arg); err != nil {
		c.JSON(nil, err)
		return
	}
	report := &like.ReserveReport{
		From:     arg.From,
		Typ:      arg.Typ,
		Oid:      arg.Oid,
		Ip:       metadata.String(c, metadata.RemoteIP),
		Platform: arg.Platform,
		Mobiapp:  arg.Mobiapp,
		Buvid:    arg.Buvid,
		Spmid:    arg.Spmid,
	}

	c.JSON(nil, service.LikeSvc.AsyncReserve(c, arg.Sid, mid, 1, report))
}

func reserveCancel(c *bm.Context) {
	arg := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(nil, service.LikeSvc.ReserveCancel(c, arg.Sid, mid))
}

func reserveFollowing(c *bm.Context) {
	arg := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	midStr, ok := c.Get("mid")
	var mid int64
	if ok {
		mid = midStr.(int64)
	}
	c.JSON(service.LikeSvc.ReserveFollowing(c, arg.Sid, mid))
}

func awardSubjectState(c *bm.Context) {
	arg := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	state, err := service.LikeSvc.AwardSubjectState(c, arg.Sid, mid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(map[string]int{"state": state}, nil)
}

func rewardSubject(c *bm.Context) {
	arg := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, service.LikeSvc.AwardSubjectReward(c, arg.Sid, mid))
}

func awardSubject(c *bm.Context) {
	arg := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
		Mid int64 `form:"mid" validate:"min=1"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(nil, service.LikeSvc.InnerAwardSubject(c, arg.Sid, arg.Mid))
}

func reserveGroupProgress(c *bm.Context) {
	arg := new(api.ActivityProgressReq)
	if err := c.Bind(arg); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	if midStr != nil {
		arg.Mid = midStr.(int64)
	}
	if arg.Time > 0 {
		arg.Time = time.Now().Unix()
	}
	c.JSON(service.LikeSvc.ActivityProgress(c, arg))
}

func reserveSendPoint(c *bm.Context) {
	arg := new(struct {
		Sid     int64 `form:"sid" validate:"min=1"`
		GroupId int64 `form:"gid" validate:"min=1"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, service.LikeSvc.SendPoints(c, mid, arg.Sid, arg.GroupId))
}

func reserveProgress(c *bm.Context) {
	arg := new(struct {
		Sid   int64  `form:"sid" validate:"min=1"`
		Rules string `form:"rules"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	midStr, _ := c.Get("mid")
	if midStr != nil {
		mid = midStr.(int64)
	}
	rs := make([]*api.ReserveProgressRule, 0, 0)
	if err := json.Unmarshal([]byte(arg.Rules), &rs); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	rules := make([]*api.ReserveProgressRule, 0, 0)
	if mid == 0 {
		for _, r := range rs {
			if r.Dimension != api.GetReserveProgressDimension_User {
				rules = append(rules, r)
			}
		}
	} else {
		rules = rs
	}
	c.JSON(service.LikeSvc.GetReserveProgress(c, &api.GetReserveProgressReq{
		Sid:   arg.Sid,
		Mid:   mid,
		Rules: rules,
	}))
}

func delReserve(c *bm.Context) {
	arg := new(struct {
		Mid int64 `form:"mid" validate:"min=1"`
		ID  int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(nil, service.LikeSvc.DelCacheReserveOnly(c, arg.ID, arg.Mid))
}

func getRelationReserveInfo(c *bm.Context) {
	arg := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(service.LikeSvc.RelationReserveInfo(c, arg.ID, mid))
}

func doRelation(c *bm.Context) {
	arg := new(struct {
		ID int64 `form:"id" validate:"min=1"`
		like.HTTPReserveReport
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	report := &like.ReserveReport{
		From:     arg.From,
		Typ:      arg.Typ,
		Oid:      arg.Oid,
		Ip:       metadata.String(c, metadata.RemoteIP),
		Platform: arg.Platform,
		Mobiapp:  arg.Mobiapp,
		Buvid:    arg.Buvid,
		Spmid:    arg.Spmid,
	}

	res, err := service.LikeSvc.DoRelation(c, arg.ID, mid, report)
	if err != nil {
		if err != activityEcode.ActivityRelationIDNoExistErr {
			err = activityEcode.ActivityDoRelationErr
		}
		c.JSON(nil, err)
		return
	}

	c.JSON(res, nil)
}

func getRelationInfo(c *bm.Context) {
	arg := new(struct {
		ID       int64  `form:"id" validate:"min=1"`
		Specific string `form:"specific"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(service.LikeSvc.GetActRelationInfo(c, arg.ID, mid, arg.Specific))
}

func UpActReserveRelationContinuing(c *bm.Context) {
	arg := new(like.UpActReserveRelationContinuingArg)
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(service.LikeSvc.UpActReserveRelationContinuing(c, mid, arg))
}

func UpActReserveRelationOthers(c *bm.Context) {
	arg := new(like.UpActReserveRelationOthersArg)
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(service.LikeSvc.UpActReserveRelationOthers(c, mid, arg))
}

func CreateUpActReserve(c *bm.Context) {
	arg := new(like.CreateUpActReserveArgs)
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(service.LikeSvc.CreateUpActReserve(c, mid, arg))
}

func UpdateUpActReserve(c *bm.Context) {
	arg := new(like.UpdateUpActReserveArgs)
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(nil, service.LikeSvc.UpdateUpActReserve(c, mid, arg))
}

func UpActReserveInfo(c *bm.Context) {
	arg := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(service.LikeSvc.UpActReserveInfoH5(c, arg.ID, mid))
}

func CreateUpActCancel(c *bm.Context) {
	arg := new(struct {
		ID   int64 `form:"id" validate:"min=1"`
		From int64 `form:"from" validate:"min=1"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	req := &api.CancelUpActReserveReq{
		Mid:  mid,
		Sid:  arg.ID,
		From: api.UpCreateActReserveFrom(arg.From),
	}
	c.JSON(service.LikeSvc.CancelUpActReserve(c, req))
}

func reserveDoveAward(c *bm.Context) {
	arg := new(struct {
		Sid   int64 `form:"sid" validate:"min=1"`
		UpMid int64 `form:"up_mid"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	midStr, _ := c.Get("mid")
	if midStr != nil {
		mid = midStr.(int64)
	}

	upname := conf.Conf.ReserveDoveAct.DefaultUpName
	upface := conf.Conf.ReserveDoveAct.DefaultUpFace

	var awards []*l.RecordDetail

	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		upinfo, err := service.LikeSvc.GetReserveUpinfo(c, arg.UpMid, arg.Sid)
		log.Infoc(ctx, "reserveDoveAward GetReserveUpinfo up_mid:%v , %v , upinfo:%v", arg.UpMid, err, upinfo)
		if err == nil && upinfo != nil {
			upname = upinfo.Info.Name
			upface = upinfo.Info.Face
		}
		return
	})

	eg.Go(func(ctx context.Context) (err error) {
		// 校验用户是否预约up主活动
		follow, err := service.LikeSvc.ReserveFollowing(c, arg.Sid, mid)
		log.Infoc(ctx, "reserveDoveAward sid:[%d] , mid:[%d] , following:%v , %v", arg.Sid, mid, follow, err)
		nowTime := time.Now()

		// 用户在活动期间预约过
		// 当前还在活动期间
		if err == nil && follow.IsFollowing &&
			follow.Ctime.Time().Unix() > conf.Conf.ReserveDoveAct.Stime &&
			follow.Ctime.Time().Unix() < conf.Conf.ReserveDoveAct.Etime &&
			conf.Conf.ReserveDoveAct.Stime < nowTime.Unix() &&
			conf.Conf.ReserveDoveAct.Etime > nowTime.Unix() {

			params := risk(c, mid, riskmdl.ActionLottery)
			awards, err = service.LotterySvc.DoLottery(c, conf.Conf.ReserveDoveAct.AwardActId, mid, params, 1, false, "")
			if err != nil {
				log.Errorc(ctx, "reserveDoveAward eg.Wait DoLottery error(%v)", err)
			}
		}
		return
	})

	if err := eg.Wait(); err == nil && awards != nil && len(awards) > 0 && awards[0].GiftType != 0 {
		log.Infoc(c, "reserveDoveAward awards%v", awards)
		award := awards[0]
		if award.Extra == nil {
			award.Extra = make(map[string]string)
		}
		award.Extra["lottery_id"] = conf.Conf.ReserveDoveAct.AwardActId
		award.Extra["address_link"] = conf.Conf.Lottery.AddressLink + conf.Conf.ReserveDoveAct.AwardActId
		award.Extra["up_name"] = upname
		award.Extra["up_face"] = upface
		award.Extra["url"] = conf.Conf.ReserveDoveAct.ShareUrl
		c.JSON(award, nil)
		return
	}

	bless := ""
	if len(conf.Conf.ReserveDoveAct.BlessingMsg) > 0 {
		bless = conf.Conf.ReserveDoveAct.BlessingMsg[time.Now().Nanosecond()/1e6%len(conf.Conf.ReserveDoveAct.BlessingMsg)]
	}
	c.JSON(l.RecordDetail{
		GiftType: 0,
		GiftName: bless,
		Extra:    map[string]string{"msg": bless, "show_gift_type": "0", "up_name": upname, "up_face": upface, "url": conf.Conf.ReserveDoveAct.ShareUrl},
	}, nil)
}

func UpActReserveRelationInfo(c *bm.Context) {
	arg := new(like.UpActReserveRelationInfoArgs)
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ids := strings.Split(arg.IDs, ",")
	if len(ids) == 0 {
		return
	}
	sids := make([]int64, 0)
	for _, sid := range ids {
		tmp, _ := strconv.ParseInt(sid, 10, 64)
		sids = append(sids, int64(tmp))
	}
	req := &api.UpActReserveRelationInfoReq{
		Mid:  mid,
		Sids: sids,
	}
	c.JSON(service.LikeSvc.UpActReserveRelationInfo(c, req))
}
