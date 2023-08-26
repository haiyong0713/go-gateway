package service

import (
	"context"
	"encoding/json"
	"go-common/library/log"
	likemdl "go-gateway/app/web-svr/activity/job/model/like"
	"go-gateway/pkg/idsafe/bvid"
	"strconv"
	"strings"

	actplatapi "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
)

// GameHolidaySyncCounterFilter 数据同步计数过滤
func (s *Service) GameHolidaySyncCounterFilter() {
	c := context.Background()
	s.gameHolidaySyncRunning.Lock()
	defer s.gameHolidaySyncRunning.Unlock()
	aids, err := s.gameholidayGetAids(c)
	if err != nil {
		log.Errorc(c, "s.gameholidayGetAid error(%v)", err)
		return
	}
	err = s.syncGameHolidayAidsToActPlat(c, aids)
	if err != nil {
		log.Errorc(c, "s.syncGameHolidayAidsToActPlat error(%v)", err)
		return
	}
	log.Infoc(c, "GameHolidaySyncCounterFilter success()")
	return
}

func (s *Service) gameholidayGetAids(c context.Context) ([]int64, error) {
	res, err := s.dao.SourceItem(context.Background(), s.c.GameHoliday.Vid)
	if err != nil {
		log.Errorc(c, "Failed to load SourceItem(%d,%v)", s.c.GameHoliday.Vid, err)
		return nil, err
	}
	tmp := new(likemdl.ArcListData)
	if err = json.Unmarshal(res, &tmp); err != nil {
		log.Errorc(c, "Failed to json unmarshal:%+v", err)
		return nil, err
	}
	aids := []int64{}
	if tmp != nil && tmp.List != nil {
		for _, v := range tmp.List {
			for _, val := range strings.Split(v.Data.Aids, ",") {
				if strings.HasPrefix(val, "BV") {
					avid, err := bvid.BvToAv(val)
					if err != nil {
						log.Errorc(c, "Failed to switch bv to av: %s %+v", val, err)
						continue
					}
					aids = append(aids, avid)
					continue
				}
				if avid, _ := strconv.ParseInt(val, 10, 64); avid > 0 {
					aids = append(aids, avid)
				}
			}
		}
	}

	log.Infoc(c, "gameholidayGetAids success")
	return aids, nil
}
func (s *Service) syncGameHolidayAidsToActPlat(c context.Context, aids []int64) error {
	values := []*actplatapi.FilterMemberInt{}
	expireTime := int64(600)
	for _, i := range aids {
		values = append(values, &actplatapi.FilterMemberInt{Value: i, ExpireTime: expireTime})
	}
	_, err := s.actplatClient.AddFilterMemberInt(c, &actplatapi.SetFilterMemberIntReq{
		Activity: s.c.GameHoliday.ActPlatActivity,
		Counter:  s.c.GameHoliday.ActPlatCounter,
		Filter:   "filter_aid_sources",
		Values:   values,
	})
	return err
}
