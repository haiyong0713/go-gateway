package college

import (
	"context"
	accountapi "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/model/college"
	"time"

	actPlat "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"

	"github.com/pkg/errors"
)

// Task 绑定学校
func (s *Service) Task(c context.Context, mid int64) (res *college.TaskReply, err error) {
	res = &college.TaskReply{}
	// 获取用户是否绑定学校
	personalCollege, err := s.getMidCollege(c, mid)
	if err != nil {
		log.Errorc(c, "s.getMidCollege(%d) error(%v)", mid, err)
		return nil, err
	}
	if personalCollege == nil || personalCollege.CollegeID == 0 {
		return nil, ecode.ActivityCollegeMidNotBindErr
	}
	allCollege, err := s.getAllCollege(c)
	if err != nil {
		log.Errorc(c, "s.getAllCollege nil")
		return res, ecode.ActivityGetAllCollegeErr
	}
	var (
		collegeInfo *college.Detail
		ok          bool
	)
	if collegeInfo, ok = allCollege[personalCollege.CollegeID]; !ok {
		return res, ecode.ActivityGetAllCollegeErr
	}
	taskList := make([]*college.Task, 0)
	var (
		followTask    *college.Task
		archiveTask   *college.Task
		inviteTask    *college.Task
		viewTask      *college.Task
		likeTask      *college.Task
		shareTask     *college.Task
		personalScore int64
		collegeRedis  *college.Detail
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		relationMid := make([]int64, 0)
		if collegeInfo.MID > 0 {
			relationMid = append(relationMid, collegeInfo.MID)
		}
		relationMid = append(relationMid, collegeInfo.RelationMid...)
		followTask, err = s.followTask(c, mid, relationMid)
		if err != nil {
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		archiveTask, err = s.videoupTask(c, mid, collegeInfo.Name)
		if err != nil {
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		inviteTask, err = s.inviteTask(c, mid)
		if err != nil {
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		viewTask, err = s.viewTask(c, mid)
		if err != nil {
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		likeTask, err = s.likeTask(c, mid)
		if err != nil {
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		shareTask, err = s.shareTask(c, mid)
		if err != nil {
			return err
		}
		return nil
	})
	// 获取总分
	eg.Go(func(ctx context.Context) (err error) {
		personalScore, err = s.getMidScore(c, mid)
		if err != nil {
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		collegeRedis, err = s.getCollegeByID(c, collegeInfo.ID)
		if err != nil {
			return ecode.ActivityCollegeGetErr
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return nil, err
	}
	taskList = append(taskList, followTask, archiveTask, inviteTask, viewTask, likeTask, shareTask)
	res.TaskList = taskList
	res.PersonalScore = personalScore
	res.TabList = collegeRedis.TabList
	return res, nil
}

// getMidScore 获取用户积分
func (s *Service) getMidScore(c context.Context, mid int64) (int64, error) {
	result, err := client.ActPlatClient.GetFormulaResult(c, &actPlat.GetFormulaResultReq{
		Activity: s.c.College.MidActivity,
		Formula:  s.c.College.MidFormula,
		Mid:      mid,
	})
	if result == nil && err != nil {
		log.Errorc(c, "s.actplatClient.GetFormulaResult mid(%d) err(%v)", mid, err)
		return 0, err
	}
	return result.Result, nil
}

// shareStatus ...
func (s *Service) shareStatus(c context.Context, mid int64) (shareCounter int64, err error) {
	shareCounter, err = s.getCounter(c, mid, college.ShareKey)
	if err != nil {
		log.Errorc(c, "s.getCounter(%d)key(%s) err(%v)", mid, college.ShareKey, err)
		return 0, err
	}
	return shareCounter, nil
}

// likeStatus ...
func (s *Service) likeStatus(c context.Context, mid int64) (likeCounter int64, err error) {
	likeCounter, err = s.getCounter(c, mid, college.LikeKey)
	if err != nil {
		log.Errorc(c, "s.getCounter(%d)key(%s) err(%v)", mid, college.LikeKey, err)
		return 0, err
	}
	return likeCounter, nil
}

// archiveStatus ...
func (s *Service) archiveStatus(c context.Context, mid int64) (archiveCounter int64, err error) {
	archiveCounter, err = s.getCounter(c, mid, college.VideoupKey)
	if err != nil {
		log.Errorc(c, "s.getCounter(%d) key (%s) err(%v)", mid, college.VideoupKey, err)
		return 0, err
	}
	return archiveCounter, nil
}

// viewStatus ...
func (s *Service) viewStatus(c context.Context, mid int64) (viewCounter int64, err error) {
	viewCounter, err = s.getCounter(c, mid, college.ViewKey)
	if err != nil {
		log.Errorc(c, "s.getCounter(%d) key (%s) err(%v)", mid, college.ViewKey, err)
		return 0, err
	}
	return viewCounter, nil
}

// inviteStatus ...
func (s *Service) inviteStatus(c context.Context, mid int64) (inviteNum map[string]int64, inviteCounter int64, err error) {
	eg := errgroup.WithContext(c)
	// 获取人数
	eg.Go(func(ctx context.Context) (err error) {
		if inviteNum, err = s.getMidInviterInfo(c, mid); err != nil {
			log.Errorc(c, "s.getMidInviterInfo(%d) err(%v)", mid, err)
			return err
		}
		return nil
	})
	// 是否邀请加分
	eg.Go(func(ctx context.Context) (err error) {
		if inviteCounter, err = s.getCounter(c, mid, college.InviteKey); err != nil {
			log.Errorc(c, "s.getCounter(%d) key (%s) err(%v)", mid, college.InviteKey, err)
			return err
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return nil, 0, err
	}
	return inviteNum, inviteCounter, nil
}

// followTask 关注任务
func (s *Service) followTask(c context.Context, mid int64, relationMid []int64) (res *college.Task, err error) {
	res = &college.Task{}
	account, midIsFollow, err := s.followStatus(c, mid, relationMid)
	if err != nil {
		return nil, err
	}
	res.Params = make(map[string]interface{})
	res.Params["followers"] = account
	res.Type = college.TaskFollow
	res.State = make(map[string]int64)
	res.State["finish"] = int64(midIsFollow)
	res.State["is_follower"] = int64(midIsFollow)
	return res, nil
}

// videoupTask 投稿任务
func (s *Service) videoupTask(c context.Context, mid int64, collegeName string) (res *college.Task, err error) {
	res = &college.Task{}
	archiveCounter, err := s.archiveStatus(c, mid)
	if err != nil {
		return nil, err
	}
	res.Params = make(map[string]interface{})
	res.Params["college"] = collegeName
	res.Type = college.TaskArchive
	res.State = make(map[string]int64)
	res.State["times"] = archiveCounter
	if archiveCounter > 0 {
		res.State["finish"] = 1
	} else {
		res.State["finish"] = 0
	}
	return res, nil
}

// inviteTask 邀请任务
func (s *Service) inviteTask(c context.Context, mid int64) (res *college.Task, err error) {
	res = &college.Task{}
	inviteNum, inviteCounter, err := s.inviteStatus(c, mid)
	if err != nil {
		return nil, err
	}
	res.Params = make(map[string]interface{})
	res.Type = college.TaskInvite
	res.State = make(map[string]int64)
	res.State = inviteNum
	res.State["times"] = inviteCounter
	return res, nil
}

// viewTask 观看任务
func (s *Service) viewTask(c context.Context, mid int64) (res *college.Task, err error) {
	res = &college.Task{}
	viewCounter, err := s.viewStatus(c, mid)
	if err != nil {
		return nil, err
	}
	res.Params = make(map[string]interface{})
	res.Params["times"] = s.c.College.ViewTimes
	res.Type = college.TaskView
	res.State = make(map[string]int64)
	res.State["times"] = viewCounter
	if viewCounter == int64(s.c.College.ViewTimes) {
		res.State["finish"] = 1
	} else {
		res.State["finish"] = 0
	}
	return res, nil
}

// likeTask 点赞任务
func (s *Service) likeTask(c context.Context, mid int64) (res *college.Task, err error) {
	res = &college.Task{}
	likeCounter, err := s.likeStatus(c, mid)
	if err != nil {
		return nil, err
	}
	res.Params = make(map[string]interface{})
	res.Params["times"] = s.c.College.LikeTimes
	res.Type = college.TaskLike
	res.State = make(map[string]int64)
	res.State["times"] = likeCounter
	if likeCounter == int64(s.c.College.LikeTimes) {
		res.State["finish"] = 1
	} else {
		res.State["finish"] = 0
	}
	return res, nil
}

// shareTask 分享任务
func (s *Service) shareTask(c context.Context, mid int64) (res *college.Task, err error) {
	res = &college.Task{}
	shareCounter, err := s.shareStatus(c, mid)
	if err != nil {
		return nil, err
	}
	res.Params = make(map[string]interface{})
	res.Params["times"] = s.c.College.ShareTimes
	res.Type = college.TaskShare
	res.State = make(map[string]int64)
	res.State["times"] = shareCounter
	if shareCounter == int64(s.c.College.ShareTimes) {
		res.State["finish"] = 1
	} else {
		res.State["finish"] = 0
	}
	return res, nil
}

// follewStatus ...
func (s *Service) followStatus(c context.Context, mid int64, relationMid []int64) (account []*college.Account, midIsFollow int, err error) {
	eg := errgroup.WithContext(c)
	var (
		memberInfo map[int64]*accountapi.Info
	)
	account = make([]*college.Account, 0)
	// 获取是否关注
	eg.Go(func(ctx context.Context) (err error) {
		if midIsFollow, err = s.college.MidIsFollow(c, mid); err != nil {
			log.Errorc(c, "s.college.MidFollow(%d) err(%v)", mid, err)
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		memberInfo, err = s.account.MemberInfo(c, relationMid)
		if err != nil {
			log.Errorc(c, "s.account.MemberInfo err(%v)", err)
			return err
		}
		for _, v := range relationMid {
			if relation, ok := memberInfo[v]; ok {
				account = append(account, &college.Account{
					Mid:  relation.Mid,
					Name: relation.Name,
					Face: relation.Face,
					Sign: relation.Sign,
					Sex:  relation.Sex,
				})
			}
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return nil, 0, err
	}
	return account, midIsFollow, nil
}

func (s *Service) getCounter(c context.Context, mid int64, counter string) (int64, error) {
	resp, err := client.ActPlatClient.GetCounterRes(c, &actPlat.GetCounterResReq{
		Counter:  counter,
		Activity: s.c.College.MemberJoinActivity,
		Mid:      mid,
		Time:     time.Now().Unix(),
	})
	if err != nil {
		log.Errorc(c, "s.actPlatClient.GetCounterRes(%v) counter(%s) error(%v)", mid, counter, err)
		err = errors.Wrapf(err, "s.actPlatClient.GetCounterRes %d counter(%s)", mid, counter)
		return 0, err
	}
	if resp == nil || len(resp.CounterList) != 1 {
		log.Errorc(c, "s.actPlatClient.GetCounterRes(%v) error(%v)", mid, err)
		err = errors.Wrapf(err, "s.actPlatClient.GetCounterRes %d return nil", mid)
		return 0, err
	}
	res := resp.CounterList[0]
	return res.Val, nil
}
