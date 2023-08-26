package http

import (
	"context"
	xecode "go-common/library/ecode"
	"go-gateway/app/web-svr/activity/interface/service"
	"strconv"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/conf"
	gameModel "go-gateway/app/web-svr/activity/interface/model/esports"
	l "go-gateway/app/web-svr/activity/interface/model/lottery_v2"
)

func addFavGame(c *bm.Context) {
	var mid int64
	args := new(struct {
		FirstGameId int64 `form:"first_game_id" validate:"min=1"`
		//FirstGameName string  `form:"first_game_name" `
		SecondGameId int64 `form:"second_game_id" default:"0"`
		//SecondGameName string   `form:"second_game_name" `
		ThirdGameId int64 `form:"third_game_id" default:"0"`
		//ThirdGameName  string  `form:"third_game_name" `
	})

	if err := c.Bind(args); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}

	c.JSON(service.EsportSvc.AddFavGame(c, mid, args.FirstGameId, args.SecondGameId, args.ThirdGameId))
}

func UserInfo(c *bm.Context) {
	var mid int64
	arg := new(struct {
		Sid        string `form:"sid" validate:"required"`
		ActionType int    `form:"action_type" default:"10" validate:"min=1" `
		Cid        int64  `form:"cid"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}

	var (
		fav  *gameModel.EsportsActFav
		incr int
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		fav, err = service.EsportSvc.UserInfo(c, mid)
		return
	})

	eg.Go(func(ctx context.Context) (err error) {
		incr, err = service.LotterySvc.CheckAddTimes(c, arg.Sid, mid, arg.Cid, arg.ActionType, 0)
		log.Infoc(ctx, "UserInfo CheckAddTimes incr:[%v] , err:[%v]", incr, err)
		if xecode.EqualError(ecode.ActivityLotteryAddTimesLimit, err) {
			incr = -1
			err = nil
		}
		return
	})

	if err := eg.Wait(); err != nil {
		c.JSON(nil, err)
		return
	}

	c.JSON(&gameModel.UserInfo{
		FavEsports:       fav,
		FavCompleted:     fav != nil && fav.FirstFavGameId > 0,
		CollectCompleted: incr <= 0,
	}, nil)
}

func settleHistory(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}

	// 查询关注关系
	followRes, err := service.LikeSvc.GetUpsRelationData(c, mid, conf.Conf.EsportsArena.UpMids)
	if err != nil {
		log.Errorc(c, "UpListRelation s.GetUpsRelationData mid(%d) error(%+v)", mid, err)
		c.JSON(nil, ecode.ActivityGetAccRelationGRPCErr)
		return
	}
	log.Infoc(c, "settleHistory mid:[%d], followRes:[%v]", mid, len(followRes))

	for _, v := range followRes {
		orderNo := strconv.FormatInt(mid, 10) + "_" + strconv.FormatInt(v.MID, 10) + "_" + strconv.FormatInt(l.TimesFollowType, 10)
		err2 := checkEsportsArenaLimit(c, conf.Conf.EsportsArena.Sid, l.TimesFollowType, 0, mid, orderNo)
		log.Infoc(c, "settleHistory  AddLotteryTimes err:%v", err2)
		if err2 == nil {
			err = service.LotterySvc.AddLotteryTimes(c, conf.Conf.EsportsArena.Sid, mid, 0, l.TimesFollowType, 0, orderNo, true)
			log.Infoc(c, "settleHistory  AddLotteryTimes my_mid:[%v] , order_no:[%v] , follow:[%v] , error(%+v)", mid, orderNo, *v, err)
		}
	}
	c.JSON(nil, nil)
}
