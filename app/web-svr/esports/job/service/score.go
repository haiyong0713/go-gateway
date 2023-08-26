package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/job/model"
)

const (
	_retry       = 3
	_scoreImgage = "https://img.scoregg.com/"
	_imageExits  = 0
	_bfsImage    = "//i0.hdslb.com/bfs/esport/"
	_scoreDomain = "img.scoregg.com"
)

func (s *Service) writeScoreInfo(p *model.ParamScore, tp string) {
	var (
		c           = context.Background()
		res, tmpRes []*model.LdInfo
		count       int
		err         error
	)
	if res, count, err = s.scoreInfo(p, tp); err != nil {
		return
	}
	for i := 2; i <= count; i++ {
		time.Sleep(time.Millisecond * 200)
		p.Pn = strconv.Itoa(i)
		if tmpRes, _, err = s.scoreInfo(p, tp); err != nil {
			return
		}
		if len(tmpRes) > 0 {
			res = append(res, tmpRes...)
		}
	}
	for _, info := range res {
		info.ImageURL = s.BfsProxy(c, info.ImageURL)
		info.Name = strings.Replace(info.Name, "'", "\\'", -1)
	}
	switch tp {
	case _scorelolChampions:
		if err = s.dao.AddLolCham(c, res); err != nil {
			log.Error("writeScoreInfo tp(%s)  s.dao.AddLolCham error(%+v)", tp, err)
		}
	}
}

func (s *Service) scoreInfo(p *model.ParamScore, tp string) (res []*model.LdInfo, count int, err error) {
	var (
		bs     []byte
		infoRs struct {
			Data struct {
				Count int `json:"count"`
				List  []*model.LdInfo
			} `json:"data"`
		}
	)
	if bs, err = s.score(p, tp); err != nil {
		return
	}
	if err = json.Unmarshal(bs, &infoRs); err != nil {
		log.Error("loadScorePages tp(%s) SerieID(%d) pn(1) json.Unmarshal rs(%s)  error(%+v)", tp, p.SerieID, string(bs), err)
	}
	if infoRs.Data.Count > 0 {
		res = append(res, infoRs.Data.List...)
	}
	count = infoRs.Data.Count/_perPage + 1
	return
}

func (s *Service) loadScorePages(p *model.ParamScore, tp string) (res []int64) {
	var (
		count  int
		tmpRes []int64
		err    error
	)
	if res, count, err = s.scoreIDs(p, tp); err != nil {
		return
	}
	for i := 2; i <= count; i++ {
		p.Pn = strconv.Itoa(i)
		if tmpRes, _, err = s.scoreIDs(p, tp); err != nil {
			return
		}
		if len(tmpRes) > 0 {
			res = append(res, tmpRes...)
		}
	}
	return
}

func (s *Service) scoreIDs(p *model.ParamScore, tp string) (res []int64, count int, err error) {
	var (
		bs     []byte
		oidsRs struct {
			Data struct {
				Count int `json:"count"`
				List  []struct {
					ID int64 `json:"id"`
				} `json:"list"`
			} `json:"data"`
		}
	)
	if bs, err = s.score(p, tp); err != nil {
		return
	}
	if err = json.Unmarshal(bs, &oidsRs); err != nil {
		log.Error("scoreIDs tp(%s) SerieID(%d) pn(1) json.Unmarshal rs(%s)  error(%+v)", tp, p.SerieID, string(bs), err)
		return
	}
	if oidsRs.Data.Count == 0 {
		return
	}
	for _, list := range oidsRs.Data.List {
		res = append(res, list.ID)
	}
	count = oidsRs.Data.Count/_perPage + 1
	return
}

func (s *Service) score(p *model.ParamScore, tp string) (rs []byte, err error) {
	params := url.Values{}
	switch tp {
	case _scoreLolGame, _scoreLiveBattleList:
		params.Set("match_id", strconv.FormatInt(p.MatchID, 10))
	case _scoreLolSeriesPlayers, _scoreLolSeriesTeams:
		params.Set("serie_id", strconv.FormatInt(p.SerieID, 10))
		params.Set("page", p.Pn)
		params.Set("per_page", strconv.FormatInt(p.Ps, 10))
	case _scoreLolPlayerStats:
		params.Set("serie_id", strconv.FormatInt(p.SerieID, 10))
		params.Set("player_id", strconv.FormatInt(p.OriginID, 10))
	case _scoreLolTeamStats:
		params.Set("serie_id", strconv.FormatInt(p.SerieID, 10))
		params.Set("team_id", strconv.FormatInt(p.OriginID, 10))
	case _scorelolChampions:
		params.Set("page", p.Pn)
		params.Set("per_page", strconv.FormatInt(p.Ps, 10))
	case _scoreLiveBattleInfo:
		params.Set("battle_string", p.BattleString)
	case _scoreOfflineTeam:
		params.Set("tournamentID", s.c.Score.OfflineTournamentID)
	}
	params.Set("api_key", s.c.Score.Key)
	params.Set("api_time", strconv.FormatInt(time.Now().Unix(), 10))
	params.Set("sign", s.scoreSign(params))
	scoreURL := s.c.Score.URL + "/" + tp + "?" + params.Encode()
	// 查看score 返回结果.
	log.Warn("score matchid(%d) url(%s)", p.MatchID, scoreURL)
	for i := 0; i < _retry; i++ {
		if rs, err = s.dao.ThirdGet(context.Background(), scoreURL); err != nil {
			time.Sleep(time.Second)
			continue
		}
		break
	}
	if err != nil {
		log.Error("score url(%s) body(%s) error(%v)", scoreURL, string(rs), err)
	}
	return
}

func (s *Service) scoreSign(params url.Values) string {
	var (
		dataParams string
		mapParams  map[string]string
		keys       []string
	)
	mapParams = make(map[string]string)
	for k, v := range params {
		if len(v[0]) > 0 {
			keys = append(keys, k)
			mapParams[k] = v[0]
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		dataParams = dataParams + k + mapParams[k]
	}
	dataParams += s.c.Score.Secret
	cipherStr := s.scoreMd5Key(dataParams)
	return s.scoreMd5Key(cipherStr)
}

func (s *Service) scoreMd5Key(p string) string {
	hasher := md5.New()
	hasher.Write([]byte(p))
	return strings.ToUpper(hex.EncodeToString(hasher.Sum(nil)))
}

func (s *Service) tournamentTeamIDs(ctx context.Context, tournamentID int64) (teamIDs []int64, err error) {
	res := struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Data    struct {
			List []struct {
				TeamID string `json:"teamID"`
			} `json:"list"`
		} `json:"data"`
	}{}
	params := url.Values{}
	params.Add("tournamentID", strconv.FormatInt(tournamentID, 10))
	err = s.getScoreData(ctx, _scoreOfflineTeam, params, &res)
	if err != nil {
		log.Errorc(ctx, "tournamentTeamIDs _scoreLolTeams s.getScoreData  tournamentID(%d) error(%+v)", tournamentID, err)
		return
	}
	for _, team := range res.Data.List {
		if intTeam, e := strconv.ParseInt(team.TeamID, 10, 64); e != nil {
			log.Errorc(ctx, "tournamentTeamIDs strconv.ParseInt tournamentID(%d) error(%+v)", tournamentID, err)
		} else {
			teamIDs = append(teamIDs, intTeam)
		}
	}
	return
}

func (s *Service) scoreTeamInfo(ctx context.Context, tournamentID, teamID int64) (res *model.ScoreTeamInfo, err error) {
	params := url.Values{}
	params.Add("tournamentID", strconv.FormatInt(tournamentID, 10))
	params.Add("teamID", strconv.FormatInt(teamID, 10))
	err = s.getScoreData(ctx, _scoreLolTeamInfo, params, &res)
	if err != nil {
		log.Errorc(ctx, "scoreTeamInfo _scoreLolTeamInfo s.getScoreData  tournamentID(%d) teamID(%d) error(%+v)", tournamentID, teamID, err)
		return
	}
	if res == nil || res.Code != _codeSuccess {
		log.Errorc(ctx, "scoreTeamInfo _scoreLolTeamInfo s.getScoreData  tournamentID(%d) teamID(%d) res(%+v)", tournamentID, teamID, res)
		err = ecode.NothingFound
	}
	return
}

// 大数据-选手数据榜.
func (s *Service) getLolDataPlayer(ctx context.Context, tournamentID int64) (res *model.LolDataPlayer, err error) {
	//https://www.yuque.com/books/share/7ac500dc-8cf1-4523-bdd7-2e43f290ec18/qlxhap
	params := url.Values{}
	params.Add("tournamentID", strconv.FormatInt(tournamentID, 10))
	err = s.getScoreData(ctx, _scoreLolDataPlayer, params, &res)
	if err != nil {
		log.Errorc(ctx, "getLolDataPlayer _scoreLolDataPlayer s.getScoreData  tournamentID(%d) error(%+v)", tournamentID, err)
		return
	}
	if res == nil || res.Code != _codeSuccess {
		log.Errorc(ctx, "getLolDataPlayer _scoreLolDataPlayer s.getScoreData  tournamentID(%d) res(%+v)", tournamentID, res)
		err = ecode.NothingFound
	}
	return
}

// 大数据-英雄数据榜-无位置.
func (s *Service) getLolDataHero2(ctx context.Context, tournamentID int64) (res *model.LolDataHero2, err error) {
	//https://www.yuque.com/books/share/7ac500dc-8cf1-4523-bdd7-2e43f290ec18/gy4gps
	params := url.Values{}
	params.Add("tournamentID", strconv.FormatInt(tournamentID, 10))
	err = s.getScoreData(ctx, _scoreLolDataHero2, params, &res)
	if err != nil {
		log.Errorc(ctx, "getLolDataHero2 _scoreLolDataHero2 s.getScoreData  tournamentID(%d) error(%+v)", tournamentID, err)
		return
	}
	if res == nil || res.Code != _codeSuccess {
		log.Errorc(ctx, "getLolDataHero2 _scoreLolDataHero2 s.getScoreData  tournamentID(%d) res(%+v)", tournamentID, res)
		err = ecode.NothingFound
	}
	return
}

// 大数据-战队与选手离线数据.
func (s *Service) getOffLineTeamAndPlayer(ctx context.Context, tournamentID int64) (res []*model.OffLineTeam, err error) {
	//https://www.yuque.com/books/share/7ac500dc-8cf1-4523-bdd7-2e43f290ec18/qhxcgr
	var offTeamRs struct {
		Code string
		Data struct {
			List []*model.OffLineTeam
		} `json:"data"`
	}
	params := url.Values{}
	params.Add("tournamentID", strconv.FormatInt(tournamentID, 10))
	err = s.getScoreData(ctx, _scoreOfflineTeam, params, &offTeamRs)
	if err != nil {
		log.Errorc(ctx, "getOffLineTeamAndPlayer _scoreOfflineTeam s.getScoreData  tournamentID(%d) error(%+v)", tournamentID, err)
		return
	}
	if offTeamRs.Code != _codeSuccess {
		log.Errorc(ctx, "getOffLineTeamAndPlayer _scoreOfflineTeam s.getScoreData  tournamentID(%d) res(%+v)", tournamentID, res)
		err = ecode.NothingFound
		return
	}
	res = offTeamRs.Data.List
	return
}
