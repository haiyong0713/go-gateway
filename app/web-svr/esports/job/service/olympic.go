package service

import (
	"context"
	"encoding/json"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/esports/job/component"
	ftpModel "go-gateway/app/web-svr/esports/job/model"
	"os"
	"strings"
	"time"
)

const (
	_olympicContestBase = 10000000
)

func (s *Service) OlympicContestLocal(ctx context.Context, file *os.File) (err error) {
	if s.c.OlympicConf == nil {
		return
	}
	if s.c.OlympicConf != nil && !s.c.OlympicConf.Open {
		return
	}
	resp, err := component.ActivityClient.GetOlympicQueryConfig(ctx, &api.GetOlympicQueryConfigReq{
		SkipCache: false,
	})
	if err != nil {
		log.Errorc(ctx, "[OlympicContestLocal][GetOlympicQueryConfig][Error], err:%+v")
		return
	}
	configs := resp.QueryConfigs
	if len(configs) == 0 {
		return
	}
	queryMapping := make(map[int64][]string)
	for _, config := range configs {
		if config.State != 1 {
			continue
		}
		if queryMapping[config.ContestId] == nil {
			queryMapping[config.ContestId] = make([]string, 0)
		}
		queryMapping[config.ContestId] = append(queryMapping[config.ContestId], config.QueryWord)
	}
	if len(queryMapping) == 0 {
		return
	}
	var olympicFile *os.File
	if s.c.OlympicConf.PreContest {
		if olympicFile, err = os.OpenFile(s.c.OlympicConf.OlympicLocalFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0766); err != nil {
			log.Error("OlympicContestLocal os.OpenFile path(%s) error(%v)", s.c.OlympicConf.OlympicLocalFile, err)
			return
		}
		defer func() {
			_ = olympicFile.Close()
		}()
	} else {
		olympicFile = file
	}
	for contestId, queryWords := range queryMapping {
		ftpTmp, errG := s.getOlympicContest(contestId, queryWords)
		if errG != nil || ftpTmp == nil {
			continue
		}
		contestsByte, errG := json.Marshal(ftpTmp)
		if errG != nil {
			log.Error("OlympicMatchesFtp json.Marshal err(%v)", errG)
			continue
		}
		if _, err = olympicFile.WriteString(string(contestsByte) + "\n"); err != nil {
			log.Error("OlympicMatchesFtp WriteString err(%v)", err)
			err = nil
		}
	}
	return
}

func (s *Service) getOlympicContest(contestId int64, queryWords []string) (ftpContest *ftpModel.FtpEsports, err error) {
	if contestId == 0 {
		return
	}
	resp, err := component.ActivityClient.GetOlympicContestDetail(ctx, &api.GetOlympicContestDetailReq{
		Id: contestId,
	})
	if err != nil {
		log.Errorc(ctx, "[getOlympicContest][GetOlympicContestDetail][Error], err:%+v", err)
		return
	}
	if resp.Id == 0 {
		return
	}
	ftpContest = &ftpModel.FtpEsports{
		ID:          resp.Id + _olympicContestBase,
		AliasSearch: strings.Join(queryWords, ","),
		StartTime:   time.Now().Unix(),
		EndTime:     time.Now().Unix(),
		Ischeck:     1,
		Status:      0,
		Title:       resp.SeasonTitle,
		Pubtime:     time.Now().Unix(),
		Lastupdate:  time.Now().Unix(),
		URL:         resp.SeasonUrl,
		SqURL:       resp.BottomUrl,
	}
	return
}

func (s *Service) OlympicFtpUpload(ctx context.Context) (err error) {

	schConfig := s.c.OlympicConf
	if schConfig == nil {
		return
	}
	if !schConfig.Open {
		return
	}
	if err = s.dao.FileMd5(schConfig.OlympicLocalFile, schConfig.OlympicLocalMD5File); err != nil {
		log.Errorc(ctx, "FtpUpload FileMd5 LocalFile(%s) LocalMD5File(%s) error(%v)", schConfig.OlympicLocalFile, schConfig.OlympicLocalMD5File, err)
		return
	}
	//upload file
	if err = s.dao.UploadFile(schConfig.OlympicLocalFile, schConfig.OlympicRemotePath, schConfig.OlympicRemoteFileName); err != nil {
		log.Errorc(ctx, "FtpUpload FileMd5 LocalFile(%s) LocalMD5File(%s) error(%v)", schConfig.OlympicLocalFile, schConfig.OlympicLocalMD5File, err)
		return
	}
	//upload md5 file
	if err = s.dao.UploadFile(schConfig.OlympicLocalMD5File, schConfig.OlympicRemotePath, schConfig.OlympicRemoteMD5FileName); err != nil {
		log.Errorc(ctx, "FtpUpload FileMd5 LocalFile(%s) LocalMD5File(%s) error(%v)", schConfig.OlympicLocalFile, schConfig.OlympicLocalMD5File, err)
		return
	}
	return nil
}
