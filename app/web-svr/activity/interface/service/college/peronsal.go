package college

import (
	"context"
	"time"

	accountapi "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/model/college"

	tagrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"

	relationapi "git.bilibili.co/bapis/bapis-go/account/service/relation"

	"github.com/pkg/errors"
)

const (
	// tagTypeVideo 稿件类型
	tagTypeVideo = 3
)

// InviterCollege 获取邀请人学校
func (s *Service) InviterCollege(c context.Context, mid int64) (res *college.InviterCollegeReply, err error) {
	inviterMid, err := s.getInviter(c, mid, activityUID)
	res = &college.InviterCollegeReply{}
	collegeInfo := &college.College{}
	account := &college.Account{}
	if inviterMid > 0 {
		personal, err := s.getMidCollege(c, inviterMid)
		if err != nil {
			log.Errorc(c, "s.getMidColleg(%d)", inviterMid)
			return res, ecode.ActivityCollegeInviterCollegeErr
		}
		if personal == nil {
			return res, nil
		}
		if personal.CollegeID > 0 {
			allCollege, err := s.getAllCollege(c)
			if err != nil {
				log.Errorc(c, "s.getAllCollege nil")
				return res, ecode.ActivityGetAllCollegeErr
			}
			var (
				collegeDetail *college.Detail
				ok            bool
			)
			if collegeDetail, ok = allCollege[personal.CollegeID]; !ok {
				return res, ecode.ActivityGetAllCollegeErr
			}
			collegeInfo.Name = collegeDetail.Name
			collegeInfo.ID = collegeDetail.ID
			collegeInfo.ProvinceID = collegeDetail.ProvinceID
			res.College = collegeInfo

			midInfo, err := s.accClient.Info3(c, &accountapi.MidReq{Mid: inviterMid})
			if err != nil {
				log.Errorc(c, "s.accClient.Info3: error(%v)", err)
				err = errors.Wrapf(err, "s.accClient.Info3")
				return res, err
			}
			if midInfo == nil || midInfo.Info == nil {
				return res, ecode.ActivityCollegeMidInfoErr
			}
			account = &college.Account{
				Mid:  midInfo.Info.Mid,
				Name: midInfo.Info.Name,
				Face: midInfo.Info.Face,
				Sign: midInfo.Info.Sign,
				Sex:  midInfo.Info.Sex,
			}
			res.Account = account
		}
	}
	return res, nil
}

// Personal 个人信息
func (s *Service) Personal(c context.Context, mid int64) (res *college.PersonalReply, err error) {
	eg := errgroup.WithContext(c)
	res = &college.PersonalReply{}
	var (
		midInfo         *accountapi.InfoReply
		personal        *college.Personal
		personalCollege *college.PersonalCollege
	)
	eg.Go(func(ctx context.Context) (err error) {
		midInfo, err = s.accClient.Info3(ctx, &accountapi.MidReq{Mid: mid})
		if err != nil {
			log.Errorc(c, "s.accClient.Info3: error(%v)", err)
			err = errors.Wrapf(err, "s.accClient.Info3")
			return err
		}
		if midInfo == nil || midInfo.Info == nil {
			return ecode.ActivityCollegeMidInfoErr
		}
		return nil
	})
	// 获取绑定关系
	eg.Go(func(ctx context.Context) (err error) {
		personalCollege, err = s.getMidCollege(c, mid)
		if err != nil {
			log.Errorc(c, "s.getMidCollege(%d) error(%v)", mid, err)
		}
		return err
	})
	// 获取用户信息,脚本计算的
	eg.Go(func(ctx context.Context) (err error) {
		personal, err = s.college.GetCollegePersonal(c, mid, s.version.Version)
		if err != nil {
			log.Errorc(c, "s.college.GetCollegePersonal (%d) version(%d)", mid, s.version.Version)
		}
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return res, ecode.ActivityCollegeMidInfoErr
	}

	var (
		ok          bool
		collegeInfo *college.Detail
	)
	var collegeID int64
	if personalCollege != nil {
		collegeID = personalCollege.CollegeID
	}
	if personal != nil {

		res.Score = personal.Score
		res.Rank = personal.Rank
		res.Diff = personal.Diff
		collegeID = personal.CollegeID
	}
	if collegeID > 0 {
		allCollege, err := s.getAllCollege(c)
		if err != nil {
			log.Errorc(c, "s.getAllCollege")
			return res, ecode.ActivityGetAllCollegeErr
		}
		if collegeInfo, ok = allCollege[collegeID]; !ok {
			return res, ecode.ActivityGetAllCollegeErr
		}
		res.College = &college.College{
			ID:         collegeInfo.ID,
			Name:       collegeInfo.Name,
			ProvinceID: collegeInfo.ProvinceID,
		}
	}
	res.Account = &college.Account{
		Mid:  midInfo.Info.Mid,
		Name: midInfo.Info.Name,
		Face: midInfo.Info.Face,
		Sign: midInfo.Info.Sign,
		Sex:  midInfo.Info.Sex,
	}

	return res, nil

}

// AidIsCollege aid是否是校园活动稿件
func (s *Service) AidIsCollege(c context.Context, mid int64, aid int64) (bool, error) {
	var (
		personal *college.PersonalCollege
		tags     []int64
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		personal, err = s.getMidCollege(c, mid)
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		tagList, err := s.tagRPC.ResTag(c, &tagrpc.ResTagReq{
			Type: 3,
			Oid:  aid,
		})
		if err != nil || tagList == nil || tagList.Tags == nil || len(tagList.Tags) == 0 {
			log.Errorc(c, "s.tagRPC.ResTag error(%v)", err)
			return nil
		}
		tags = make([]int64, 0)
		for _, v := range tagList.Tags {
			tags = append(tags, v.Id)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return false, err
	}
	if personal == nil || personal.CollegeID == 0 {
		return false, nil
	}
	allCollege, err := s.getAllCollege(c)
	if err != nil {
		log.Errorc(c, "s.getAllCollege nil")
		return false, ecode.ActivityGetAllCollegeErr
	}
	var (
		collegeInfo *college.Detail
		ok          bool
	)
	if collegeInfo, ok = allCollege[personal.CollegeID]; !ok {
		return false, ecode.ActivityGetAllCollegeErr
	}
	for _, v := range tags {
		if v == collegeInfo.TagID {
			return true, nil
		}
	}
	return false, nil

}

// sendPoint
func (s *Service) sendPoint(c context.Context, points int64, timesstamp int64, mid int64, source int64, business string) error {
	data := &college.ActPlatActivityPoints{
		Points:    points,
		Mid:       mid,
		Source:    source,
		Activity:  s.c.College.MemberJoinActivity,
		Business:  business,
		Timestamp: timesstamp,
	}
	err := s.college.SendPoint(c, mid, data)
	if err != nil {
		log.Errorc(c, " s.college.actCounterIncr(%d,%v)", mid, *data)
		return err
	}
	return nil
}

// Follow ...
func (s *Service) Follow(c context.Context, mid int64) (res *college.FollowReply, err error) {
	// 判断用户是否关注过
	isFollow, err := s.college.MidIsFollow(c, mid)
	if isFollow == 1 {
		return nil, nil
	}
	// 获取用户所在学校
	personal, err := s.getMidCollege(c, mid)
	if err != nil || personal == nil {
		return nil, ecode.ActivityCollegeMidNoBindCollegeErr
	}
	collegeID := personal.CollegeID
	allCollege, err := s.getAllCollege(c)
	var (
		collegeInfo *college.Detail
		ok          bool
	)
	if err != nil {
		log.Errorc(c, "s.getAllCollege nil")
		return res, ecode.ActivityGetAllCollegeErr
	}
	if collegeInfo, ok = allCollege[collegeID]; !ok {
		return res, ecode.ActivityGetAllCollegeErr
	}
	follow := make([]int64, 0)
	if collegeInfo.MID > 0 {
		follow = append(follow, collegeInfo.MID)
	}
	if collegeInfo.RelationMid != nil && len(collegeInfo.RelationMid) > 0 {
		follow = append(follow, collegeInfo.RelationMid...)
	}
	if len(follow) == 0 {
		return nil, nil
	}
	followingReply, err := s.relationClient.BatchAddFollowingAsync(c, &relationapi.BatchAddFollowingsReq{Mid: mid, Fid: follow})
	if err != nil || followingReply == nil || followingReply.AllSucceed == false {
		log.Errorc(c, "s.relationClient.BatchAddFollowingAsync(%d,%v) error(%v) followingReply(%v)", mid, follow, err, followingReply)
		err = ecode.ActivityCollegeMidFolloweErr
		return nil, err
	}
	// redis记录+积分
	eg := errgroup.WithContext(c)
	// redis
	eg.Go(func(ctx context.Context) (err error) {
		if err = s.college.MidFollow(c, mid); err != nil {
			log.Errorc(c, "s.college.MidFollow(%d) err(%v)", mid, err)
			return err
		}
		return nil
	})
	// 加积分
	eg.Go(func(ctx context.Context) (err error) {
		if err = s.sendPoint(c, s.c.College.FollowPoint, time.Now().Unix(), mid, follow[0], college.FollowKey); err != nil {
			log.Errorc(c, "s.sendPoint(%d) (%s) err(%v)", mid, college.FollowKey, err)
			return err
		}
		log.Infoc(c, "s.sendPoint(%d) (%s) err(%v)", mid, college.FollowKey, err)
		return nil
	})

	if err := eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return nil, err
	}
	return nil, nil
}
