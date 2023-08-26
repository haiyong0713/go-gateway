package grpc

import (
	"context"
	"go-gateway/app/web-svr/activity/interface/model/bwsonline"
	"strings"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/rpc/warden/ratelimiter/quota"
	"go-gateway/app/web-svr/activity/ecode"
	pb "go-gateway/app/web-svr/activity/interface/api"
	con "go-gateway/app/web-svr/activity/interface/conf"
	like "go-gateway/app/web-svr/activity/interface/model/like"
	lmdl "go-gateway/app/web-svr/activity/interface/model/like"
	rankmdl "go-gateway/app/web-svr/activity/interface/model/rank_v3"
	riskmdl "go-gateway/app/web-svr/activity/interface/model/risk"
	"go-gateway/app/web-svr/activity/interface/service"
	bnj2021 "go-gateway/app/web-svr/activity/interface/service/newyear2021"
	"go-gateway/app/web-svr/activity/interface/service/wishes_2021_spring"
	"strconv"

	api "git.bilibili.co/bapis/bapis-go/account/service"
)

// New new a grpc server.
func New(c *warden.ServerConfig, quotaCfg *quota.Config) *warden.Server {
	limiter := quota.New(quotaCfg)
	ws := warden.NewServer(c)
	ws.Use(limiter.Limit())
	svc := &activityService{}
	pb.InitLocalActivityServer(svc)
	pb.RegisterActivityServer(ws.Server(), svc)
	ws, err := ws.Start()
	if err != nil {
		panic(err)
	}
	return ws
}

type activityService struct {
}

var _ pb.ActivityServer = &activityService{}

const (
	_singleLottery = 1
	_isInternal    = true
)

func (s *activityService) CheckReserveDoveAct(ctx context.Context, req *pb.CheckReserveDoveActReq) (*pb.CheckReserveDoveActReply, error) {
	return service.LikeSvc.CheckReserveDoveAct(ctx, req)
}
func (s *activityService) CommonActivityUserCommit(ctx context.Context, req *pb.CommonActivityUserCommitReq) (
	reply *pb.CommonActivityUserCommitReply, err error) {
	reply = new(pb.CommonActivityUserCommitReply)
	{
		reply.Status = 0
	}
	err = wishes_2021_spring.InsertUserCommit2DB(ctx, req)
	if err != nil {
		reply.Status = 1
	}

	return
}

func (s *activityService) IncrStockInCache(ctx context.Context, req *pb.GiftStockReq) (replay *pb.NoReply, err error) {
	replay = new(pb.NoReply)
	if IncrNum, err := service.BwsOnlineSvc.IncrReserveStock(ctx, req); err != nil {
		log.Errorc(ctx, "IncrStockInCache  failed:%+v , replay:%v", err, IncrNum)
	}
	return
}

func (s *activityService) SyncGiftStockInCache(ctx context.Context, req *pb.GiftStockReq) (replay *pb.SyncGiftStockResp, err error) {
	log.Infoc(ctx, "SyncGiftStockInCache req:%v", *req)
	if replay, err = service.BwsOnlineSvc.SyncStock(ctx, req); err != nil {
		log.Errorc(ctx, "SyncGiftStockInCache  failed:%+v , replay:%v", err, replay)
	}
	return
}

func (s *activityService) BwParkBeginReserveList(ctx context.Context, req *pb.BwParkBeginReserveReq) (replay *pb.BwParkBeginReserveResp, err error) {
	log.Infoc(ctx, "BwParkBeginReserveList req:%v", *req)
	replay = new(pb.BwParkBeginReserveResp)
	replay.ReserveList, err = service.BwsOnlineSvc.GetBeginReserve(ctx, req)
	return
}

func (s *activityService) CommonActivityAuditPub(ctx context.Context, req *pb.CommonActivityAuditPubReq) (
	reply *pb.NoReply, err error) {

	reply = new(pb.NoReply)
	err = wishes_2021_spring.ActAuditMertialPub(ctx, req)
	return
}

func (s *activityService) ModuleConfig(c context.Context, r *pb.ModuleConfigReq) (rs *pb.ModuleConfigReply, err error) {
	return nil, xecode.NothingFound
}

// NatInfoFromForeign .
func (s *activityService) NatInfoFromForeign(c context.Context, r *pb.NatInfoFromForeignReq) (rs *pb.NatInfoFromForeignReply, err error) {
	return &pb.NatInfoFromForeignReply{}, nil
}

// ActSubProtocol .
func (s *activityService) ActSubProtocol(c context.Context, r *pb.ActSubProtocolReq) (rs *pb.ActSubProtocolReply, err error) {
	var res *lmdl.SubProtocol
	if res, err = service.LikeSvc.ActProtocol(c, &lmdl.ArgActProtocol{Sid: r.Sid}); err != nil {
		return
	}
	rs = &pb.ActSubProtocolReply{}
	rs.Subject = &pb.Subject{}
	rs.Subject.DeepCopyFromSubjectItem(res.SubjectItem)
	if res.ActSubjectProtocol != nil {
		rs.Protocol = &pb.ActSubjectProtocol{}
		rs.Protocol.DeepCopyFromActSubjectProtocol(res.ActSubjectProtocol)
	}
	if len(res.Rules) > 0 {
		for _, v := range res.Rules {
			if v == nil {
				continue
			}
			reserveRule := &pb.ReserveRule{}
			reserveRule.DeepCopyFromActSubjectProtocol(v)
			rs.Rules = append(rs.Rules, reserveRule)
		}
	}
	return
}

// ActSubProtocol .
func (s *activityService) ActSubsProtocol(c context.Context, r *pb.ActSubsProtocolReq) (rs *pb.ActSubsProtocolReply, err error) {
	var res map[int64]*lmdl.SubProtocol
	if res, err = service.LikeSvc.ActsProtocol(c, r.Sids); err != nil {
		return
	}
	rs = &pb.ActSubsProtocolReply{}
	rs.List = make(map[int64]*pb.ActSubProtocolReply)
	for k, v := range res {
		tmp := &pb.ActSubProtocolReply{}
		tmp.Subject = &pb.Subject{}
		tmp.Subject.DeepCopyFromSubjectItem(v.SubjectItem)
		if v.ActSubjectProtocol != nil {
			tmp.Protocol = &pb.ActSubjectProtocol{}
			tmp.Protocol.DeepCopyFromActSubjectProtocol(v.ActSubjectProtocol)
		}
		if len(v.Rules) > 0 {
			for _, v := range v.Rules {
				if v == nil {
					continue
				}
				reserveRule := &pb.ReserveRule{}
				reserveRule.DeepCopyFromActSubjectProtocol(v)
				tmp.Rules = append(tmp.Rules, reserveRule)
			}
		}
		rs.List[k] = tmp
	}
	return
}

// ActSubject .
func (s *activityService) ActSubject(c context.Context, r *pb.ActSubjectReq) (rs *pb.ActSubjectReply, err error) {
	var sub *lmdl.SubjectItem
	if sub, err = service.LikeSvc.ActSubject(c, r.Sid); err != nil {
		return
	}
	if sub == nil || sub.ID == 0 {
		err = ecode.ActivityHasOffLine
		return
	}
	rs = &pb.ActSubjectReply{}
	tp := &pb.Subject{}
	tp.DeepCopyFromSubjectItem(sub)
	rs.Subject = tp
	return
}

// ActSubjects .
func (s *activityService) ActSubjects(c context.Context, r *pb.ActSubjectsReq) (rs *pb.ActSubjectsReply, err error) {
	var sub map[int64]*lmdl.SubjectItem
	if sub, err = service.LikeSvc.ActSubjects(c, r.Sids); err != nil {
		return
	}
	rs = &pb.ActSubjectsReply{}
	rs.List = make(map[int64]*pb.Subject, len(sub))
	for _, v := range sub {
		tp := &pb.Subject{}
		tp.DeepCopyFromSubjectItem(v)
		rs.List[v.ID] = tp
	}
	return
}

// NatConfig .
func (s *activityService) NatConfig(c context.Context, r *pb.NatConfigReq) (rs *pb.NatConfigReply, err error) {
	return nil, ecode.NativePageOffline
}

// NatConfig .
func (s *activityService) BaseConfig(c context.Context, r *pb.BaseConfigReq) (rs *pb.BaseConfigReply, err error) {
	return nil, ecode.NativePageOffline
}

// ActLiked .
func (s *activityService) ActLiked(c context.Context, r *pb.ActLikedReq) (rs *pb.ActLikedReply, err error) {
	var act *lmdl.ActReply
	if act, err = service.LikeSvc.LikeAct(c, &lmdl.ParamAddLikeAct{Sid: r.Sid, Lid: r.Lid, Score: r.Score}, r.Mid); err != nil {
		return
	}
	if act == nil {
		err = xecode.NothingFound
		return
	}
	rs = &pb.ActLikedReply{Lid: act.Lid, Score: act.Score}
	return
}

// ModuleMixExt .
func (s *activityService) ModuleMixExt(c context.Context, r *pb.ModuleMixExtReq) (rs *pb.ModuleMixExtReply, err error) {
	return nil, xecode.NothingFound
}

// ModuleMixExt .
func (s *activityService) ModuleMixExts(c context.Context, r *pb.ModuleMixExtsReq) (rs *pb.ModuleMixExtsReply, err error) {
	return &pb.ModuleMixExtsReply{}, nil
}

// ActLikes .
func (s *activityService) ActLikes(c context.Context, r *pb.ActLikesReq) (rs *pb.LikesReply, err error) {
	var list *lmdl.ActLikes
	if r.Pn > 0 {
		r.Offset = -1
	}
	if list, err = service.LikeSvc.ActLikes(c, &lmdl.ArgActLikes{Sid: r.Sid, Mid: r.Mid, Offset: r.Offset, SortType: int(r.SortType), Pn: int(r.Pn), Ps: int(r.Ps)}); err != nil {
		return
	}
	if list == nil {
		err = xecode.NothingFound
		return
	}
	rs = &pb.LikesReply{Total: list.Total, HasMore: list.HasMore, Offset: list.Offset}
	if list.Sub != nil {
		rs.Subject = &pb.Subject{}
		rs.Subject.DeepCopyFromSubjectItem(list.Sub)
	}
	rs.List = make([]*pb.ItemObj, 0, len(list.List))
	for _, v := range list.List {
		rs.List = append(rs.List, convertModelList(v))
	}
	return
}

func convertModelList(itemObj *lmdl.ItemObj) *pb.ItemObj {
	if itemObj == nil || itemObj.Item == nil {
		return nil
	}
	tem := &pb.Item{}
	tem.DeepCopyFromItem(itemObj.Item)
	return &pb.ItemObj{
		Item:     tem,
		Score:    itemObj.Score,
		HasLiked: itemObj.HasLiked,
	}
}

// GuessAdd  add guess.
func (s *activityService) GuessAdd(c context.Context, r *pb.GuessAddReq) (rs *pb.NoReply, err error) {
	rs, err = service.LikeSvc.GuessAdd(c, r)
	return
}

// GuessEdit  edit  guess.
func (s *activityService) GuessEdit(c context.Context, r *pb.GuessEditReq) (rs *pb.NoReply, err error) {
	rs, err = service.LikeSvc.GuessEdit(c, r)
	return
}

// GuessGroupDel  del guess group.
func (s *activityService) GuessGroupDel(c context.Context, r *pb.GuessGroupDelReq) (rs *pb.GuessGroupReply, err error) {
	rs, err = service.LikeSvc.GuessGroupDel(c, r)
	return
}

// GuessUpResult  update guess result.
func (s *activityService) GuessUpResult(c context.Context, r *pb.GuessUpResultReq) (rs *pb.NoReply, err error) {
	rs, err = service.LikeSvc.GuessUpResult(c, r)
	return
}

// GuessAllList  match guess all list.
func (s *activityService) GuessAllList(c context.Context, r *pb.GuessListReq) (rs *pb.GuessListAllReply, err error) {
	rs, err = service.LikeSvc.GuessAllList(c, r)
	return
}

// GuessList  match guess list.
func (s *activityService) GuessList(c context.Context, r *pb.GuessListReq) (rs *pb.GuessListReply, err error) {
	rs, err = service.LikeSvc.GuessList(c, r)
	return
}

// GuessLists  match guess lists.
func (s *activityService) GuessLists(c context.Context, r *pb.GuessListsReq) (rs *pb.GuessListsReply, err error) {
	var tmp map[int64]*pb.GuessListReply
	rs = &pb.GuessListsReply{}
	tmp = make(map[int64]*pb.GuessListReply, len(r.Oids))
	if tmp, err = service.LikeSvc.GuessLists(c, r); err != nil {
		return
	}
	rs = &pb.GuessListsReply{MatchGuesses: tmp}
	return
}

// GuessUserAdd user add guess.
func (s *activityService) GuessUserAdd(c context.Context, r *pb.GuessUserAddReq) (rs *pb.NoReply, err error) {
	rs, err = service.LikeSvc.UserAddGuess(c, r)
	return
}

// GuessUserList user add guess.
func (s *activityService) UserGuessList(c context.Context, r *pb.UserGuessListReq) (rs *pb.UserGuessListReply, err error) {
	var (
		tmp   []*pb.GuessUserGroup
		count int64
	)
	rs = &pb.UserGuessListReply{}
	if tmp, count, err = service.LikeSvc.UserGuessList(c, r); err != nil {
		return
	}
	rs = &pb.UserGuessListReply{UserGroup: tmp, Page: &pb.PageInfo{Pn: r.Pn, Ps: r.Ps, Total: count}}
	return
}

// UserGuessGroup user guess group.
func (s *activityService) UserGuessGroup(c context.Context, r *pb.UserGuessGroupReq) (rs *pb.GuessUserGroup, err error) {
	rs, err = service.LikeSvc.UserGuessGroup(c, r)
	return
}

// GuessUserData user guess data.
func (s *activityService) UserGuessData(c context.Context, r *pb.UserGuessDataReq) (rs *pb.UserGuessDataReply, err error) {
	rs, err = service.LikeSvc.UserLogData(c, r)
	return
}

// UserGuessMatch user guess match list.
func (s *activityService) UserGuessMatch(c context.Context, r *pb.UserGuessMatchReq) (rs *pb.UserGuessMatchReply, err error) {
	rs, err = service.LikeSvc.UserGuessMatch(c, r)
	return
}

// UserGuessResult user guess result.
func (s *activityService) UserGuessResult(c context.Context, r *pb.UserGuessResultReq) (rs *pb.UserGuessResultReply, err error) {
	rs, err = service.LikeSvc.UserGuessResult(c, r)
	return
}

// UserGuessMatchs user guess more match list.
func (s *activityService) UserGuessMatchs(c context.Context, r *pb.UserGuessMatchsReq) (rs *pb.UserGuessMatchsReply, err error) {
	var (
		userGroup []*pb.GuessUserGroup
		count     int64
	)
	rs = &pb.UserGuessMatchsReply{}
	if userGroup, count, err = service.LikeSvc.UserGuessMatchs(c, r); err != nil {
		return
	}
	rs.UserGroup = userGroup
	rs.Page = &pb.PageInfo{Total: count, Pn: r.Pn, Ps: r.Ps}
	return
}

func (s *activityService) AddReserve(c context.Context, r *pb.AddReserveReq) (rs *pb.NoReply, err error) {
	report := &like.ReserveReport{
		From:     r.From,
		Typ:      r.Typ,
		Oid:      r.Oid,
		Ip:       r.Ip,
		Platform: r.Platform,
		Mobiapp:  r.Mobiapp,
		Buvid:    r.Buvid,
		Spmid:    r.Spmid,
	}
	if err = service.LikeSvc.AsyncReserve(c, r.Sid, r.Mid, 1, report); err != nil {
		return
	}
	rs = &pb.NoReply{}
	return
}

func (s *activityService) DelReserve(c context.Context, r *pb.DelReserveReq) (rs *pb.NoReply, err error) {
	if err = service.LikeSvc.ReserveCancel(c, r.Sid, r.Mid); err != nil {
		return
	}
	rs = &pb.NoReply{}
	return
}

func (s *activityService) ReserveFollowing(c context.Context, r *pb.ReserveFollowingReq) (rs *pb.ReserveFollowingReply, err error) {
	var (
		rly *lmdl.ActFollowingReply
	)
	if rly, err = service.LikeSvc.ReserveFollowing(c, r.Sid, r.Mid); err != nil {
		return
	}
	rs = &pb.ReserveFollowingReply{}
	if rly != nil {
		rs.IsFollow = rly.IsFollowing
		rs.Total = rly.Total
		rs.Mtime = rly.Mtime
		rs.Order = rly.Order
	}
	return
}

func (s *activityService) ReserveFollowings(c context.Context, r *pb.ReserveFollowingsReq) (rs *pb.ReserveFollowingsReply, err error) {
	var (
		rly map[int64]*lmdl.ActFollowingReply
	)
	if rly, err = service.LikeSvc.ReserveFollowings(c, r.Sids, r.Mid); err != nil {
		return
	}
	rs = &pb.ReserveFollowingsReply{}
	rs.List = make(map[int64]*pb.ReserveFollowingReply)
	for k, v := range rly {
		rs.List[k] = &pb.ReserveFollowingReply{IsFollow: v.IsFollowing, Total: v.Total, Order: v.Order}
	}
	return
}

func (s *activityService) UpActDoTask(c context.Context, r *pb.UpActDoTaskReq) (rs *pb.UpActDoTaskReply, err error) {
	rs = &pb.UpActDoTaskReply{}
	var days int64
	if days, err = service.LikeSvc.UpActDo(c, r.Sid, r.Mid, r.Totaltime, r.Matchedpercent); err != nil {
		return
	}
	rs.Days = days
	return
}

func (s *activityService) UpActInfo(c context.Context, r *pb.UpActInfoReq) (rs *pb.UpActInfoReply, err error) {
	rs = &pb.UpActInfoReply{}
	act := &pb.UpActInfo{}
	if act, err = service.LikeSvc.UpActInfo(c, r.Aid); err != nil {
		return
	}
	rs.UpActInfo = act
	return
}

func (s *activityService) NativePages(c context.Context, r *pb.NativePagesReq) (*pb.NativePagesReply, error) {
	return &pb.NativePagesReply{}, nil
}

func (s *activityService) NativeLoadPages(c context.Context, r *pb.NativePagesReq) (*pb.NativePagesReply, error) {
	return &pb.NativePagesReply{}, nil
}

func (s *activityService) NativePagesExt(c context.Context, r *pb.NativePagesExtReq) (*pb.NativePagesExtReply, error) {
	return &pb.NativePagesExtReply{}, nil
}

func (s *activityService) NativeValidPagesExt(c context.Context, r *pb.NativeValidPagesExtReq) (*pb.NativeValidPagesExtReply, error) {
	return &pb.NativeValidPagesExtReply{}, nil
}

func (s *activityService) NativePage(c context.Context, r *pb.NativePageReq) (*pb.NativePageReply, error) {
	return &pb.NativePageReply{}, nil
}

func (s *activityService) ClockInTag(c context.Context, r *pb.ClockInTagReq) (*pb.ClockInTagReply, error) {
	tags, err := service.LikeSvc.ClockInTag(c, r.Mid)
	if err != nil {
		return nil, err
	}
	return &pb.ClockInTagReply{Tags: tags}, nil
}

func (s *activityService) ActLikeCount(c context.Context, r *pb.ActLikeCountReq) (*pb.ActLikeCountReply, error) {
	total, err := service.LikeSvc.ActLikeCount(c, r.Sid)
	if err != nil {
		return nil, err
	}
	return &pb.ActLikeCountReply{Total: total}, nil
}

func (s *activityService) AwardSubjectState(c context.Context, r *pb.AwardSubjectStateReq) (rs *pb.AwardSubjectStateReply, err error) {
	rs = &pb.AwardSubjectStateReply{}
	state, err := service.LikeSvc.AwardSubjectStateByID(c, r.Id, r.Mid)
	if err != nil {
		return
	}
	rs.State = int32(state)
	return
}

func (s *activityService) RewardSubject(c context.Context, r *pb.RewardSubjectReq) (rs *pb.NoReply, err error) {
	rs = &pb.NoReply{}
	if err = service.LikeSvc.AwardSubjectRewardByID(c, r.Id, r.Mid); err != nil {
		return
	}
	return
}

// NatTabModules .
func (s *activityService) NatTabModules(c context.Context, r *pb.NatTabModulesReq) (*pb.NatTabModulesReply, error) {
	return &pb.NatTabModulesReply{}, nil
}

func (s *activityService) NativePagesTab(c context.Context, r *pb.NativePagesTabReq) (*pb.NativePagesTabReply, error) {
	return &pb.NativePagesTabReply{}, nil
}

func (s *activityService) FissionLotteryDo(c context.Context, r *pb.FissionLotteryDoReq) (*pb.FissionLotteryDoReply, error) {
	rly, err := service.LikeSvc.FissionDoLottery(c, r.Sid, r.Mid)
	if err != nil {
		return nil, err
	}
	if rly == nil || len(rly) == 0 || rly[0] == nil {
		return &pb.FissionLotteryDoReply{}, nil
	}
	return &pb.FissionLotteryDoReply{Record: &pb.LotteryRecordDetail{
		ID:       rly[0].ID,
		Mid:      rly[0].Mid,
		Num:      int64(rly[0].Num),
		GiftId:   rly[0].GiftID,
		GiftName: rly[0].GiftName,
		GiftType: int64(rly[0].GiftType),
		ImgUrl:   rly[0].ImgURL,
		Type:     int64(rly[0].Type),
		Ctime:    rly[0].Ctime,
		Cid:      rly[0].CID,
	}}, nil
}

func (s *activityService) FissionLotteryUpNum(c context.Context, r *pb.FissionLotteryUpNumReq) (*pb.FissionLotteryUpNumReply, error) {
	rly, err := service.LikeSvc.FissionUpLotteryNum(c, r.Sid, r.Num)
	if err != nil {
		return nil, err
	}
	return &pb.FissionLotteryUpNumReply{Affected: rly}, nil
}

func (s *activityService) LotteryUserRecord(c context.Context, r *pb.LotteryUserRecordReq) (*pb.LotteryUserRecordReply, error) {
	rly, err := service.LikeSvc.GetMyList(c, r.Sid, int(r.Pn), int(r.Ps), r.Mid, false)
	if err != nil {
		return nil, err
	}
	if rly == nil || rly.Page == nil {
		return &pb.LotteryUserRecordReply{}, nil
	}
	reply := &pb.LotteryUserRecordReply{Total: int64(rly.Page.Total)}
	for _, v := range rly.List {
		if v == nil {
			continue
		}
		reply.List = append(reply.List, &pb.LotteryRecord{
			ID:       v.ID,
			Mid:      v.Mid,
			Num:      int64(v.Num),
			GiftId:   v.GiftID,
			GiftName: v.GiftName,
			GiftType: int64(v.GiftType),
			ImgUrl:   v.ImgURL,
			Type:     int64(v.Type),
			Ctime:    v.Ctime,
			Cid:      v.CID,
		})
	}
	return reply, nil
}

func (s *activityService) ArcSubTypeCount(c context.Context, r *pb.ArcSubTypeCountReq) (*pb.ActSubTypeCountReply, error) {
	rly, err := service.LikeSvc.LikeArcTypeCount(c, r.Sid)
	if err != nil {
		return nil, err
	}
	return &pb.ActSubTypeCountReply{Counts: rly}, nil
}

func (s *activityService) SendBwsOnlinePiece(ctx context.Context, r *pb.SendBwsOnlinePieceReq) (*pb.NoReply, error) {
	if err := service.BwsOnlineSvc.SendSpecialPiece(ctx, r.Mid, r.Id, r.Token, con.Conf.BwsOnline.DefaultBid); err != nil {
		return nil, err
	}
	return &pb.NoReply{}, nil
}

func (s *activityService) WxLotteryAward(c context.Context, r *pb.WxLotteryAwardReq) (*pb.WxLotteryAwardReply, error) {
	show, jumpURL, err := service.LikeSvc.WxLotteryAwardRedDot(c, r.Mid)
	if err != nil {
		return nil, err
	}
	return &pb.WxLotteryAwardReply{Show: show, URL: jumpURL}, nil
}

func (s *activityService) SyncSubjectRules(c context.Context, r *pb.SyncSubjectRulesReq) (*pb.NoReply, error) {
	err := service.LikeSvc.SyncSubjectRules(c, r.SID, r.Counter)
	return &pb.NoReply{}, err
}

func (s *activityService) SyncUserState(c context.Context, r *pb.SyncUserStateReq) (*pb.NoReply, error) {
	return &pb.NoReply{}, service.LikeSvc.SyncUserState(c, r)
}

func (s *activityService) SyncUserScore(c context.Context, r *pb.SyncUserScoreReq) (*pb.NoReply, error) {
	return &pb.NoReply{}, service.LikeSvc.SyncUserScore(c, r)
}

func (s *activityService) LotteryUnusedTimes(c context.Context, r *pb.LotteryUnusedTimesdReq) (*pb.LotteryUnusedTimesReply, error) {
	isBnj, num := bnj2021.ARDrawQuota(c, r.Mid, r.Sid)
	if isBnj {
		return &pb.LotteryUnusedTimesReply{Times: int64(num)}, nil
	}
	newlottery, err := service.LotterySvc.InitLottery(c, r.Sid)
	if err != nil {
		return nil, err
	}
	if !newlottery {
		rly, err := service.LikeSvc.GetUnusedTimes(c, r.Sid, r.Mid)
		if err != nil {
			return nil, err
		}
		return &pb.LotteryUnusedTimesReply{Times: int64(rly.Times)}, nil
	}
	rly, err := service.LotterySvc.GetUnusedTimes(c, r.Sid, r.Mid)
	if err != nil {
		return nil, err
	}
	return &pb.LotteryUnusedTimesReply{Times: int64(rly.Times)}, nil

}

func (s *activityService) LotteryAddTimes(c context.Context, r *pb.LotteryAddTimesReq) (*pb.LotteryAddTimesReply, error) {
	newlottery, err := service.LotterySvc.InitLottery(c, r.Sid)
	if err != nil {
		return nil, err
	}
	if !newlottery {
		err := service.LikeSvc.AddLotteryTimes(c, r.Sid, r.Mid, r.Cid, int(r.ActionType), 0, r.OrderNo, false)
		if err != nil {
			return nil, err
		}
	}
	err = service.LotterySvc.AddLotteryTimes(c, r.Sid, r.Mid, r.Cid, int(r.ActionType), 0, r.OrderNo, false)
	if err != nil {
		return nil, err
	}
	return &pb.LotteryAddTimesReply{}, nil
}

func (s *activityService) LotteryWinList(c context.Context, r *pb.LotteryWinListReq) (*pb.LotteryWinListReply, error) {
	newlottery, err := service.LotterySvc.InitLottery(c, r.Sid)
	res := &pb.LotteryWinListReply{}
	res.List = make([]*pb.LotteryWinList, 0)
	if err != nil {
		return nil, err
	}
	if !newlottery {
		list, err := service.LikeSvc.WinList(c, r.Sid, r.Num, r.NeedCache)
		if err != nil {
			return nil, err
		}
		if list != nil {
			for _, v := range list {
				res.List = append(res.List, &pb.LotteryWinList{
					Name:     v.Name,
					GiftID:   v.GiftID,
					GiftName: v.GiftName,
					Mid:      v.Mid,
					Ctime:    int64(v.Ctime),
				})
			}
		}
	}
	list, err := service.LotterySvc.WinList(c, r.Sid, r.Num, r.NeedCache)
	if err != nil {
		return nil, err
	}
	if list != nil {
		for _, v := range list {
			res.List = append(res.List, &pb.LotteryWinList{
				Name:     v.Name,
				GiftID:   v.GiftID,
				GiftName: v.GiftName,
				Mid:      v.Mid,
				Ctime:    int64(v.Ctime),
			})
		}
	}

	return res, nil
}

// DoLottery user do lottery
func (s *activityService) DoLottery(c context.Context, r *pb.DoLotteryReq) (*pb.DoLotteryReply, error) {
	if r.Nums == 0 {
		r.Nums = 1
	}
	// 检查orderno
	if r.OrderNo != "" {
		if find := strings.Contains(r.OrderNo, "@"); find {
			return nil, ecode.ActivityLotteryOrderNoErr
		}
	}

	data := make([]*pb.LotteryRecordDetail, 0)
	newlottery, err := service.LotterySvc.InitLottery(c, r.Sid)
	if err != nil {
		return nil, err
	}
	if !newlottery {
		lotteryRecordList, err := service.LikeSvc.DoLottery(c, r.Sid, r.Mid, int(r.Nums), _isInternal)
		if err != nil {
			return nil, err
		}
		if len(lotteryRecordList) > 0 {
			for _, v := range lotteryRecordList {
				if v == nil {
					continue
				}
				lottery := &pb.LotteryRecordDetail{
					ID:       v.ID,
					Mid:      v.Mid,
					Num:      int64(v.Num),
					GiftId:   v.GiftID,
					GiftName: v.GiftName,
					GiftType: int64(v.GiftType),
					ImgUrl:   v.ImgURL,
					Type:     int64(v.Type),
					Ctime:    v.Ctime,
					Cid:      v.CID,
				}
				data = append(data, lottery)
			}
		}
	} else {
		var risk *riskmdl.Base
		if r.Risk != nil {
			risk = &riskmdl.Base{
				Buvid:     r.Risk.Buvid,
				Origin:    r.Risk.Origin,
				Referer:   r.Risk.Referer,
				IP:        r.Risk.Ip,
				Ctime:     time.Now().Format("2006-01-02 15:04:05"),
				UserAgent: r.Risk.UserAgent,
				Build:     r.Risk.Build,
				Platform:  r.Risk.Platform,
				Action:    riskmdl.ActionLottery,
				MID:       r.Mid,
				API:       r.Risk.Api,
				EsTime:    time.Now().Unix(),
			}
		}
		lotteryRecordList, err := service.LotterySvc.DoLottery(c, r.Sid, r.Mid, risk, int(r.Nums), false, r.OrderNo)
		if err != nil {
			return nil, err
		}
		if len(lotteryRecordList) > 0 {
			for _, v := range lotteryRecordList {
				if v == nil {
					continue
				}
				lottery := &pb.LotteryRecordDetail{
					ID:       v.ID,
					Mid:      v.Mid,
					Num:      int64(v.Num),
					GiftId:   v.GiftID,
					GiftName: v.GiftName,
					GiftType: int64(v.GiftType),
					ImgUrl:   v.ImgURL,
					Type:     int64(v.Type),
					Ctime:    v.Ctime,
					Cid:      v.CID,
					Extra:    v.Extra,
				}
				data = append(data, lottery)
			}
		}
	}
	return &pb.DoLotteryReply{Data: data}, nil
}

// DoLottery user do lottery
func (s *activityService) LotteryGift(c context.Context, r *pb.LotteryGiftReq) (*pb.LotteryGiftReply, error) {
	gift := make([]*pb.LotteryGift, 0)
	lotteryGift, err := service.LotterySvc.Gift(c, r.Sid)
	if err != nil {
		return nil, err
	}
	if lotteryGift != nil {
		for _, v := range lotteryGift {
			gift = append(gift, &pb.LotteryGift{
				ID:        v.ID,
				Name:      v.Name,
				Type:      int64(v.Type),
				ImgUrl:    v.ImgURL,
				SendNum:   v.SendNum,
				Num:       v.Num,
				Efficient: int64(v.Efficient),
			})
		}
	}
	return &pb.LotteryGiftReply{Gift: gift}, nil
}

// up主发起活动白名单接口 .
func (s *activityService) IsUpActUid(c context.Context, r *pb.IsUpActUidReq) (*pb.IsUpActUidReply, error) {
	return &pb.IsUpActUidReply{}, nil
}

// up主发起有效活动列表接口 .
func (s *activityService) UpActNativePages(c context.Context, r *pb.UpActNativePagesReq) (*pb.UpActNativePagesReply, error) {
	return &pb.UpActNativePagesReply{}, nil
}

// up主发起活动-进审核态
func (s *activityService) UpActNativePageBind(c context.Context, r *pb.UpActNativePageBindReq) (*pb.UpActNativePageBindReply, error) {
	return nil, xecode.RequestErr
}

// 根据活动id查询用户稿件情况
func (s *activityService) ListActivityArcs(ctx context.Context, req *pb.ListActivityArcsReq) (*pb.ListActivityArcsReply, error) {
	return service.LikeSvc.ListActivityArcs(ctx, req)
}

// CollegeAidIsActivity 开学季活动是否活动稿件
func (s *activityService) CollegeAidIsActivity(c context.Context, r *pb.CollegeAidIsActivityReq) (*pb.CollegeAidIsActivityRes, error) {
	res, err := service.CollegeSvc.AidIsCollege(c, r.Mid, r.Aid)
	return &pb.CollegeAidIsActivityRes{IsActivity: res}, err
}

// GetReserveProgress 获取预约数据
func (s *activityService) GetReserveProgress(ctx context.Context, req *pb.GetReserveProgressReq) (*pb.GetReserveProgressRes, error) {
	return service.LikeSvc.GetReserveProgress(ctx, req)
}

func (s *activityService) SponsorNativePages(ctx context.Context, req *pb.SponsorNativePagesReq) (*pb.SponsorNativePagesReply, error) {
	return &pb.SponsorNativePagesReply{}, nil
}

// GetReserveProgress 获取预约数据
func (s *activityService) GetNatProgressParams(ctx context.Context, req *pb.GetNatProgressParamsReq) (*pb.GetNatProgressParamsReply, error) {
	return &pb.GetNatProgressParamsReply{}, nil
}

func (s *activityService) ActRelationInfo(ctx context.Context, req *pb.ActRelationInfoReq) (*pb.ActRelationInfoReply, error) {
	return service.LikeSvc.ActRelationInfo(ctx, req)
}

func (s *activityService) ActRelationReserve(ctx context.Context, req *pb.ActRelationReserveReq) (*pb.ActRelationReserveReply, error) {
	return service.LikeSvc.ActRelationReserve(ctx, req)
}

func (s *activityService) ActRelationReserveInfo(ctx context.Context, req *pb.ActRelationReserveInfoReq) (*pb.ActRelationReserveInfoReply, error) {
	return service.LikeSvc.ActRelationReserveInfo(ctx, req)
}

func (s *activityService) GRPCDoRelation(ctx context.Context, req *pb.GRPCDoRelationReq) (*pb.NoReply, error) {
	return service.LikeSvc.GRPCDoRelation(ctx, req)
}

func (s *activityService) BwsGamePlayable(ctx context.Context, req *pb.BwsGamePlayableReq) (*pb.NoReply, error) {
	_, _, _, err := service.BwsSvc.UserPlayable(ctx, req.Mid, req.Bid, req.GameId)
	res := &pb.NoReply{}
	return res, err
}

func (s *activityService) BwsGamePlay(ctx context.Context, req *pb.BwsGamePlayReq) (*pb.NoReply, error) {
	err := service.BwsSvc.UserPlayGame(ctx, req.Mid, req.Bid, req.GameId, req.Star, req.Pass)
	res := &pb.NoReply{}
	return res, err
}

func getMidDate(ctx context.Context, mid int64, day string) (int64, string) {
	if service.BwsSvc.IsTest(ctx) {
		if service.BwsSvc.IsVip(ctx) {
			mid, day = service.BwsSvc.GetVipMidDate(ctx)
		} else {
			mid, day = service.BwsSvc.GetNormalMidDate(ctx)
		}
	}
	return mid, day
}

func (s *activityService) Bws2020Member(ctx context.Context, req *pb.Bws2020MemberReq) (*pb.Bws2020MemberReply, error) {
	if req.BwsDate == "" {
		req.BwsDate = time.Now().Format("20060102")
	}
	var isVip bool
	if service.BwsSvc.IsWhiteMid(ctx, req.Mid) {
		_, err := service.BwsSvc.GetUserToken(ctx, req.Bid, req.Mid)
		if err != nil {
			return nil, err
		}
		isVip = true
	} else {
		mid, d := getMidDate(ctx, req.Mid, req.BwsDate)
		today, _ := strconv.ParseInt(d, 10, 64)
		ticketRes, err := service.BwsOnlineSvc.HasVipTickets(ctx, mid, today)
		if err != nil {
			return nil, err
		}

		if len(ticketRes) > 0 {
			isVip = true
		}
	}

	userDetail, err := service.BwsSvc.UserDetail(ctx, req.Mid, req.Bid, req.BwsDate, isVip)
	if err != nil {
		return nil, err
	}
	res := &pb.Bws2020MemberReply{
		Mid:           userDetail.Mid,
		Bid:           userDetail.Bid,
		Heart:         userDetail.Heart,
		Star:          userDetail.Star,
		BwsDate:       userDetail.BwsDate,
		StarLastTime:  userDetail.StarLastTime,
		StarGame:      userDetail.StarGame,
		Rank:          userDetail.Rank,
		LotteryRemain: userDetail.LotteryRemain,
	}
	return res, err
}

func (s *activityService) RelationReserveCancel(ctx context.Context, req *pb.RelationReserveCancelReq) (*pb.NoReply, error) {
	return service.LikeSvc.RelationReserveCancel(ctx, req)
}

func (s *activityService) InternalSyncActRelationInfoDB2Cache(ctx context.Context, req *pb.InternalSyncActRelationInfoDB2CacheReq) (*pb.InternalSyncActRelationInfoDB2CacheReply, error) {
	return service.LikeSvc.InternalSyncActRelationInfoDB2Cache(ctx, req)
}

func (s *activityService) InternalUpdateItemDataWithCache(ctx context.Context, req *pb.InternalUpdateItemDataWithCacheReq) (*pb.InternalUpdateItemDataWithCacheReply, error) {
	return service.LikeSvc.InternalUpdateItemDataWithCache(ctx, req)
}

func (s *activityService) InternalSyncActSubjectInfoDB2Cache(ctx context.Context, req *pb.InternalSyncActSubjectInfoDB2CacheReq) (*pb.InternalSyncActSubjectInfoDB2CacheReply, error) {
	return service.LikeSvc.InternalSyncActSubjectInfoDB2Cache(ctx, req)
}

func (s *activityService) InternalSyncActSubjectReserveIDsInfoDB2Cache(ctx context.Context, req *pb.InternalSyncActSubjectReserveIDsInfoDB2CacheReq) (*pb.InternalSyncActSubjectReserveIDsInfoDB2CacheReply, error) {
	return service.LikeSvc.InternalSyncActSubjectReserveIDsInfoDB2Cache(ctx, req)
}

func (s *activityService) UpList(ctx context.Context, req *pb.UpListReq) (*pb.UpListReply, error) {
	params := &lmdl.ParamList{
		Sid:  req.Sid,
		Type: req.Type,
		Pn:   int(req.Pn),
		Ps:   int(req.Ps),
	}
	res := &pb.UpListReply{}
	res.Page = &pb.UpListPage{}
	res.List = make([]*pb.UpListItem, 0)

	data, err := service.LikeSvc.UpList(ctx, params, req.Mid)
	if err != nil {
		return res, err
	}
	if data == nil {
		return res, nil
	}
	if data.Page != nil {
		res.Page = &pb.UpListPage{
			Num:   int64(data.Page.Num),
			Ps:    int64(data.Page.Size),
			Total: data.Page.Total,
		}
	}
	if data.List != nil {
		for _, v := range data.List {
			object, ok := v.Object.(map[string]interface{})
			if !ok {
				continue
			}
			account, ok := object["act"]
			if !ok {
				continue
			}
			content, ok := object["cont"]
			if !ok {
				continue
			}
			likeContent, ok := content.(*like.LikeContent)
			if !ok {
				continue
			}
			acc, ok := account.(struct {
				*api.Info
				Following    bool  `json:"following"`
				FollowerNum  int64 `json:"follower_num"`
				FollowingNum int64 `json:"following_num"`
			})
			if !ok {
				continue
			}
			if acc.Info == nil || likeContent == nil {
				continue
			}
			accPb := &pb.AccountInfo{
				Mid:       acc.Mid,
				Name:      acc.Name,
				Sex:       acc.Sex,
				Face:      acc.Face,
				Sign:      acc.Sign,
				Rank:      acc.Rank,
				Following: acc.Following,
			}
			contentPb := &pb.LikeContent{
				ID:      likeContent.ID,
				Message: likeContent.Message,
				IP:      likeContent.IP,
				Plat:    likeContent.Plat,
				Device:  likeContent.Device,
				Image:   likeContent.Image,
				Reply:   likeContent.Reply,
				Link:    likeContent.Link,
				ExName:  likeContent.ExName,
				IPv6:    likeContent.IPv6,
			}
			item := &pb.Item{}
			item.DeepCopyFromItem(v.Item)
			res.List = append(res.List, &pb.UpListItem{
				Item:    item,
				Account: accPb,
				Content: contentPb,
			})
		}
	}
	return res, nil
}

func (s *activityService) ActReserveTag(ctx context.Context, req *pb.ActReserveTagReq) (reply *pb.ActReserveTagReply, err error) {
	return service.LikeSvc.GetActReserveTag(ctx, req)
}

func (s *activityService) UpActReserveRelationInfo(ctx context.Context, req *pb.UpActReserveRelationInfoReq) (*pb.UpActReserveRelationInfoReply, error) {
	return service.LikeSvc.UpActReserveRelationInfo(ctx, req)
}

func (s *activityService) CreateUpActReserveRelation(ctx context.Context, req *pb.CreateUpActReserveRelationReq) (*pb.CreateUpActReserveRelationReply, error) {
	return service.LikeSvc.CreateUpActReserveRelation(ctx, req)
}

func (s *activityService) CancelUpActReserve(ctx context.Context, req *pb.CancelUpActReserveReq) (*pb.CancelUpActReserveReply, error) {
	return service.LikeSvc.CancelUpActReserve(ctx, req)
}

func (s *activityService) UpActReserveInfo(ctx context.Context, req *pb.UpActReserveInfoReq) (*pb.UpActReserveInfoReply, error) {
	return service.LikeSvc.UpActReserveInfo(ctx, req)
}

func (s *activityService) CanUpCreateActReserve(ctx context.Context, req *pb.CanUpCreateActReserveReq) (*pb.CanUpCreateActReserveReply, error) {
	return service.LikeSvc.CanUpCreateActReserve(ctx, req)
}

func (s *activityService) SpringFestival2021InviteToken(ctx context.Context, req *pb.SpringFestival2021InviteTokenReq) (reply *pb.SpringFestival2021InviteTokenReply, err error) {
	reply = &pb.SpringFestival2021InviteTokenReply{}
	tokenRes, err := service.SpringFestival2021Svc.InviteShare(ctx, req.Mid)
	if err != nil {
		return nil, err
	}
	reply.Token = tokenRes.Token
	return reply, nil
}

func (s *activityService) SpringFestival2021SendCardToken(ctx context.Context, req *pb.SpringFestival2021SendCardTokenReq) (reply *pb.SpringFestival2021SendCardTokenReply, err error) {
	reply = &pb.SpringFestival2021SendCardTokenReply{}
	tokenRes, err := service.SpringFestival2021Svc.CardShare(ctx, req.Mid, req.CardID)
	if err != nil {
		return nil, err
	}
	reply.Token = tokenRes.Token
	return reply, nil
}

func (s *activityService) SpringFestival2021MidCard(ctx context.Context, req *pb.SpringFestival2021MidCardReq) (reply *pb.SpringFestival2021MidCardReply, err error) {
	reply = &pb.SpringFestival2021MidCardReply{}
	res, err := service.SpringFestival2021Svc.Cards(ctx, req.Mid)
	if err != nil {
		return nil, err
	}
	if res.Cards != nil {
		reply.CardID1 = res.Cards.Card1
		reply.CardID2 = res.Cards.Card2
		reply.CardID3 = res.Cards.Card3
		reply.CardID4 = res.Cards.Card4
		reply.CardID5 = res.Cards.Card5
		reply.Compose = res.Cards.Compose
	}
	return reply, nil
}

func (s *activityService) InviteToken(ctx context.Context, req *pb.InviteTokenReq) (reply *pb.InviteTokenReply, err error) {
	reply = &pb.InviteTokenReply{}
	tokenRes, err := service.CardSvc.InviteShare(ctx, req.Mid, req.Activity)
	if err != nil {
		return nil, err
	}
	reply.Token = tokenRes.Token
	return reply, nil
}

func (s *activityService) SendCardToken(ctx context.Context, req *pb.SendCardTokenReq) (reply *pb.SendCardTokenReply, err error) {
	reply = &pb.SendCardTokenReply{}
	tokenRes, err := service.CardSvc.CardShare(ctx, req.Mid, req.CardID, req.Activity)
	if err != nil {
		return nil, err
	}
	reply.Token = tokenRes.Token
	return reply, nil
}

func (s *activityService) Cards2021MidCard(ctx context.Context, req *pb.CardsMidCardReq) (reply *pb.CardsMidCardReply, err error) {
	reply = &pb.CardsMidCardReply{}
	res, err := service.CardSvc.Cards(ctx, req.Mid, req.Activity)
	if err != nil {
		return nil, err
	}
	if res.Cards != nil {
		reply.CardID1 = res.Cards.Card1
		reply.CardID2 = res.Cards.Card2
		reply.CardID3 = res.Cards.Card3
		reply.CardID4 = res.Cards.Card4
		reply.CardID5 = res.Cards.Card5
		reply.CardID6 = res.Cards.Card6
		reply.CardID7 = res.Cards.Card7
		reply.CardID8 = res.Cards.Card8
		reply.CardID9 = res.Cards.Card9
		reply.Compose = res.Cards.Compose
	}
	return reply, nil
}
func (s *activityService) UpActReserveCanBindList(ctx context.Context, req *pb.UpActReserveCanBindListReq) (*pb.UpActReserveCanBindListReply, error) {
	return service.LikeSvc.UpActReserveCanBindList(ctx, req)
}

func (s *activityService) UpActReserveBindList(ctx context.Context, req *pb.UpActReserveBindListReq) (*pb.UpActReserveBindListReply, error) {
	return service.LikeSvc.UpActReserveBindList(ctx, req)
}

func (s *activityService) BindActReserve(ctx context.Context, req *pb.BindActReserveReq) (*pb.BindActReserveReply, error) {
	return service.LikeSvc.BindActReserve(ctx, req)
}

func (s *activityService) CreateUpActReserve(ctx context.Context, req *pb.CreateUpActReserveReq) (*pb.CreateUpActReserveReply, error) {
	args := &like.CreateUpActReserveArgs{
		Title:             req.Title,
		Type:              int64(req.Type),
		From:              int64(req.From),
		LivePlanStartTime: int64(req.LivePlanStartTime),
		Oid:               req.Oid,
		CreateDynamic:     req.CreateDynamic,
		LotteryID:         req.LotteryID,
		LotteryType:       int64(req.LotteryType),
	}
	reply := new(pb.CreateUpActReserveReply)
	sid, err := service.LikeSvc.CreateUpActReserve(ctx, req.Mid, args)
	if err != nil {
		return nil, err
	}
	reply.Sid = sid
	return reply, nil
}

func (s *activityService) GetActReserveTotal(ctx context.Context, req *pb.GetActReserveTotalReq) (*pb.GetActReserveTotalReply, error) {
	return service.LikeSvc.GetActReserveTotal(ctx, req.Sid)
}

func (s *activityService) UpActUserSpaceCard(ctx context.Context, req *pb.UpActUserSpaceCardReq) (*pb.UpActUserSpaceCardReply, error) {
	return service.LikeSvc.GetUpActUserSpaceCard(ctx, req)
}

func (s *activityService) ActivityProgress(ctx context.Context, req *pb.ActivityProgressReq) (*pb.ActivityProgressReply, error) {
	if req.Type == 1 && req.Sid == 0 {
		return nil, xecode.Error(xecode.RequestErr, "sid不能为空")
	}
	if req.Type == 2 && len(req.Gids) == 0 {
		return nil, xecode.Error(xecode.RequestErr, "gid不能为空")
	}
	return service.LikeSvc.ActivityProgress(ctx, req)
}

func (s *activityService) UpActReserveVerification4Cancel(ctx context.Context, req *pb.UpActReserveVerification4CancelReq) (*pb.UpActReserveVerification4CancelReply, error) {
	return service.LikeSvc.UpActReserveVerification4Cancel(ctx, req)
}

func (s *activityService) UpActReserveRelationInfoByTime(ctx context.Context, req *pb.UpActReserveRelationInfoByTimeReq) (*pb.UpActReserveRelationInfoByTimeReply, error) {
	return service.LikeSvc.UpActReserveRelationInfoByTime(ctx, req)
}

func (s *activityService) UpActReserveRelationDBInfoByCondition(ctx context.Context, req *pb.UpActReserveRelationDBInfoByConditionReq) (*pb.UpActReserveRelationDBInfoByConditionReply, error) {
	return service.LikeSvc.UpActReserveRelationDBInfoByCondition(ctx, req)
}

func (s *activityService) UpActReserveLiveStateExpire(ctx context.Context, req *pb.UpActReserveLiveStateExpireReq) (*pb.UpActReserveLiveStateExpireReply, error) {
	return service.LikeSvc.UpActReserveLiveStateExpire(ctx, req)
}

func (s *activityService) UpActReserveRelationInfo4Live(ctx context.Context, req *pb.UpActReserveRelationInfo4LiveReq) (*pb.UpActReserveRelationInfo4LiveReply, error) {
	return service.LikeSvc.UpActReserveRelationInfo4Live(ctx, req)
}

func (s *activityService) GetSidAndDynamicIDByOid(ctx context.Context, req *pb.GetSidAndDynamicIDByOidReq) (*pb.GetSidAndDynamicIDByOidReply, error) {
	return service.LikeSvc.GetSidAndDynamicIDByOid(ctx, req)
}

func (s *activityService) RankResult(ctx context.Context, req *pb.RankResultReq) (res *pb.RankResultResp, err error) {
	res = new(pb.RankResultResp)
	var ret *rankmdl.ResultList
	ret, err = service.Rankv3Svc.GetRankByID(ctx, req.RankID, int(req.Pn), int(req.Ps))
	if ret != nil {
		res.StatisticsType = int64(ret.StatisticsType)
		res.BatchTime = ret.ShowBatchTime
		if ret.Page != nil {
			res.Page = &pb.PageInfo{
				Pn:    req.Pn,
				Ps:    req.Ps,
				Total: int64(ret.Page.Total),
			}
		}
		if len(ret.List) > 0 {
			for _, v := range ret.List {
				var account = &pb.Account{}
				var tag = &pb.Tag{}
				var archive = make([]*pb.ArchiveInfo, 0)
				if v.Account != nil {
					account.MID = v.Account.Mid
					account.Name = v.Account.Name
					account.Face = v.Account.Face
					vl := v.Account.VipInfo.Label
					vipLabel := pb.VipLabel{
						Path:        vl.Path,
						LabelTheme:  vl.LabelTheme,
						TextColor:   vl.TextColor,
						BgStyle:     vl.BgStyle,
						BgColor:     vl.BgColor,
						BorderColor: vl.BorderColor,
					}
					account.Vip = pb.VipInfo{
						Type:               v.Account.VipInfo.Type,
						Status:             v.Account.VipInfo.Status,
						DueDate:            v.Account.VipInfo.DueDate,
						VipPayType:         v.Account.VipInfo.VipPayType,
						ThemeType:          v.Account.VipInfo.ThemeType,
						AvatarSubscript:    v.Account.VipInfo.AvatarSubscript,
						NicknameColor:      v.Account.VipInfo.NicknameColor,
						Role:               v.Account.VipInfo.Role,
						Label:              vipLabel,
						AvatarSubscriptUrl: v.Account.VipInfo.AvatarSubscriptUrl,
					}
					account.Official = pb.OfficialInfo{
						Role:  v.Account.Official.Role,
						Title: v.Account.Official.Title,
						Desc:  v.Account.Official.Desc,
						Type:  v.Account.Official.Type,
					}
				}
				if v.Tag != nil {
					tag.TagID = v.Tag.TID
					tag.Name = v.Tag.Name
				}
				if len(v.Archive) > 0 {
					for _, arc := range v.Archive {
						if arc != nil && arc.Account != nil {
							arcAccount := &pb.Account{
								MID:  arc.Account.Mid,
								Name: arc.Account.Name,
								Face: arc.Account.Face,
							}
							archive = append(archive, &pb.ArchiveInfo{
								BvID:      arc.Bvid,
								Score:     arc.Score,
								Tname:     arc.TypeName,
								Title:     arc.Title,
								Desc:      arc.Desc,
								Duration:  arc.Duration,
								Pic:       arc.Pic,
								View:      int64(arc.View),
								Like:      int64(arc.Like),
								Danmaku:   int64(arc.Danmaku),
								Reply:     int64(arc.Reply),
								Fav:       int64(arc.Fav),
								Coin:      int64(arc.Coin),
								Share:     int64(arc.Share),
								Ctime:     int64(arc.PubDate),
								ShowScore: arc.ShowScore,
								Account:   arcAccount,
								ShowLink:  arc.ShowLink,
							})
						}

					}
				}
				res.List = append(res.List, &pb.RankResult{
					ObjectType: int64(v.ObjectType),
					Score:      v.Score,
					ShowScore:  v.ShowScore,
					Account:    account,
					Archive:    archive,
					Tag:        tag,
				})
			}
		}
	}
	return
}

func (s *activityService) CanUpActReserve4Dynamic(ctx context.Context, req *pb.CanUpActReserve4DynamicReq) (*pb.CanUpActReserve4DynamicReply, error) {
	return service.LikeSvc.CanUpActReserve4Dynamic(ctx, req)
}

func (s *activityService) UpActReserveRecord(ctx context.Context, req *pb.UpActReserveRecordReq) (*pb.UpActReserveRecordReply, error) {
	return service.LikeSvc.UpActReserveRecord(ctx, req)
}

func (s *activityService) QuestionAnswerAll(ctx context.Context, req *pb.QuestionAnswerAllReq) (*pb.QuestionAnswerAllReply, error) {
	ret, err := service.LikeSvc.QuestionAnswerAll(ctx, req.Sid, req.PoolId, req.Mid, req.Answer)
	if ret == nil {
		ret = new(pb.QuestionAnswerAllReply)
	}
	return ret, err
}
func (s *activityService) UpActReserveRelationDependAudit(ctx context.Context, req *pb.UpActReserveRelationDependAuditReq) (*pb.UpActReserveRelationDependAuditReply, error) {
	return service.LikeSvc.UpActReserveRelationDependAudit(ctx, req)
}

func (s *activityService) CanUpActReserveByType(ctx context.Context, req *pb.CanUpActReserveByTypeReq) (*pb.CanUpActReserveByTypeReply, error) {
	return service.LikeSvc.CanUpActReserveByType(ctx, req)
}

func (s *activityService) DelKnowledgeCache(ctx context.Context, req *pb.DelKnowledgeCacheReq) (*pb.NoReply, error) {
	return service.KnowledgeSvr.DelKnowledgeCache(ctx, req)
}

func (c *activityService) CheckBindBWParkTicket(ctx context.Context, req *pb.CheckBindBWParkTicketReq) (replay *pb.CheckBindBWParkTicketResp, err error) {
	replay = new(pb.CheckBindBWParkTicketResp)
	var id int64
	if isWhite := service.BwsSvc.IsWhiteMid(ctx, req.Mid); isWhite {
		replay.Hasbind = true
		return
	}
	if id, err = service.BwsOnlineSvc.CheckBind(ctx, req.Mid); id > 0 && err == nil {
		replay.Hasbind = true
	}
	return
}

func (c *activityService) BatchCacheBindRecords(ctx context.Context, req *pb.BatchCacheBindRecordsReq) (replay *pb.BatchCacheBindRecordsResp, err error) {
	replay = new(pb.BatchCacheBindRecordsResp)
	var records []*bwsonline.TicketBindRecord
	if records, err = service.BwsOnlineSvc.BatchCacheBindRecords(ctx, req.StartIndex, req.Limit); err == nil {
		for _, v := range records {
			replay.RecordIds = append(replay.RecordIds, v.Id)
		}
	}
	return
}

func (s *activityService) CanUpActReserveFull(ctx context.Context, req *pb.CanUpActReserveFullReq) (*pb.CanUpActReserveFullReply, error) {
	return service.LikeSvc.CanUpActReserveFull(ctx, req)
}

func (s *activityService) CanUpRelateOthersActReserve(ctx context.Context, req *pb.CanUpRelateOthersActReserveReq) (*pb.CanUpRelateOthersActReserveReply, error) {
	return service.LikeSvc.CanUpRelateOthersActReserve(ctx, req)
}

func (s *activityService) CanUpRelateReserveAuth(ctx context.Context, req *pb.CanUpRelateReserveAuthReq) (*pb.CanUpRelateReserveAuthReply, error) {
	return service.LikeSvc.CanUpRelateReserveAuth(ctx, req)
}
