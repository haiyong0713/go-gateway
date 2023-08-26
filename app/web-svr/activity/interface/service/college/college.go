package college

import (
	"context"
	"fmt"
	"go-common/library/log"
	"strings"

	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/model/college"
)

// GetAllCollege 获得所有学校
func (s *Service) GetAllCollege(c context.Context, key string) (res *college.AllCollegeReply, err error) {
	if s.allCollege == nil || len(s.allCollege) == 0 {
		return nil, ecode.ActivityGetAllCollegeErr
	}
	collegeList := s.allCollege
	res = &college.AllCollegeReply{}
	collegeBatch := make([]*college.Base, 0)
	for _, v := range collegeList {
		if len(collegeBatch) == s.c.College.SearchShowLength {
			break
		}
		if strings.Contains(v.Name, key) {
			collegeBatch = append(collegeBatch, &college.Base{
				ID:      v.ID,
				Name:    v.Name,
				Initial: v.Initial,
			})
		}
	}
	res.College = collegeBatch
	return res, nil
}

// Detail 学校详情
func (s *Service) Detail(c context.Context, collegeID int64) (res *college.DetailReply, err error) {
	res = &college.DetailReply{}
	allCollege, err := s.getAllCollege(c)
	if err != nil {
		log.Errorc(c, "s.getAllCollege nil")
		return res, ecode.ActivityGetAllCollegeErr
	}
	if _, ok := allCollege[collegeID]; !ok {
		return res, ecode.ActivityGetAllCollegeErr
	}
	collegeRedis, err := s.getCollegeByID(c, collegeID)
	if err != nil {
		return nil, ecode.ActivityCollegeGetErr
	}
	var nationRankStr, provinceRank string
	nationRankInt := collegeRedis.NationwideRank
	provinceRankInt := collegeRedis.ProvinceRank
	if nationRankInt > 100 {
		nationRankStr = fmt.Sprintf("100+")
	} else {
		nationRankStr = fmt.Sprintf("%d", nationRankInt)
	}
	if provinceRankInt > 100 {
		provinceRank = fmt.Sprintf("100+")
	} else {
		provinceRank = fmt.Sprintf("%d", provinceRankInt)
	}
	collegeDetail := &college.DetailCollege{
		Score:      collegeRedis.Score,
		Nationwide: nationRankStr,
		Province:   provinceRank,
		TabList:    collegeRedis.TabList,
		Name:       collegeRedis.Name,
		ID:         collegeRedis.ID,
	}
	res.College = collegeDetail
	return res, nil
}
