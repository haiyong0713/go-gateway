package http

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/model/bml"
	l "go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/app/web-svr/activity/interface/service"
	"time"
)

func bmlOnlineMyGuessList(ctx *bm.Context) {
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	myGuessList, err := service.BmlSvc.MyGuessList(ctx, mid)
	ctx.JSON(map[string]interface{}{
		"success_guess_list": myGuessList,
	}, err)
}

func bmlOnlineGuessDo(ctx *bm.Context) {
	v := new(struct {
		GuessType   int    `json:"guess_type" form:"guess_type" validate:"min=1"`
		GuessAnswer string `json:"guess_answer" form:"guess_answer" validate:"min=1"`
	})
	var err error
	if err = ctx.Bind(v); err != nil {
		return
	}

	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	var rewardConf *bml.RewardConf
	log.Infoc(ctx, "bmlOnlineGuessDo start guess:%v", mid)
	if rewardConf, err = service.BmlSvc.DoGuess(ctx, mid, v.GuessType, v.GuessAnswer); err != nil {
		ctx.JSON(nil, err)
		return
	}
	if v.GuessType == bml.GuessTypeCommon {
		// 查询关注关系
		var followRes []*l.FollowReply
		followRes, err := service.LikeSvc.GetUpsRelationData(ctx, mid, conf.Conf.BMLGuessAct.FollowMids)
		if err != nil {
			log.Errorc(ctx, "bmlOnlineGuessDo GetUpsRelationData mid(%d) error(%+v)", mid, err)
			ctx.JSON(nil, ecode.ActivityGetAccRelationGRPCErr)
			return
		}
		log.Infoc(ctx, "bmlOnlineGuessDo mid:[%d], followRes:[%v]", mid, len(followRes))
		if len(followRes) > 0 && followRes[0].MID > 0 {
			glist, err2 := service.BmlSvc.SendReward(ctx, mid, v.GuessType, v.GuessAnswer, &bml.RewardConf{
				RewardId:      conf.Conf.BMLGuessAct.CommonForeverRewardId,
				RewardVersion: conf.Conf.BMLGuessAct.CommonForeverStockVersion,
				StockLimit:    conf.Conf.BMLGuessAct.CommonForeverStock,
			})
			if err2 != ecode.BMLGuessRewardSendOutError {
				ctx.JSON(map[string]interface{}{
					"guess_result": glist,
				}, err2)
				return
			}
			log.Infoc(ctx, "bmlOnlineGuessDo , SendReward , glist:%v , err2:%v", glist, err2)
		}
	}

	glist, err2 := service.BmlSvc.SendReward(ctx, mid, v.GuessType, v.GuessAnswer, rewardConf)
	if err2 == ecode.BMLGuessRewardSendOutError {
		err3 := service.BmlSvc.CacheUserAnswerRecord(ctx, &bml.GuessRecordItem{
			GuessType: v.GuessType,
			GuessTime: time.Now().Unix(),
		}, mid)
		log.Infoc(ctx, "CacheUserAnswerRecord mid:%v ,err:%v", mid, err3)
	}
	ctx.JSON(map[string]interface{}{
		"guess_result": glist,
	}, err2)
}
