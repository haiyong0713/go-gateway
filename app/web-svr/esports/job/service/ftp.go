package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"go-common/library/log"
	ftpModel "go-gateway/app/web-svr/esports/job/model"
)

const (
	//ContLimit is used for getting ugc value 50 records every time
	_ContLimit = 50
	_hasBlock  = 1
	//_hasCheck 字段过滤
	_hasCheck     = 1
	_officialTeam = 1
	_jumpURL      = "https://www.bilibili.com/h5/match/data/schedule?time=%d&sids=%d&cid=%d&navhide=1"
)

// FtpUpload .
func (s *Service) FtpUpload() (err error) {
	schConfig := s.c.Search
	if err = s.dao.FileMd5(schConfig.LocalFile, schConfig.LocalMD5File); err != nil {
		log.Error("FtpUpload FileMd5 LocalFile(%s) LocalMD5File(%s) error(%v)", schConfig.LocalFile, schConfig.LocalMD5File, err)
		return
	}
	//upload file
	if err = s.dao.UploadFile(schConfig.LocalFile, schConfig.RemotePath, schConfig.RemoteFileName); err != nil {
		log.Error("FtpUpload FileMd5 LocalFile(%s) LocalMD5File(%s) error(%v)", schConfig.LocalFile, schConfig.LocalMD5File, err)
		return
	}
	//upload md5 file
	if err = s.dao.UploadFile(schConfig.LocalMD5File, schConfig.RemotePath, schConfig.RemoteMD5FileName); err != nil {
		log.Error("FtpUpload FileMd5 LocalFile(%s) LocalMD5File(%s) error(%v)", schConfig.LocalFile, schConfig.LocalMD5File, err)
		return
	}
	// 单独的奥运赛程ftp上传
	_ = s.OlympicFtpUpload(ctx)
	return nil
}

// LoadSeasons .
func (s *Service) LoadSeasons() (err error) {
	var (
		cnt       int
		cycle     int
		seasons   []*ftpModel.FtpSeason
		id        int64
		mapSeason map[int64]*ftpModel.FtpSeason
	)
	c := context.Background()
	if cnt, err = s.dao.SeasonCount(c); err != nil {
		log.Error("LoadSeasons error(%v)", err)
		return
	}
	cycle = cnt / _ContLimit
	if cnt%_ContLimit != 0 {
		cycle = cnt/_ContLimit + 1
	}
	seasons = make([]*ftpModel.FtpSeason, 0, cnt)
	//cycle get sql value
	for i := 0; i < cycle; i++ {
		var (
			seasonsTmp []*ftpModel.FtpSeason
		)
		if i == 0 {
			id = 0
		} else {
			id = seasons[len(seasons)-1].ID
		}
		if seasonsTmp, err = s.dao.Season(c, id, _ContLimit); err != nil {
			log.Error("LoadSeasons Season err(%v)", err)
			return
		}
		seasons = append(seasons, seasonsTmp...)
	}
	mapSeason = make(map[int64]*ftpModel.FtpSeason, len(seasons))
	for _, v := range seasons {
		mapSeason[v.ID] = v
	}
	s.mapSeason = mapSeason
	return
}

// LoadMatchs .
func (s *Service) LoadMatchs() (err error) {
	var (
		cnt       int
		cycle     int
		matchs    []*ftpModel.FtpMatchs
		id        int64
		mapMatchs map[int64]*ftpModel.FtpMatchs
	)
	c := context.Background()
	if cnt, err = s.dao.FtpMatchsCount(c); err != nil {
		log.Error("LoadMatchs error(%v)", err)
		return
	}
	cycle = cnt / _ContLimit
	if cnt%_ContLimit != 0 {
		cycle = cnt/_ContLimit + 1
	}
	matchs = make([]*ftpModel.FtpMatchs, 0, cnt)
	//cycle get sql value
	for i := 0; i < cycle; i++ {
		var (
			matchsTmp []*ftpModel.FtpMatchs
		)
		if i == 0 {
			id = 0
		} else {
			id = matchs[len(matchs)-1].ID
		}
		if matchsTmp, err = s.dao.FtpMatchs(c, id, _ContLimit); err != nil {
			log.Error("LoadMatchs err(%v)", err)
			return
		}
		matchs = append(matchs, matchsTmp...)
	}
	mapMatchs = make(map[int64]*ftpModel.FtpMatchs, len(matchs))
	for _, v := range matchs {
		mapMatchs[v.ID] = v
	}
	s.mapMatchs = mapMatchs
	return
}

// LoadTeams .
func (s *Service) LoadTeams() (err error) {
	var (
		cnt      int
		cycle    int
		teams    []*ftpModel.FtpTeams
		id       int64
		mapTeams map[int64]*ftpModel.FtpTeams
	)
	c := context.Background()
	if cnt, err = s.dao.FtpTeamsCount(c); err != nil {
		log.Error("LoadTeams error(%v)", err)
		return
	}
	cycle = cnt / _ContLimit
	if cnt%_ContLimit != 0 {
		cycle = cnt/_ContLimit + 1
	}
	teams = make([]*ftpModel.FtpTeams, 0, cnt)
	//cycle get sql value
	for i := 0; i < cycle; i++ {
		var (
			teamsTmp []*ftpModel.FtpTeams
		)
		if i == 0 {
			id = 0
		} else {
			id = teams[len(teams)-1].ID
		}
		if teamsTmp, err = s.dao.FtpTeams(c, id, _ContLimit); err != nil {
			log.Error("LoadTeams err(%v)", err)
			return
		}
		teams = append(teams, teamsTmp...)
	}
	mapTeams = make(map[int64]*ftpModel.FtpTeams, len(teams))
	for _, v := range teams {
		mapTeams[v.ID] = v
	}
	s.mapTeam = mapTeams
	return
}

// LoadContests .
func (s *Service) LoadContests() (err error) {
	var (
		cnt   int
		cycle int
		id    int64
		file  *os.File
	)
	if file, err = os.OpenFile(s.c.Search.LocalFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0766); err != nil {
		log.Error("LoadContests os.OpenFile path(%s) error(%v)", s.c.Search.LocalFile, err)
		return
	}
	defer file.Close()
	c := context.Background()
	if cnt, err = s.dao.FtpContestsCount(c); err != nil {
		log.Error("LoadContests FtpContestsCount error(%v)", err)
		return
	}
	cycle = cnt / _ContLimit
	if cnt%_ContLimit != 0 {
		cycle = cnt/_ContLimit + 1
	}
	var (
		contestsTmp  []*ftpModel.FtpContest
		contestsByte []byte
	)
	//cycle get sql value
	for i := 0; i < cycle; i++ {
		if i == 0 {
			id = 0
		} else {
			id = contestsTmp[len(contestsTmp)-1].ID
		}
		if contestsTmp, err = s.dao.FtpContests(c, id, _ContLimit); err != nil {
			log.Error("LoadContests FtpContests err(%v)", err)
			return
		}
		ftpTmp := s.FtpTransfer(contestsTmp)
		for _, v := range ftpTmp {
			if contestsByte, err = json.Marshal(v); err != nil {
				log.Error("LoadContests json.Marshal err(%v)", err)
				return
			}
			if _, err = file.WriteString(string(contestsByte) + "\n"); err != nil {
				log.Error("LoadContests WriteString err(%v)", err)
				err = nil
			}
		}
	}
	// 奥运赛程同步
	_ = s.OlympicContestLocal(ctx, file)
	return
}

// FtpTransfer .
func (s *Service) FtpTransfer(contests []*ftpModel.FtpContest) (res []*ftpModel.FtpEsports) {
	for _, v := range contests {
		var (
			homeTeam                   *ftpModel.FtpTeams
			awayTeam                   *ftpModel.FtpTeams
			season                     *ftpModel.FtpSeason
			match                      *ftpModel.FtpMatchs
			ok                         bool
			homeTeamName, awayTeamName string
		)
		tmp := &ftpModel.FtpEsports{
			ID:        v.ID,
			StartTime: v.Stime,
			EndTime:   v.Etime,
			Status:    v.Special,
			Spid:      v.HomeID,
			SalerID:   v.AwayID,
			TpID:      v.Sid,
		}
		if homeTeam, ok = s.mapTeam[v.HomeID]; ok {
			//主队 对应后台战队简称
			tmp.IBrandname = homeTeam.Title
			//别名 对应后台战队全称
			tmp.SBrandname = homeTeam.SubTitle
			//有一个战队不为空 就可以为官方战队
			if homeTeam.TeamType == _officialTeam {
				tmp.IsBlock = _hasBlock
			}
		} else {
			log.Error("FtpTransfer mapTeam Miss homeTeam(%d) cid(%d)", v.HomeID, v.ID)
		}
		if awayTeam, ok = s.mapTeam[v.AwayID]; ok {
			//客队 对应后台战队简称
			tmp.ICategory = awayTeam.Title
			//别名 对应后台战队全称
			tmp.SCategory = awayTeam.SubTitle
			if awayTeam.TeamType == _officialTeam {
				tmp.IsBlock = _hasBlock
			}
		} else {
			log.Error("FtpTransfer mapTeam Miss awayTeam(%d) cid(%d)", v.AwayID, v.ID)
		}
		if season, ok = s.mapSeason[v.Sid]; ok {
			tmp.Title = season.Title
			tmp.Pubtime = season.Stime
			tmp.Lastupdate = season.Etime
		} else {
			log.Error("FtpTransfer mapSeason Miss season(%d)", v.Sid)
		}
		if match, ok = s.mapMatchs[v.Mid]; ok {
			//赛事别名 赛事全称+赛事简称
			tmp.AliasTitle = match.SubTitle
			if match.Title != "" {
				tmp.AliasTitle = match.SubTitle + "," + match.Title
			}
		} else {
			log.Error("FtpTransfer mapMatchs Miss match(%d) cid(%d)", v.Mid, v.ID)
		}
		if homeTeam != nil {
			homeTeamName = homeTeam.Title
		}
		if awayTeam != nil {
			awayTeamName = awayTeam.Title
		}
		if homeTeamName == "" || awayTeamName == "" {
			continue
		}
		//赛程名称 {主队名} VS {客队名}
		tmp.Gname = homeTeamName + "VS" + awayTeamName
		//1 主队名VS客队名
		AliasSearch1 := homeTeamName + "VS" + awayTeamName
		//2 客队名VS主队名
		AliasSearch2 := awayTeamName + "VS" + homeTeamName
		//3 主队名 客队名（中间加空格）
		AliasSearch3 := homeTeamName + " " + awayTeamName
		//4 客队名 主队（中间加空格）
		AliasSearch4 := awayTeamName + " " + homeTeamName
		//5 主队名 VS 客队名（VS两边加空格）
		AliasSearch5 := homeTeamName + " VS " + awayTeamName
		//6 客队名 VS 主队名（VS两边加空格）
		AliasSearch6 := awayTeamName + " VS " + homeTeamName
		tmp.AliasSearch = fmt.Sprintf("%s,%s,%s,%s,%s,%s", AliasSearch1, AliasSearch2, AliasSearch3, AliasSearch4, AliasSearch5, AliasSearch6)
		tmp.URL = fmt.Sprintf(_jumpURL, time.Unix(v.Stime, 0).UnixNano()/1e6, v.Sid, v.ID)
		tmp.SqURL = tmp.URL
		//无直播地址 并且 无比赛回放地址 并且 无集锦 地址(且) 则为0 这里取反 因为默认为0
		if v.Playback != "" || v.CollectionURL != "" || v.LiveRoom != 0 {
			tmp.Ischeck = _hasCheck
		}
		res = append(res, tmp)
	}
	return
}
