package service

import (
	"context"
	"encoding/json"
	"fmt"
	bm "go-common/library/net/http/blademaster"
	v12 "go-gateway/app/web-svr/esports/service/api/v1"
	"strconv"
	"time"

	"go-common/library/conf/env"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	errGroup "go-common/library/sync/errgroup.v2"
	"go-common/library/xstr"
	actmdl "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/esports/admin/client"
	"go-gateway/app/web-svr/esports/admin/component"
	"go-gateway/app/web-svr/esports/admin/model"
	"go-gateway/app/web-svr/esports/ecode"
	v1 "go-gateway/app/web-svr/esports/interface/api/v1"

	tunnelCommon "git.bilibili.co/bapis/bapis-go/platform/common/tunnel"

	bGroup "git.bilibili.co/bapis/bapis-go/platform/service/bgroup/v2"
	tunnelmdl "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"
	tunnelV2Mdl "git.bilibili.co/bapis/bapis-go/platform/service/tunnel/v2"
)

var (
	_emptyContestList       = make([]*model.Contest, 0)
	_emptyContestData       = make([]*model.ContestData, 0)
	_rollBackErrorMsg       = "[Service][Contest][Add][DB][Transaction][Rollback][Error], err:(%+v)"
	_rollBackErrorMsgUpdate = "[Service][Contest][Update][DB][Transaction][Rollback][Error], err:(%+v)"
	_clearCacheErrorMsg     = "contest component ClearComponentContestCacheGRPC SeasonID(%+v) ContestID(%d) error(%+v)"
)

const (
	_sortDesc                  = 1
	_sortASC                   = 2
	_ReplyTypeContest          = "27"
	_platform                  = 3
	_pushClose                 = 0
	_pushOpen                  = 1
	cacheKey4AutoSubscribeList = "auto_subscribe"
	_imagePre                  = "https://i0.hdslb.com"
	_imagePreUat               = "https://uat-i0.hdslb.com"
	_contestFreeze             = 1
	_contestInUse              = 0
)

// ContestInfo .
func (s *Service) ContestInfo(c context.Context, id int64) (data *model.ContestInfo, err error) {
	var (
		gameMap map[int64][]*model.Game
		teamMap map[int64]*model.Team
		teamIDs []int64
		hasTeam bool
	)
	contest := new(model.Contest)
	if err = s.dao.DB.Where("id=?", id).First(&contest).Error; err != nil {
		log.Error("ContestInfo Error (%v)", err)
		return
	}
	if gameMap, err = s.gameList(model.TypeContest, []int64{id}); err != nil {
		return
	}
	if contest.HomeID > 0 {
		teamIDs = append(teamIDs, contest.HomeID)
	}
	if contest.AwayID > 0 {
		teamIDs = append(teamIDs, contest.AwayID)
	}
	if contest.SuccessTeam > 0 {
		teamIDs = append(teamIDs, contest.SuccessTeam)
	}
	if ids := unique(teamIDs); len(ids) > 0 {
		var teams []*model.Team
		if err = s.dao.DB.Model(&model.Team{}).Where("id IN (?)", ids).Find(&teams).Error; err != nil {
			log.Error("ContestList team Error (%v)", err)
			return
		}
		if len(teams) > 0 {
			hasTeam = true
		}
		teamMap = make(map[int64]*model.Team, len(teams))
		for _, v := range teams {
			teamMap[v.ID] = v
		}
	}
	data = &model.ContestInfo{Contest: contest}
	if len(gameMap) > 0 {
		if games, ok := gameMap[id]; ok {
			data.Games = games
		}
	}
	if len(data.Games) == 0 {
		data.Games = _emptyGameList
	}
	if hasTeam {
		if team, ok := teamMap[contest.HomeID]; ok {
			data.HomeName = team.Title
		}
		if team, ok := teamMap[contest.AwayID]; ok {
			data.AwayName = team.Title
		}
		if team, ok := teamMap[contest.SuccessTeam]; ok {
			data.SuccessName = team.Title
		}
	}
	var cDatas []*model.ContestData
	if err = s.dao.DB.Model(&model.ContestData{}).Where(map[string]interface{}{"is_deleted": _notDeleted}).Where("cid IN (?)", []int64{id}).Find(&cDatas).Error; err != nil {
		log.Error("ContestInfo Find ContestData Error (%v)", err)
		return
	}
	// 获取赛程队列列表
	contestTeams, err := s.GetContestTeams(c, id)
	if err != nil {
		log.Errorc(c, "[Service][Contest][GetContestTeams][Error], err:(%+v)", err)
		return
	}
	teamIds := make([]int64, 0)
	for _, team := range contestTeams {
		teamIds = append(teamIds, team.TeamId)
	}
	data.TeamIds = xstr.JoinInts(teamIds)
	data.Data = cDatas
	if contest.SeriesID > 0 {
		if d, err := s.FetchContestSeriesByID(c, contest.SeriesID); err == nil {
			data.Series = d
		}
	}

	return
}

// ContestList .
func (s *Service) ContestList(c context.Context, mid, sid, pn, ps, srt, teamid, guessType int64) (list []*model.ContestInfo, count int64, err error) {
	var contests []*model.Contest
	source := s.dao.DB.Model(&model.Contest{})
	if srt == _sortDesc {
		source = source.Order("stime DESC")
	} else if srt == _sortASC {
		source = source.Order("stime ASC")
	}
	if mid > 0 {
		source = source.Where("mid=?", mid)
	}
	if sid > 0 {
		source = source.Where("sid=?", sid)
	}
	if teamid > 0 {
		source = source.Where("home_id = ? OR away_id = ?", teamid, teamid)
	}
	if guessType >= 0 {
		source = source.Where("guess_type=?", guessType)
	}
	source.Count(&count)
	if err = source.Offset((pn - 1) * ps).Limit(ps).Find(&contests).Error; err != nil {
		log.Error("ContestList Error (%v)", err)
		return
	}
	if len(contests) == 0 {
		contests = _emptyContestList
		return
	}
	if list, err = s.contestInfos(contests, true); err != nil {
		log.Error("s.contestInfos Error (%v)", err)
	}
	return
}

func (s *Service) contestInfos(contests []*model.Contest, useGame bool) (list []*model.ContestInfo, err error) {
	var (
		conIDs, teamIDs            []int64
		gameMap                    map[int64][]*model.Game
		teamMap                    map[int64]*model.Team
		cDataMap                   map[int64][]*model.ContestData
		hasGame, hasTeam, hasCData bool
	)
	for _, v := range contests {
		conIDs = append(conIDs, v.ID)
		if v.HomeID > 0 {
			teamIDs = append(teamIDs, v.HomeID)
		}
		if v.AwayID > 0 {
			teamIDs = append(teamIDs, v.AwayID)
		}
		if v.SuccessTeam > 0 {
			teamIDs = append(teamIDs, v.SuccessTeam)
		}
	}
	if useGame {
		if gameMap, err = s.gameList(model.TypeContest, conIDs); err != nil {
			return
		} else if len(gameMap) > 0 {
			hasGame = true
		}
	}
	if ids := unique(teamIDs); len(ids) > 0 {
		var teams []*model.Team
		if err = s.dao.DB.Model(&model.Team{}).Where("id IN (?)", ids).Find(&teams).Error; err != nil {
			log.Error("ContestList team Error (%v)", err)
			return
		}
		if len(teams) > 0 {
			hasTeam = true
		}
		teamMap = make(map[int64]*model.Team, len(teams))
		for _, v := range teams {
			teamMap[v.ID] = v
		}
	}
	if len(conIDs) > 0 {
		var cDatas []*model.ContestData
		if err = s.dao.DB.Model(&model.ContestData{}).Where(map[string]interface{}{"is_deleted": _notDeleted}).Where("cid IN (?)", conIDs).Find(&cDatas).Error; err != nil {
			log.Error("ContestList Find ContestData Error (%v)", err)
			return
		}
		if len(cDatas) > 0 {
			hasCData = true
		}
		cDataMap = make(map[int64][]*model.ContestData, len(cDatas))
		for _, v := range cDatas {
			cDataMap[v.CID] = append(cDataMap[v.CID], v)
		}
	}
	for _, v := range contests {
		contest := &model.ContestInfo{Contest: v}
		if hasGame {
			if games, ok := gameMap[v.ID]; ok {
				contest.Games = games
			}
		}
		if len(contest.Games) == 0 {
			contest.Games = _emptyGameList
		}
		if hasTeam {
			if team, ok := teamMap[v.HomeID]; ok {
				contest.HomeName = team.Title
			}
			if team, ok := teamMap[v.AwayID]; ok {
				contest.AwayName = team.Title
			}
			if team, ok := teamMap[v.SuccessTeam]; ok {
				contest.SuccessName = team.Title
			}
		}
		if hasCData {
			if cData, ok := cDataMap[v.ID]; ok {
				contest.Data = cData
			}
		} else {
			contest.Data = _emptyContestData
		}
		list = append(list, contest)
	}
	return
}

// AddContest .
func (s *Service) AddContest(c context.Context, param *model.Contest, gids []int64) (err error) {
	season, err := s.contestParamsCheck(c, param)
	if err != nil {
		return
	}
	// check game idsEsportsCDataErr
	var (
		games       []*model.Game
		gidMaps     []*model.GIDMap
		contestData []*model.ContestData
	)
	if param.DataType == 0 {
		param.Data = ""
		param.MatchID = 0
	}
	if param.Data != "" {
		if err = json.Unmarshal([]byte(param.Data), &contestData); err != nil {
			err = ecode.EsportsContestDataErr
			return
		}
	}
	if err = s.dao.DB.Model(&model.Game{}).Where("status=?", _statusOn).Where("id IN (?)", gids).Find(&games).Error; err != nil {
		log.Error("AddContest check game ids Error (%v)", err)
		return
	}
	if len(games) == 0 {
		log.Error("AddContest games(%v) not found", gids)
		err = xecode.RequestErr
		return
	}
	pushSwitch := param.LiveRoom > 0 && param.Special == 0
	if pushSwitch {
		param.PushSwitch = _pushOpen
	}
	tx := s.dao.DB.Begin()
	log.Infoc(c, "[Service][Contest][Add][DB][Transaction][Begin]")
	if err = tx.Model(&model.Contest{}).Create(param).Error; err != nil {
		log.Error("AddContest tx.Model Create(%+v) error(%v)", param, err)
		if errR := tx.Rollback().Error; errR != nil {
			log.Errorc(c, _rollBackErrorMsg, errR)
		}
		err = xecode.Errorf(xecode.RequestErr, "新增时保存es_contests表失败(%+v)", err)
		return
	}
	for _, v := range games {
		gidMaps = append(gidMaps, &model.GIDMap{Type: model.TypeContest, Oid: param.ID, Gid: v.ID})
	}
	sql, sqlParam := model.GidBatchAddSQL(gidMaps)
	if err = tx.Model(&model.GIDMap{}).Exec(sql, sqlParam...).Error; err != nil {
		log.Error("AddContest tx.Model Create(%+v) error(%v)", param, err)
		if err = tx.Rollback().Error; err != nil {
			log.Errorc(c, _rollBackErrorMsg, err)
		}
		err = xecode.Errorf(xecode.RequestErr, "新增时[GidBatchAddSQL]保存es_gid_map表失败(%+v)", err)
		return
	}
	if len(contestData) > 0 {
		sql, sqlParam := model.BatchAddCDataSQL(param.ID, contestData)
		if err = tx.Model(&model.Module{}).Exec(sql, sqlParam...).Error; err != nil {
			log.Error("AddContest Module tx.Model Create(%+v) error(%v)", sqlParam, err)
			if err = tx.Rollback().Error; err != nil {
				log.Errorc(c, _rollBackErrorMsg, err)
			}
			err = xecode.Errorf(xecode.RequestErr, "新增[BatchAddCDataSQL]保存es_gid_map表失败(%+v)", err)
			return
		}
	}
	if err = s.addNewEventAndGroup(c, param, season); err != nil {
		log.Errorc(c, "AddContest s.addNewEventAndGroup contestID(%d) error(%+v)", param.ID, err)
		if err = tx.Rollback().Error; err != nil {
			log.Errorc(c, _rollBackErrorMsg, err)
		}
		err = xecode.Errorf(xecode.RequestErr, "开始时间与结束时间不能为空，为空时天马订阅卡无法推送。stime(%d) etime(%d)", param.Stime, param.Etime)
		return
	}

	// 添加赛程队伍
	if err = s.ContestTeamsAdd(c, param.ID, param.Sid, param.TeamIds); err != nil {
		if errR := tx.Rollback().Error; errR != nil {
			log.Errorc(c, _rollBackErrorMsg, errR)
		}
		return
	}

	// 填写直接间并且是普通赛程发push
	if pushSwitch {
		if param.Stime == 0 || param.Etime == 0 {
			err = fmt.Errorf("开始时间与结束时间不能为空，为空时天马订阅卡无法推送。stime(%d) etime(%d)", param.Stime, param.Etime)
			return
		}
		if s.isNewBGroup(param.LiveRoom) { // 根据直播间区分建小卡
			err = s.addNewSmallCard(c, param, season) // 新小卡
		} else {
			err = s.addOldEventAndCard(c, param, season) // 老天马卡
		}
		if err != nil {
			log.Errorc(c, "[Service][Contest][Add][AddCard][Error], err:(%+v)", err)
			if err = tx.Rollback().Error; err != nil {
				log.Errorc(c, _rollBackErrorMsg, err)
			}
			err = xecode.Errorf(xecode.RequestErr, "增加天马下卡失败")
			return
		}
	}
	if err = tx.Commit().Error; err != nil {
		log.Errorc(c, "[Service][Contest][Add][DB][Transaction][Commit][Error], err:(%+v)", err)
		err = xecode.Errorf(xecode.RequestErr, "数据库异常，请重试")
		return
	}
	log.Infoc(c, "[Service][Contest][Update][DB][Transaction][Commit][Error]")
	s.AsyncRebuildContestTeamsCache(c, param)
	go func() {
		_ = s.ClearESportCacheByType(v1.ClearCacheType_CONTEST, []int64{param.ID})
	}()

	// 删除赛程组件缓存.
	s.cache.Do(c, func(c context.Context) {
		if e := s.ClearComponentContestCacheByGRPC(&v1.ClearComponentContestCacheRequest{SeasonID: param.Sid, ContestID: param.ID, ContestHome: param.HomeID, ContestAway: param.AwayID}); e != nil {
			log.Errorc(c, _clearCacheErrorMsg, param.Sid, param.ID, err)
		}
	})

	conn := component.GlobalAutoSubCache.Get(c)
	autoSubK4Home := model.AutoSubscribeDetail{
		SeasonID:  param.Sid,
		TeamId:    param.HomeID,
		ContestID: param.ID,
	}
	bs4Home, _ := json.Marshal(autoSubK4Home)
	autoSub4Away := model.AutoSubscribeDetail{
		SeasonID:  param.Sid,
		TeamId:    param.AwayID,
		ContestID: param.ID,
	}
	bs4Away, _ := json.Marshal(autoSub4Away)
	if _, pushErr := conn.Do("LPUSH", cacheKey4AutoSubscribeList, string(bs4Home)); pushErr != nil {
		fmt.Println("LPUSH", cacheKey4AutoSubscribeList, string(bs4Home), pushErr)
	}
	if _, pushErr := conn.Do("LPUSH", cacheKey4AutoSubscribeList, string(bs4Away)); pushErr != nil {
		fmt.Println("LPUSH", cacheKey4AutoSubscribeList, string(bs4Away), pushErr)
	}
	_ = conn.Close()
	// register reply
	if err = s.dao.RegReply(c, param.ID, param.Adid, _ReplyTypeContest); err != nil {
		err = nil
	}

	var updateGuessBizErr error
	req4UpdateGuessBiz := new(v1.UpdateSeasonGuessVersionRequest)
	{
		req4UpdateGuessBiz.MatchId = param.ID
	}
	for i := 0; i < 3; i++ {
		_, updateGuessBizErr = s.espClient.UpdateSeasonGuessVersion(c, req4UpdateGuessBiz)
		if updateGuessBizErr == nil {
			break
		}
	}
	if updateGuessBizErr != nil {
		log.Errorc(c, "s.espClient.UpdateSeasonGuessVersion param(%+v) error(%v)", param.ID, err)
	}

	if _, err := s.espClient.RefreshContestDataPageCache(c, &v1.RefreshContestDataPageCacheRequest{
		Cids: []int64{param.ID},
	}); err != nil {
		log.Errorc(c, "s.espClient.RefreshContestDataPageCache  param(%+v) error(%v)", param.ID, err)
		return err
	}
	return
}

func (s *Service) contestParamsCheck(ctx context.Context, param *model.Contest) (season *model.Season, err error) {
	if param.LiveRoom > 0 && !isLiveRoomValid(ctx, []int64{param.LiveRoom}) {
		err = ecode.EsportsMatchLiveInvalid
		return
	}
	seasonId := param.Sid
	seasonInfo, err := s.SeasonInfo(ctx, seasonId)
	if err != nil {
		err = xecode.Errorf(xecode.RequestErr, "赛季信息获取失败，请重试")
		return
	}
	if seasonInfo.SeasonType == model.SeasonTypeEscape && param.Special != model.ContestSpecial {
		err = xecode.Errorf(xecode.RequestErr, "赛季的比赛类型为大逃杀类，添加赛程时只能添加特殊赛程")
		return
	}

	// check sid
	season = new(model.Season)
	if err = s.dao.DB.Where("id=?", param.Sid).Where("status=?", _statusOn).First(&season).Error; err != nil {
		log.Errorc(ctx, "AddOrEditContest s.dao.DB.Where id(%d) error(%d)", param.Sid, err)
		return
	}
	// check mid
	match := new(model.Match)
	if err = s.dao.DB.Where("id=?", param.Mid).Where("status=?", _statusOn).First(&match).Error; err != nil {
		log.Errorc(ctx, "AddOrEditContest s.dao.DB.Where id(%d) error(%d)", param.Mid, err)
		return
	}
	return
}

func (s *Service) isNewBGroup(liveID int64) bool {
	if s.c.TunnelBGroup.SendNew == 1 {
		return true
	}
	return liveID == s.c.TunnelBGroup.NewCardLiveID
}

func (s *Service) addOldEventAndCard(c context.Context, param *model.Contest, season *model.Season) (err error) {
	if err = s.addEvent(c, param); err != nil {
		return err
	}
	// 更新卡片
	if err = s.upsertCard(c, param, season); err != nil {
		if xecode.Cause(err).Code() == _noAddEvent {
			// 注册事件
			if err = s.addEvent(c, param); err != nil {
				return
			}
			// 更新卡片
			if err = s.upsertCard(c, param, season); err != nil {
				err = fmt.Errorf("更新卡片出错(%+v),请重新添加赛程", err)
				return
			}
		}
	}
	return
}

func (s *Service) createBGroup(c context.Context, contest *model.Contest) (err error) {
	// 创建人群包 https://info.bilibili.co/pages/viewpage.action?pageId=184996626#id-%E4%BA%BA%E7%BE%A4%E5%8C%85service%E6%9C%8D%E5%8A%A1%E6%8E%A5%E5%8F%A3%E6%96%87%E6%A1%A3-%E4%BA%BA%E7%BE%A4%E5%8C%85%E5%88%9B%E5%BB%BA
	req := &bGroup.AddBGroupReq{
		Type:       3,
		Name:       strconv.FormatInt(contest.ID, 10),
		AppName:    "pink",
		Business:   s.c.TunnelBGroup.NewBusiness,
		Creator:    s.c.TunnelBGroup.NewBusiness,
		Definition: "{\"oid\":" + strconv.FormatInt(contest.ID, 10) + "}",
		Dimension:  1,
	}
	_, err = client.BGroupClient.AddBGroup(c, req)
	if xecode.Cause(err).Code() == _bgroupExits { //人群包已经存在
		err = nil
		return
	}
	if err != nil {
		log.Errorc(c, "AddContest addNewEventAndCard  AddBGroup contestID(%d) error(%+v)", contest.ID, err)
		err = fmt.Errorf("创建人群包出错(%+v)", err)
		return
	}
	return
}

func (s *Service) createEvent(c context.Context, contest *model.Contest, season *model.Season) (err error) {
	// 创建事件.
	if err = s.addTunnelNewEvent(c, contest, season); err != nil {
		log.Errorc(c, "AddContest addNewEventAndCard  createEvent contestID(%d) error(%+v)", contest.ID, err)
		err = fmt.Errorf("创建事件出错(%+v)", err)
		return
	}
	return
}

func (s *Service) addNewSmallCard(c context.Context, contest *model.Contest, season *model.Season) (err error) {
	if contest.Stime < time.Now().Unix() { // 如果开始时间小于当前时间不建卡
		log.Infoc(c, "AddContest addNewEventAndCard createEvent  stime less now not create card contestID(%d)", contest.ID)
		return nil // 不建卡，因为建卡也无订阅人数，不需要推卡
	}
	// 创建卡片.
	if err = s.UpsertTunnelNewCard(c, contest, season); err != nil {
		log.Errorc(c, "AddContest addNewEventAndCard  s.UpsertTunnelNewCard() contestID(%d) error(%+v)", contest.ID, err)
		err = fmt.Errorf("创建小卡出错(%+v)", err)
		return
	}
	return
}

func (s *Service) addNewEventAndGroup(c context.Context, contest *model.Contest, season *model.Season) (err error) {
	eg := errGroup.WithContext(c)
	eg.Go(func(c context.Context) error {
		// 创建人群包.
		return s.createBGroup(c, contest)
	})
	eg.Go(func(c context.Context) error {
		// 创建事件.
		return s.createEvent(c, contest, season)
	})
	if err = eg.Wait(); err != nil {
		log.Errorc(c, "AddContest addNewEventAndGroup eg.Wait() contestID(%d) error(%+v)", contest.ID, err)
		return
	}
	return
}

func (s *Service) addTunnelNewEvent(c context.Context, contest *model.Contest, season *model.Season) (err error) {
	// 注册事件 https://info.bilibili.co/pages/viewpage.action?pageId=184989531#id-%E4%B8%9C%E9%A3%8E%E7%BB%9F%E4%B8%80%E5%8D%8F%E8%AE%AE%E4%B8%9A%E5%8A%A1%E5%AF%B9%E6%8E%A5-1.1%E6%B3%A8%E5%86%8C%E4%BA%8B%E4%BB%B6
	tunnelV2Req := &tunnelV2Mdl.AddEventReq{
		BizId:    s.c.TunnelBGroup.TunnelBizID,
		UniqueId: contest.ID,
		Title:    fmt.Sprintf("赛事订阅直播提醒%d(%s-%s)", contest.ID, season.Title, contest.GameStage),
	}
	_, err = client.TunnelV2Client.AddEvent(c, tunnelV2Req)
	if xecode.Cause(err).Code() == _eventAlready { // 事件已注册不用返回错误
		err = nil
	}
	if err != nil {
		log.Errorc(c, "AddContest addTunnelNewEvent AddEvent contestID(%d) error(%+v)", contest.ID, err)
		return fmt.Errorf("小卡注册事件出错(%+v)", err)
	}
	return
}

func (s *Service) UpsertTunnelNewCard(c context.Context, contest *model.Contest, season *model.Season) (err error) {
	// 新增/更新Feed订阅卡-模板模式 https://info.bilibili.co/pages/viewpage.action?pageId=184989531#id-%E4%B8%9C%E9%A3%8E%E7%BB%9F%E4%B8%80%E5%8D%8F%E8%AE%AE%E4%B8%9A%E5%8A%A1%E5%AF%B9%E6%8E%A5-%E6%96%B0%E5%A2%9E/%E6%9B%B4%E6%96%B0Feed%E8%AE%A2%E9%98%85%E5%8D%A1-%E6%A8%A1%E6%9D%BF%E6%A8%A1%E5%BC%8F%E6%96%B0%E5%A2%9E/%E6%9B%B4%E6%96%B0Feed%E8%AE%A2%E9%98%85%E5%8D%A1-%E6%A8%A1%E6%9D%BF%E6%A8%A1%E5%BC%8F
	var (
		teamIDs          []int64
		teams            map[int64]*model.Team
		params           = make(map[string]string)
		imgPre           string
		cartInitialState = "active"
	)
	if env.DeployEnv == env.DeployEnvUat {
		imgPre = _imagePreUat
	} else {
		imgPre = _imagePre
	}
	if contest.Status == 1 {
		cartInitialState = "frozen"
	}
	teamIDs = append(teamIDs, contest.HomeID)
	teamIDs = append(teamIDs, contest.AwayID)
	teams, err = s.getTeams(teamIDs)
	if err != nil {
		log.Errorc(c, "upsertCard s.getTeams error(%+v)", err)
		return
	}
	// 主客队不存在不发送
	if len(teams) < 2 {
		return
	}
	homeTeam, ok := teams[contest.HomeID]
	if !ok {
		return
	}
	awayTeam, ok := teams[contest.AwayID]
	if !ok {
		return
	}
	// params.
	params["teamA"] = homeTeam.Title
	params["teamB"] = awayTeam.Title
	params["stage"] = contest.GameStage
	params["season"] = season.Title
	// cardContent.
	twoLink := fmt.Sprintf(s.c.TunnelBGroup.Link, contest.LiveRoom)
	cardContent := &tunnelCommon.FeedTemplateCardContent{
		TemplateId: s.c.TunnelBGroup.NewTemplateID,
		Params:     params,
		Link:       twoLink,
		Icon:       imgPre + season.Logo,
		Button: &tunnelCommon.FeedButton{
			Type: "text",
			Text: s.c.TunnelBGroup.NewCardText,
			Link: twoLink,
		},
		Trace: &tunnelCommon.FeedTrace{
			SubGoTo:  "esports",
			Param:    season.ID,
			SubParam: contest.ID,
		},
		ShowTimeTag: tunnelCommon.HideTimeTag, // 不展示时间.
	}
	userInfoStruct := struct {
		Name     string `json:"name"`
		Business string `json:"business"`
	}{
		strconv.FormatInt(contest.ID, 10),
		s.c.TunnelBGroup.NewBusiness}
	userInfo, _ := json.Marshal(userInfoStruct)
	feedTemplateReq := &tunnelV2Mdl.UpsertCardFeedTemplateReq{
		BizId:        s.c.TunnelBGroup.TunnelBizID,
		UniqueId:     contest.ID,
		CardUniqueId: contest.ID,
		TriggerType:  "time",
		StartTime:    time.Unix(contest.Stime, 0).Format("2006-01-02 15:04:05"),
		EndTime:      time.Unix(contest.Etime, 0).Format("2006-01-02 15:04:05"),
		TargetUserGroup: &tunnelCommon.TargetUserGroup{
			UserType: tunnelCommon.BGroup,
			UserInfo: string(userInfo),
		},
		CardContent:  cardContent,
		Description:  fmt.Sprintf("赛季(%d-%s),赛程(%d),阶段(%s)", season.ID, season.Title, contest.ID, contest.GameStage),
		InitialState: cartInitialState,
	}
	if err = retry.WithAttempts(c, "contest_add_card", 3, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err = client.TunnelV2Client.UpsertCardFeedTemplate(c, feedTemplateReq)
		errCode := xecode.Cause(err).Code()
		if errCode == _cardNotExists || errCode == _cardStatusErr {
			err = nil
		}
		return err
	}); err != nil {
		log.Errorc(c, "AddContest UpsertTunnelNewCard AddEvent contestID(%d) error(%+v)", contest.ID, err)
		return fmt.Errorf("新增/更新Feed订阅卡-模板模式出错(%+v)", err)
	}
	return
}

func genAutoSubUniqKey(seasonID, teamID int64) string {
	return fmt.Sprintf("%v_%v", seasonID, teamID)
}

// EditContest .
func (s *Service) EditContest(c context.Context, param *model.Contest, gids []int64) (err error) {
	season, err := s.contestParamsCheck(c, param)
	if err != nil {
		return
	}
	var (
		games                    []*model.Game
		preGidMaps, addGidMaps   []*model.GIDMap
		upGidMapAdd, upGidMapDel []int64
		pushSwitch               = param.LiveRoom > 0 && param.Special == 0
	)
	preData := new(model.Contest)
	if err = s.dao.DB.Where("id=?", param.ID).First(&preData).Error; err != nil {
		log.Errorc(c, "EditContest s.dao.DB.Where id(%d) error(%d)", param.ID, err)
		return
	}
	if err = s.dao.DB.Model(&model.Game{}).Where("status=?", _statusOn).Where("id IN (?)", gids).Find(&games).Error; err != nil {
		log.Errorc(c, "EditContest check game ids Error (%v)", err)
		return
	}
	if len(games) == 0 {
		err = xecode.Errorf(xecode.RequestErr, "关联游戏不准为空")
		return
	}
	if err = s.dao.DB.Model(&model.GIDMap{}).Where("oid=?", param.ID).Where("type=?", model.TypeContest).Find(&preGidMaps).Error; err != nil {
		log.Errorc(c, "EditContest games(%v) not found", gids)
		return
	}
	var (
		newCData []*model.ContestData
	)
	// save after GuessType also 1
	if preData.GuessType > 0 {
		param.GuessType = preData.GuessType
	}
	if param.DataType == 0 {
		param.Data = ""
		param.MatchID = 0
	}
	if param.Data != "" {
		if err = json.Unmarshal([]byte(param.Data), &newCData); err != nil {
			err = ecode.EsportsContestDataErr
			return
		}
		for _, c := range newCData {
			if c.URL != "" && c.AvCID != 0 && c.PointData == 0 {
				err = fmt.Errorf("回放链接、cid和gameID必须全部填写才能提交(%+v)", c.AvCID)
				return
			}
		}
	}
	if pushSwitch {
		param.PushSwitch = _pushOpen
	}
	gidsMap := make(map[int64]int64, len(gids))
	preGidsMap := make(map[int64]int64, len(preGidMaps))
	for _, v := range gids {
		gidsMap[v] = v
	}
	for _, v := range preGidMaps {
		preGidsMap[v.Gid] = v.Gid
		if _, ok := gidsMap[v.Gid]; ok {
			if v.IsDeleted == 1 {
				upGidMapAdd = append(upGidMapAdd, v.ID)
			}
		} else {
			upGidMapDel = append(upGidMapDel, v.ID)
		}
	}
	for _, gid := range gids {
		if _, ok := preGidsMap[gid]; !ok {
			addGidMaps = append(addGidMaps, &model.GIDMap{Type: model.TypeContest, Oid: param.ID, Gid: gid})
		}
	}
	tx := s.dao.DB.Begin()
	log.Infoc(c, "[Service][Contest][Update][DB][Transaction][Begin]")
	if err = tx.Error; err != nil {
		log.Errorc(c, "EditContest s.dao.DB.Begin error(%v)", err)
		return
	}
	if err = tx.Model(&model.Contest{}).Save(param).Error; err != nil {
		log.Error("EditContest Update(%+v) error(%v)", param, err)
		if err = tx.Rollback().Error; err != nil {
			log.Errorc(c, _rollBackErrorMsgUpdate, err)
		}
		err = xecode.Errorf(xecode.RequestErr, "编辑时保存es_contests表失败(%+v)", err)
		return
	}
	if len(upGidMapAdd) > 0 {
		if err = tx.Model(&model.GIDMap{}).Where("id IN (?)", upGidMapAdd).Updates(map[string]interface{}{"is_deleted": _notDeleted}).Error; err != nil {
			log.Error("EditContest GIDMap Add(%+v) error(%v)", upGidMapAdd, err)
			if err = tx.Rollback().Error; err != nil {
				log.Errorc(c, _rollBackErrorMsgUpdate, err)
			}
			err = xecode.Errorf(xecode.RequestErr, "编辑时保存es_gid_map表失败(%+v)", err)
			return
		}
	}
	if len(upGidMapDel) > 0 {
		if err = tx.Model(&model.GIDMap{}).Where("id IN (?)", upGidMapDel).Updates(map[string]interface{}{"is_deleted": _deleted}).Error; err != nil {
			log.Error("EditContest GIDMap Del(%+v) error(%v)", upGidMapDel, err)
			if err = tx.Rollback().Error; err != nil {
				log.Errorc(c, _rollBackErrorMsgUpdate, err)
			}
			err = xecode.Errorf(xecode.RequestErr, "编辑时保存es_gid_map表del失败(%+v)", err)
			return
		}
	}
	if len(addGidMaps) > 0 {
		sql, sqlParam := model.GidBatchAddSQL(addGidMaps)
		if err = tx.Model(&model.GIDMap{}).Exec(sql, sqlParam...).Error; err != nil {
			log.Error("EditContest GIDMap Create(%+v) error(%v)", addGidMaps, err)
			if err = tx.Rollback().Error; err != nil {
				log.Errorc(c, _rollBackErrorMsgUpdate, err)
			}
			err = xecode.Errorf(xecode.RequestErr, "编辑时保存es_gid_map表add失败(%+v)", err)
			return
		}
	}
	var (
		mapOldCData, mapNewCData    map[int64]*model.ContestData
		upCData, addCData, oldCData []*model.ContestData
		delCData                    []int64
	)
	if len(newCData) > 0 {
		// check module
		if err = s.dao.DB.Model(&model.ContestData{}).Where("cid=?", param.ID).Where("is_deleted=?", _notDeleted).Find(&oldCData).Error; err != nil {
			log.Error("EditContest s.dao.DB.Model Find (%+v) error(%v)", param.ID, err)
			return
		}
		mapOldCData = make(map[int64]*model.ContestData, len(oldCData))
		for _, v := range oldCData {
			mapOldCData[v.ID] = v
		}
		//新数据在老数据中 更新老数据。新的数据不在老数据 添加新数据
		for _, cData := range newCData {
			if _, ok := mapOldCData[cData.ID]; ok {
				upCData = append(upCData, cData)
			} else {
				addCData = append(addCData, cData)
			}
		}
		mapNewCData = make(map[int64]*model.ContestData, len(oldCData))
		for _, v := range newCData {
			mapNewCData[v.ID] = v
		}
		//老数据在新中 上面已经处理。老数据不在新数据中 删除老数据
		for _, cData := range oldCData {
			if _, ok := mapNewCData[cData.ID]; !ok {
				delCData = append(delCData, cData.ID)
			}
		}
		if len(upCData) > 0 {
			sql, sqlParam := model.BatchEditCDataSQL(upCData)
			if err = tx.Model(&model.ContestData{}).Exec(sql, sqlParam...).Error; err != nil {
				log.Error("EditContest s.dao.DB.Model tx.Model Exec(%+v) error(%v)", upCData, err)
				if err = tx.Rollback().Error; err != nil {
					log.Errorc(c, _rollBackErrorMsgUpdate, err)
				}
				err = xecode.Errorf(xecode.RequestErr, "编辑时保存es_contests_data表update失败(%+v)", err)
				return
			}
		}
		if len(delCData) > 0 {
			if err = tx.Model(&model.ContestData{}).Where("id IN (?)", delCData).Updates(map[string]interface{}{"is_deleted": _deleted}).Error; err != nil {
				log.Error("EditContest s.dao.DB.Model Updates(%+v) error(%v)", delCData, err)
				if err = tx.Rollback().Error; err != nil {
					log.Errorc(c, _rollBackErrorMsgUpdate, err)
				}
				err = xecode.Errorf(xecode.RequestErr, "编辑时保存es_contests_data表del失败(%+v)", err)
				return
			}
		}
		if len(addCData) > 0 {
			sql, sqlParam := model.BatchAddCDataSQL(param.ID, addCData)
			if err = tx.Model(&model.ContestData{}).Exec(sql, sqlParam...).Error; err != nil {
				log.Error("EditContest s.dao.DB.Model Create(%+v) error(%v)", addCData, err)
				if err = tx.Rollback().Error; err != nil {
					log.Errorc(c, _rollBackErrorMsgUpdate, err)
				}
				err = xecode.Errorf(xecode.RequestErr, "编辑时保存es_contests_data表add失败(%+v)", err)
				return
			}
		}
	} else {
		if err = tx.Model(&model.ContestData{}).Where("cid = ?", param.ID).Updates(map[string]interface{}{"is_deleted": _deleted}).Error; err != nil {
			log.Error("EditContest s.dao.DB.Model Updates(%+v) error(%v)", param.ID, err)
			if err = tx.Rollback().Error; err != nil {
				log.Errorc(c, _rollBackErrorMsgUpdate, err)
			}
			err = xecode.Errorf(xecode.RequestErr, "编辑时保存es_contests_data表del all失败(%+v)", err)
			return
		}
	}

	// 更新赛程队伍信息
	if err = s.ContestTeamsUpdate(c, param, tx); err != nil {
		log.Errorc(c, "[Service][Contest][Update][ContestTeamsUpdate][Error] err:(%+v)", err)
		message := err.Error()
		err = tx.Rollback().Error
		if err != nil {
			log.Errorc(c, "[Service][Contest][Update][DB][Transaction][RollBack][Error] err:(%+v)", err)
		}
		err = xecode.Errorf(xecode.RequestErr, message)
		return
	}

	// 更新卡片
	log.Infoc(c, "[Service][Contest][Update][PushCard][Begin]")
	if pushSwitch {
		if s.isNewBGroup(param.LiveRoom) {
			// 更新小卡.
			err = s.editNewSmallCard(c, param, season)
		} else {
			// 更新老卡.
			err = s.editOldEventAndCard(c, param, season, preData)
		}
		if err != nil {
			log.Errorc(c, "[Service][Contest][Update][EditCard][Error], err:(%+v)", err)
			if err = tx.Rollback().Error; err != nil {
				log.Errorc(c, _rollBackErrorMsgUpdate, err)
			}
			return
		}
	} else { // 取消推送，要删除小卡.
		if preData.LiveRoom > 0 && param.LiveRoom == 0 {
			if err = s.deleteNewCard(c, param); err != nil {
				log.Errorc(c, "[Service][Contest][Update][Card][Delete][Error], err:(%+v)", err)
				if err = tx.Rollback().Error; err != nil {
					log.Errorc(c, _rollBackErrorMsgUpdate, err)
				}
				err = xecode.Errorf(xecode.RequestErr, "取消推送，删除小卡失败")
				return
			}
		}
	}
	log.Infoc(c, "[Service][Contest][Update][PushCard][End]")
	if err = tx.Commit().Error; err != nil {
		log.Errorc(c, "[Service][Contest][Update][DB][Transaction][Commit][Error]EditContest tx.Commit cid(%d) error(%v)", param.ID, err)
		return
	}
	log.Infoc(c, "[Service][Contest][Update][DB][Transaction][Commit][Error]")
	s.AsyncRebuildContestTeamsCache(c, param)
	go func() {
		_ = s.ClearESportCacheByType(v1.ClearCacheType_CONTEST, []int64{param.ID})
	}()

	// 删除赛程组件缓存.
	s.cache.Do(c, func(c context.Context) {
		if e := s.ClearComponentContestCacheByGRPC(&v1.ClearComponentContestCacheRequest{SeasonID: param.Sid, ContestID: param.ID, ContestHome: param.HomeID, ContestAway: param.AwayID}); e != nil {
			log.Errorc(c, _clearCacheErrorMsg, param.Sid, param.ID, e)
		}
	})

	if param.GuessType > 0 && (preData.Stime != param.Stime || preData.Stime != param.Etime) {
		editReq := &actmdl.GuessEditReq{Business: int64(actmdl.GuessBusiness_esportsType), Oid: param.ID, Stime: param.Stime, Etime: param.Etime}
		if _, e := s.actClient.GuessEdit(c, editReq); e != nil {
			log.Error("s.actClient.GuessEdit  param(%+v) error(%v)", editReq, e)
			return
		}
	}
	if _, err := s.espClient.RefreshContestDataPageCache(c, &v1.RefreshContestDataPageCacheRequest{
		Cids: []int64{param.ID},
	}); err != nil {
		log.Errorc(c, "s.espClient.RefreshContestDataPageCache  param(%+v) error(%v)", param.ID, err)
		return err
	}

	//刷新阶段积分表和树状图
	err = s.RefreshContestSeriesExtraInfo(c, param)
	if err != nil {
		return
	}
	return
}

func (s *Service) editOldEventAndCard(c context.Context, param *model.Contest, season *model.Season, preData *model.Contest) (err error) {
	if preData.PushSwitch == _pushClose {
		if err = s.addEvent(c, param); err != nil {
			if xecode.Cause(err).Code() == _eventAlready { // 事件已注册不用返回错误
				err = nil
			} else {
				return err
			}
		}
	}
	if err = s.upsertCard(c, param, season); err != nil {
		if xecode.Cause(err).Code() == _noAddEvent {
			// 注册事件
			if err = s.addEvent(c, param); err != nil {
				return
			}
			// 更新卡片
			if err = s.upsertCard(c, param, season); err != nil {
				err = fmt.Errorf("更新卡片出错(%+v),请重新更新赛程", err)
				return
			}
		}
	}
	return
}

func (s *Service) editNewSmallCard(c context.Context, contest *model.Contest, season *model.Season) (err error) {
	// 修改卡片.
	if contest.Etime < time.Now().Unix() { // 如果结束时间小于当前时间
		log.Infoc(c, "EditContest  stime(%d) less now not create card contestID(%d) live_room(%d) ", contest.Stime, contest.ID, contest.LiveRoom)
		return s.deleteNewCard(c, contest) // 删除小卡
	} else {
		return s.UpsertTunnelNewCard(c, contest, season)
	}
}

func (s *Service) cancelNewCard(c context.Context, contest *model.Contest) (err error) {
	if err = retry.WithAttempts(c, "tunnel_card_cancel_event", 3, netutil.DefaultBackoffConfig, func(c context.Context) error {
		cancelReq := &tunnelV2Mdl.OperateCardReq{
			BizId:        s.c.TunnelBGroup.TunnelBizID,
			UniqueId:     contest.ID,
			CardUniqueId: contest.ID,
		}
		_, err = client.TunnelV2Client.CancelCard(c, cancelReq)
		errCode := xecode.Cause(err).Code()
		if errCode == _cardNotExists || errCode == _cardStatusErr {
			err = nil
		}
		return err
	}); err != nil {
		log.Errorc(c, "EditContest or ForbidContest CancelCard contestID(%d) error(%+v)", contest.ID, err)
		return fmt.Errorf("冻结小卡出错(%+v)", err)
	}
	return
}

func (s *Service) deleteNewCard(c context.Context, contest *model.Contest) (err error) {
	if err = retry.WithAttempts(c, "tunnel_card_cancel_event", 3, netutil.DefaultBackoffConfig, func(c context.Context) error {
		delReq := &tunnelV2Mdl.OperateCardReq{
			BizId:        s.c.TunnelBGroup.TunnelBizID,
			UniqueId:     contest.ID,
			CardUniqueId: contest.ID,
		}
		_, err = client.TunnelV2Client.DeleteCard(c, delReq)
		errCode := xecode.Cause(err).Code()
		if errCode == _cardNotExists || errCode == _cardStatusErr {
			err = nil
		}
		return err
	}); err != nil {
		log.Errorc(c, "EditContest deleteNewCard contestID(%d) error(%+v)", contest.ID, err)
		return fmt.Errorf("删除小卡出错(%+v)", err)
	}
	return
}

func (s *Service) activeNewCard(c context.Context, contest *model.Contest) (err error) {
	if err = retry.WithAttempts(c, "tunnel_card_active_event", 3, netutil.DefaultBackoffConfig, func(c context.Context) error {
		cancelReq := &tunnelV2Mdl.OperateCardReq{
			BizId:        s.c.TunnelBGroup.TunnelBizID,
			UniqueId:     contest.ID,
			CardUniqueId: contest.ID,
		}
		_, err = client.TunnelV2Client.ActiveCard(c, cancelReq)
		errCode := xecode.Cause(err).Code()
		if errCode == _cardNotExists || errCode == _cardStatusErr {
			err = nil
		}
		return err
	}); err != nil {
		log.Errorc(c, "ForbidContest ActiveCard contestID(%d) error(%+v)", contest.ID, err)
		return fmt.Errorf("激活小卡出错(%+v)", err)
	}
	return
}

// ForbidContest .
func (s *Service) ForbidContest(c context.Context, id int64, state int) (err error) {
	preContest := new(model.Contest)
	if err = s.dao.DB.Where("id=?", id).First(&preContest).Error; err != nil {
		log.Error("ContestForbid s.dao.DB.Where id(%d) error(%d)", id, err)
		return
	}
	tx := s.dao.DB.Begin()
	if err = tx.Error; err != nil {
		log.Error("ForbidContest s.dao.DB.Begin error(%v)", err)
		return
	}
	if err = tx.Model(&model.Contest{}).Where("id=?", id).Update(map[string]int{"status": state}).Error; err != nil {
		log.Error("ForbidContest Update(%d) error(%v)", id, err)
		err = fmt.Errorf("冻结es_contests表失败(%+v)", err)
		tx.Rollback()
		return
	}
	if state == _contestFreeze {
		// 冻结卡片.
		if err = s.cancelNewCard(c, preContest); err != nil {
			tx.Rollback()
			return
		}
	} else if state == _contestInUse {
		// 激活卡片.
		if err = s.activeNewCard(c, preContest); err != nil {
			tx.Rollback()
			return
		}
	}
	if err = tx.Commit().Error; err != nil {
		log.Error("ForbidContest tx.Commit cid(%d) error(%v)", id, err)
		return
	}

	// 删除赛程组件缓存.
	s.cache.Do(c, func(c context.Context) {
		if e := s.ClearComponentContestCacheByGRPC(&v1.ClearComponentContestCacheRequest{SeasonID: preContest.Sid, ContestID: preContest.ID, ContestHome: preContest.HomeID, ContestAway: preContest.AwayID}); e != nil {
			log.Errorc(c, _clearCacheErrorMsg, preContest.Sid, preContest.ID, e)
		}
	})

	if _, err := s.espClient.RefreshContestDataPageCache(c, &v1.RefreshContestDataPageCacheRequest{
		Cids: []int64{id},
	}); err != nil {
		log.Errorc(c, "s.espClient.RefreshContestDataPageCache  param(%+v) error(%v)", id, err)
		return err
	}

	//刷新阶段积分表和树状图
	err = s.RefreshContestSeriesExtraInfo(c, preContest)
	if err != nil {
		return
	}
	return
}

func (s *Service) addEvent(ctx context.Context, contest *model.Contest) (err error) {
	// 注册事件
	var (
		eventRly *tunnelmdl.AddEventReply
		title    = fmt.Sprintf("赛事订阅直播提醒%d", contest.ID)
	)
	eventArg := &tunnelmdl.AddEventReq{BizId: s.c.TunnelPush.TunnelBizID, UniqueId: contest.ID, Title: title, Platform: _platform}
	if eventRly, err = s.tunnelClient.AddEvent(ctx, eventArg); err != nil {
		if xecode.Cause(err).Code() == _eventAlready { // 事件已注册不用返回错误
			err = nil
		} else {
			log.Errorc(ctx, "s.tunnelClient.AddEvent error(%v)", err)
			return fmt.Errorf("注册事件出错(%+v)", err)
		}
	}
	if eventRly != nil {
		log.Errorc(ctx, "addEvent success event_id(%+v)", eventRly.EventId)
	}
	return
}

func (s *Service) upsertCard(ctx context.Context, contest *model.Contest, season *model.Season) (err error) {
	var (
		aiCard  *tunnelmdl.AiCommonCard
		params  = make(map[string]string)
		teamIDs []int64
		imgPre  string
	)
	teamIDs = append(teamIDs, contest.HomeID)
	teamIDs = append(teamIDs, contest.AwayID)
	teams, err := s.getTeams(teamIDs)
	if err != nil {
		log.Errorc(ctx, "upsertCard s.getTeams error(%+v)", err)
		return
	}
	// 主客队不存在不发送
	if len(teams) < 2 {
		return
	}
	homeTeam, ok := teams[contest.HomeID]
	if !ok {
		return
	}
	awayTeam, ok := teams[contest.AwayID]
	if !ok {
		return
	}
	params["t1"] = homeTeam.Title
	params["t2"] = awayTeam.Title
	params["content1"] = season.Title
	params["content2"] = contest.GameStage
	if env.DeployEnv == env.DeployEnvUat {
		imgPre = _imagePreUat
	} else {
		imgPre = _imagePre
	}
	aiCard = &tunnelmdl.AiCommonCard{
		TemplateId: s.c.TunnelPush.TemplateID,
		Link:       fmt.Sprintf(s.c.TunnelPush.Link, contest.LiveRoom),
		Icon:       imgPre + season.Logo,
		Params:     params,
	}
	arg := &tunnelmdl.UpsertCardReq{
		BizId:    s.c.TunnelPush.TunnelBizID,
		UniqueId: contest.ID,
		Platform: _platform,
		AiCard:   aiCard,
	}
	if _, err = s.tunnelClient.UpsertCard(ctx, arg); err != nil {
		log.Errorc(ctx, "s.tunnelClient.UpsertCard arg(%+v) error(%v)", arg, err)
		return
	} else {
		log.Errorc(ctx, "UpsertCard success id(%d)", contest.ID)
	}
	return
}

func (s *Service) getTeams(teamIDs []int64) (res map[int64]*model.Team, err error) {
	var teams []*model.Team
	if err = s.dao.DB.Model(&model.Team{}).Where("id IN (?)", unique(teamIDs)).Find(&teams).Error; err != nil {
		log.Error("getTeams Team ids(%+v) error(%v)", teamIDs, err)
		return
	}
	res = make(map[int64]*model.Team, len(teamIDs))
	for _, v := range teams {
		res[v.ID] = v
	}
	return
}

func (s *Service) MatchFix(ctx context.Context, matchID int64) (err error) {
	var rly interface{}
	conn := component.GlobalAutoSubCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()
	rkey := fmt.Sprintf("match_id_%d", matchID)
	if rly, err = conn.Do("SET", rkey, "1", "EX", s.c.Rule.MatchFixLimit, "NX"); err != nil {
		log.Error("conn.Do(GET key(%s)) error(%v)", rkey, err)
		err = fmt.Errorf("操作失败(%+v)", err)
		return
	}
	if err != nil {
		err = fmt.Errorf("数据更新失败，请稍后再操作(%+v)", err)
		return
	}
	if rly == nil {
		err = fmt.Errorf("数据在更新中，请稍等")
		return
	}
	// 因为修复需要很久
	go s.dao.FixMatchUseJob(matchID)
	return
}

func (s *Service) BatchRefreshContestDataPageCache(c context.Context, cids []int64) error {
	if len(cids) == 0 {
		return nil
	}
	if _, err := s.espClient.RefreshContestDataPageCache(c, &v1.RefreshContestDataPageCacheRequest{
		Cids: cids,
	}); err != nil {
		return err
	}
	return nil

}

// 刷新阶段积分表和树状图
func (s *Service) RefreshContestSeriesExtraInfo(ctx context.Context, contest *model.Contest) (err error) {
	cs, err := model.FindContestSeriesByID(contest.SeriesID)
	if err != nil {
		return
	}
	switch cs.Type {
	case 1:
		_, err = client.EsportsGrpcClient.RefreshSeriesPointMatchInfo(ctx, &v1.RefreshSeriesPointMatchInfoReq{
			SeriesId: cs.ID,
		})
	case 2:
		_, err = client.EsportsGrpcClient.RefreshSeriesKnockoutMatchInfo(ctx, &v1.RefreshSeriesKnockoutMatchInfoReq{
			SeriesId: cs.ID,
		})

	}
	return
}

func (s *Service) SaveContestByGrpc(ctx context.Context, param *model.Contest, gameIds []int64) (err error) {
	contest, contestData, teamIds, err := s.formatSaveContestReq(ctx, param)
	if err != nil {
		return
	}
	_, err = client.EsportsServiceClient.SaveContest(
		ctx,
		&v12.SaveContestReq{
			Contest:     contest,
			GameIds:     gameIds,
			TeamIds:     teamIds,
			ContestData: contestData,
			AdId:        param.Adid,
		},
	)
	if err != nil {
		return
	}
	s.AsyncRebuildContestTeamsCache(ctx, param)
	return
}

func (s *Service) UpdateCheck(ctx *bm.Context, param *model.Contest) (err error) {
	if param.ID == 0 {
		return
	}
	preData := new(model.Contest)
	if err = s.dao.DB.Where("id=?", param.ID).First(&preData).Error; err != nil {
		log.Errorc(ctx, "[updateCheck][Orm][Error], err:%+v", err)
		return
	}
	if preData.GuessType == 0 {
		return
	}
	// 有竞猜时无法更改主客队，或有白名单更改主客队
	if preData.HomeID != param.HomeID || preData.AwayID != param.AwayID {
		if err = s.forceUpdateContestCheck(ctx); err != nil {
			return
		}
	}
	return
}

func (s *Service) FreezeCheck(ctx *bm.Context, state int, contestId int64) (err error) {
	if contestId == 0 {
		return
	}
	preData := new(model.Contest)
	if err = s.dao.DB.Where("id=?", contestId).First(&preData).Error; err != nil {
		log.Errorc(ctx, "[updateCheck][Orm][Error], err:%+v", err)
		return
	}
	if preData.GuessType == 0 {
		return
	}
	if preData.Status == _contestInUse && state == _contestFreeze {
		if err = s.forceUpdateContestCheck(ctx); err != nil {
			return
		}
	}
	return
}

func (s *Service) forceUpdateContestCheck(ctx *bm.Context) (err error) {
	username, ok := ctx.Get("username")
	if !ok {
		err = xecode.Errorf(xecode.RequestErr, "配置了竞猜的赛程不允许修改主客队，且不可被冻结，如有需求请先流局竞猜")
		return
	}
	if len(s.c.Rule.UserNameLimit) == 0 {
		err = xecode.Errorf(xecode.RequestErr, "配置了竞猜的赛程不允许修改主客队，且不可被冻结，如有需求请先流局竞猜")
		return
	}
	log.Infoc(ctx, "[ForceUpdateContestCheck], username:%+v", username)
	for _, user := range s.c.Rule.UserNameLimit {
		if username == user {
			return
		}
	}
	err = xecode.Errorf(xecode.RequestErr, "配置了竞猜的赛程不允许修改主客队，且不可被冻结，如有需求请先流局竞猜")
	return
}

func (s *Service) formatSaveContestReq(ctx context.Context, param *model.Contest) (contest *v12.ContestModel, contestDataReq []*v12.ContestDataModel, teamIds []int64, err error) {
	contestData := make([]*model.ContestData, 0)
	if param.Data != "" {
		if err = json.Unmarshal([]byte(param.Data), &contestData); err != nil {
			err = xecode.Errorf(ecode.EsportsContestDataErr, "比赛数据不正确")
			return
		}
	}
	contestDataReq = make([]*v12.ContestDataModel, 0)
	for _, v := range contestData {
		contestDataReq = append(contestDataReq,
			&v12.ContestDataModel{
				ID:        v.ID,
				Cid:       v.CID,
				Url:       v.URL,
				PointData: v.PointData,
				AvCid:     v.AvCID,
			})
	}

	contest = &v12.ContestModel{
		ID:            param.ID,
		GameStage:     param.GameStage,
		Stime:         param.Stime,
		Etime:         param.Etime,
		HomeID:        param.HomeID,
		AwayID:        param.AwayID,
		HomeScore:     param.HomeScore,
		AwayScore:     param.AwayScore,
		LiveRoom:      param.LiveRoom,
		Aid:           param.Aid,
		Collection:    param.Collection,
		Dic:           param.Dic,
		Status:        param.Status,
		Sid:           param.Sid,
		Mid:           param.Mid,
		Special:       param.Special,
		SuccessTeam:   param.SuccessTeam,
		SpecialName:   param.SpecialName,
		SpecialTips:   param.SpecialTips,
		SpecialImage:  param.SpecialImage,
		Playback:      param.Playback,
		CollectionURL: param.CollectionURL,
		LiveURL:       param.LiveURL,
		DataType:      param.DataType,
		MatchID:       param.MatchID,
		GuessType:     param.GuessType,
		GameStage1:    param.GameStage1,
		GameStage2:    param.GameStage2,
		SeriesId:      param.SeriesID,
		PushSwitch:    param.PushSwitch,
		ActivePush:    0,
		ContestStatus: param.ContestStatus,
		ExternalID:    param.ExternalId,
	}
	teamIds, err = s.teamsStringSplit(ctx, param.TeamIds)
	if err != nil {
		err = xecode.Errorf(xecode.RequestErr, "填写的战队信息解析失败，请先检查战队填写")
		return
	}
	return
}
