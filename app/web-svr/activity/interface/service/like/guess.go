package like

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/net/netutil"
	retryLib "go-common/library/retry"
	"go-gateway/app/web-svr/activity/ecode"
	pb "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/model/guess"

	accmdl "git.bilibili.co/bapis/bapis-go/account/service"
	coinmdl "git.bilibili.co/bapis/bapis-go/community/service/coin"
	espServiceApi "go-gateway/app/web-svr/esports/service/api/v1"

	"go-common/library/sync/errgroup.v2"
)

var _empUserRecord = make([]*pb.GuessUserGroup, 0)

const (
	_blocked     = 1
	_guessReason = "竞猜投注"
	_correct     = 1
	_guess       = 1
	_twoDecimal  = 100.00
	_notFinish   = 0
	_finish      = 1
	_oidGuessMax = 10
	_deleted     = 1
)

// GuessAdd  add guess.
func (s *Service) GuessAdd(c context.Context, p *pb.GuessAddReq) (rs *pb.NoReply, err error) {
	var mainIDs, tmpIDs []*guess.MainID
	rs = &pb.NoReply{}
	if tmpIDs, err = s.guessDao.RawOidMIDs(c, p.Oid, p.Business); err != nil {
		err = ecode.ActGuessDataFail
		return
	}
	for _, main := range tmpIDs {
		if main.IsDeleted == _deleted {
			continue
		}
		mainIDs = append(mainIDs, main)
	}
	if len(mainIDs)+len(p.Groups) > _oidGuessMax {
		err = ecode.ActGuessAddMax
		return
	}
	if err = s.guessDao.AddMatchGuess(c, p); err != nil {
		log.Error("GuessAdd s.guessDao.AddGroupGuess error(%+v)", err)
		return
	}

	go func() {
		if tmpList, resetErr := s.guessDao.ResetMainIDListInCacheByOID(context.Background(), p.Oid, p.Business); resetErr == nil {
			for _, v := range tmpList {
				_ = s.guessDao.ResetGuessAggregationInfoInCacheByMainID(context.Background(), v.ID, p.Business)
			}
		}
	}()

	return rs, nil
}

// GuessEdit  edit guess.
func (s *Service) GuessEdit(c context.Context, p *pb.GuessEditReq) (rs *pb.NoReply, err error) {
	rs = &pb.NoReply{}
	if _, err = s.guessDao.UpGuess(c, p); err != nil {
		log.Error("GuessEdit s.guessDao.UpGuess error(%+v)", err)
		return
	}

	go func() {
		if _, resetErr := s.guessDao.ResetMainIDListInCacheByOID(context.Background(), p.Oid, p.Business); resetErr != nil {
			_ = s.guessDao.DeleteMainIDCache(context.Background(), p.Oid, p.Business)
		}
	}()

	return rs, nil
}

// GuessGroupDel  del guess group.
func (s *Service) GuessGroupDel(c context.Context, r *pb.GuessGroupDelReq) (rs *pb.GuessGroupReply, err error) {
	rs = &pb.GuessGroupReply{}
	var (
		mainGuess       *guess.MainGuess
		affected        int64
		mainIDs, tmpIDs []*guess.MainID
	)
	if mainGuess, err = s.guessDao.GuessMain(c, r.MainID); err != nil {
		err = ecode.ActGuessNotExist
		return
	}
	if affected, err = s.guessDao.DelGroup(c, r.MainID); err != nil {
		log.Error("GuessGroupDel s.guessDao.DelGroup mainID(%d) error(%+v)", r.MainID, err)
		return
	}
	if affected == 0 {
		err = ecode.ActGuessDelFail
		return
	}
	if tmpIDs, err = s.guessDao.RawOidMIDs(c, mainGuess.Oid, mainGuess.Business); err != nil {
		log.Error("GuessGroupDel s.guessDao.RawOidMIDs oid(%d) business(%d) error(%+v)", mainGuess.Oid, mainGuess.Business, err)
		return
	}

	go func() {
		ctx := context.Background()
		if _, resetErr := s.guessDao.ResetMainIDListInCacheByOID(ctx, mainGuess.Oid, mainGuess.Business); resetErr != nil {
			_ = s.guessDao.DeleteMainIDCache(ctx, mainGuess.Oid, mainGuess.Business)
			_ = s.guessDao.DeleteGuessAggregationInfoCache(ctx, r.MainID, mainGuess.Business)
		}
	}()

	for _, main := range tmpIDs {
		if main.IsDeleted == _deleted {
			continue
		}
		mainIDs = append(mainIDs, main)
	}
	if len(mainIDs) > 0 {
		rs.HaveGuess = 1
	}
	return
}

// GuessUpResult  update guess result.
func (s *Service) GuessUpResult(c context.Context, r *pb.GuessUpResultReq) (rs *pb.NoReply, err error) {
	var (
		count     int64
		haveGuess *guess.DetailGuess
	)
	rs = &pb.NoReply{}
	if haveGuess, err = s.guessDao.HaveGuess(c, r.MainID, r.DetailID); err != nil {
		log.Error("GuessUpResult s.guessDao.HaveGuess mainID(%d) detailID(%d) error(%+v)", r.MainID, r.DetailID, err)
		return
	}
	if haveGuess.ID == 0 {
		err = ecode.ActGuessNotExist
		return
	}
	if count, err = s.guessDao.UpGuessResult(c, r.MainID, r.DetailID); err != nil {
		log.Error("GuessUpResult s.guessDao.UpGuessResult mainID(%d) detailID(%d) error(%+v)", r.MainID, r.DetailID, err)
		return
	}
	if count == 0 {
		log.Error("GuessUpResult s.guessDao.UpGuessResult mainID(%d) detailID(%d) fail", r.MainID, r.DetailID)
		err = ecode.ActGuessResFail
	}

	if err == nil {
		go func() {
			ctx := context.Background()
			if d, err := s.guessDao.RawGuessMain(ctx, r.MainID); err == nil {
				if tmpErr := s.guessDao.ResetGuessAggregationInfoInCacheByMainID(ctx, r.MainID, d.Business); tmpErr != nil {
					_ = s.guessDao.DeleteGuessAggregationInfoCache(ctx, r.MainID, d.Business)
				}
			}
		}()
	}

	return
}

// GuessAllList business guess all list.
func (s *Service) GuessAllList(c context.Context, r *pb.GuessListReq) (rs *pb.GuessListAllReply, err error) {
	var (
		mainIDs    []*guess.MainID
		mdRes      map[int64]*guess.MainRes
		matchGuess []*pb.GuessAllList
	)
	rs = &pb.GuessListAllReply{}
	if mainIDs, err = s.guessDao.RawOidMIDs(c, r.Oid, r.Business); err != nil {
		log.Error("s.guessDao.RawOidMIDs oid(%d) business(%d) error(%v)", r.Oid, r.Business, err)
		return
	}
	if mdRes, err = s.mainRes(c, r.Business, mainIDs, false); err != nil {
		log.Error("s.mainRes oid(%d) business(%d) error(%v)", r.Oid, r.Business, err)
		return
	}
	mainDetail := s.mainAllDetail(mdRes)
	for _, main := range mainIDs {
		if md, ok := mainDetail[main.ID]; ok {
			matchGuess = append(matchGuess, md)
		}
	}
	rs = &pb.GuessListAllReply{MatchGuess: matchGuess}
	return
}

func (s *Service) mainAllDetail(mdRes map[int64]*guess.MainRes) (rs map[int64]*pb.GuessAllList) {
	rs = make(map[int64]*pb.GuessAllList, len(mdRes))
	for _, md := range mdRes {
		var detail []*pb.GuessDetail
		for _, d := range md.Details {
			odds := float32(d.Odds) / _twoDecimal
			detail = append(detail, &pb.GuessDetail{Id: d.ID, Odds: odds, Option: d.Option})
		}
		rs[md.ID] = &pb.GuessAllList{
			Id:           md.ID,
			Title:        md.Title,
			GuessCount:   md.GuessCount,
			ResultId:     md.ResultID,
			StakeType:    md.StakeType,
			Details:      detail,
			TemplateType: md.TemplateType,
		}
	}
	return
}

func (s *Service) NewMDResult(c context.Context, business, oid int64, haveDel bool) (rs map[int64]*guess.MainRes, mainIDs []*guess.MainID, err error) {
	if mainIDs, err = s.guessDao.MainIDListByOID(c, oid, business); err != nil {
		log.Error("s.guessDao.MainIDListByOID oid(%d) business(%d) error(%v)", oid, business, err)
		return
	}

	if rs, err = s.NewMainRes(c, business, mainIDs, haveDel); err != nil {
		log.Error("s.NewMainRes oid(%d) business(%d) error(%v)", oid, business, err)
	}

	return
}

func (s *Service) oMDResult(c context.Context, business, oid int64, haveDel bool) (rs map[int64]*guess.MainRes, mainIDs []*guess.MainID, err error) {
	if mainIDs, err = s.guessDao.OidMIDs(c, oid, business); err != nil {
		log.Error("s.guessDao.OidMIDs oid(%d) business(%d) error(%v)", oid, business, err)
		return
	}
	if rs, err = s.mainRes(c, business, mainIDs, haveDel); err != nil {
		log.Error("s.mainRes oid(%d) business(%d) error(%v)", oid, business, err)
	}
	return
}

// GuessList business guess list.
func (s *Service) GuessList(c context.Context, r *pb.GuessListReq) (rs *pb.GuessListReply, err error) {
	var (
		mainIDs []*guess.MainID
		mdRes   map[int64]*guess.MainRes
		mds     []*pb.GuessList
		rsTmp   map[int64][]*pb.GuessList
		ok      bool
	)
	rs = &pb.GuessListReply{}

	mdRes, mainIDs, ok = hotMainListAndDetailListByOIDAndBusiness(r.Oid, r.Business)
	if !ok {
		if mdRes, mainIDs, err = s.NewMDResult(c, r.Business, r.Oid, false); err != nil {
			log.Error("s.oMDResult business(%d) oid(%d) error(%v)", r.Oid, r.Business, err)
			err = ecode.ActGuessesFail
			return
		}
		count := len(mdRes)
		if count == 0 {
			err = xecode.NothingFound
			return
		}
	}

	rsTmp = s.mdResult(c, mdRes, r.Mid, mainIDs)
	for _, main := range mainIDs {
		mds = append(mds, rsTmp[main.ID]...)
	}
	rs.MatchGuess = mds
	return
}

// GuessLists business guess list.
func (s *Service) GuessLists(c context.Context, r *pb.GuessListsReq) (rs map[int64]*pb.GuessListReply, err error) {
	var (
		rsMap   map[int64][]*guess.MainID
		mdRes   map[int64]*guess.MainRes
		mainIDs []*guess.MainID
	)
	if rsMap, err = s.guessDao.OidsMIDs(c, r.Oids, r.Business); err != nil {
		log.Error("s.guessDao.RawOidsMids oids(%d) business(%d) error(%v)", r.Oids, r.Business, err)
		err = ecode.ActGuessesFail
		return
	}
	count := len(rsMap)
	if count == 0 {
		err = xecode.NothingFound
		return
	}
	for _, mains := range rsMap {
		mainIDs = append(mainIDs, mains...)
	}
	if mdRes, err = s.mainRes(c, r.Business, mainIDs, false); err != nil {
		log.Error("s.mainRes oids(%d) business(%d) error(%v)", r.Oids, r.Business, err)
		err = ecode.ActGuessesFail
		return
	}
	rs = make(map[int64]*pb.GuessListReply, count)
	rsTmp := s.mdResult(c, mdRes, r.Mid, mainIDs)
	for oid, mains := range rsMap {
		var tmpList []*pb.GuessList
		for _, main := range mains {
			if guessLists, ok := rsTmp[main.ID]; ok {
				tmpList = append(tmpList, guessLists...)
			}
		}
		if len(tmpList) > 0 {
			rs[oid] = &pb.GuessListReply{MatchGuess: tmpList}
		}
	}
	return
}

func (s *Service) NewMainRes(c context.Context, businessID int64, mainIDs []*guess.MainID, haveDel bool) (mdRes map[int64]*guess.MainRes, err error) {
	var ids []int64
	if len(mainIDs) == 0 {
		return
	}
	for _, mainID := range mainIDs {
		if !haveDel && mainID.IsDeleted == _deleted {
			continue
		}
		ids = append(ids, mainID.ID)
	}
	if mdRes, err = s.guessDao.DetailListByMainIDList(c, ids, businessID); err != nil {
		log.Error("s.guessDao.MDsResult ids(%+v) business(%d) error(%v)", ids, businessID, err)
		err = ecode.ActGuessesFail
		return
	}
	return
}

func (s *Service) mainRes(c context.Context, businessID int64, mainIDs []*guess.MainID, haveDel bool) (mdRes map[int64]*guess.MainRes, err error) {
	var ids []int64
	if len(mainIDs) == 0 {
		return
	}
	for _, mainID := range mainIDs {
		if !haveDel && mainID.IsDeleted == _deleted {
			continue
		}
		ids = append(ids, mainID.ID)
	}
	if mdRes, err = s.guessDao.MDsResult(c, ids, businessID); err != nil {
		log.Error("s.guessDao.MDsResult ids(%+v) business(%d) error(%v)", ids, businessID, err)
		err = ecode.ActGuessesFail
		return
	}
	return
}

func (s *Service) mdResult(c context.Context, mds map[int64]*guess.MainRes, mid int64, mainIDs []*guess.MainID) (rs map[int64][]*pb.GuessList) {
	var (
		list     []*pb.GuessList
		ids      []int64
		userLogs map[int64]*guess.UserGuessLog
		e        error
	)
	count := len(mds)
	if count == 0 {
		return
	}
	if mid > 0 {
		for _, main := range mainIDs {
			ids = append(ids, main.ID)
		}
		if userLogs, e = s.guessDao.UserGuess(c, ids, mid); e != nil {
			log.Error("GuessList s.guessDao.UserGuess mid(%d) mainIDs(%+v) error(%+v)", mid, ids, e)
		}
	}
	rs = make(map[int64][]*pb.GuessList, count)
	for mainID, md := range mds {
		detailRs, isGuess := s.groupGuess(md, userLogs)
		list = append(list, &pb.GuessList{
			Id:           mainID,
			Title:        md.Title,
			StakeType:    md.StakeType,
			IsGuess:      isGuess,
			Details:      detailRs,
			TemplateType: md.TemplateType,
			RightOption:  md.RightOption,
		})
		rs[mainID] = list
		list = nil
	}
	return
}

func (s *Service) groupGuess(oneMain *guess.MainRes, userLogs map[int64]*guess.UserGuessLog) (rs []*pb.GuessDetail, isGuess int64) {
	var (
		income  float32
		stake   int64
		correct int64
	)
	dStakes := make(map[int64]int64, len(oneMain.Details))
	for _, md := range oneMain.Details {
		dStakes[md.ID] = md.TotalStake
	}
	for _, d := range oneMain.Details {
		if oneMain.ResultID > 0 {
			d.Odds = float32(d.Odds) / _twoDecimal
		} else {
			if d.TotalStake > 0 {
				odds := decimal(1+(otherStakes(d.ID, dStakes)/float32(d.TotalStake))*s.c.Rule.GuessPercent, 2)
				if odds > s.c.Rule.GuessMaxOdds {
					odds = s.c.Rule.GuessMaxOdds
				}
				d.Odds = odds
			}
		}
		if userLog, ok := userLogs[oneMain.ID]; ok && userLog.DetailID == d.ID {
			isGuess = _guess
			income = decimal(userLog.Income/_twoDecimal, 1)
			stake = userLog.Stake
		} else {
			income = 0
			stake = 0
		}
		correct = 0
		if d.ID == oneMain.ResultID {
			correct = _correct
		}
		rs = append(rs, &pb.GuessDetail{Id: d.ID, Odds: d.Odds, Option: d.Option, Income: income, Stake: stake, Correct: correct})
	}
	return
}

func decimal(f float32, n int) float32 {
	n10 := math.Pow10(n)
	rs64 := math.Trunc((float64(f)+0.5/n10)*n10) / n10
	return float32(rs64)
}

func otherStakes(id int64, dStakes map[int64]int64) (rs float32) {
	var stakes int64
	for k, stake := range dStakes {
		if k != id {
			stakes = stakes + stake
		}
	}
	return float32(stakes)
}

// UserAddGuess user add guess.
func (s *Service) UserAddGuess(c context.Context, r *pb.GuessUserAddReq) (rs *pb.NoReply, err error) {
	var (
		mainGuess  *guess.MainGuess
		list       map[int64]*guess.UserGuessLog
		countReply *coinmdl.UserCoinsReply
		coinErr    error
		userLogID  int64
		user       *accmdl.Profile
		ip         = metadata.String(c, metadata.RemoteIP)
		check      bool
	)
	rs = &pb.NoReply{}
	if check, err = s.guessDao.RsSetNX(c, fmt.Sprintf("add_%d", r.Mid), 1); err != nil || !check {
		log.Warn("UserAddGuess mid:%d to fast err:%v", r.Mid, err)
		err = ecode.ActGuessFail
		return
	}
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) (err error) {
		if user, err = s.userInfo(ctx, r.Mid); err != nil {
			err = ecode.ActGuessFail
		}
		return err
	})
	group.Go(func(ctx context.Context) (err error) {
		if mainGuess, err = s.guessDao.GuessMain(c, r.MainID); err != nil {
			log.Errorc(c, "UserAddGuess s.guessDao.MainGuess(%d) error(%+v) ", r.DetailID, err)
			err = ecode.ActGuessFail
		}
		return err
	})
	group.Go(func(ctx context.Context) (err error) {
		mainIDs := []int64{r.MainID}
		if list, err = s.guessDao.UserGuess(c, mainIDs, r.Mid); err != nil {
			log.Errorc(c, "UserAddGuess s.guessDao.UserGuess mid(%d) mainID(%d) error(%+v)", r.Mid, r.MainID, err)
			err = ecode.ActGuessFail
		}
		return err
	})
	if err = group.Wait(); err != nil {
		return
	}
	if user.Silence == _blocked {
		err = ecode.ActGuessDisabled
		return
	}
	if mainGuess == nil || mainGuess.ID == 0 {
		err = ecode.ActGuessNotExist
		return
	}
	if time.Now().Unix() > mainGuess.Stime {
		err = ecode.ActGuessOverEnd
		return
	}
	if mainGuess.ResultID > 0 || time.Now().Unix() >= mainGuess.Etime {
		err = ecode.ActGuessOverEnd
		return
	}
	if r.Stake > mainGuess.MaxStake {
		err = ecode.ActGuessOverMax
		return
	}
	if mainGuess.StakeType == int64(pb.StakeType_coinType) {
		if countReply, err = s.coinClient.UserCoins(c, &coinmdl.UserCoinsReq{Mid: r.Mid}); err != nil {
			log.Error("s.coinClient.UserCoins(%d) error(%v)", r.Mid, err)
			return
		}
		if countReply.Count < float64(r.Stake) {
			log.Error("s.coinClient.UserCoins(%d) countReply(%v) MaxStake(%d)", r.Mid, countReply.Count, mainGuess.MaxStake)
			err = ecode.ActGuessCoinFail
			return
		}
	}
	if len(list) > 0 {
		err = ecode.ActUserGuessAlready
		return
	}
	if userLogID, err = s.guessDao.UserAddGuess(c, mainGuess.Business, mainGuess.ID, r); err != nil {
		log.Error("UserAddGuess s.guessDao.UserAddGuess business(%d)  mid(%d) mainID(%d) error(%+v)", mainGuess.Business, r.Mid, r.MainID, err)
		err = ecode.ActGuessFail
		return
	}
	if r.StakeType == int64(pb.StakeType_coinType) {
		loseCoin := float64(-r.Stake)
		if _, coinErr = s.coinClient.ModifyCoins(
			c,
			&coinmdl.ModifyCoinsReq{
				Mid:      r.Mid,
				Count:    loseCoin,
				Reason:   _guessReason,
				IP:       ip,
				UniqueID: fmt.Sprintf("%v_%v_%v", r.MainID, r.DetailID, r.Mid),
				Caller:   "esports_activity",
				Operator: "LeeLei"}); coinErr != nil {
			log.Error("s.coinClient.ModifyCoin(%d,%d,%s) error(%v)", r.Mid, r.Stake, ip, coinErr)
			err = nil
		}
	}
	// update cache
	s.cache.Do(c, func(c context.Context) {
		var userLog []*guess.UserGuessLog
		userLog = append(userLog, &guess.UserGuessLog{ID: userLogID, Mid: r.Mid, MainID: r.MainID, DetailID: r.DetailID, StakeType: r.StakeType, Stake: r.Stake})
		s.guessDao.DelGuessCache(c, mainGuess.Oid, mainGuess.Business, r.MainID, r.Mid, r.StakeType)
		s.guessDao.AppendCacheUserGuessList(c, r.Mid, userLog, mainGuess.Business)
		s.guessReportContest(c, mainGuess.Oid, r.Mid)
	})
	return
}

func (s *Service) guessReportContest(ctx context.Context, oid, mid int64) {
	arg := &espServiceApi.GetContestRequest{
		Mid: mid,
		Cid: oid,
	}
	contestInfo, err := client.EspServiceClient.GetContestInfo(ctx, arg)
	if err != nil {
		log.Errorc(ctx, "sendGuessDatabus arg(%+v) error(%+v)", arg, err)
		return
	}
	s.sendGuessDatabus(ctx, mid, contestInfo)
}

func (s *Service) sendGuessDatabus(ctx context.Context, mid int64, contestInfo *espServiceApi.ContestInfo) {
	if contestInfo == nil || contestInfo.Contest == nil {
		return
	}
	reqParam := struct {
		Timestamp int64 `json:"timestamp"`
		Mid       int64 `json:"mid"`
		ContestId int64 `json:"contest_id"`
		Stime     int64 `json:"stime"`
		Etime     int64 `json:"etime"`
		HomeId    int64 `json:"home_id"`
		AwayId    int64 `json:"away_id"`
		SeasonId  int64 `json:"season_id"`
		MatchId   int64 `json:"match_id"`
		SeriesId  int64 `json:"series_id"`
		LiveRoom  int64 `json:"live_room"`
		GameId    int64 `json:"game_id"`
	}{
		time.Now().Unix(),
		mid,
		contestInfo.Contest.ID,
		contestInfo.Contest.Stime,
		contestInfo.Contest.Etime,
		contestInfo.Contest.HomeID,
		contestInfo.Contest.AwayID,
		contestInfo.Contest.Season.ID,
		contestInfo.Contest.MatchID,
		contestInfo.Contest.SeriesId,
		contestInfo.Contest.LiveRoom,
		contestInfo.Contest.GameId,
	}
	key := fmt.Sprintf("guess_report_mid_%v_contest_%v", mid, contestInfo.Contest.ID)
	if err := retryLib.WithAttempts(ctx, "interface_guess_send_contest_info", 3, netutil.DefaultBackoffConfig, func(c context.Context) error {
		return component.ActGuessProducer.Send(ctx, key, reqParam)
	}); err != nil {
		log.Errorc(ctx, "sendGuessDatabus component.ActGuessProducer.Send mid(%d) reqParam(%+v) error(%+v)", mid, reqParam, err)
		return
	}
	log.Infoc(ctx, "sendGuessDatabus component.ActGuessProducer.Send mid(%d) reqParam(%+v) success", mid, reqParam)
	return
}

func (s *Service) userInfo(c context.Context, mid int64) (*accmdl.Profile, error) {
	arg := &accmdl.MidReq{
		Mid: mid,
	}
	res, err := s.accClient.Profile3(c, arg)
	if err != nil {
		log.Error("s.accClient.Profile3(%d) error(%v)", mid, err)
		return nil, err
	}
	return res.Profile, nil
}

// UserGuessList user guess list.
func (s *Service) UserGuessList(c context.Context, r *pb.UserGuessListReq) (rs []*pb.GuessUserGroup, count int64, err error) {
	var (
		tmp, gu, list []*guess.UserGuessLog
		mainIDs       []int64
		mdRes         map[int64]*guess.MainRes
	)
	if tmp, err = s.guessDao.UserGuessList(c, r.Mid, r.Business); err != nil {
		log.Error("UserGuessList s.guessDao.UserGuessList mid(%d) business(%d) error(%+v)", r.Mid, r.Business, err)
		err = ecode.ActGuessDataFail
		return
	}
	if r.Status == 1 {
		//进行中
		for _, v := range tmp {
			if v.Status == _notFinish {
				gu = append(gu, v)
			}
		}
	} else if r.Status == 2 {
		//已结束
		for _, v := range tmp {
			if v.Status == _finish {
				gu = append(gu, v)
			}
		}
	} else {
		//全量数据
		gu = tmp
	}
	count = int64(len(gu))
	if count == 0 {
		err = xecode.NothingFound
		return
	}
	start := (r.Pn - 1) * r.Ps
	end := start + r.Ps - 1
	if count < start {
		return
	}
	if count > end+1 {
		list = gu[start : end+1]
	} else {
		list = gu[start:]
	}
	for _, mainsUser := range list {
		mainIDs = append(mainIDs, mainsUser.MainID)
	}
	if mdRes, err = s.guessDao.MDsResult(c, mainIDs, r.Business); err != nil {
		log.Error("s.guessDao.MDsResult mainIDs(%+v) business(%d) error(%v)", mainIDs, r.Business, err)
		return
	}
	rs = s.guessUserGroup(list, mdRes)
	return
}

func (s *Service) guessUserGroup(list []*guess.UserGuessLog, mdRes map[int64]*guess.MainRes) (rs []*pb.GuessUserGroup) {
	var (
		mdGroup *guess.MainRes
		ok      bool
		odds    float32
		option  string
	)
	for _, userLog := range list {
		var (
			correct     int64
			income      float32
			rightOption string
		)
		if mdGroup, ok = mdRes[userLog.MainID]; !ok {
			continue
		} else {
			for _, detail := range mdGroup.Details {
				if detail.ID == userLog.DetailID {
					option = detail.Option
					odds = detail.Odds / _twoDecimal
				}

				if detail.ID == mdGroup.ResultID {
					rightOption = detail.Option
				}
			}
		}
		if userLog.DetailID == mdGroup.ResultID {
			correct = 1
		}
		if userLog.Income > 0 {
			income = decimal(float32(userLog.Stake)+userLog.Income/_twoDecimal, 1)
		}
		rs = append(rs, &pb.GuessUserGroup{
			Id:           userLog.ID,
			Stake:        userLog.Stake,
			Income:       income,
			Status:       userLog.Status,
			Ctime:        userLog.Ctime,
			MainID:       mdGroup.ID,
			Oid:          mdGroup.Oid,
			Title:        mdGroup.Title,
			StakeType:    mdGroup.StakeType,
			ResultId:     mdGroup.ResultID,
			DetailID:     userLog.DetailID,
			Odds:         odds,
			Option:       option,
			Stime:        mdGroup.Stime,
			Etime:        mdGroup.Etime,
			Correct:      correct,
			IsDeleted:    mdGroup.IsDeleted,
			TemplateType: mdGroup.TemplateType,
			RightOption:  rightOption,
		})
	}
	return
}

// UserGuessGroup  user guess mainID.
func (s *Service) UserGuessGroup(c context.Context, r *pb.UserGuessGroupReq) (rs *pb.GuessUserGroup, err error) {
	var (
		listMap map[int64]*guess.UserGuessLog
		list    []*guess.UserGuessLog
		mdRes   *guess.MainRes
		tmpRs   []*pb.GuessUserGroup
		mdMap   map[int64]*guess.MainRes
	)
	rs = &pb.GuessUserGroup{}
	ids := []int64{r.MainId}
	if listMap, err = s.guessDao.UserGuess(c, ids, r.Mid); err != nil {
		log.Error("UserGuessGroup s.guessDao.UserGuess mid(%d) mainIDs(%+v) error(%+v)", r.Mid, ids, err)
		err = ecode.ActGuessDataFail
		return
	}
	if len(listMap) == 0 {
		err = xecode.NothingFound
		return
	}
	if mdRes, err = s.guessDao.MDResult(c, r.MainId, r.Business); err != nil {
		log.Error("UserGuessGroup s.guessDao.MDResult MainId(%+v) business(%d) error(%v)", r.MainId, r.Business, err)
		err = ecode.ActGuessDataFail
		return
	}
	mdMap = make(map[int64]*guess.MainRes)
	mdMap[r.MainId] = mdRes
	for _, userLog := range listMap {
		list = append(list, userLog)
	}
	tmpRs = s.guessUserGroup(list, mdMap)
	rs = tmpRs[0]
	return
}

// UserLogData user guess data.
func (s *Service) UserLogData(c context.Context, r *pb.UserGuessDataReq) (rs *pb.UserGuessDataReply, err error) {
	if rs, err = s.guessDao.UserStat(c, r.Mid, r.StakeType, r.Business); err != nil {
		log.Error("UserLogData s.guessDao.UserLog mid(%d) StakeType(%d) business(%d) error(%+v)", r.Mid, r.StakeType, r.Business, err)
		err = ecode.ActGuessDataFail
		return
	}
	if rs.Id == 0 {
		err = xecode.NothingFound
	}
	return
}

// UserGuessMatch user guess match list.
func (s *Service) UserGuessMatch(c context.Context, r *pb.UserGuessMatchReq) (rs *pb.UserGuessMatchReply, err error) {
	var (
		listMap map[int64]*guess.UserGuessLog
		list    []*guess.UserGuessLog
		mainIDs []*guess.MainID
		ids     []int64
		mdRes   map[int64]*guess.MainRes
		tmpRs   []*pb.GuessUserGroup
	)
	rs = &pb.UserGuessMatchReply{}
	if mdRes, mainIDs, err = s.oMDResult(c, r.Business, r.Oid, true); err != nil {
		log.Error("s.oMDResult oid(%d) business(%d) error(%v)", r.Oid, r.Business, err)
		err = ecode.ActGuessDataFail
		return
	}
	if len(mainIDs) == 0 {
		err = xecode.NothingFound
		return
	}
	for _, main := range mainIDs {
		if main.OID == r.Oid {
			ids = append(ids, main.ID)
		}
	}
	if listMap, err = s.guessDao.UserGuess(c, ids, r.Mid); err != nil {
		log.Error("UserGuessMatch s.guessDao.UserGuess mid(%d) mainIDs(%+v) error(%+v)", r.Mid, ids, err)
		err = ecode.ActGuessDataFail
		return
	}
	if len(listMap) == 0 {
		err = xecode.NothingFound
		return
	}
	for _, userLog := range listMap {
		list = append(list, userLog)
	}
	tmpRs = s.guessUserGroup(list, mdRes)
	sort.Slice(tmpRs, func(i, j int) bool { return tmpRs[i].Id > tmpRs[j].Id })
	rs.UserGroup = tmpRs
	return
}

// UserGuessResult user guess right wrong result.
func (s *Service) UserGuessResult(c context.Context, r *pb.UserGuessResultReq) (rs *pb.UserGuessResultReply, err error) {
	var (
		total, success int64
		tmpList        []*pb.GuessUserGroup
	)
	if tmpList, total, err = s.userGuessRecords(c, r.Business, r.Mid, r.Oids); err != nil {
		return
	}
	for _, record := range tmpList {
		if record.Correct == _correct {
			success++
		}
	}
	rs = &pb.UserGuessResultReply{
		Business:     r.Business,
		Mid:          r.Mid,
		TotalGuess:   total,
		TotalSuccess: success,
	}
	if have, e := s.dao.RsNXGet(c, suitsCheckKey(r.Mid, r.Sid)); e != nil {
		log.Error("UserGuessResult s.dao.RsNXGet mid(%d) sid(%d) error(%v)", r.Mid, r.Sid, e)
	} else if have == "1" {
		rs.HaveSuit = 1
	}
	return
}

// UserGuessMatchs user guess more match list.
func (s *Service) UserGuessMatchs(c context.Context, r *pb.UserGuessMatchsReq) (rs []*pb.GuessUserGroup, count int64, err error) {
	var tmpRs []*pb.GuessUserGroup
	if tmpRs, count, err = s.userGuessRecords(c, r.Business, r.Mid, r.Oids); err != nil {
		return
	}
	start := (r.Pn - 1) * r.Ps
	end := start + r.Ps - 1
	if count < start {
		return
	}
	sort.Slice(tmpRs, func(i, j int) bool { return tmpRs[i].Id > tmpRs[j].Id })
	if count > end+1 {
		tmpRs = tmpRs[start : end+1]
	} else {
		tmpRs = tmpRs[start:]
	}
	rs = tmpRs
	return
}

func (s *Service) userGuessRecords(c context.Context, business, mid int64, oids []int64) (rs []*pb.GuessUserGroup, count int64, err error) {
	var (
		rsMap   map[int64][]*guess.MainID
		ids     []int64
		mainIDs []*guess.MainID
		listMap map[int64]*guess.UserGuessLog
		list    []*guess.UserGuessLog
		mdRes   map[int64]*guess.MainRes
	)
	if rsMap, err = s.guessDao.OidsMIDs(c, oids, business); err != nil {
		log.Error("s.guessDao.RawOidsMids oids(%d) business(%d) error(%v)", oids, business, err)
		err = ecode.ActGuessesFail
		return
	}
	if len(rsMap) == 0 {
		rs = _empUserRecord
		return
	}
	for _, mains := range rsMap {
		for _, main := range mains {
			if _, ok := rsMap[main.OID]; ok {
				ids = append(ids, main.ID)
				mainIDs = append(mainIDs, main)
			}
		}
	}
	if len(ids) == 0 || len(mainIDs) == 0 {
		rs = _empUserRecord
		return
	}
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		if mdRes, err = s.mainRes(ctx, business, mainIDs, true); err != nil {
			log.Error("s.mainRes oids(%d) business(%d) error(%v)", oids, business, err)
			err = ecode.ActGuessesFail
		}
		return err
	})
	group.Go(func(ctx context.Context) error {
		if listMap, err = s.guessDao.UserGuess(ctx, ids, mid); err != nil {
			log.Error("UserGuessMatchs s.guessDao.UserGuess mid(%d) mainIDs(%+v) error(%+v)", mid, ids, err)
			err = ecode.ActGuessDataFail
		}
		return err
	})
	if err = group.Wait(); err != nil {
		return
	}
	count = int64(len(listMap))
	if count == 0 {
		rs = _empUserRecord
		return
	}
	for _, userLog := range listMap {
		list = append(list, userLog)
	}
	rs = s.guessUserGroup(list, mdRes)
	return
}
