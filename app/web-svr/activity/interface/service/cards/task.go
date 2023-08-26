package cards

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/client"
	actmdl "go-gateway/app/web-svr/activity/interface/model/actplat"
	cardsmdl "go-gateway/app/web-svr/activity/interface/model/cards"
	riskmdl "go-gateway/app/web-svr/activity/interface/model/risk"

	accountapi "git.bilibili.co/bapis/bapis-go/account/service"

	relationapi "git.bilibili.co/bapis/bapis-go/account/service/relation"
)

const (
	// ActivityUID 风险uid
	ActivityUID = "activity_ny_card"
)

// ClickTask 点击任务
func (s *Service) ClickTask(ctx context.Context, mid int64, business string, risk *riskmdl.Base, mobiApp string) (err error) {
	// 给邀请人完成任务
	var isAllow bool
	for _, v := range s.c.Cards.AllowClickFinish {
		if business == v {
			isAllow = true
		}
	}
	if !isAllow {
		return ecode.SpringFestivalTaskErr
	}
	if business == signBusiness {
		spRisk := &riskmdl.Sf21Sign{
			Base:        *risk,
			Mid:         mid,
			ActivityUID: s.c.Cards.ActivityUID,
			MobiApp:     mobiApp,
		}
		_, err := s.risk(ctx, mid, riskmdl.ActionCardsSign, spRisk, spRisk.EsTime)
		if err != nil {
			log.Errorc(ctx, "s.risk mid(%d) sign err(%v)", mid, err)
		}
	}
	return s.actSend(ctx, mid, mid, business)
}

// actSend
func (s *Service) actSend(ctx context.Context, mid int64, source int64, business string) (err error) {
	timeStamp := time.Now().Unix()
	activityPoints := &actmdl.ActivityPoints{
		Timestamp: timeStamp,
		Mid:       mid,
		Source:    source,
		Activity:  s.c.Cards.Activity,
		Business:  business,
	}
	err = s.actDao.Send(ctx, mid, activityPoints)
	if err != nil {
		log.Errorc(ctx, "s.actDao.Send end error data(%+v) err(%v) ", activityPoints, err)
		content := fmt.Sprintf("timestamp:%d,mid:%d,source:%d,Activity:%s,Business:%s,err:%v", timeStamp, mid, source, s.c.Cards.Activity, business, err)
		s.cache.SyncDo(ctx, func(ctx context.Context) {
			if s.c.Cards.NeedRetry == 0 {
				err2 := s.wechatdao.SendWeChat(ctx, s.c.Wechat.PublicKey, "[青你3集卡任务完成推送失败]", content, "zhangtinghua")
				if err2 != nil {
					log.Errorc(ctx, " s.wechatdao.SendWeChat(%v)", err2)
				}
				return
			}
			for i := 0; i < retry; i++ {
				err = s.actDao.Send(ctx, mid, activityPoints)
				if err == nil {
					return
				}
				log.Errorc(ctx, "retry info:%v", err)
				time.Sleep(timeSleep)
			}
			if err != nil {
				log.Errorc(ctx, "s.actDao.Send end error data(%+v) err(%v) ", activityPoints, err)
				content := fmt.Sprintf("timestamp:%d,mid:%d,source:%d,Activity:%s,Business:%s,err:%v", timeStamp, mid, source, s.c.Cards.Activity, business, err)
				err2 := s.wechatdao.SendWeChat(ctx, s.c.Wechat.PublicKey, "[青你3集卡任务完成推送失败 重试失败]", content, "zhangtinghua")
				if err2 != nil {
					log.Errorc(ctx, " s.wechatdao.SendWeChat(%v)", err2)
				}
			}

		})
	}
	return nil
}

func (s *Service) initTask(ctx context.Context) error {
	task, err := s.dao.RawTaskList(ctx, s.c.Cards.ActivityID)
	if err != nil {
		log.Errorc(ctx, "s.dao.RawTaskList err(%v)", err)
		return err
	}
	s.allTask = task
	return nil
}

func today() string {
	return time.Now().Format("2006-01-02")
}

func (s *Service) todayFollowMid() []*cardsmdl.FollowMid {
	newMids := make([]*cardsmdl.FollowMid, 0)
	if s.followMid != nil {
		return s.followMid
	}
	return newMids
}

func (s *Service) todayOgvLink(ctx context.Context) string {
	today := today()
	if s.ogvLink != nil {
		for _, v := range s.ogvLink {
			if v.Date != "" {
				if v.Date == today {
					return v.Link
				}

			}
		}
	}
	return ""
}

func (s *Service) initFollowMid(ctx context.Context) error {
	today := today()
	mids, err := s.dao.FollowMid(ctx, s.c.Cards.FollowMidUri, time.Now().Unix())
	if err != nil {
		log.Errorc(ctx, "s.dao.FollowMid err(%v)", err)
		return err
	}
	newMids := make([]*cardsmdl.FollowMid, 0)
	midAlready := make(map[int64]struct{})
	if mids != nil {
		for _, v := range mids {
			if v.Date != "" {
				date := strings.Split(v.Date, ",")
				if len(date) > 0 {
					for _, day := range date {
						if day == today {
							if _, ok := midAlready[v.Mid]; ok {
								continue
							}
							newMids = append(newMids, v)
							midAlready[v.Mid] = struct{}{}
						}
					}
				}
			}
		}
	}
	s.followMid = newMids
	return nil
}

func (s *Service) initOgvLink(ctx context.Context) error {
	ogvLink, err := s.dao.OgvLink(ctx, s.c.SpringFestival2021.OgvLinkUri, time.Now().Unix())
	if err != nil {
		log.Errorc(ctx, "s.dao.OgvLink err(%v)", err)
		return err
	}
	s.ogvLink = ogvLink
	return nil
}

func (s *Service) todayFollow() []int64 {
	follow := s.todayFollowMid()
	mids := make([]int64, 0)
	if len(follow) == 0 {
		return mids
	}
	for _, v := range follow {
		mids = append(mids, v.Mid)
	}
	return mids
}

// Follower 获取关注人信息
func (s *Service) Follower(ctx context.Context, mid int64) (res *cardsmdl.FollowerReply, err error) {
	mids := s.todayFollow()
	follow := s.todayFollowMid()
	res = &cardsmdl.FollowerReply{}
	res.List = make([]*cardsmdl.Follower, 0)
	var (
		memberInfo  map[int64]*accountapi.Info
		midIsFollow map[int64]*relationapi.FollowingReply
	)
	eg := errgroup.WithContext(ctx)

	eg.Go(func(ctx context.Context) (err error) {
		memberInfo, err = s.account.MemberInfo(ctx, mids)
		if err != nil {
			log.Errorc(ctx, "s.account.MemberInfo mids(%v),err(%v)", mids, err)
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		newMids := make([]int64, 0)
		for _, v := range mids {
			if v != mid {
				newMids = append(newMids, v)
			}
		}
		midIsFollow, err = s.midIsFollow(ctx, mid, newMids)
		if err != nil {
			log.Errorc(ctx, "s.midIsFollow mids(%v),err(%v)", mids, err)
			return err
		}
		return nil
	})

	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "eg.Wait error(%v)", err)
		return res, err
	}
	var allFollow = true
	for _, v := range follow {
		if member, ok := memberInfo[v.Mid]; ok {
			var midsIsFollow bool

			if v.Mid == mid {
				midsIsFollow = false
			} else {
				if isFollow, ok1 := midIsFollow[v.Mid]; ok1 && isFollow != nil {
					if isFollow.Attribute < 128 {
						midsIsFollow = true
					} else {
						allFollow = false
					}
				} else {
					allFollow = false

				}
			}
			account := s.accountToAccount(ctx, member)
			res.List = append(res.List, &cardsmdl.Follower{
				Account:  account,
				Desc:     v.Desc,
				IsFollow: midsIsFollow,
			})
		}
	}
	res.AllFollow = allFollow
	return res, nil
}

// Follow ...
func (s *Service) Follow(ctx context.Context, mid int64, risk *riskmdl.Base, mobiApp string) (err error) {
	mids := make([]int64, 0)
	follow := s.todayFollowMid()
	for _, v := range follow {
		if v.Mid != mid {
			mids = append(mids, v.Mid)
		}
	}
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		if len(mids) > 0 {
			err = s.follow(ctx, mid, mids)
			return
		}
		return nil

	})
	eg.Go(func(ctx context.Context) (err error) {
		return s.actSend(ctx, mid, mid, followBusiness)
	})
	eg.Go(func(ctx context.Context) (err error) {
		spRisk := &riskmdl.Sf21Follow{
			Base:        *risk,
			Mid:         mid,
			ActivityUID: ActivityUID,
			MobiApp:     mobiApp,
			Fid:         xstr.JoinInts(mids),
		}
		_, err = s.risk(ctx, mid, riskmdl.ActionCardsFollow, spRisk, spRisk.EsTime)
		if err != nil {
			log.Errorc(ctx, "s.risk mid(%d) follow err(%v)", mid, err)
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "eg.Wait error(%v)", err)
		return err
	}
	return
}

// finishFollow 完成关注任务
func (s *Service) finishFollow(ctx context.Context, mid int64) (midsIsAllFollow bool, err error) {
	mids := s.todayFollow()
	newMids := make([]int64, 0)
	for _, v := range mids {
		if v != mid {
			newMids = append(newMids, v)
		}
	}
	var midsNotAllFollow = false
	if len(newMids) > 0 {
		midIsFollow, err := s.midIsFollow(ctx, mid, newMids)
		if err != nil {
			log.Errorc(ctx, "s.midIsFollow err(%v)", err)
			return false, err
		}
		for _, v := range mids {
			if v == mid {
				continue
			}
			isFollow, ok := midIsFollow[v]
			if !ok {
				midsNotAllFollow = true
				break
			}

			if isFollow == nil {
				midsNotAllFollow = true
				break
			}
			if isFollow.Attribute >= 128 {
				midsNotAllFollow = true
				break
			}
		}
	}

	if !midsNotAllFollow {
		// 完成任务
		midsIsAllFollow = true
		err = s.actSend(ctx, mid, mid, followBusiness)
	}
	return midsIsAllFollow, err
}

func (s *Service) follow(ctx context.Context, mid int64, follower []int64) (err error) {
	followingReply, err := client.RelationClient.BatchAddFollowingAsync(ctx, &relationapi.BatchAddFollowingsReq{Mid: mid, Fid: follower})
	if err != nil || followingReply == nil || followingReply.AllSucceed == false {
		log.Errorc(ctx, "s.relationClient.BatchAddFollowingAsync(%d) error(%v) followingReply(%v)", mid, err, followingReply)
		err = ecode.ActivityCollegeMidFolloweErr
		return err
	}
	return nil
}

func (s *Service) accountToAccount(c context.Context, midInfo *accountapi.Info) *cardsmdl.Account {
	return &cardsmdl.Account{
		Mid:  midInfo.Mid,
		Name: midInfo.Name,
		Face: midInfo.Face,
		Sign: midInfo.Sign,
		Sex:  midInfo.Sex,
	}
}

// midIsFollow ...
func (s *Service) midIsFollow(c context.Context, mid int64, followers []int64) (map[int64]*relationapi.FollowingReply, error) {
	followingMapReply, err := client.RelationClient.Relations(c, &relationapi.RelationsReq{Mid: mid, Fid: followers})
	if err != nil || followingMapReply == nil {
		log.Error("s.relationClient.Relations(%d,%v) error(%v)", mid, followers, err)
		return nil, err
	}
	return followingMapReply.FollowingMap, nil
}

// getCounter ...
func (s *Service) getCounter(c context.Context, mid int64, activity string, counter string) (count int64, err error) {
	count, err = s.actDao.GetCounter(c, mid, activity, counter)
	if err != nil {
		log.Errorc(c, "s.actDao.GetCounter counter(%s) err(%v)", counter, err)
		return
	}
	return
}

// Task 任务列表
func (s *Service) Task(ctx context.Context, mid int64) (res *cardsmdl.TaskReply, err error) {
	allTask := s.allTask
	eg := errgroup.WithContext(ctx)
	taskList := make([]*cardsmdl.TaskMember, 0)
	res = &cardsmdl.TaskReply{}
	resTaskList := make([]*cardsmdl.TaskDetail, 0)
	res.List = resTaskList
	var (
		mutex sync.Mutex
	)
	for _, task := range allTask {
		taskCounter := task.Counter
		taskFinishTimes := task.FinishTimes
		activity := task.Activity
		eg.Go(func(ctx context.Context) (err error) {
			counter, err := s.getCounter(ctx, mid, activity, taskCounter)
			if err != nil {
				log.Errorc(ctx, "s.getCounter err(%v)", err)
				return err
			}
			var state int

			if counter >= taskFinishTimes {
				state = cardsmdl.StateFinish
				counter = taskFinishTimes
			}
			mutex.Lock()
			taskList = append(taskList, &cardsmdl.TaskMember{
				Count:   counter,
				State:   state,
				Counter: taskCounter,
			})
			mutex.Unlock()
			return
		})

	}
	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "eg.Wait error(%v)", err)
		return res, err
	}
	taskMapList := make(map[string]*cardsmdl.TaskMember)
	for _, v := range taskList {
		taskMapList[v.Counter] = v
	}
	for _, task := range allTask {
		link := task.Link
		if task.Counter == ogvBusiness {
			newLink := s.todayOgvLink(ctx)
			if newLink != "" {
				link = newLink
			}
		}
		var taskMember = &cardsmdl.TaskMember{}
		member, ok := taskMapList[task.Counter]
		if ok {
			taskMember = member
			if task.Counter == followBusiness {
				if taskMember.State == 0 {
					isAllFollow, err := s.finishFollow(ctx, mid)
					if err != nil {
						log.Errorc(ctx, "s.finishFollow mid(%d)", mid)
					}
					if isAllFollow {
						taskMember.State = 1
						taskMember.Count = 1
					}
				}
			}

		}
		taskDetail := &cardsmdl.TaskDetail{
			Task: &cardsmdl.SimpleTask{
				TaskName:    task.TaskName,
				LinkName:    task.LinkName,
				Desc:        task.Desc,
				Link:        link,
				FinishTimes: task.FinishTimes,
			},
			Member: taskMember,
		}
		resTaskList = append(resTaskList, taskDetail)
	}
	res.List = resTaskList
	return

}
