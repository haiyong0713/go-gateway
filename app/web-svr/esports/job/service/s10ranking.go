package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-gateway/app/web-svr/esports/job/dao"
	"gopkg.in/go-playground/validator.v9"
)

var (
	isS10RankingDataSyncProcRunning bool = false
	ctx                             context.Context
	validate                        = validator.New()
	s10RankingLiveOffLineImage      map[string]bool
)

type RankingData struct {
	Stage       int              `json:"stage"`
	ISNullPoint bool             `json:"is_null_point"`
	PointList   []*pointInfo     `json:"point_list"`
	Tree        []*roundTreeNode `json:"tree"`
	Mtime       int64            `json:"mtime"`
}

type S10RankingInterventionData struct {
	TournamentID  string `json:"tournament_id" form:"tournament_id"`
	CurrentRound  string `json:"current_round" form:"current_round"`
	FinalistRound string `json:"finalist_round" form:"finalist_round"`
}

func (s *Service) InitSyncS10RankingData() {
	ctx = trace.SimpleServerTrace(context.Background(), "syncS10RankingData")
}

func (s *Service) S10RankingDataIntervention(ctx context.Context, p *S10RankingInterventionData) error {
	return globalMemcache.Set(ctx, &memcache.Item{
		Flags:  memcache.FlagJSON,
		Key:    s.c.RankingDataWatch.InterventionCacheKey,
		Object: p,
	})
}

func (s *Service) SyncS10RankingData(roundID ...string) (ret interface{}) {
	if isS10RankingDataSyncProcRunning {
		return
	}
	// 手动忽略抢锁导致的同时执行问题
	isS10RankingDataSyncProcRunning = true
	defer func() {
		isS10RankingDataSyncProcRunning = false
		log.Infoc(ctx, "syncS10RankingData: syncS10RankingData end syncing.")
	}()
	log.Infoc(ctx, "syncS10RankingData: syncS10RankingData start syncing.")

	// 加载干预数据，加载失败不干预
	interventionData := S10RankingInterventionData{}
	if err := globalMemcache.Get(ctx, s.c.RankingDataWatch.InterventionCacheKey).Scan(&interventionData); err != nil {
		log.Errorc(ctx, "syncS10RankingData: globalMemcache.Get err[%v]", err)
	}

	// 拉取联赛信息
	tournament, err := s.getTournamentFromScore(ctx, interventionData.TournamentID)
	if err != nil {
		return
	}
	if tournament == nil {
		log.Errorc(ctx, "syncS10RankingData: fail get tournament info from score.")
		return
	}
	// 备份落地数据
	if err := s.dao.SaveThirdResourceData(ctx, fmt.Sprintf(dao.ResourceIDTournament, s.c.RankingDataWatch.TournamentID), tournament); err != nil {
		log.Errorc(ctx, "syncS10RankingData: s.dao.SaveThirdResourceData err[%v]", err)
	}
	// 拉取联赛信息
	rl, err := s.getRoundListFromScore(ctx, s.c.RankingDataWatch.TournamentID)
	if err != nil {
		return
	}
	if len(rl) == 0 {
		log.Errorc(ctx, "syncS10RankingData: fail get tournament round info from score.")
		return
	}
	// 备份落地数据
	if err := s.dao.SaveThirdResourceData(ctx, fmt.Sprintf(dao.ResourceIDRoundInfo, s.c.RankingDataWatch.TournamentID), rl); err != nil {
		log.Errorc(ctx, "syncS10RankingData: s.dao.SaveThirdResourceData err[%v]", err)
	}
	// 拉取树状图信息
	rtm := make(map[string][]*roundTreeNode)
	for _, r := range rl {
		if r.IsUseTree == "1" {
			rt, err := s.getRoundTreeFromScore(ctx, r.TournamentID, r.RoundID)
			if err != nil {
				return
			}
			if rt == nil {
				log.Errorc(ctx, "syncS10RankingData: fail get tournament tree info from score with round [%s]", r.RoundID)
				return
			}
			rtm[r.RoundID] = rt
		}
	}
	for roundID, roundTree := range rtm {
		// 备份落地数据
		if err := s.dao.SaveThirdResourceData(ctx, fmt.Sprintf(dao.ResourceIDRoundTree, s.c.RankingDataWatch.TournamentID, roundID), roundTree); err != nil {
			log.Errorc(ctx, "syncS10RankingData: s.dao.SaveThirdResourceData err[%v]", err)
		}
	}

	// 计算当前轮次
	var round *roundInfo
	if len(roundID) > 0 {
		for _, round = range rl {
			if round.RoundID == roundID[0] {
				break
			}
		}
		if round.RoundID != roundID[0] {
			round = nil
		}
	}
	if round == nil && interventionData.CurrentRound != "" {
		for _, round = range rl {
			if round.RoundID == interventionData.CurrentRound {
				break
			}
		}
		if round.RoundID != interventionData.CurrentRound {
			round = nil
		}
	}
	// 干预失败，或者未进行干预，根据数据自动计算round
	if round == nil {
		if tournament.Status == 2 {
			// 比赛已结束时score会重制轮次信息，我们仍然需要展示最终结果，所以需要忽略is_now_week，直接展示最终一个轮次信息
			round = rl[len(rl)-1]
		} else {
			// 非结束情况优先使用is_now_week，如果没有就取第一个
			for _, round = range rl {
				if round.IsNowWeek == 1 {
					break
				}
			}
			if round.IsNowWeek != 1 {
				round = rl[0]
			}
		}
	}

	pointList := round.PointList

	// 奇怪得兼容逻辑
	if round.RoundID == "409" {
		for _, r := range rl {
			if r.RoundID == "408" {
				pointList = r.PointList
				break
			}
		}
	}

	treeList := rtm[round.RoundID]

	// 计算空积分状态
	var isNullPoint = true
	if len(pointList) > 0 {
		for _, p := range pointList {
			for _, t := range p.List {
				if t.WLNum != 0 || t.WLNum != "0" || t.Win != "0" || t.Los != "0" {
					isNullPoint = false
					break
				}
			}
		}
	}

	finalistRound := interventionData.FinalistRound
	stage := 2
	if round.RoundID == finalistRound || finalistRound == "" {
		stage = 1
	}

	rankingData := RankingData{
		Stage:       stage,
		ISNullPoint: isNullPoint,
		PointList:   pointList,
		Tree:        treeList,
		Mtime:       time.Now().Unix(),
	}

	// 加载本地图片列表
	s10RankingLiveOffLineImage = s.LoadLiveOffLineImageMap()
	// 替换图片
	if err := s.s10RankingDataReplaceImg(ctx, &rankingData); err != nil {
		log.Errorc(ctx, "syncS10RankingData: s.s10RankingDataReplaceImg err[%v]", err)
		return
	}

	// 备份落地数据
	if err := s.dao.SaveThirdResourceData(ctx, fmt.Sprintf(dao.ResourceIDRankingData, round.RoundID), rankingData); err != nil {
		log.Errorc(ctx, "syncS10RankingData: s.dao.SaveThirdResourceData err[%v]", err)
	}

	// 更新round信息到mc
	if err := globalMemcache.Set(ctx, &memcache.Item{
		Flags:  memcache.FlagJSON,
		Key:    fmt.Sprint(s.c.RankingDataWatch.RoundDataCacheKeyPre, round.RoundID),
		Object: rankingData,
	}); err != nil {
		log.Errorc(ctx, "syncS10RankingData: globalMemcache.Set err[%v]", err)
		return
	}
	// 更新roundID列表到mc中
	roundList := make([]string, 0, 0)
	if err := globalMemcache.Get(ctx, s.c.RankingDataWatch.RoundIDListCacheKey).Scan(&roundList); err != nil {
		if err != memcache.ErrNotFound {
			log.Errorc(ctx, "syncS10RankingData: globalMemcache.Get err[%v]", err)
			return
		} else {
			roundList = make([]string, 0, 1)
		}
	}
	find := false
	for _, rid := range roundList {
		if rid == round.RoundID {
			find = true
			break
		}
	}
	if !find {
		roundList = append(roundList, round.RoundID)
		if err := globalMemcache.Set(ctx, &memcache.Item{
			Flags:  memcache.FlagJSON,
			Key:    s.c.RankingDataWatch.RoundIDListCacheKey,
			Object: roundList,
		}); err != nil {
			log.Errorc(ctx, "syncS10RankingData: globalMemcache.Set err[%v]", err)
			return
		}
	}

	if len(roundID) == 0 {
		// 更新当前轮次信息到mc
		if err := globalMemcache.Set(ctx, &memcache.Item{
			Flags:  memcache.FlagJSON,
			Key:    s.c.RankingDataWatch.CurrentRoundIDCacheKey,
			Object: round.RoundID,
		}); err != nil {
			log.Errorc(ctx, "syncS10RankingData: globalMemcache.Set err[%v]", err)
		}
	}

	return map[string]interface{}{
		"tournament":       tournament,
		"roundListInfo":    rl,
		"roundTree":        rtm,
		"rankingData":      rankingData,
		"roundList":        roundList,
		"currentRound":     round.RoundID,
		"interventionData": interventionData,
	}
}

func (s *Service) s10RankingDataReplaceImg(ctx context.Context, data interface{}) (err error) {
	ptr := reflect.ValueOf(data)
	// 必须是指针，否则无法设置修改值
	if ptr.Kind() != reflect.Ptr {
		return nil
	}
	// 获取指针指向的对象
	v := ptr.Elem()
	// 获取数据类型，interface的话需要多取一次真实类型
	t := v.Kind()
	if t == reflect.Interface {
		t = v.Elem().Kind()
	}
	// 判断数据类型，选择处理方案
	switch t {
	case reflect.Struct:
		{
			// 遍历结构体属性进行处理
			numFiled := v.NumField()
			for i := 0; i < numFiled; i++ {
				field := v.Field(i)
				if field.Kind() == reflect.Ptr {
					err = s.s10RankingDataReplaceImg(ctx, field.Interface())
					if err != nil {
						return
					}
				} else {
					err = s.s10RankingDataReplaceImg(ctx, field.Addr().Interface())
					if err != nil {
						return
					}
				}
			}
			break
		}
	case reflect.String:
		{
			val := v.Interface().(string)
			// 字符串字段判断是否是url
			if e := validate.Var(val, "url"); e == nil {
				// 判断是否包含score域名
				if strings.Contains(val, _scoreDomain) {
					// 上传图片更新域名
					v.SetString(s.replaceImg(ctx, val, "RankingDataImage", s10RankingLiveOffLineImage))
				}
			}
			break
		}
	case reflect.Slice:
		{
			// 遍历数组，一个一个处理
			l := v.Len()
			for i := 0; i < l; i++ {
				one := v.Index(i)
				if one.Kind() == reflect.Ptr {
					err = s.s10RankingDataReplaceImg(ctx, one.Interface())
					if err != nil {
						return
					}
				} else {
					err = s.s10RankingDataReplaceImg(ctx, one.Addr().Interface())
					if err != nil {
						return
					}
				}
			}
			break
		}
	case reflect.Map:
		{
			// 遍历map，判断类型处理
			for _, key := range v.MapKeys() {
				val := v.MapIndex(key)
				if val.Kind() == reflect.Interface {
					val = val.Elem()
				}
				switch val.Kind() {
				case reflect.Ptr:
					{
						err = s.s10RankingDataReplaceImg(ctx, val.Interface())
						if err != nil {
							return
						}
						break
					}
				default:
					// 拷贝一份指针变量，传入处理，然后重置
					nVal := reflect.New(val.Type())
					nVal.Elem().Set(val)
					err = s.s10RankingDataReplaceImg(ctx, nVal.Interface())
					if err != nil {
						return
					}
					v.SetMapIndex(key, nVal.Elem())
				}
			}
		}
	case reflect.Ptr:
		{
			return s.s10RankingDataReplaceImg(ctx, v.Interface())
		}
	}
	return
}

func (s *Service) getTournamentFromScore(ctx context.Context, tournamentID string) (t *TournamentList, err error) {
	// 拉取联赛列表
	tournamentList, err := s.getTournamentListFromScore(ctx)
	if err != nil {
		return
	}
	if tournamentID == "" {
		tournamentID = s.c.RankingDataWatch.TournamentID
	}
	for _, tournament := range tournamentList {
		if tournament.TournamentID == tournamentID {
			return tournament, nil
		}
	}
	return nil, nil
}

type TournamentList struct {
	TournamentID string `json:"tournamentID"`
	Name         string `json:"name"`
	NameEn       string `json:"name_en"`
	Image        string `json:"image"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	Status       int    `json:"status"`
}

func (s *Service) getTournamentListFromScore(ctx context.Context) (t []*TournamentList, err error) {
	res := struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Data    struct {
			List []*TournamentList `json:"list"`
		} `json:"data"`
	}{}
	err = s.getScoreData(ctx, "b/tournament_list.php", url.Values{}, &res)
	t = res.Data.List
	return
}

type pointInfo struct {
	Letter  string `json:"letter"`
	GroupID string `json:"group_id"`
	List    []*struct {
		TeamID        string      `json:"team_id"`
		TeamShortName string      `json:"team_short_name"`
		TeamImage     string      `json:"team_image"`
		Win           string      `json:"win"`
		Los           string      `json:"los"`
		Percent       int         `json:"percent"`
		WinLose       string      `json:"win_lose"`
		WLNum         interface{} `json:"w_l_num"`
		Sorting       int         `json:"sorting"`
	} `json:"list"`
}

type pointList []*pointInfo

func (p *pointList) UnmarshalJSON(data []byte) error {
	var q []*pointInfo
	json.Unmarshal(data, &q)
	*p = q
	return nil
}

type roundInfo struct {
	RoundID      string    `json:"roundID"`
	TournamentID string    `json:"tournamentID"`
	Name         string    `json:"name"`
	NameEn       string    `json:"name_en"`
	IsNowWeek    int       `json:"is_now_week"`
	IsUseTree    string    `json:"is_use_tree"`
	IsUsePoints  string    `json:"is_use_points"`
	RType        string    `json:"r_type"`
	PointList    pointList `json:"point_list"`
}

func (s *Service) getRoundListFromScore(ctx context.Context, tournamentID string) (r []*roundInfo, err error) {
	res := struct {
		Code    string       `json:"code"`
		Message string       `json:"message"`
		Data    []*roundInfo `json:"data"`
	}{}
	params := url.Values{}
	params.Add("tournamentID", tournamentID)
	err = s.getScoreData(ctx, "b/round_list.php", params, &res)
	r = res.Data
	return
}

type roundTreeNode struct {
	MatchID        string `json:"match_id"`
	TeamID         string `json:"team_id"`
	Remark         string `json:"remark"`
	Children       []*roundTreeNode
	MatchStatus    string `json:"match_status"`
	TeamShortName  string `json:"team_short_name"`
	TeamImage      string `json:"team_image"`
	TeamAID        string `json:"team_a_id"`
	TeamAShortName string `json:"team_a_short_name"`
	TeamAImage     string `json:"team_a_image"`
	TeamAWin       string `json:"team_a_win"`
	TeamBID        string `json:"team_b_id"`
	TeamBShortName string `json:"team_b_short_name"`
	TeamBImage     string `json:"team_b_image"`
	TeamBWin       string `json:"team_b_win"`
	MatchTime      string `json:"match_time"`
}

func (s *Service) getRoundTreeFromScore(ctx context.Context, tournamentID string, roundID string) (t []*roundTreeNode, err error) {
	res := struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Data []*roundTreeNode `json:"data"`
		} `json:"data"`
	}{}
	params := url.Values{}
	params.Add("tournamentID", tournamentID)
	params.Add("roundID", roundID)
	err = s.getScoreData(ctx, "b/round_tree.php", params, &res)
	t = res.Data.Data
	return
}

func (s *Service) getScoreData(ctx context.Context, api string, params url.Values, res interface{}) (err error) {
	params.Set("api_key", s.c.Score.Key)
	params.Set("api_time", strconv.FormatInt(time.Now().Unix(), 10))
	params.Set("sign", s.scoreSign(params))
	scoreURL := s.c.Score.URL + "/" + api + "?" + params.Encode()
	for i := 0; i < _retry; i++ {
		var rs []byte
		if rs, err = s.dao.ThirdGet(ctx, scoreURL); err != nil {
			log.Errorc(ctx, "getScoreData s.dao.ThirdGet error api(%s) params(%v) error(%+v)", api, params, err)
			time.Sleep(time.Second)
			continue
		}
		if err = json.Unmarshal(rs, res); err != nil {
			log.Errorc(ctx, "getScoreData json.Unmarshal error api(%s) params(%v) error(%+v)", api, params, err)
			return
		}
		return nil
	}
	return
}
