package college

import (
	"context"
	"fmt"

	accountapi "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/model/college"
	"go-gateway/app/web-svr/activity/interface/model/rank"
	"go-gateway/pkg/idsafe/bvid"

	relationapi "git.bilibili.co/bapis/bapis-go/account/service/relation"
)

// ProvinceRank 获取省排行
func (s *Service) ProvinceRank(c context.Context, provinceID int64, ps, pn int) (res *college.ProvinceCollegeRankReply, err error) {
	res = &college.ProvinceCollegeRankReply{}
	res.CollegeList = make([]*college.RankReply, 0)
	// 获取省信息
	allProvince, err := s.getAllProvince(c)
	if err != nil {
		return nil, ecode.ActivityGetProvinceErr
	}
	var (
		province *college.Province
		ok       bool
		start    = ((pn - 1) * ps)
		end      = start + ps
	)
	if province, ok = allProvince[provinceID]; !ok {
		return res, ecode.ActivityGetProvinceErr
	}
	res.Province = province
	// 获取省排行
	provinceRank, err := s.getProvinceRank(c, provinceID)
	if err != nil {
		log.Errorc(c, "s.getProvinceRank(%d) err(%v)", provinceID, err)
		return nil, ecode.ActivityGetProvinceRankErr
	}

	page := &college.Page{}
	page.Total = len(provinceRank)
	page.Num = pn
	page.Size = ps
	res.Page = page

	if len(provinceRank)-1 < start {
		return res, nil
	}
	if end > len(provinceRank)-1 {
		end = len(provinceRank)
	}
	rank := provinceRank[start:end]
	rankReply, err := s.rankCollegeDataReply(c, rank)
	if err != nil {
		log.Errorc(c, " s.rankCollegeDataReply err(%v)", err)
		return nil, err
	}
	res.CollegeList = rankReply
	res.Time = s.version.Time
	return res, nil
}

// rankDataReply 排行记录整理
func (s *Service) rankCollegeDataReply(c context.Context, rankRedis []*rank.Redis) ([]*college.RankReply, error) {
	if rankRedis == nil {
		return nil, nil
	}
	var (
		archive map[int64]*api.Arc
		err     error
	)
	res := make([]*college.RankReply, 0)
	aids := make([]int64, 0)
	for _, v := range rankRedis {
		aids = append(aids, v.Aids...)
	}
	archive, err = s.archive.ArchiveInfo(c, aids)
	if err != nil {
		log.Errorc(c, "s.archive.ArchiveInfo err(%v)", err)
		return nil, err
	}
	for _, v := range rankRedis {
		archiveBatch := make([]*college.Archive, 0)
		for _, aid := range v.Aids {
			if arc, ok := archive[aid]; ok {
				author := &college.Author{
					Mid:  arc.Author.Mid,
					Face: arc.Author.Face,
					Name: arc.Author.Name,
				}
				var bvidStr string
				if bvidStr, err = bvid.AvToBv(arc.Aid); err != nil || bvidStr == "" {
					continue
				}
				archiveBatch = append(archiveBatch, &college.Archive{
					TypeName: arc.TypeName,
					Title:    arc.Title,
					Desc:     arc.Desc,
					Duration: arc.Duration,
					Pic:      arc.Pic,
					View:     arc.Stat.View,
					Author:   author,
					Bvid:     bvidStr,
					Danmaku:  arc.Stat.Danmaku,
					Ctime:    arc.PubDate,
				})
			}
		}
		allCollege, err := s.getAllCollege(c)
		if err != nil {
			log.Errorc(c, "s.getAllCollege(c), err(%v)", err)
			continue
		}
		if collegeDetail, ok := allCollege[v.Mid]; ok {
			res = append(res, &college.RankReply{
				ID:       v.Mid,
				Score:    v.Score,
				Archive:  archiveBatch,
				Province: collegeDetail.Province,
				Name:     collegeDetail.Name,
				Initial:  collegeDetail.Initial,
			})
		}
	}
	return res, nil
}

// NationwideRank 获取省排行
func (s *Service) NationwideRank(c context.Context, ps, pn int) (res *college.ProvinceCollegeRankReply, err error) {
	res = &college.ProvinceCollegeRankReply{}
	res.CollegeList = make([]*college.RankReply, 0)
	var (
		start = ((pn - 1) * ps)
		end   = start + ps
	)

	// 获取全国排行
	nationwideRank, err := s.getNationwideCollegeRank(c)
	if err != nil {
		log.Errorc(c, "s.getNationwideCollegeRank() err(%v)", err)
		return nil, ecode.ActivityGetNationwideRankErr
	}

	page := &college.Page{}
	page.Total = len(nationwideRank)
	page.Num = pn
	page.Size = ps
	res.Page = page
	if len(nationwideRank)-1 < start {
		return res, nil
	}
	if end > len(nationwideRank)-1 {
		end = len(nationwideRank)
	}
	rank := nationwideRank[start:end]
	rankReply, err := s.rankCollegeDataReply(c, rank)
	if err != nil {
		log.Errorc(c, " s.rankCollegeDataReply err(%v)", err)
		return nil, err
	}
	res.CollegeList = rankReply
	res.Time = s.version.Time
	return res, nil
}

func (s *Service) collegeMidRankKey(c context.Context, collegeID int64, version int) string {
	return fmt.Sprintf("college_mid:%d:%d", version, collegeID)
}

// CollegePeopleRank 校内用户排行
func (s *Service) CollegePeopleRank(c context.Context, collegeID int64, ps, pn int) (res *college.PeopleRankReply, err error) {
	res = &college.PeopleRankReply{}
	res.MemberList = make([]*college.MemberReply, 0)
	var (
		start = ((pn - 1) * ps)
		end   = start + ps
		ok    bool
	)
	// 获取省信息
	allCollege, err := s.getAllCollege(c)
	if err != nil {
		return nil, ecode.ActivityGetAllCollegeErr
	}
	if _, ok = allCollege[collegeID]; !ok {
		return res, ecode.ActivityGetAllCollegeErr
	}
	redis, err := s.rank.GetRank(c, s.collegeMidRankKey(c, collegeID, s.version.Version))
	if err != nil {
		log.Errorc(c, "s.rank.GetRank collegeId(%d) error(%v)", collegeID, err)
		return
	}
	page := &college.Page{}
	page.Total = len(redis)
	page.Num = pn
	page.Size = ps
	res.Page = page
	if len(redis)-1 < start {
		return res, nil
	}
	if end > len(redis)-1 {
		end = len(redis)
	}
	rank := redis[start:end]
	rankReply, err := s.rankDataReply(c, rank)
	if err != nil {
		log.Errorc(c, " s.rankDataReply err(%v)", err)
		return nil, err
	}
	res.MemberList = rankReply
	res.Time = s.version.Time
	return res, nil
}

// rankDataReply 排行记录整理
func (s *Service) rankDataReply(c context.Context, rankRedis []*rank.Redis) ([]*college.MemberReply, error) {
	if rankRedis == nil {
		return nil, nil
	}
	var (
		memberInfo map[int64]*accountapi.Info
		err        error
	)
	res := make([]*college.MemberReply, 0)
	mids := make([]int64, 0)
	for _, v := range rankRedis {
		mids = append(mids, v.Mid)
	}
	memberInfo, err = s.account.MemberInfo(c, mids)
	if err != nil {
		log.Errorc(c, "s.account.MemberInfo err(%v)", err)
		return nil, err
	}
	for _, v := range rankRedis {
		if member, ok := memberInfo[v.Mid]; ok {
			memberReply := &college.MemberReply{}
			account := &college.Account{
				Mid:  member.Mid,
				Name: member.Name,
				Face: member.Face,
				Sign: member.Sign,
				Sex:  member.Sex,
			}
			memberReply.Account = account
			memberReply.Score = v.Score
			res = append(res, memberReply)
		}
	}
	return res, nil
}

// ArchiveList ...
func (s *Service) ArchiveList(c context.Context, mid int64, collegeID int64, tab, ps, pn int) (res *college.ArchiveListReply, err error) {
	allCollege, err := s.getAllCollege(c)
	if err != nil {
		log.Errorc(c, "s.getAllCollege nil")
		return res, ecode.ActivityGetAllCollegeErr
	}
	if _, ok := allCollege[collegeID]; !ok {
		return res, ecode.ActivityGetAllCollegeErr
	}
	aids, err := s.college.GetArchiveTabArchive(c, collegeID, tab)
	if err != nil {
		log.Errorc(c, "s.college.GetArchiveTabArchive collegeId(%d) error(%v)", collegeID, err)
		return
	}
	res = &college.ArchiveListReply{}
	var (
		start = ((pn - 1) * ps)
		end   = start + ps
	)
	page := &college.Page{}
	page.Total = len(aids)
	page.Num = pn
	page.Size = ps
	res.Page = page
	if len(aids)-1 < start {
		return res, nil
	}
	if end > len(aids)-1 {
		end = len(aids)
	}
	rank := aids[start:end]
	var (
		archive map[int64]*api.Arc
	)
	mids := make([]int64, 0)
	archive, err = s.archive.ArchiveInfo(c, aids)
	if err != nil {
		log.Errorc(c, "s.archive.ArchiveInfo err(%v)", err)
		return nil, err
	}
	archiveInfoBatch := make([]*college.ArchiveInfo, 0)
	for _, aid := range rank {
		if arc, ok := archive[aid]; ok {
			author := &college.Author{
				Mid:  arc.Author.Mid,
				Face: arc.Author.Face,
				Name: arc.Author.Name,
			}
			mids = append(mids, arc.Author.Mid)
			var bvidStr string
			if bvidStr, err = bvid.AvToBv(arc.Aid); err != nil || bvidStr == "" {
				continue
			}
			archiveInfoBatch = append(archiveInfoBatch, &college.ArchiveInfo{Archive: &college.Archive{
				TypeName: arc.TypeName,
				Title:    arc.Title,
				Desc:     arc.Desc,
				Duration: arc.Duration,
				Pic:      arc.Pic,
				View:     arc.Stat.View,
				Author:   author,
				Bvid:     bvidStr,
				Ctime:    arc.PubDate,
				Danmaku:  arc.Stat.Danmaku,
			}})
		}

	}
	// 获取用户是否关注
	res.ArchiveInfo = archiveInfoBatch
	if mid > 0 {
		mapFollower, err := s.midIsFollow(c, mid, mids)
		if err != nil {
			log.Errorc(c, "s.midIsFollow")
			return res, nil
		}
		for k, v := range archiveInfoBatch {
			author := v.Archive.Author.Mid
			if isFollower, ok := mapFollower.FollowingMap[author]; ok && isFollower != nil {
				if isFollower.Attribute < 128 {
					archiveInfoBatch[k].IsFollower = 1
				}
			}
		}
		res.ArchiveInfo = archiveInfoBatch
	}
	return res, nil
}

// midIsFollow ...
func (s *Service) midIsFollow(c context.Context, mid int64, followers []int64) (*relationapi.FollowingMapReply, error) {
	followingMapReply, err := s.relationClient.Relations(c, &relationapi.RelationsReq{Mid: mid, Fid: followers})
	if err != nil || followingMapReply == nil {
		log.Error("s.relationClient.Relations(%d,%v) error(%v)", mid, followers, err)
		return nil, err
	}
	return followingMapReply, nil
}
