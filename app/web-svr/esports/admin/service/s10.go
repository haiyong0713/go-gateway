package service

import (
	"go-common/library/cache/memcache"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/esports/admin/component"
	"go-gateway/app/web-svr/esports/admin/model"
)

func (s *Service) RankDataInterventionSave(c *bm.Context, req *model.S10RankDataInterventionReq) (interface{}, error) {
	interventionData := new(model.S10RankingInterventionData)
	if err := component.GlobalMemcached.Get(c, s.c.RankingDataWatch.InterventionCacheKey).Scan(&interventionData); err != nil {
		log.Errorc(c, "RankDataInterventionGet: globalMemcache.Get(%s) err[%v]", s.c.RankingDataWatch.InterventionCacheKey, err)
		return nil, err
	}
	if _, ok := c.Request.Form["tournament_id"]; ok {
		interventionData.TournamentID = req.TournamentID
	}
	if _, ok := c.Request.Form["current_round"]; ok {
		interventionData.CurrentRound = req.CurrentRound
	}
	if _, ok := c.Request.Form["promote_num"]; ok {
		interventionData.PromoteNum = req.PromoteNum
	}
	if _, ok := c.Request.Form["eliminate_num"]; ok {
		interventionData.EliminateNum = req.EliminateNum
	}
	if _, ok := c.Request.Form["final_promote_num"]; ok {
		interventionData.FinalPromoteNum = req.FinalPromoteNum
	}
	if _, ok := c.Request.Form["final_eliminate_num"]; ok {
		interventionData.FinalEliminateNum = req.FinalEliminateNum
	}
	if req.FinalistRound != "" {
		interventionData.FinalistRound = req.FinalistRound
	}
	if req.UpdatePic > 0 {
		interventionData.RoundInfo = []model.S10RankingInterventionRoundInfo{
			{
				RoundID: req.FinalistRound,
				H5Pic:   req.FinalistH5Pic,
				WebPic:  req.FinalistWebPic,
			},
			{
				RoundID: req.FinalRound,
				H5Pic:   req.FinalH5Pic,
				WebPic:  req.FinalWebPic,
			},
		}
	}
	err := component.GlobalMemcached.Set(c, &memcache.Item{
		Flags:  memcache.FlagJSON,
		Key:    s.c.RankingDataWatch.InterventionCacheKey,
		Object: interventionData,
	})
	return interventionData, err
}

func (s *Service) RankDataInterventionGet(c *bm.Context) (interface{}, error) {
	interventionData := new(model.S10RankingInterventionData)
	if err := component.GlobalMemcached.Get(c, s.c.RankingDataWatch.InterventionCacheKey).Scan(&interventionData); err != nil {
		log.Errorc(c, "RankDataInterventionGet: globalMemcache.Get(%s) err[%v]", s.c.RankingDataWatch.InterventionCacheKey, err)
		return nil, err
	}
	data := map[string]interface{}{
		"raw": interventionData,
	}
	if len(interventionData.RoundInfo) == 2 {
		data["finalist_round"] = interventionData.RoundInfo[0].RoundID
		data["finalist_h_5_pic"] = interventionData.RoundInfo[0].H5Pic
		data["finalist_web_pic"] = interventionData.RoundInfo[0].WebPic
		data["final_round"] = interventionData.RoundInfo[1].RoundID
		data["final_h_5_pic"] = interventionData.RoundInfo[1].H5Pic
		data["final_web_pic"] = interventionData.RoundInfo[1].WebPic
	}
	return data, nil
}
