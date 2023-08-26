package college

import (
	"context"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/model/college"
	mdlrank "go-gateway/app/web-svr/activity/interface/model/rank"
)

const (
	// allCollegeLimit db取大学数
	allCollegeLimit = 500
)

// loadVersion 获取最新版本号
func (s *Service) loadVersion() {
	c := context.Background()
	if s.version == nil {
		s.version = &college.Version{}
	}
	version, err := s.college.GetCollegeVersion(c)
	if err != nil {
		log.Errorc(c, "s.college.GetCollegeVersion(c) err(%v)", err)
		return
	}
	if version.Version > s.version.Version {
		s.version = version
	}
	log.Infoc(c, "loadVersion version(%v)", version)

}

// UpdateVersion 更新版本号
func (s *Service) UpdateVersion() (error, error) {
	s.loadVersion()
	return nil, nil
}

// getCollegeByID 获取学校信息
func (s *Service) getCollegeByID(c context.Context, collegeID int64) (res *college.Detail, err error) {
	res, err = s.college.CacheGetCollegeDetail(c, collegeID, s.version.Version)
	if err != nil {
		allCollege, err := s.getAllCollege(c)
		if err != nil {
			log.Errorc(c, "s.getAllCollege err(%v)", err)
			return nil, err
		}
		if collegeInfo, ok := allCollege[collegeID]; ok {
			return collegeInfo, nil
		}
	}
	return
}

func (s *Service) collegeCollegeRankKey(c context.Context, rankType int64, version int) string {
	return fmt.Sprintf("college:%d:%d", version, rankType)
}

// loadAllProvince 加载所有省
func (s *Service) loadAllProvince() {
	c := context.Background()
	province, err := s.college.GetAllProvince(c)
	if err != nil {
		log.Errorc(c, "s.college.GetAllProvince err(%v)", err)
		return
	}
	s.allProvince = province
}

// getAllProvince 获取所有省份
func (s *Service) getAllProvince(c context.Context) (map[int64]*college.Province, error) {
	if s.allProvince == nil {
		return nil, ecode.ActivityGetProvinceErr
	}
	var mapAllProvince = make(map[int64]*college.Province)

	for _, v := range s.allProvince {
		mapAllProvince[v.ID] = v
	}
	return mapAllProvince, nil
}

func (s *Service) getAllCollege(c context.Context) (map[int64]*college.Detail, error) {
	if s.allCollege == nil {
		return nil, ecode.ActivityGetAllCollegeErr
	}
	var mapAllCollege = make(map[int64]*college.Detail)
	for _, v := range s.allCollege {
		mapAllCollege[v.ID] = v
	}
	return mapAllCollege, nil
}

// loadProvinceRank 加载省排行
func (s *Service) loadProvinceRank() {
	if s.provinceLastVersion >= s.version.Version {
		return
	}
	c := context.Background()
	province, err := s.getAllProvince(c)
	if err != nil {
		return
	}
	// 获取所有省份的排行
	for _, v := range province {
		redis, err := s.rank.GetRank(c, s.collegeCollegeRankKey(c, v.ID, s.version.Version))
		if err != nil {
			log.Errorc(c, "s.rank.GetRank provinceID(%d) error(%v)", v.ID, err)
			return
		}
		s.provinceCollegeRank.Set(v.ID, redis)
	}
	s.provinceLastVersion = s.version.Version
	log.Infoc(c, "loadProvinceRank() success version(%d)", s.provinceLastVersion)
}

func (s *Service) loadNationwideRank() {
	if s.nationLastVersion >= s.version.Version {
		return
	}
	c := context.Background()
	// 获取所有省份的排行
	redis, err := s.rank.GetRank(c, s.collegeCollegeRankKey(c, college.NationWideRankType, s.version.Version))
	if err != nil {
		log.Errorc(c, "s.rank.GetRank provinceID(%d) error(%v)", college.NationWideRankType, err)
		return
	}
	s.nationWideCollegeRank = redis
	s.nationLastVersion = s.version.Version
	log.Infoc(c, "loadNationwideRank() success version(%d)", s.nationLastVersion)
}

// loadAllCollege 获取所有学校
func (s *Service) loadAllCollege() {
	c := context.Background()
	var offset int64
	collegeList := make([]*college.Detail, 0)
	for {
		college, err := s.college.GetAllCollege(c, offset, allCollegeLimit)
		if err != nil {
			log.Errorc(c, "s.college.GetAllCollege error(%v)", err)
			return
		}
		if len(college) > 0 {
			collegeList = append(collegeList, college...)
		}
		if len(college) < allCollegeLimit {
			break
		}
		offset += allCollegeLimit
	}
	s.allCollege = collegeList
	log.Infoc(c, "loadAllCollege() success")
}

// getProvinceRank 获取省排行
func (s *Service) getProvinceRank(c context.Context, provinceID int64) (data []*mdlrank.Redis, err error) {
	if s.provinceCollegeRank == nil {
		return nil, ecode.ActivityGetProvinceRankErr
	}
	data = s.provinceCollegeRank.Get(provinceID)
	if data == nil {
		err = ecode.ActivityGetProvinceRankErr
	}
	return
}

// getNationwideCollegeRank ...
func (s *Service) getNationwideCollegeRank(c context.Context) (data []*mdlrank.Redis, err error) {
	if s.nationWideCollegeRank == nil {
		return nil, ecode.ActivityGetNationwideRankErr
	}
	return s.nationWideCollegeRank, nil
}
