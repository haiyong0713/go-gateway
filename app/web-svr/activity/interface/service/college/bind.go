package college

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/model/college"
	"time"

	passportinfoapi "git.bilibili.co/bapis/bapis-go/passport/service/user"
	actplatapi "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	"github.com/pkg/errors"
)

const (
	activityUID = "2020college"
)

// Bind 绑定学校
func (s *Service) Bind(c context.Context, mid int64, buvid string, collegeID int64, year int) (res *college.BindReply, err error) {
	// 判断大学id是否正确
	allCollege, err := s.getAllCollege(c)
	if err != nil {
		log.Errorc(c, "s.getAllCollege nil")
		return res, ecode.ActivityGetAllCollegeErr
	}
	var (
		collegeInfo *college.Detail
		ok          bool
	)
	if collegeInfo, ok = allCollege[collegeID]; !ok {
		return res, ecode.ActivityGetAllCollegeErr
	}
	// 检查是否保定过
	res = &college.BindReply{}
	college := &college.College{}
	res.College = college
	checkBind, err := s.getMidCollege(c, mid)
	if err != nil {
		return res, ecode.ActivityGetBindCollegeErr
	}
	if checkBind != nil && checkBind.CollegeID != 0 {
		res.College.ID = checkBind.CollegeID
		res.College.Name = checkBind.CollegeName
		return res, ecode.ActivityGetBindCollegeErr
	}

	// 获取邀请人信息
	inviterMid, err := s.getInviter(c, mid, activityUID)
	if err != nil {
		log.Errorc(c, "s.getInviter(%d) error(%v)", mid, err)
	}

	err = s.doBind(c, inviterMid, mid, buvid, collegeID, collegeInfo.Name, year)
	if err != nil {
		log.Errorc(c, "s.doBind err(%v)", err)
		return res, ecode.ActivityBindCollegeErr
	}
	return res, nil
}

func (s *Service) telHash(tel string) string {
	hash := md5.New()
	hash.Write([]byte(tel))
	if s.c.Invite.TelSalt != "" {
		hash.Write([]byte(s.c.Invite.TelSalt))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func (s *Service) getInviter(c context.Context, mid int64, activityUID string) (int64, error) {
	// 获取邀请人的hash手机号
	hashReply, err := s.passportClient.UserTelHash(c, &passportinfoapi.UserTelHashReq{Mid: mid})
	if hashReply == nil || err != nil || hashReply.TelHash == "" {
		log.Errorc(c, "s.passportClient.UserTelHash(%d) hashReply(%v) error(%v)", mid, hashReply, err)
		return 0, err
	}
	log.Infoc(c, "getInviter mid (%d) telHash (%s)", mid, hashReply.TelHash)
	// 获取邀请人信息
	inviter, err := s.invite.GetMidBindInviter(c, hashReply.TelHash, activityUID)
	if err != nil {
		log.Errorc(c, "s.college.GetMidBindInviter mid(%d) err(%v)", mid, err)
	}
	if err == nil {
		return inviter, nil
	}
	inviterInfo, err := s.invite.GetInviteByTelHash(c, mid, activityUID, hashReply.TelHash)

	if err != nil {
		log.Errorc(c, "s.college.GetInviteByTelHash mid(%d) err(%v)", mid, err)
		return 0, ecode.ActivityInviterGetErr
	}
	err = s.invite.SetMidBindInviter(c, hashReply.TelHash, inviterInfo.Mid, activityUID)
	if err != nil {
		log.Errorc(c, "s.SetMidBindInviter telhash(%s)  mid(%d) err(%v)", hashReply.TelHash, mid, err)
	}
	return inviterInfo.Mid, nil
}

func (s *Service) doBind(c context.Context, inviter int64, mid int64, buvid string, collegeID int64, name string, year int) error {
	// 获取当前用户是否三新
	var infoReply *passportinfoapi.CheckFreshUserReply
	infoReply, err := s.passportClient.CheckFreshUser(c, &passportinfoapi.CheckFreshUserReq{Mid: mid, Buvid: buvid, Period: s.c.College.FreshMidPeriod})
	if err != nil || infoReply == nil {
		log.Errorc(c, "s.passportClient.CheckFreshUser(%d) infoReply(%v) error(%v)", mid, infoReply, err)
		err = errors.Wrapf(err, "s.passportClient.CheckFreshUser %d", mid)
		return err
	}
	var midType int
	if infoReply.IsNew {
		midType = college.MidTypeIsNew
	} else {
		midType = college.MidTypeIsOld

	}
	_, err = s.college.MidBindCollege(c, mid, midType, collegeID, inviter, year)
	if err != nil {
		log.Errorc(c, "s.college.MidBindCollege(%d,%d,%d,%d,%d) err(%v)", mid, midType, collegeID, inviter, year, err)
		return err
	}
	eg := errgroup.WithContext(c)
	// 通知用户 加入filter
	eg.Go(func(ctx context.Context) (err error) {
		_, err = client.ActPlatClient.AddFilterMemberInt(c, &actplatapi.SetFilterMemberIntReq{
			Activity: s.c.College.MemberJoinActivity,
			Counter:  s.c.College.MemberJoinCounter,
			Filter:   s.c.College.MemberJoinFilter,
			// todo 过期时间
			Values: []*actplatapi.FilterMemberInt{{Value: mid}},
		})
		if err != nil {
			log.Errorc(c, "s.actplatClient.AddFilterMemberInt mid(%d) err(%v)", mid, err)
			return err
		}
		return
	})
	// 给邀请人加积分
	if inviter > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			if err = s.sendPoint(c, s.c.College.InviterPoint, time.Now().Unix(), inviter, mid, college.InviteKey); err != nil {
				log.Errorc(c, "s.sendPoint(%d) (%s) err(%v)", mid, college.InviteKey, err)
				return err
			}
			log.Infoc(c, "s.sendPoint(%d) (%s) err(%v)", mid, college.InviteKey, err)
			return
		})
	}
	// todo 设置绑定缓存
	eg.Go(func(ctx context.Context) (err error) {
		midBind := &college.PersonalCollege{MID: mid, CollegeID: collegeID, CollegeName: name}
		if err = s.college.CacheSetMidCollege(c, mid, midBind); err != nil {
			log.Errorc(c, "s.college.CacheSetMidColleg(%d) err(%v)", mid, err)
		}
		return nil
	})
	// 删除邀请人缓存
	eg.Go(func(ctx context.Context) (err error) {
		if err = s.college.DelCacheMidInviter(c, inviter); err != nil {
			log.Errorc(c, "s.college.DelCacheMidInviter(%d) err(%v)", inviter, err)
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return err
	}
	return nil
}

// getMidCollege 检查是否绑定过
func (s *Service) getMidCollege(c context.Context, mid int64) (*college.PersonalCollege, error) {
	midBind, err := s.college.CacheGetMidCollege(c, mid)
	if err != nil {
		log.Errorc(c, "s.college.CacheGetMidCollege mid(%d) err(%v)", mid, err)
	}
	if err == nil && midBind != nil {
		return midBind, nil
	}
	midBind, err = s.college.GetMidBindCollege(c, mid)
	if err != nil {
		log.Errorc(c, "s.college.GetMidBindCollege mid(%d) err(%v)", mid, err)
		return nil, err
	}
	err = s.college.CacheSetMidCollege(c, mid, midBind)
	if err != nil {
		log.Errorc(c, "s.college.CacheSetMidCollege mid(%d) err(%v)", mid, err)
	}
	return midBind, nil

}

// getMidInviterInfo 获取用户邀请信息
func (s *Service) getMidInviterInfo(c context.Context, mid int64) (map[string]int64, error) {
	inviterInfo, err := s.college.CacheMidInviter(c, mid)
	res := make(map[string]int64)
	res[college.MidTypeIsNewStr] = 0
	res[college.MidTypeIsAllStr] = 0
	if err == nil && len(inviterInfo) == 2 {
		res = inviterInfo
		return res, nil
	}
	var (
		new int
		old int
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		new, err = s.college.CountInviterNum(c, mid, college.MidTypeIsNew)
		if err != nil {
			log.Errorc(c, "s.CountInviterNum mid(%d) err(%v)", mid, err)
			return err
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		old, err = s.college.CountInviterNum(c, mid, college.MidTypeIsOld)
		if err != nil {
			log.Errorc(c, "s.CountInviterNum mid(%d) err(%v)", mid, err)
			return err
		}
		return
	})
	if err = eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return nil, err
	}
	res[college.MidTypeIsNewStr] = int64(new)
	res[college.MidTypeIsAllStr] = int64(new + old)
	err = s.college.AddCacheMidInviter(c, mid, res)
	if err != nil {
		log.Errorc(c, "s.college.AddCacheMidInviter (%d) err(%v)", mid, err)
	}
	return res, nil
}
