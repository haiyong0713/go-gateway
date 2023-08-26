package service

import (
	"context"
	"time"

	"go-common/library/log"
	"go-common/library/net/metadata"
	arcecode "go-gateway/app/app-svr/archive/ecode"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/web/interface/model"

	coinmdl "git.bilibili.co/bapis/bapis-go/community/service/coin"
	thumbmdl "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
)

var _emptyCoinArcList = make([]*model.CoinArc, 0)

// Coins get archive User added coins.
func (s *Service) Coins(c context.Context, mid, aid int64) (res *model.ArchiveUserCoins, err error) {
	var rs *coinmdl.ItemUserCoinsReply
	if rs, err = s.coinGRPC.ItemUserCoins(c, &coinmdl.ItemUserCoinsReq{Mid: mid, Aid: aid, Business: model.CoinArcBusiness}); err != nil {
		log.Error("s.coinGRPC.ItemUserCoins(%d,%d) error(%v)", mid, aid, err)
		return
	}
	res = new(model.ArchiveUserCoins)
	if rs != nil {
		res.Multiply = rs.Number
	}
	return
}

// AddCoin add coin to archive.
func (s *Service) AddCoin(ctx context.Context, aid, mid, upID, multiply, avtype int64, business string, selectLike int, riskParams *model.RiskManagement) (res *model.AddCoinRes, err error) {
	res = &model.AddCoinRes{
		Like:        false,
		IsRisk:      false,
		GaiaResType: model.GaiaResponseType_Default,
	}
	var (
		pubTime int64
		typeID  int32
		maxCoin int64 = 2
	)
	var arcReply *arcmdl.ArcReply
	if avtype == model.CoinAddArcType {
		if arcReply, err = s.arcGRPC.Arc(ctx, &arcmdl.ArcRequest{Aid: aid}); err != nil {
			log.Error("s.arcGRPC.Arc(%v) error(%v)", aid, err)
			return
		}
		if arcReply == nil || arcReply.Arc == nil || !arcReply.Arc.IsNormal() {
			err = arcecode.ArchiveNotExist
			return
		}
		//pgc 稿件没有maxcoin限制
		if arcReply.Arc.AttrVal(arcmdl.AttrBitIsPGC) != arcmdl.AttrYes && arcReply.Arc.Copyright == int32(arcmdl.CopyrightCopy) {
			maxCoin = 1
		}
		upID = arcReply.Arc.Author.Mid
		typeID = arcReply.Arc.TypeID
		pubTime = int64(arcReply.Arc.PubDate)
		action := _coinAddAction
		scene := _coinAddScene
		if selectLike == 1 {
			action = _coinToLikeAction
			scene = _coinToLikeScene
		}
		riskParams.Action = action
		riskParams.Scene = scene
		riskParams.Pubtime = arcReply.Arc.PubDate.Time().Format("2006-01-02 15:04:05")
		riskParams.Title = arcReply.Arc.Title
		riskParams.PlayNum = arcReply.Arc.Stat.View
		riskResult := s.RiskVerifyAndManager(ctx, riskParams)
		if riskResult != nil {
			res.GaiaResType = riskResult.GaiaResType
			res.IsRisk = riskResult.IsRisk
			res.GaiaData = riskResult.GaiaData
			return res, nil
		}
	}
	ip := metadata.String(ctx, metadata.RemoteIP)
	arg := &coinmdl.AddCoinReq{
		IP:       ip,
		Mid:      mid,
		Upmid:    upID,
		MaxCoin:  maxCoin,
		Aid:      aid,
		Business: business,
		Number:   multiply,
		Typeid:   typeID,
		PubTime:  pubTime,
		Platform: "pc",
	}
	if _, err = s.coinGRPC.AddCoin(ctx, arg); err == nil && avtype == model.CoinAddArcType && selectLike == 1 {
		if _, err = s.thumbupGRPC.Like(ctx, &thumbmdl.LikeReq{Business: _businessLike, Mid: mid, UpMid: upID, MessageID: aid, Action: thumbmdl.Action_ACTION_LIKE, IP: ip, Platform: "pc"}); err != nil {
			log.Error("AddCoin s.thumbupGRPC.Like  mid(%d) upID(%d) aid(%d) error(%+v)", mid, upID, aid, err)
			err = nil
		} else {
			res.Like = true
		}
	}
	return
}

// CoinExp get coin exp today
func (s *Service) CoinExp(c context.Context, mid int64) (exp int64, err error) {
	var todayExp *coinmdl.TodayExpReply
	if todayExp, err = s.coinGRPC.TodayExp(c, &coinmdl.TodayExpReq{Mid: mid}); err != nil {
		log.Error("CoinExp s.coinGRPC.TodayExp mid(%d) error(%v)", mid, err)
		err = nil
		return
	}
	exp = todayExp.Exp
	return
}

// CoinList get coin list.
// nolint:gomnd
func (s *Service) CoinList(c context.Context, mid int64, pn, ps int) (list []*model.CoinArc, count int, err error) {
	var (
		coinReply *coinmdl.ListReply
		aids      []int64
		arcsReply *arcmdl.ArcsReply
	)
	if coinReply, err = s.coinGRPC.List(c, &coinmdl.ListReq{Mid: mid, Business: model.CoinArcBusiness, Ts: time.Now().Unix()}); err != nil {
		log.Error("CoinList s.coinGRPC.List(%d) error(%v)", mid, err)
		err = nil
		list = _emptyCoinArcList
		return
	}
	existAids := make(map[int64]int64, len(coinReply.List))
	afVideos := make(map[int64]*coinmdl.ModelList, len(coinReply.List))
	for _, v := range coinReply.List {
		if _, ok := existAids[v.Aid]; ok {
			afVideos[v.Aid].Number += v.Number
			continue
		}
		afVideos[v.Aid] = v
		aids = append(aids, v.Aid)
		existAids[v.Aid] = v.Aid
	}
	count = len(aids)
	start := (pn - 1) * ps
	end := pn * ps
	switch {
	case start > count:
		aids = aids[:0]
	case end >= count:
		aids = aids[start:]
	default:
		aids = aids[start:end]
	}
	if len(aids) == 0 {
		list = _emptyCoinArcList
		return
	}
	if arcsReply, err = s.arcGRPC.Arcs(c, &arcmdl.ArcsRequest{Aids: aids}); err != nil {
		log.Error("CoinList s.arcGRPC.Arcs(%v) error(%v)", aids, err)
		err = nil
		list = _emptyCoinArcList
		return
	}
	for _, aid := range aids {
		if arc, ok := arcsReply.Arcs[aid]; ok && arc.IsNormal() {
			if arc.Access >= 10000 {
				arc.Stat.View = 0
			}
			if item, ok := afVideos[aid]; ok {
				model.ClearAttrAndAccess(arc)
				list = append(list, &model.CoinArc{Arc: arc, Bvid: s.avToBv(arc.Aid), Coins: item.Number, Time: item.Ts})
			}
		}
	}
	if len(list) == 0 {
		list = _emptyCoinArcList
	}
	return
}
